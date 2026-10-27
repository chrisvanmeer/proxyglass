package server

import (
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
	for _, expected := range []string{"<!DOCTYPE html>", "/static/css/output.css", "/static/img/proxyglass-mark.svg", "id=\"content\"", "id=\"filterPanel\"", "id=\"filterToggle\""} {
		if !strings.Contains(body, expected) {
			t.Errorf("full dashboard HTML does not contain %q", expected)
		}
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
