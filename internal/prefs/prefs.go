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

// Prefs are persisted in a cookie. Filters are stored as denylists: only items
// the user explicitly disabled are kept. This keeps the cookie small and makes
// newly discovered targets/frontends/backends/servers visible by default.
//
// The Show* fields and the *Set flags are legacy allowlists from older cookies.
// They are only read to migrate an existing cookie and are never written again.
type Prefs struct {
	Initialized      bool            `json:"i"`
	ShowFrontendsSet bool            `json:"fs,omitempty"`
	ShowBackendsSet  bool            `json:"bs,omitempty"`
	ShowServersSet   bool            `json:"vs,omitempty"`
	ShowTargetsSet   bool            `json:"ts,omitempty"`
	ShowFrontends    map[string]bool `json:"sf,omitempty"`
	ShowBackends     map[string]bool `json:"sb,omitempty"`
	ShowServers      map[string]bool `json:"ss,omitempty"`
	ShowTargets      map[string]bool `json:"st,omitempty"`
	HideFrontends    map[string]bool `json:"hf,omitempty"`
	HideBackends     map[string]bool `json:"hb,omitempty"`
	HideServers      map[string]bool `json:"hv,omitempty"`
	HideTargets      map[string]bool `json:"ht,omitempty"`
	HideMode         bool            `json:"hm,omitempty"`
	SortKey          string          `json:"sk"`
	SortKeySet       bool            `json:"sks"`
	SortAsc          bool            `json:"sa"`
	Refresh          int             `json:"r"` // seconds, 0 disabled
}

func (p *Prefs) TargetVisible(key string) bool   { return !p.HideTargets[key] }
func (p *Prefs) FrontendVisible(key string) bool { return !p.HideFrontends[key] }
func (p *Prefs) BackendVisible(key string) bool  { return !p.HideBackends[key] }
func (p *Prefs) ServerVisible(key string) bool   { return !p.HideServers[key] }

func Default() *Prefs {
	return &Prefs{
		HideFrontends: map[string]bool{},
		HideBackends:  map[string]bool{},
		HideServers:   map[string]bool{},
		HideTargets:   map[string]bool{},
		HideMode:      true,
		SortKey:       "config",
		SortAsc:       true,
		Refresh:       5,
	}
}

func Get(r *http.Request) *Prefs {
	return GetWithDefault(r, 5)
}

func GetWithDefault(r *http.Request, defaultRefresh int) *Prefs {
	c := readPrefsCookie(r)
	if c == nil {
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
	p.ensureMaps()
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

func (p *Prefs) ensureMaps() {
	if p.HideFrontends == nil {
		p.HideFrontends = map[string]bool{}
	}
	if p.HideBackends == nil {
		p.HideBackends = map[string]bool{}
	}
	if p.HideServers == nil {
		p.HideServers = map[string]bool{}
	}
	if p.HideTargets == nil {
		p.HideTargets = map[string]bool{}
	}
}

// ParseFromForm builds preferences from a filter drawer submission. The
// supplied keys describe every item currently known to the server; anything
// not present in the submitted form becomes hidden.
func ParseFromForm(r *http.Request, targets, frontends, backends, servers []string) *Prefs {
	p := Default()
	if c := readPrefsCookie(r); c != nil {
		if val, err := url.QueryUnescape(c.Value); err == nil {
			var base Prefs
			if json.Unmarshal([]byte(val), &base) == nil {
				p = &base
				p.ensureMaps()
			}
		}
	}
	_ = r.ParseForm()
	if _, hasFilters := r.Form["filters"]; hasFilters {
		p.Initialized = true
		p.HideTargets = mergeHidden(p.HideTargets, r, "tg_", targets)
		p.HideFrontends = mergeHidden(p.HideFrontends, r, "fe_", frontends)
		p.HideBackends = mergeHidden(p.HideBackends, r, "be_", backends)
		p.HideServers = mergeHidden(p.HideServers, r, "sv_", servers)
		p.clearLegacy()
		p.HideMode = true
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

func mergeHidden(old map[string]bool, r *http.Request, prefix string, keys []string) map[string]bool {
	present := map[string]bool{}
	for name := range r.Form {
		if strings.HasPrefix(name, prefix) {
			present[strings.TrimPrefix(name, prefix)] = true
		}
	}
	hidden := map[string]bool{}
	inUniverse := map[string]bool{}
	for _, key := range keys {
		inUniverse[key] = true
		if !present[key] {
			hidden[key] = true
		}
	}
	// Keep hidden entries for items that are temporarily absent from the
	// snapshot (for example a target that could not be fetched right now).
	for key := range old {
		if !inUniverse[key] {
			hidden[key] = true
		}
	}
	return hidden
}

// Normalize migrates legacy allowlist cookies to the denylist format. It
// returns true when the cookie should be persisted again.
func (p *Prefs) Normalize(targets, frontends, backends, servers []string) bool {
	p.ensureMaps()
	if p.HideMode {
		return false
	}
	p.HideTargets = legacyHidden(targets, p.ShowTargets, p.ShowTargetsSet || len(p.ShowTargets) > 0)
	p.HideFrontends = legacyHidden(frontends, p.ShowFrontends, p.ShowFrontendsSet || len(p.ShowFrontends) > 0)
	p.HideBackends = legacyHidden(backends, p.ShowBackends, p.ShowBackendsSet || len(p.ShowBackends) > 0)
	p.HideServers = legacyHidden(servers, p.ShowServers, p.ShowServersSet || len(p.ShowServers) > 0)
	p.clearLegacy()
	p.HideMode = true
	return true
}

func legacyHidden(keys []string, shown map[string]bool, active bool) map[string]bool {
	hidden := map[string]bool{}
	if !active {
		return hidden
	}
	for _, key := range keys {
		if !shown[key] {
			hidden[key] = true
		}
	}
	return hidden
}

func (p *Prefs) clearLegacy() {
	p.ShowFrontends = nil
	p.ShowBackends = nil
	p.ShowServers = nil
	p.ShowTargets = nil
	p.ShowFrontendsSet = false
	p.ShowBackendsSet = false
	p.ShowServersSet = false
	p.ShowTargetsSet = false
}

func readPrefsCookie(r *http.Request) *http.Cookie {
	var best *http.Cookie
	for _, c := range r.Cookies() {
		if c.Name != cookieName {
			continue
		}
		if best == nil {
			best = c
			continue
		}
		if preferCookie(c, best) {
			best = c
		}
	}
	return best
}

func preferCookie(candidate, current *http.Cookie) bool {
	candidateVal, _ := url.QueryUnescape(candidate.Value)
	currentVal, _ := url.QueryUnescape(current.Value)
	var candidatePrefs, currentPrefs Prefs
	candidateValid := json.Unmarshal([]byte(candidateVal), &candidatePrefs) == nil
	currentValid := json.Unmarshal([]byte(currentVal), &currentPrefs) == nil
	if candidateValid && currentValid && candidatePrefs.HideMode != currentPrefs.HideMode {
		return candidatePrefs.HideMode
	}
	return len(candidateVal) > len(currentVal)
}
