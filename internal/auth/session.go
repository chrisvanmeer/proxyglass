package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]time.Time
	ttl      time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = 8 * time.Hour
	}
	return &SessionManager{
		sessions: make(map[string]time.Time),
		ttl:      ttl,
	}
}

func (s *SessionManager) Create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	now := time.Now()
	s.sweep(now)
	s.mu.Lock()
	s.sessions[token] = now.Add(s.ttl)
	s.mu.Unlock()
	return token, nil
}

func (s *SessionManager) Valid(token string) bool {
	if token == "" {
		return false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.sessions[token]
	if !ok {
		return false
	}
	if now.After(exp) {
		delete(s.sessions, token)
		return false
	}
	return true
}

func (s *SessionManager) Revoke(token string) {
	if token == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *SessionManager) sweep(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, exp := range s.sessions {
		if now.After(exp) {
			delete(s.sessions, token)
		}
	}
}
