package prefs

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const cookieName = "proxyglass_prefs"

type Prefs struct {
	Initialized      bool            `json:"i"`
	ShowFrontendsSet bool            `json:"fs"`
	ShowBackendsSet  bool            `json:"bs"`
	ShowServersSet   bool            `json:"vs"`
	ShowTargetsSet   bool            `json:"ts"`
	ShowFrontends    map[string]bool `json:"sf"`
	ShowBackends     map[string]bool `json:"sb"`
	ShowServers      map[string]bool `json:"ss"`
	ShowTargets      map[string]bool `json:"st"`
	SortKey          string          `json:"sk"`
	SortKeySet       bool            `json:"sks"`
	SortAsc          bool            `json:"sa"`
	Refresh          int             `json:"r"` // seconds, 0 disabled
}

// FilterActive methods preserve compatibility with preference cookies created
// before per-filter flags existed: a non-empty old map is treated as active,
// while an empty legacy map means "show everything".
func (p *Prefs) FrontendFilterActive() bool {
	return p.ShowFrontendsSet || len(p.ShowFrontends) > 0
}

func (p *Prefs) BackendFilterActive() bool {
	return p.ShowBackendsSet || len(p.ShowBackends) > 0
}

func (p *Prefs) ServerFilterActive() bool {
	return p.ShowServersSet || len(p.ShowServers) > 0
}

func (p *Prefs) TargetFilterActive() bool {
	return p.ShowTargetsSet || len(p.ShowTargets) > 0
}

func Default() *Prefs {
	return &Prefs{
		ShowFrontends: map[string]bool{},
		ShowBackends:  map[string]bool{},
		ShowServers:   map[string]bool{},
		ShowTargets:   map[string]bool{},
		SortKey:       "config",
		SortAsc:       true,
		Refresh:       5,
	}
}

func Get(r *http.Request) *Prefs {
	return GetWithDefault(r, 5)
}

func GetWithDefault(r *http.Request, defaultRefresh int) *Prefs {
	c, err := r.Cookie(cookieName)
	if err != nil {
		p := Default()
		p.Refresh = defaultRefresh
		return p
	}
	val, err := url.QueryUnescape(c.Value)
	if err != nil {
		p := Default()
		p.Refresh = defaultRefresh
		return p
	}
	var p Prefs
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		p := Default()
		p.Refresh = defaultRefresh
		return p
	}
	if p.ShowFrontends == nil {
		p.ShowFrontends = map[string]bool{}
	}
	if p.ShowBackends == nil {
		p.ShowBackends = map[string]bool{}
	}
	if p.ShowServers == nil {
		p.ShowServers = map[string]bool{}
	}
	if p.ShowTargets == nil {
		p.ShowTargets = map[string]bool{}
	}
	if !p.SortKeySet {
		p.SortKey = "config"
	}
	if p.Refresh < 0 {
		p.Refresh = 0
	}
	return &p
}

func Set(w http.ResponseWriter, p *Prefs) {
	b, _ := json.Marshal(p)
	val := url.QueryEscape(string(b))
	c := &http.Cookie{
		Name:     cookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().AddDate(1, 0, 0),
		MaxAge:   60 * 60 * 24 * 365,
	}
	http.SetCookie(w, c)
}

func KeyFE(name, target string) string { return "fe:" + target + ":" + name }
func KeyBE(name, target string) string { return "be:" + target + ":" + name }
func KeyTarget(target string) string   { return "target:" + target }
func KeySV(name, be, target string) string {
	return "sv:" + target + ":" + be + ":" + name
}

func ParseFromForm(r *http.Request) *Prefs {
	p := Default()
	// read existing as base
	if c, err := r.Cookie(cookieName); err == nil {
		if val, err := url.QueryUnescape(c.Value); err == nil {
			var base Prefs
			if json.Unmarshal([]byte(val), &base) == nil {
				p = &base
				if p.ShowFrontends == nil {
					p.ShowFrontends = map[string]bool{}
				}
				if p.ShowBackends == nil {
					p.ShowBackends = map[string]bool{}
				}
				if p.ShowServers == nil {
					p.ShowServers = map[string]bool{}
				}
				if p.ShowTargets == nil {
					p.ShowTargets = map[string]bool{}
				}
			}
		}
	}
	_ = r.ParseForm()
	if _, hasFilters := r.Form["filters"]; hasFilters {
		p.Initialized = true
		_, p.ShowFrontendsSet = r.Form["frontends_filter"]
		_, p.ShowBackendsSet = r.Form["backends_filter"]
		_, p.ShowServersSet = r.Form["servers_filter"]
		_, p.ShowTargetsSet = r.Form["targets_filter"]
		p.ShowFrontends = map[string]bool{}
		p.ShowBackends = map[string]bool{}
		p.ShowServers = map[string]bool{}
		p.ShowTargets = map[string]bool{}
		for k := range r.Form {
			if strings.HasPrefix(k, "fe_") {
				p.ShowFrontends[strings.TrimPrefix(k, "fe_")] = true
			} else if strings.HasPrefix(k, "be_") {
				p.ShowBackends[strings.TrimPrefix(k, "be_")] = true
			} else if strings.HasPrefix(k, "sv_") {
				p.ShowServers[strings.TrimPrefix(k, "sv_")] = true
			} else if strings.HasPrefix(k, "tg_") {
				p.ShowTargets[strings.TrimPrefix(k, "tg_")] = true
			}
		}
	}
	if sk := r.FormValue("sort"); sk != "" {
		p.SortKey = sk
		p.SortKeySet = true
	}
	if sa := r.FormValue("asc"); sa == "1" {
		p.SortAsc = true
	} else if sa == "0" {
		p.SortAsc = false
	}
	if r.FormValue("refresh") != "" {
		if v, err := strconv.Atoi(r.FormValue("refresh")); err == nil {
			if v < 0 {
				v = 0
			}
			p.Refresh = v
		}
	} else if _, ok := r.Form["refresh"]; ok {
		p.Refresh = 0
	}
	if !p.SortKeySet {
		p.SortKey = "config"
	}
	return p
}
