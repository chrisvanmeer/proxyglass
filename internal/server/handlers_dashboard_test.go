package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/model"
	"github.com/chrisvanmeer/proxyglass/internal/prefs"
)

func TestDashboardRendersFullLayoutAndBundledBrandAssets(t *testing.T) {
	s, err := NewServer(&config.AppConfig{Server: config.Server{DefaultRefresh: 5}})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	s.handleDashboard(response, httptest.NewRequest("GET", "/", nil))
	body := response.Body.String()
	for _, expected := range []string{"<!DOCTYPE html>", `rel="icon" type="image/svg+xml" href="/static/img/proxyglass-mark.svg"`, "/static/css/output.css", "/static/img/proxyglass-mark.svg", "id=\"content\"", "id=\"filterPanel\"", "id=\"filterToggle\"", "id=\"topologyPanel\"", "id=\"topologyToggle\"", "Topology overview", ".pg-fleet-grid{display:grid!important;grid-template-columns:repeat(4,minmax(0,1fr))"} {
		if !strings.Contains(body, expected) {
			t.Errorf("full dashboard HTML does not contain %q", expected)
		}
	}
}

func TestBundledFaviconIsServed(t *testing.T) {
	s, err := NewServer(&config.AppConfig{})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	s.httpSrv.Handler.ServeHTTP(response, httptest.NewRequest("GET", "/static/img/proxyglass-mark.svg", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("favicon request returned HTTP %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("favicon content type = %q, want image/svg+xml", response.Header().Get("Content-Type"))
	}
}

func TestRefreshPreferenceCanBeDisabled(t *testing.T) {
	s, err := NewServer(&config.AppConfig{Server: config.Server{DefaultRefresh: 5}})
	if err != nil {
		t.Fatal(err)
	}
	save := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/prefs", strings.NewReader("refresh=0"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.handleSavePrefs(save, req)

	var prefCookie *http.Cookie
	for _, cookie := range save.Result().Cookies() {
		if cookie.Name == "proxyglass_prefs" {
			prefCookie = cookie
		}
	}
	if prefCookie == nil {
		t.Fatal("saving refresh=0 did not set the preferences cookie")
	}

	page := httptest.NewRecorder()
	pageReq := httptest.NewRequest("GET", "/", nil)
	pageReq.AddCookie(prefCookie)
	s.handleDashboard(page, pageReq)
	body := page.Body.String()
	if !strings.Contains(body, `name="refresh" value="0"`) {
		t.Fatal("disabled refresh preference was not rendered")
	}
	if strings.Contains(body, `hx-trigger="every 5s"`) {
		t.Fatal("disabled refresh preference still enables polling")
	}
}

func TestLegacyPreferencesDoNotHideNewlyParsedServers(t *testing.T) {
	snapshot := &model.Snapshot{
		Targets:   []string{"edge-a"},
		Frontends: []model.Frontend{{Name: "web", Target: "edge-a", Status: "OPEN"}},
		Backends:  []model.Backend{{Name: "pool", Target: "edge-a", Status: "UP", Servers: []model.Server{{Name: "web1", Backend: "pool", Target: "edge-a", Status: "UP"}}}},
	}
	legacy := &prefs.Prefs{Initialized: true, ShowServers: map[string]bool{}, ShowFrontends: map[string]bool{}, ShowBackends: map[string]bool{}, ShowTargets: map[string]bool{}}
	view := buildView(snapshot, legacy, 5)
	if len(view.Backends) != 1 || !view.Backends[0].Shown || len(view.Backends[0].Servers) != 1 || !view.Backends[0].Servers[0].Shown {
		t.Fatalf("legacy prefs unexpectedly hid backend servers: %#v", view.Backends)
	}
}

func TestDashboardPlacesFrontendsFirstAndGroupsServerFilters(t *testing.T) {
	snapshot := &model.Snapshot{
		Targets:   []string{"edge-a"},
		Frontends: []model.Frontend{{Name: "web-front", Target: "edge-a", Status: "OPEN"}},
		Backends: []model.Backend{{
			Name: "web-pool", Target: "edge-a", Status: "UP",
			Servers: []model.Server{{Name: "web1", Backend: "web-pool", Target: "edge-a", Status: "UP"}},
		}},
	}
	view := buildView(snapshot, prefs.Default(), 5)
	var rendered strings.Builder
	if err := dashboardTemplate().ExecuteTemplate(&rendered, "dash", view); err != nil {
		t.Fatal(err)
	}
	html := rendered.String()
	front := strings.Index(html, `>web-front</h2>`)
	backend := strings.Index(html, `>web-pool</h2>`)
	if front < 0 || backend < 0 || front > backend {
		t.Fatal("frontend cards should appear before backend cards")
	}
	if !strings.Contains(html, `data-backend-toggle="be:edge-a:web-pool"`) || !strings.Contains(html, `data-backend-server="be:edge-a:web-pool"`) {
		t.Fatal("backend filter checkbox is not linked to its child server checkbox")
	}
}

func TestBuildViewConfigurationSortUsesConfiguredTargetOrder(t *testing.T) {
	snapshot := &model.Snapshot{
		Targets: []string{"edge-b", "edge-a"},
		Frontends: []model.Frontend{
			{Name: "frontend-a", Target: "edge-a", Status: "OPEN"},
			{Name: "frontend-b", Target: "edge-b", Status: "OPEN"},
		},
		Backends: []model.Backend{
			{Name: "backend-a", Target: "edge-a", Status: "UP", Servers: []model.Server{
				{Name: "a-second", Backend: "backend-a", Target: "edge-a", Status: "UP"},
				{Name: "a-first", Backend: "backend-a", Target: "edge-a", Status: "UP"},
			}},
			{Name: "backend-b", Target: "edge-b", Status: "UP", Servers: []model.Server{
				{Name: "b-second", Backend: "backend-b", Target: "edge-b", Status: "UP"},
				{Name: "b-first", Backend: "backend-b", Target: "edge-b", Status: "UP"},
			}},
		},
	}
	preferences := prefs.Default()
	preferences.SortKey = "config"
	preferences.SortAsc = false // Configuration order remains fixed; direction only affects other sort modes.
	view := buildView(snapshot, preferences, 5)

	if len(view.Frontends) != 2 || view.Frontends[0].Target != "edge-b" || view.Frontends[1].Target != "edge-a" {
		t.Fatalf("frontends do not follow target config order: %#v", view.Frontends)
	}
	if len(view.Backends) != 2 || view.Backends[0].Target != "edge-b" || view.Backends[1].Target != "edge-a" {
		t.Fatalf("backends do not follow target config order: %#v", view.Backends)
	}
	if got := view.Backends[0].Servers; len(got) != 2 || got[0].Name != "b-second" || got[1].Name != "b-first" {
		t.Fatalf("servers should retain HAProxy row order in config sort: %#v", got)
	}
	var rendered strings.Builder
	if err := dashboardTemplate().ExecuteTemplate(&rendered, "dash", view); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.String(), `id="sortDirection" class="pg-select mt-3" data-select data-submit="true" hidden`) {
		t.Fatal("sort direction control should be hidden for fixed configuration order")
	}
}

func TestBuildViewLinksTopologyOnlyForConfiguredRoutes(t *testing.T) {
	snapshot := &model.Snapshot{
		Targets:   []string{"edge-a"},
		Frontends: []model.Frontend{{Name: "public", Target: "edge-a", Status: "OPEN"}},
		Backends: []model.Backend{{Name: "app-pool", Target: "edge-a", Status: "UP", Servers: []model.Server{{
			Name: "app-01", Backend: "app-pool", Target: "edge-a", Status: "UP",
		}}}},
	}
	view := buildView(snapshot, prefs.Default(), 5, config.TopologyRoute{
		Target: "edge-a", Frontend: "public", Backends: []string{"app-pool"},
	})

	routeEdges, serverEdges := 0, 0
	anchors := map[string]bool{}
	for _, edge := range view.Topology.Edges {
		if edge.Kind == "route" {
			routeEdges++
		}
		if edge.Kind == "server" {
			serverEdges++
		}
	}
	for _, node := range view.Topology.Nodes {
		if node.Anchor != "" {
			anchors[node.Anchor] = true
		}
	}
	if routeEdges != 1 || serverEdges != 1 {
		t.Fatalf("topology edges route=%d server=%d; want one each: %#v", routeEdges, serverEdges, view.Topology.Edges)
	}
	for _, anchor := range []string{"frontend-fe:edge-a:public", "backend-be:edge-a:app-pool", "server-sv:edge-a:app-pool:app-01"} {
		if !anchors[anchor] {
			t.Errorf("topology is missing clickable row anchor %q", anchor)
		}
	}
	var rendered strings.Builder
	if err := dashboardTemplate().ExecuteTemplate(&rendered, "dash", view); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.String(), `class="pg-topology-edge pg-topology-edge-route"`) || !strings.Contains(rendered.String(), `data-topology-anchor="server-sv:edge-a:app-pool:app-01"`) {
		t.Fatal("topology SVG should render configured routes and clickable server anchors")
	}

	withoutMapping := buildView(snapshot, prefs.Default(), 5)
	for _, edge := range withoutMapping.Topology.Edges {
		if edge.Kind == "route" {
			t.Fatal("topology must not invent frontend-to-backend routes without a mapping")
		}
	}
}

func TestRefreshRendersOnlyHTMXFragment(t *testing.T) {
	s, err := NewServer(&config.AppConfig{Server: config.Server{DefaultRefresh: 5}})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	s.handleRefresh(response, httptest.NewRequest("GET", "/api/refresh", nil))
	if strings.Contains(response.Body.String(), "<!DOCTYPE html>") {
		t.Fatal("HTMX refresh should return a content fragment, not a nested page")
	}
	if !strings.Contains(response.Body.String(), `id="topologyContent" hx-swap-oob="innerHTML"`) {
		t.Fatal("HTMX refresh should also update topology content out of band")
	}
}

func TestHumanBytes(t *testing.T) {
	for input, want := range map[int64]string{0: "0 B", 1024: "1.0 KB", 1048576: "1.0 MB"} {
		if got := humanBytes(input); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestLastSessionAge(t *testing.T) {
	for raw, want := range map[string]string{"0": "Just now", "1": "1s ago", "386": "386s ago", "-": "No session yet", "": "No session yet"} {
		if got := lastSessionAge(raw); got != want {
			t.Errorf("lastSessionAge(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestDashboardSortComparison(t *testing.T) {
	cmp := compareViewValues("scur", "a", "UP", 4, 1, "b", "DOWN", 2, 9)
	if !ordered(cmp, false) || ordered(cmp, true) {
		t.Fatalf("descending session sort did not rank the higher value first: cmp=%d", cmp)
	}
	if got := compareViewValues("status", "a", "UP", 0, 0, "b", "DOWN", 0, 0); got <= 0 {
		t.Fatalf("status comparison = %d, want UP > DOWN lexically", got)
	}
}

func TestActivityTimeTracksChangedHAProxyCounters(t *testing.T) {
	server := &Server{}
	start := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	snapshot := &model.Snapshot{
		FetchedAt: start,
		Frontends: []model.Frontend{{Name: "front", Target: "edge", SessionsTot: 4, BytesOut: 100}},
		Backends:  []model.Backend{{Name: "pool", Target: "edge", SessionsTot: 4, BytesOut: 100}},
	}
	server.annotateActivity(snapshot)
	if !snapshot.Frontends[0].LastActivity.Equal(start) || !snapshot.Backends[0].LastActivity.Equal(start) {
		t.Fatal("first traffic observation should initialize activity time")
	}

	next := &model.Snapshot{
		FetchedAt: start.Add(30 * time.Second),
		Frontends: []model.Frontend{{Name: "front", Target: "edge", SessionsTot: 4, BytesOut: 100}},
		Backends:  []model.Backend{{Name: "pool", Target: "edge", SessionsTot: 4, BytesOut: 100}},
	}
	server.annotateActivity(next)
	if !next.Frontends[0].LastActivity.Equal(start) || !next.Backends[0].LastActivity.Equal(start) {
		t.Fatal("unchanged traffic counters should retain their last activity time")
	}

	next.FetchedAt = start.Add(time.Minute)
	next.Frontends[0].BytesOut++
	next.Backends[0].BytesOut++
	server.annotateActivity(next)
	if !next.Frontends[0].LastActivity.Equal(next.FetchedAt) || !next.Backends[0].LastActivity.Equal(next.FetchedAt) {
		t.Fatal("changed traffic counters should advance last activity time")
	}
}
