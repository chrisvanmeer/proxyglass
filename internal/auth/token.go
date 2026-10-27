package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"
)

const tokenByteLength = 16

var ErrTokenNotFound = errors.New("token not found")

type TokenInfo struct {
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	Label     string `json:"label,omitempty"`
}

type TokenStore struct {
	Tokens map[string]TokenInfo `json:"tokens"`
}

func LoadTokens(tokensFile string) (*TokenStore, error) {
	store := &TokenStore{Tokens: make(map[string]TokenInfo)}
	data, err := os.ReadFile(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	var raw struct {
		Tokens map[string]json.RawMessage `json:"tokens"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for token, rawInfo := range raw.Tokens {
		var legacy int64
		if err := json.Unmarshal(rawInfo, &legacy); err == nil {
			store.Tokens[token] = TokenInfo{CreatedAt: legacy}
			continue
		}
		var info TokenInfo
		if err := json.Unmarshal(rawInfo, &info); err != nil {
			return nil, fmt.Errorf("invalid token entry for %q: %w", token, err)
		}
		store.Tokens[token] = info
	}
	return store, nil
}

func SaveTokens(tokensFile string, store *TokenStore) error {
	if store == nil {
		store = &TokenStore{Tokens: make(map[string]TokenInfo)}
	}
	if store.Tokens == nil {
		store.Tokens = make(map[string]TokenInfo)
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(tokensFile), 0750); err != nil {
		return err
	}
	if err := os.WriteFile(tokensFile, data, 0640); err != nil {
		return err
	}
	if u, err := user.Lookup("proxyglass"); err == nil {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)
		_ = os.Chown(tokensFile, uid, gid)
	}
	return nil
}

func GenerateToken(tokensFile, label string, ttl time.Duration) (string, TokenInfo, error) {
	b := make([]byte, tokenByteLength)
	if _, err := rand.Read(b); err != nil {
		return "", TokenInfo{}, err
	}
	token := hex.EncodeToString(b)
	store, err := LoadTokens(tokensFile)
	if err != nil {
		return "", TokenInfo{}, err
	}
	now := time.Now().Unix()
	info := TokenInfo{CreatedAt: now, Label: label}
	if ttl > 0 {
		info.ExpiresAt = now + int64(ttl.Seconds())
	}
	store.Tokens[token] = info
	if err := SaveTokens(tokensFile, store); err != nil {
		return "", TokenInfo{}, err
	}
	return token, info, nil
}

func RevokeToken(tokensFile string, token string) error {
	store, err := LoadTokens(tokensFile)
	if err != nil {
		return err
	}
	if _, ok := store.Tokens[token]; !ok {
		return ErrTokenNotFound
	}
	delete(store.Tokens, token)
	return SaveTokens(tokensFile, store)
}

func (s *TokenStore) Valid(token string) bool {
	info, ok := s.Tokens[token]
	if !ok {
		return false
	}
	if info.ExpiresAt > 0 && info.ExpiresAt <= time.Now().Unix() {
		return false
	}
	return true
}
