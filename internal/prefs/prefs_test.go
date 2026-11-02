package prefs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestConfigOrderIsDefaultAndMigratesLegacySortPreference(t *testing.T) {
	if got := GetWithDefault(httptest.NewRequest("GET", "/", nil), 5).SortKey; got != "config" {
		t.Fatalf("default sort = %q, want config", got)
	}

	legacyCookie := &http.Cookie{Name: cookieName, Value: url.QueryEscape(`{"sk":"name","sa":true,"r":5}`)}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(legacyCookie)
	if got := GetWithDefault(req, 5).SortKey; got != "config" {
		t.Fatalf("legacy cookie sort = %q, want migrated config default", got)
	}
}

func TestExplicitSortPreferenceOverridesConfigDefault(t *testing.T) {
	form := url.Values{"sort": {"name"}}
	post := httptest.NewRequest("POST", "/api/prefs", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	preferences := ParseFromForm(post, nil, nil, nil, nil)
	if preferences.SortKey != "name" || !preferences.SortKeySet {
		t.Fatalf("explicit sort preference = %#v, want name marked as user-selected", preferences)
	}
}

func TestConfiguredRefreshDefaultAndZeroPersist(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if got := GetWithDefault(req, 17).Refresh; got != 17 {
		t.Fatalf("default refresh = %d, want 17", got)
	}

	form := url.Values{"refresh": {"0"}}
	post := httptest.NewRequest("POST", "/api/prefs", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	prefs := ParseFromForm(post, nil, nil, nil, nil)
	if prefs.Refresh != 0 {
		t.Fatalf("refresh = %d, want disabled (0)", prefs.Refresh)
	}
}

func TestEmptyFilterSelectionHidesEveryKnownItem(t *testing.T) {
	form := url.Values{
		"filters": {"1"},
		"refresh": {"5"},
	}
	post := httptest.NewRequest("POST", "/api/prefs", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	p := ParseFromForm(post, []string{"target:a"}, []string{"fe:a:f"}, []string{"be:a:b"}, []string{"sv:a:b:s"})
	if !p.Initialized {
		t.Fatal("expected initialized filters")
	}
	if p.TargetVisible("target:a") || p.FrontendVisible("fe:a:f") || p.BackendVisible("be:a:b") || p.ServerVisible("sv:a:b:s") {
		t.Fatalf("expected every known item to be hidden, got %#v", p)
	}
}

func TestParseFromFormOnlyHidesUncheckedItems(t *testing.T) {
	form := url.Values{
		"filters":     {"1"},
		"tg_target:a": {"on"},
		"fe_fe:a:f":   {"on"},
	}
	post := httptest.NewRequest("POST", "/api/prefs", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	p := ParseFromForm(post, []string{"target:a", "target:b"}, []string{"fe:a:f"}, []string{"be:a:b"}, []string{"sv:a:b:s"})
	if !p.TargetVisible("target:a") || !p.FrontendVisible("fe:a:f") {
		t.Fatal("checked items should stay visible")
	}
	if p.TargetVisible("target:b") || p.BackendVisible("be:a:b") || p.ServerVisible("sv:a:b:s") {
		t.Fatal("unchecked/absent items should be hidden")
	}
}

func TestNormalizeMigratesLegacyAllowlistWithoutChangingSelections(t *testing.T) {
	p := &Prefs{
		Initialized:    true,
		ShowServers:    map[string]bool{"sv:a": true},
		ShowServersSet: true,
	}
	changed := p.Normalize(nil, nil, nil, []string{"sv:a", "sv:b"})
	if !changed {
		t.Fatal("expected migration to report a change")
	}
	if !p.HideMode {
		t.Fatal("expected denylist mode to be enabled")
	}
	if !p.ServerVisible("sv:a") {
		t.Fatal("previously selected server should stay visible")
	}
	if p.ServerVisible("sv:b") {
		t.Fatal("legacy migration must keep previously unselected servers hidden")
	}
	if len(p.ShowServers) != 0 {
		t.Fatalf("legacy allowlist should be cleared, got %#v", p.ShowServers)
	}
}

func TestNormalizeLegacyEmptyFilterMeansShowEverything(t *testing.T) {
	p := &Prefs{Initialized: true}
	if !p.Normalize(nil, nil, []string{"be:a"}, nil) {
		t.Fatal("expected migration to report a change")
	}
	if !p.BackendVisible("be:a") {
		t.Fatal("legacy cookie without filters should show everything")
	}
}

func TestNewlyAddedItemsAreVisibleByDefault(t *testing.T) {
	p := &Prefs{
		HideMode:    true,
		HideServers: map[string]bool{"sv:disabled": true},
	}
	if !p.Normalize(nil, nil, nil, []string{"sv:disabled", "sv:new"}) {
		// Normalize only migrates legacy cookies; a denylist cookie is already normalized.
	}
	if p.ServerVisible("sv:disabled") {
		t.Fatal("explicitly disabled item must stay disabled")
	}
	if !p.ServerVisible("sv:new") {
		t.Fatal("newly discovered item must be visible by default")
	}
}

func TestDenylistCookieStaysSmall(t *testing.T) {
	p := Default()
	p.HideServers["sv:disabled"] = true
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "sv:visible") || strings.Contains(string(data), "sf") || strings.Contains(string(data), "kf") {
		t.Fatalf("cookie should only contain explicitly hidden items, got %s", data)
	}
	if len(data) > 200 {
		t.Fatalf("denylist cookie grew unexpectedly large: %d bytes (%s)", len(data), data)
	}
}
