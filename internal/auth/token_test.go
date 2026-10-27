package auth

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateValidateAndRevokeToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "tokens.json")
	token, _, err := GenerateToken(path, "test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	store, err := LoadTokens(path)
	if err != nil || !store.Valid(token) {
		t.Fatalf("generated token not valid: store=%v err=%v", store, err)
	}
	if err := RevokeToken(path, token); err != nil {
		t.Fatal(err)
	}
	store, err = LoadTokens(path)
	if err != nil || store.Valid(token) {
		t.Fatalf("revoked token still valid: store=%v err=%v", store, err)
	}
}
