package server

import "testing"

func TestGroupMatchesOrbitronKeycloakBehavior(t *testing.T) {
	tests := []struct {
		name    string
		claim   string
		allowed string
		want    bool
	}{
		{name: "exact name", claim: "admins", allowed: "admins", want: true},
		{name: "leading slash", claim: "/admins", allowed: "admins", want: true},
		{name: "allowed leading slash", claim: "admins", allowed: "/admins", want: true},
		{name: "nested Keycloak path matches leaf", claim: "/orbitron/platform/admins", allowed: "admins", want: true},
		{name: "full nested path exact match", claim: "/orbitron/platform/admins", allowed: "/orbitron/platform/admins", want: true},
		{name: "different leaf denied", claim: "/orbitron/platform/users", allowed: "admins", want: false},
		{name: "substring is not a match", claim: "super-admins", allowed: "admins", want: false},
		{name: "whitespace is ignored", claim: " /orbitron/admins ", allowed: " admins ", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := groupMatches(test.claim, test.allowed); got != test.want {
				t.Errorf("groupMatches(%q, %q) = %t, want %t", test.claim, test.allowed, got, test.want)
			}
		})
	}
}

func TestHasAllowedGroupChecksEveryClaimAndConfiguredGroup(t *testing.T) {
	if !hasAllowedGroup([]string{"users", "/orbitron/ops/admins"}, []string{"platform-admins", "admins"}) {
		t.Fatal("expected any matching allowed Keycloak group to authorize the user")
	}
	if hasAllowedGroup([]string{"users", "developers"}, []string{"admins", "operators"}) {
		t.Fatal("user without an allowed group must be denied")
	}
}
