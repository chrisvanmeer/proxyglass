package haproxy

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/logger"
	"github.com/chrisvanmeer/proxyglass/internal/model"
)

const targetRequestTimeout = 10 * time.Second

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) FetchAll(ctx context.Context, targets []config.Target) *model.Snapshot {
	snap := &model.Snapshot{
		Targets:   make([]string, 0, len(targets)),
		Frontends: []model.Frontend{},
		Backends:  []model.Backend{},
		FetchedAt: time.Now(),
		Errors:    map[string]error{},
	}
	for _, target := range targets {
		snap.Targets = append(snap.Targets, target.Name)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(tgt config.Target) {
			defer wg.Done()
			tr := http.DefaultTransport.(*http.Transport).Clone()
			verifyTLS := tgt.VerifyTLS == nil || *tgt.VerifyTLS
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: !verifyTLS} //nolint:gosec
			client := &http.Client{Timeout: targetRequestTimeout, Transport: tr}
			defer client.CloseIdleConnections()
			req, err := http.NewRequestWithContext(ctx, "GET", tgt.URL, nil)
			if err != nil {
				logger.Warn("HAProxy target %q request could not be created: %s", tgt.Name, requestErrorMessage(err))
				mu.Lock()
				snap.Errors[tgt.Name] = err
				mu.Unlock()
				return
			}
			if tgt.Username != "" || tgt.Password != "" {
				req.SetBasicAuth(tgt.Username, tgt.Password)
			}
			req.Header.Set("User-Agent", "proxyglass/1.0")
			resp, err := client.Do(req)
			if err != nil {
				if isRequestTimeout(err) {
					logger.Warn("HAProxy target %q request timed out after %s", tgt.Name, targetRequestTimeout)
				} else {
					logger.Warn("HAProxy target %q request failed: %s", tgt.Name, requestErrorMessage(err))
				}
				mu.Lock()
				snap.Errors[tgt.Name] = err
				mu.Unlock()
				return
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != 200 {
				logger.Warn("HAProxy target %q returned HTTP status %s", tgt.Name, resp.Status)
				mu.Lock()
				snap.Errors[tgt.Name] = fmt.Errorf("bad status: %s", resp.Status)
				mu.Unlock()
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Warn("HAProxy target %q response body could not be read: %s", tgt.Name, requestErrorMessage(err))
				mu.Lock()
				snap.Errors[tgt.Name] = err
				mu.Unlock()
				return
			}
			fes, bes, err := parseCSV(string(body), tgt.Name)
			if err != nil {
				logger.Warn("HAProxy target %q returned invalid stats CSV: %v", tgt.Name, err)
				mu.Lock()
				snap.Errors[tgt.Name] = err
				mu.Unlock()
				return
			}
			mu.Lock()
			snap.Frontends = append(snap.Frontends, fes...)
			snap.Backends = append(snap.Backends, bes...)
			mu.Unlock()
		}(t)
	}
	wg.Wait()
	return snap
}

func isRequestTimeout(err error) bool {
	var requestErr *url.Error
	if errors.As(err, &requestErr) {
		err = requestErr.Err
	}
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout())
}

// requestErrorMessage omits the request URL from net/url errors because URLs
// can contain credentials in userinfo, even though that is discouraged.
func requestErrorMessage(err error) string {
	var requestErr *url.Error
	if errors.As(err, &requestErr) {
		return requestErr.Err.Error()
	}
	return err.Error()
}

func parseCSV(data, target string) ([]model.Frontend, []model.Backend, error) {
	r := csv.NewReader(strings.NewReader(data))
	r.FieldsPerRecord = -1
	// HAProxy CSV uses commas
	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	backendMap := map[string]*model.Backend{}
	var frontends []model.Frontend
	indices := map[string]int{}
	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		if strings.HasPrefix(rec[0], "#") {
			header := strings.TrimSpace(strings.TrimPrefix(rec[0], "#"))
			if header == "pxname" && len(rec) > 1 {
				for i, name := range rec {
					name = strings.TrimSpace(strings.TrimPrefix(name, "#"))
					indices[name] = i
				}
			}
			continue
		}
		if len(indices) > 0 {
			if idx, ok := indices["type"]; !ok || idx >= len(rec) {
				continue
			}
		}
		pxname := rec[0]
		svname := rec[1]
		qtype := field(rec, indices, "type", 32)
		// The conventional HAProxy stats listener itself appears as a synthetic
		// frontend/backend in CSV. It is a monitoring endpoint, not an app proxy.
		if pxname == "stats" {
			continue
		}
		if qtype == "0" && svname == "FRONTEND" {
			fe := model.Frontend{
				Name:        pxname,
				Target:      target,
				Status:      field(rec, indices, "status", 17),
				SessionsCur: toInt64(field(rec, indices, "scur", 4)),
				SessionsMax: toInt64(field(rec, indices, "smax", 5)),
				SessionsTot: toInt64(field(rec, indices, "stot", 7)),
				Rate:        toInt64(field(rec, indices, "rate", 33)),
				RateMax:     toInt64(field(rec, indices, "rate_max", 35)),
				BytesIn:     toInt64(field(rec, indices, "bin", 8)),
				BytesOut:    toInt64(field(rec, indices, "bout", 9)),
				DeniedReq:   toInt64(field(rec, indices, "dreq", 10)),
				DeniedResp:  toInt64(field(rec, indices, "dresp", 11)),
				ErrorsReq:   toInt64(field(rec, indices, "ereq", 12)),
			}
			frontends = append(frontends, fe)
			continue
		}
		if qtype == "1" && svname == "BACKEND" {
			key := pxname + "|" + target
			b := backendMap[key]
			if b == nil {
				b = &model.Backend{Name: pxname, Target: target, Servers: []model.Server{}}
				backendMap[key] = b
			}
			b.Status = field(rec, indices, "status", 17)
			b.CheckStatus = field(rec, indices, "check_status", 36)
			b.CheckCode = field(rec, indices, "check_code", 37)
			b.LastChk = field(rec, indices, "last_chk", 56)
			b.LastChkDur = field(rec, indices, "check_duration", 38)
			b.SessionsCur = toInt64(field(rec, indices, "scur", 4))
			b.SessionsMax = toInt64(field(rec, indices, "smax", 5))
			b.SessionsTot = toInt64(field(rec, indices, "stot", 7))
			b.LastSess = field(rec, indices, "lastsess", 55)
			b.BytesIn = toInt64(field(rec, indices, "bin", 8))
			b.BytesOut = toInt64(field(rec, indices, "bout", 9))
			continue
		}
		if qtype == "2" {
			key := pxname + "|" + target
			b, ok := backendMap[key]
			if !ok {
				b = &model.Backend{
					Name:    pxname,
					Target:  target,
					Servers: []model.Server{},
				}
				backendMap[key] = b
			}
			s := model.Server{
				Name:        svname,
				Backend:     pxname,
				Target:      target,
				Status:      field(rec, indices, "status", 17),
				CheckStatus: field(rec, indices, "check_status", 36),
				CheckCode:   field(rec, indices, "check_code", 37),
				LastChk:     field(rec, indices, "last_chk", 56),
				LastChkDur:  field(rec, indices, "check_duration", 38),
				LastSess:    field(rec, indices, "lastsess", 55),
				SessionsCur: toInt64(field(rec, indices, "scur", 4)),
				SessionsMax: toInt64(field(rec, indices, "smax", 5)),
				SessionsLim: field(rec, indices, "slim", 6),
				BytesIn:     toInt64(field(rec, indices, "bin", 8)),
				BytesOut:    toInt64(field(rec, indices, "bout", 9)),
				ErrorsReq:   field(rec, indices, "ereq", 12),
				ErrorsConn:  field(rec, indices, "econ", 13),
				ErrorsResp:  field(rec, indices, "eresp", 14),
				Wredis:      field(rec, indices, "wredis", 15),
				Wretr:       field(rec, indices, "wretr", 16),
			}
			b.Servers = append(b.Servers, s)
		}
	}
	var backends []model.Backend
	for _, b := range backendMap {
		backends = append(backends, *b)
	}
	return frontends, backends, nil
}

func get(rec []string, idx int) string {
	if idx < 0 || idx >= len(rec) {
		return ""
	}
	return rec[idx]
}

func field(rec []string, indices map[string]int, name string, fallback int) string {
	if i, ok := indices[name]; ok {
		return get(rec, i)
	}
	return get(rec, fallback)
}

func toInt64(s string) int64 {
	var v int64
	// simple atoi
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			v = v*10 + int64(s[i]-'0')
		}
	}
	return v
}
