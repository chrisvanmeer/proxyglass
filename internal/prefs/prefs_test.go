package prefs

import (
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
	preferences := ParseFromForm(post)
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
	prefs := ParseFromForm(post)
	if prefs.Refresh != 0 {
		t.Fatalf("refresh = %d, want disabled (0)", prefs.Refresh)
	}
}

func TestEmptyFilterSelectionIsNotInterpretedAsDefaultAll(t *testing.T) {
	form := url.Values{
		"filters":          {"1"},
		"targets_filter":   {"1"},
		"frontends_filter": {"1"},
		"backends_filter":  {"1"},
		"servers_filter":   {"1"},
		"refresh":          {"5"},
	}
	post := httptest.NewRequest("POST", "/api/prefs", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	prefs := ParseFromForm(post)
	if !prefs.Initialized || !prefs.TargetFilterActive() || !prefs.ServerFilterActive() || len(prefs.ShowTargets) != 0 {
		t.Fatalf("expected initialized empty filters, got %#v", prefs)
	}
}

func TestLegacyCookieWithNoServerFiltersDefaultsToAllServers(t *testing.T) {
	p := &Prefs{Initialized: true, ShowServers: map[string]bool{}}
	if p.ServerFilterActive() {
		t.Fatal("legacy preference cookie with no server entries should default to showing servers")
	}
}
