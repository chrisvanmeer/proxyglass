package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chrisvanmeer/proxyglass/internal/haproxy"
	"github.com/chrisvanmeer/proxyglass/internal/model"
	"github.com/chrisvanmeer/proxyglass/internal/prefs"
)

func contextTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func (s *Server) getSnapshot() *model.Snapshot {
	s.mu.RLock()
	if s.lastSnap != nil && time.Since(s.lastFetch) < 2*time.Second {
		snap := s.lastSnap
		s.mu.RUnlock()
		return snap
	}
	s.mu.RUnlock()
	client := haproxy.New()
	ctx, cancel := contextTimeout(10 * time.Second)
	defer cancel()
	snap := client.FetchAll(ctx, s.cfg.Targets)
	s.annotateActivity(snap)
	s.mu.Lock()
	s.lastSnap = snap
	s.lastFetch = time.Now()
	s.mu.Unlock()
	return snap
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	p := prefs.GetWithDefault(r, s.cfg.Server.DefaultRefresh)
	snap := s.getSnapshot()
	data := buildView(snap, p, s.cfg.Server.DefaultRefresh)
	tmpl := dashboardTemplate()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "dash", data)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	client := haproxy.New()
	ctx, cancel := contextTimeout(10 * time.Second)
	defer cancel()
	snap := client.FetchAll(ctx, s.cfg.Targets)
	s.annotateActivity(snap)
	s.mu.Lock()
	s.lastSnap = snap
	s.lastFetch = time.Now()
	s.mu.Unlock()
	p := prefs.GetWithDefault(r, s.cfg.Server.DefaultRefresh)
	data := buildView(snap, p, s.cfg.Server.DefaultRefresh)
	tmpl := dashboardTemplate()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "content", data)
}

func (s *Server) handleSavePrefs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		p := prefs.ParseFromForm(r)
		prefs.Set(w, p)
		data := buildView(s.getSnapshot(), p, s.cfg.Server.DefaultRefresh)
		tmpl := dashboardTemplate()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.ExecuteTemplate(w, "content", data)
		return
	}
	w.WriteHeader(400)
}

type ViewData struct {
	Frontends  []FView
	Backends   []BView
	Prefs      *prefs.Prefs
	Targets    []string
	Errors     map[string]error
	DefaultRef int
	Fleet      FleetView
}

type FleetView struct {
	Proxies      int
	ActiveSess   int64
	ServersUp    int
	ServersTotal int
	BytesInTotal int64
}

type FView struct {
	Key            string
	Name           string
	Target         string
	Status         string
	SCur           int64
	SMax           int64
	STot           int64
	Rate           int64
	RateMax        int64
	BIn            int64
	BOut           int64
	Denied         int64
	Errors         int64
	LastActivity   string
	LastActivityAt string
	TargetShown    bool
	Shown          bool
}

type BView struct {
	Key            string
	Name           string
	Target         string
	Status         string
	SCur           int64
	SMax           int64
	BIn            int64
	BOut           int64
	LastActivity   string
	LastActivityAt string
	Servers        []SView
	VisibleServers bool
	TargetShown    bool
	Shown          bool
}

type SView struct {
	Key          string
	Name         string
	Backend      string
	Target       string
	Status       string
	CheckStatus  string
	CheckCode    string
	LastChk      string
	LastChkDur   string
	LastActivity string
	SCur         int64
	SMax         int64
	SLim         string
	BIn          int64
	BOut         int64
	ErrorsReq    string
	ErrorsConn   string
	ErrorsResp   string
	Shown        bool
}

func (s *Server) annotateActivity(snap *model.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activity == nil {
		s.activity = make(map[string]activityStamp)
	}
	for i := range snap.Frontends {
		frontend := &snap.Frontends[i]
		key := "frontend\x00" + frontend.Target + "\x00" + frontend.Name
		frontend.LastActivity = s.observeActivity(key, trafficSignature{
			sessions: frontend.SessionsCur,
			total:    frontend.SessionsTot,
			bytesIn:  frontend.BytesIn,
			bytesOut: frontend.BytesOut,
		}, snap.FetchedAt)
	}
	for i := range snap.Backends {
		backend := &snap.Backends[i]
		key := "backend\x00" + backend.Target + "\x00" + backend.Name
		backend.LastActivity = s.observeActivity(key, trafficSignature{
			sessions: backend.SessionsCur,
			total:    backend.SessionsTot,
			bytesIn:  backend.BytesIn,
			bytesOut: backend.BytesOut,
		}, snap.FetchedAt)
		if seconds, err := strconv.ParseInt(backend.LastSess, 10, 64); err == nil && seconds >= 0 {
			backend.LastActivity = snap.FetchedAt.Add(-time.Duration(seconds) * time.Second)
		}
	}
}

// observeActivity stores when Proxyglass first observed traffic counters and
// updates that time whenever HAProxy reports new sessions or bytes.
// Caller must hold s.mu.
func (s *Server) observeActivity(key string, signature trafficSignature, at time.Time) time.Time {
	if signature == (trafficSignature{}) {
		delete(s.activity, key)
		return time.Time{}
	}
	if old, ok := s.activity[key]; ok && old.signature == signature {
		return old.at
	}
	s.activity[key] = activityStamp{signature: signature, at: at}
	return at
}

func activityAge(at, now time.Time) string {
	if at.IsZero() {
		return "No activity yet"
	}
	seconds := int64(now.Sub(at).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	switch {
	case seconds < 5:
		return "Just now"
	case seconds < 60:
		return strconv.FormatInt(seconds, 10) + "s ago"
	case seconds < 3600:
		return strconv.FormatInt(seconds/60, 10) + "m ago"
	case seconds < 86400:
		return strconv.FormatInt(seconds/3600, 10) + "h ago"
	default:
		return strconv.FormatInt(seconds/86400, 10) + "d ago"
	}
}

func activityTimestamp(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.Local().Format(time.RFC3339)
}

func lastSessionAge(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return "No session yet"
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds < 0 {
		return "—"
	}
	if seconds == 0 {
		return "Just now"
	}
	return strconv.FormatInt(seconds, 10) + "s ago"
}

func dashboardTemplate() *template.Template {
	funcs := template.FuncMap{
		"humanBytes": humanBytes,
		"mulf":       func(a, b float64) float64 { return a * b },
		"divf": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"toFloat64": func(v interface{}) float64 {
			switch t := v.(type) {
			case int:
				return float64(t)
			case int64:
				return float64(t)
			case float64:
				return t
			}
			return 0
		},
		"printf": fmt.Sprintf,
	}
	return template.Must(template.New("dash").Funcs(funcs).Parse(dashboardHTML))
}

func buildView(snap *model.Snapshot, p *prefs.Prefs, defRef int) *ViewData {
	vd := &ViewData{
		Frontends:  []FView{},
		Backends:   []BView{},
		Prefs:      p,
		Targets:    snap.Targets,
		Errors:     snap.Errors,
		DefaultRef: defRef,
	}
	targetOrder := make(map[string]int, len(snap.Targets))
	for index, target := range snap.Targets {
		targetOrder[target] = index
	}
	configTargetIndex := func(target string) int {
		if index, ok := targetOrder[target]; ok {
			return index
		}
		return len(targetOrder)
	}
	fv := FleetView{Proxies: len(snap.Targets)}
	for _, be := range snap.Backends {
		for _, sv := range be.Servers {
			fv.ServersTotal++
			if strings.EqualFold(sv.Status, "UP") {
				fv.ServersUp++
			}
		}
	}
	for _, fe := range snap.Frontends {
		fv.ActiveSess += fe.SessionsCur
		fv.BytesInTotal += fe.BytesIn
	}
	vd.Fleet = fv
	for _, fe := range snap.Frontends {
		k := prefs.KeyFE(fe.Name, fe.Target)
		targetShown := !p.TargetFilterActive() || p.ShowTargets[prefs.KeyTarget(fe.Target)]
		shown := targetShown &&
			(!p.FrontendFilterActive() || p.ShowFrontends[k])
		vd.Frontends = append(vd.Frontends, FView{Key: k, Name: fe.Name, Target: fe.Target, Status: fe.Status, SCur: fe.SessionsCur, SMax: fe.SessionsMax, STot: fe.SessionsTot, Rate: fe.Rate, RateMax: fe.RateMax, BIn: fe.BytesIn, BOut: fe.BytesOut, Denied: fe.DeniedReq + fe.DeniedResp, Errors: fe.ErrorsReq, LastActivity: activityAge(fe.LastActivity, snap.FetchedAt), LastActivityAt: activityTimestamp(fe.LastActivity), TargetShown: targetShown, Shown: shown})
	}
	for _, be := range snap.Backends {
		k := prefs.KeyBE(be.Name, be.Target)
		targetShown := !p.TargetFilterActive() || p.ShowTargets[prefs.KeyTarget(be.Target)]
		shown := targetShown &&
			(!p.BackendFilterActive() || p.ShowBackends[k])
		bv := BView{Key: k, Name: be.Name, Target: be.Target, Status: be.Status, SCur: be.SessionsCur, SMax: be.SessionsMax, BIn: be.BytesIn, BOut: be.BytesOut, LastActivity: activityAge(be.LastActivity, snap.FetchedAt), LastActivityAt: activityTimestamp(be.LastActivity), TargetShown: targetShown, Shown: shown}
		for _, sv := range be.Servers {
			sk := prefs.KeySV(sv.Name, sv.Backend, sv.Target)
			sshown := !p.ServerFilterActive() || p.ShowServers[sk]
			if sshown {
				bv.VisibleServers = true
			}
			bv.Servers = append(bv.Servers, SView{Key: sk, Name: sv.Name, Backend: sv.Backend, Target: sv.Target, Status: sv.Status, CheckStatus: sv.CheckStatus, CheckCode: sv.CheckCode, LastChk: sv.LastChk, LastChkDur: sv.LastChkDur, LastActivity: lastSessionAge(sv.LastSess), SCur: sv.SessionsCur, SMax: sv.SessionsMax, SLim: sv.SessionsLim, BIn: sv.BytesIn, BOut: sv.BytesOut, ErrorsReq: sv.ErrorsReq, ErrorsConn: sv.ErrorsConn, ErrorsResp: sv.ErrorsResp, Shown: sshown})
		}
		if p.SortKey != "config" {
			sort.Slice(bv.Servers, func(i, j int) bool {
				if p.SortKey == "status" {
					if p.SortAsc {
						return bv.Servers[i].Status < bv.Servers[j].Status
					}
					return bv.Servers[i].Status > bv.Servers[j].Status
				}
				if p.SortKey == "scur" {
					if p.SortAsc {
						return bv.Servers[i].SCur < bv.Servers[j].SCur
					}
					return bv.Servers[i].SCur > bv.Servers[j].SCur
				}
				if p.SortKey == "bin" {
					if p.SortAsc {
						return bv.Servers[i].BIn < bv.Servers[j].BIn
					}
					return bv.Servers[i].BIn > bv.Servers[j].BIn
				}
				if p.SortAsc {
					return bv.Servers[i].Name < bv.Servers[j].Name
				}
				return bv.Servers[i].Name > bv.Servers[j].Name
			})
		}
		vd.Backends = append(vd.Backends, bv)
	}
	sort.SliceStable(vd.Frontends, func(i, j int) bool {
		a, b := vd.Frontends[i], vd.Frontends[j]
		if p.SortKey == "config" {
			return configTargetIndex(a.Target) < configTargetIndex(b.Target)
		}
		if less := compareViewValues(p.SortKey, a.Name, a.Status, a.SCur, a.BIn, b.Name, b.Status, b.SCur, b.BIn); less != 0 {
			return ordered(less, p.SortAsc)
		}
		return a.Target < b.Target
	})
	sort.SliceStable(vd.Backends, func(i, j int) bool {
		a, b := vd.Backends[i], vd.Backends[j]
		if p.SortKey == "config" {
			return configTargetIndex(a.Target) < configTargetIndex(b.Target)
		}
		if less := compareViewValues(p.SortKey, a.Name, a.Status, a.SCur, a.BIn, b.Name, b.Status, b.SCur, b.BIn); less != 0 {
			return ordered(less, p.SortAsc)
		}
		return a.Target < b.Target
	})
	return vd
}

// compareViewValues returns -1, 0, or 1 for a pair of rows using the chosen
// dashboard sort field. The ordering direction is applied separately.
func compareViewValues(key, nameA, statusA string, sessionsA, bytesA int64, nameB, statusB string, sessionsB, bytesB int64) int {
	switch key {
	case "status":
		return strings.Compare(statusA, statusB)
	case "scur":
		return compareInt64(sessionsA, sessionsB)
	case "bin":
		return compareInt64(bytesA, bytesB)
	default:
		return strings.Compare(nameA, nameB)
	}
}

func compareInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func ordered(cmp int, ascending bool) bool {
	if ascending {
		return cmp < 0
	}
	return cmp > 0
}

func humanBytes(b int64) string {
	if b < 1024 {
		return strconv.FormatInt(b, 10) + " B"
	}
	div := int64(1024)
	exp := 0
	for n := b / div; n >= 1024 && exp < 6; n /= div {
		div *= 1024
		exp++
	}
	val := float64(b) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB", "PB", "EB"}[exp]
	return strconv.FormatFloat(val, 'f', 1, 64) + " " + suffix
}
