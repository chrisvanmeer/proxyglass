package server

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chrisvanmeer/proxyglass/internal/auth"
	"github.com/chrisvanmeer/proxyglass/internal/config"
	"github.com/chrisvanmeer/proxyglass/internal/haproxy"
	"github.com/chrisvanmeer/proxyglass/internal/logger"
	"github.com/chrisvanmeer/proxyglass/internal/model"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	cfg          *config.AppConfig
	httpSrv      *http.Server
	client       *haproxy.Client
	sessions     *auth.SessionManager
	oidcProvider *oidc.Provider
	oidcOAuth    *oauth2.Config
	oidcStates   map[string]oidcState
	mu           sync.RWMutex
	lastSnap     *model.Snapshot
	lastFetch    time.Time
	activity     map[string]activityStamp
}

type trafficSignature struct {
	sessions int64
	total    int64
	bytesIn  int64
	bytesOut int64
}

type activityStamp struct {
	signature trafficSignature
	at        time.Time
}

type oidcState struct {
	nonce   string
	expires time.Time
}

func NewServer(cfg *config.AppConfig) (*Server, error) {
	s := &Server{
		cfg:      cfg,
		client:   haproxy.New(),
		sessions: auth.NewSessionManager(cfg.OIDC.SessionTTL()),
	}
	if cfg.OIDC.EnabledAndConfigured() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		provider, err := oidc.NewProvider(ctx, cfg.OIDC.Issuer)
		if err != nil {
			return nil, err
		}
		redirectURL := cfg.OIDC.RedirectURI
		if redirectURL == "" {
			redirectURL = "https://proxyglass.invalid/ui/oidc/callback"
		}
		s.oidcProvider = provider
		s.oidcOAuth = &oauth2.Config{
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		}
		s.oidcStates = make(map[string]oidcState)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/health", s.handleHealth)
	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/ui/oidc/login", s.handleOIDCLogin)
	mux.HandleFunc("/ui/oidc/callback", s.handleOIDCCallback)
	mux.HandleFunc("/", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("/api/refresh", s.requireAuth(s.handleRefresh))
	mux.HandleFunc("/api/prefs", s.requireAuth(s.handleSavePrefs))

	s.httpSrv = &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: mux,
	}
	return s, nil
}

func NewServerFromPath(path string) (*Server, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return NewServer(cfg)
}

func (s *Server) Start() error {
	err := s.httpSrv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.RequireAuth {
			next(w, r)
			return
		}
		if c, err := r.Cookie("proxyglass_session"); err == nil {
			if s.sessions.Valid(c.Value) {
				next(w, r)
				return
			}
		}
		authh := r.Header.Get("Authorization")
		if strings.HasPrefix(authh, "Bearer ") {
			token := strings.TrimPrefix(authh, "Bearer ")
			store, err := auth.LoadTokens(s.cfg.TokensFile)
			if err == nil && store.Valid(token) {
				next(w, r)
				return
			}
		}
		if strings.HasPrefix(authh, "Basic ") {
			enc := strings.TrimPrefix(authh, "Basic ")
			if dec, err := base64.StdEncoding.DecodeString(enc); err == nil {
				parts := strings.SplitN(string(dec), ":", 2)
				if len(parts) == 2 && parts[0] == "token" {
					store, err := auth.LoadTokens(s.cfg.TokensFile)
					if err == nil && store.Valid(parts[1]) {
						next(w, r)
						return
					}
				}
			}
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		token := r.FormValue("token")
		if token == "" {
			authh := r.Header.Get("Authorization")
			if strings.HasPrefix(authh, "Bearer ") {
				token = strings.TrimPrefix(authh, "Bearer ")
			}
		}
		store, err := auth.LoadTokens(s.cfg.TokensFile)
		if err == nil && store.Valid(token) {
			tok, _ := s.sessions.Create()
			http.SetCookie(w, &http.Cookie{Name: "proxyglass_session", Value: tok, Path: "/", HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(s.cfg.OIDC.SessionTTL())})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	tmpl := template.Must(template.New("login").Parse(loginHTML))
	_ = tmpl.Execute(w, map[string]interface{}{"OIDC": s.cfg.OIDC.EnabledAndConfigured()})
}

const loginHTML = `<!DOCTYPE html>
<html lang="en" class="dark"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Login - Proxyglass</title><link rel="icon" type="image/svg+xml" href="/static/img/proxyglass-mark.svg"><link rel="stylesheet" href="/static/css/output.css?v=0.1.2"><script src="/static/js/htmx.min.js"></script><style>html,body{min-height:100%;margin:0}body{font-family:Inter,ui-sans-serif,system-ui,sans-serif;background:#070b16;color:#e2e8f0;display:flex;align-items:center;justify-content:center}.pg-login{width:min(420px,calc(100% - 32px));padding:30px;border-radius:22px;background:linear-gradient(145deg,rgba(21,31,51,.96),rgba(12,19,34,.94));border:1px solid rgba(148,163,184,.2);box-shadow:0 24px 80px rgba(0,0,0,.45)}.pg-login input,.pg-login button{box-sizing:border-box;width:100%;padding:12px 14px;border-radius:10px;background:#101a2d;color:#e2e8f0;border:1px solid rgba(255,255,255,.12);margin:6px 0}.pg-login button{background:linear-gradient(100deg,#0891b2,#7c3aed);cursor:pointer}.pg-login img{width:260px;max-width:100%;height:auto;margin-bottom:16px}</style></head>
<body class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 text-slate-100 font-sans flex items-center justify-center">
<div class="fixed inset-0 overflow-hidden pointer-events-none"><div class="absolute -top-40 -right-40 w-80 h-80 bg-cyan-500/10 rounded-full blur-3xl"></div><div class="absolute -bottom-40 -left-40 w-80 h-80 bg-violet-500/10 rounded-full blur-3xl"></div></div>
<div class="relative z-10 w-full max-w-md p-8">
<div class="pg-login backdrop-blur-xl bg-white/5 border border-white/10 rounded-2xl shadow-2xl p-8">
<img src="/static/img/proxyglass-logo.svg" alt="Proxyglass">
<p class="text-slate-400 mb-6">Sign in to access dashboard</p>
<form method="post" action="/login" class="space-y-4">
<input type="password" name="token" placeholder="Bearer token" class="w-full px-4 py-2 rounded-lg bg-white/5 border border-white/10 focus:border-cyan-400/50 focus:outline-none" required>
<button class="w-full px-4 py-2 rounded-lg bg-gradient-to-r from-cyan-500/20 to-violet-500/20 border border-cyan-400/30 hover:border-cyan-400/60 transition">Sign in with token</button>
</form>
{{if .OIDC}}<div class="mt-4"><a href="/ui/oidc/login" class="block w-full text-center px-4 py-2 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 transition">Sign in with SSO</a></div>{{end}}
</div></div>
</body></html>`

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("proxyglass_session"); err == nil {
		s.sessions.Revoke(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "proxyglass_session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	if s.oidcOAuth == nil {
		http.Error(w, "OIDC not configured", 400)
		return
	}
	state, err := randomString(32)
	if err != nil {
		http.Error(w, "unable to start OIDC login", http.StatusInternalServerError)
		return
	}
	nonce, err := randomString(32)
	if err != nil {
		http.Error(w, "unable to start OIDC login", http.StatusInternalServerError)
		return
	}
	s.mu.Lock()
	for key, entry := range s.oidcStates {
		if time.Now().After(entry.expires) {
			delete(s.oidcStates, key)
		}
	}
	s.oidcStates[state] = oidcState{nonce: nonce, expires: time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	oauthCfg := *s.oidcOAuth
	if s.cfg.OIDC.RedirectURI == "" {
		proto := "http"
		if requestIsHTTPS(r) {
			proto = "https"
		}
		oauthCfg.RedirectURL = proto + "://" + r.Host + "/ui/oidc/callback"
	}
	http.Redirect(w, r, oauthCfg.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	if s.oidcOAuth == nil {
		logger.Warn("OIDC: callback received while OIDC is not configured")
		http.Error(w, "OIDC not configured", http.StatusBadRequest)
		return
	}
	if idpError := r.URL.Query().Get("error"); idpError != "" {
		logger.Warn("OIDC: identity provider returned an authorization error: %q", idpError)
		http.Error(w, "OIDC login failed", http.StatusUnauthorized)
		return
	}
	state := r.URL.Query().Get("state")
	s.mu.Lock()
	entry, ok := s.oidcStates[state]
	delete(s.oidcStates, state)
	s.mu.Unlock()
	if !ok || time.Now().After(entry.expires) {
		logger.Warn("OIDC: callback has an unknown or expired state")
		http.Error(w, "OIDC login failed", http.StatusUnauthorized)
		return
	}
	oauthCfg := *s.oidcOAuth
	if s.cfg.OIDC.RedirectURI == "" {
		proto := "http"
		if requestIsHTTPS(r) {
			proto = "https"
		}
		oauthCfg.RedirectURL = proto + "://" + r.Host + "/ui/oidc/callback"
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	token, err := oauthCfg.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		logger.Warn("OIDC: authorization-code exchange failed: %v", err)
		http.Error(w, "OIDC token exchange failed", http.StatusUnauthorized)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		logger.Warn("OIDC: token response did not include an ID token")
		http.Error(w, "OIDC response did not include an ID token", http.StatusUnauthorized)
		return
	}
	verifier := s.oidcProvider.Verifier(&oidc.Config{ClientID: s.cfg.OIDC.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.Warn("OIDC: ID token verification failed: %v", err)
		http.Error(w, "OIDC ID token verification failed", http.StatusUnauthorized)
		return
	}
	var claims struct {
		Nonce             string   `json:"nonce"`
		Subject           string   `json:"sub"`
		Email             string   `json:"email"`
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		logger.Warn("OIDC: could not read verified ID token claims: %v", err)
		http.Error(w, "OIDC nonce validation failed", http.StatusUnauthorized)
		return
	}
	identity := oidcIdentity(claims.Email, claims.PreferredUsername, claims.Subject)
	if claims.Nonce != entry.nonce {
		logger.Warn("OIDC: nonce validation failed for user %s", identity)
		http.Error(w, "OIDC nonce validation failed", http.StatusUnauthorized)
		return
	}
	if len(s.cfg.OIDC.AllowedGroups) > 0 && !hasAllowedGroup(claims.Groups, s.cfg.OIDC.AllowedGroups) {
		logger.Warn("OIDC: user %s (groups %q) is not a member of an allowed group %q; denying access",
			identity, claims.Groups, s.cfg.OIDC.AllowedGroups)
		http.Error(w, "user is not a member of an allowed group", http.StatusForbidden)
		return
	}
	session, err := s.sessions.Create()
	if err != nil {
		logger.Error("OIDC: failed to create session for user %s: %v", identity, err)
		http.Error(w, "unable to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "proxyglass_session", Value: session, Path: "/", HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(s.cfg.OIDC.SessionTTL())})
	logger.Info("OIDC: session established for user %s", identity)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func requestIsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func hasAllowedGroup(groups, allowed []string) bool {
	for _, group := range groups {
		for _, match := range allowed {
			if groupMatches(group, match) {
				return true
			}
		}
	}
	return false
}

// groupMatches accepts exact group paths and Keycloak's default full-path
// group claims when the configured allowed group is just the leaf name. This
// mirrors Orbitron's group matching behavior.
func groupMatches(claim, allowed string) bool {
	claim = strings.TrimSpace(claim)
	allowed = strings.TrimSpace(allowed)
	if claim == allowed {
		return true
	}
	trimmedClaim := strings.TrimPrefix(claim, "/")
	trimmedAllowed := strings.TrimPrefix(allowed, "/")
	if trimmedClaim == trimmedAllowed {
		return true
	}
	if i := strings.LastIndex(trimmedClaim, "/"); i >= 0 && trimmedClaim[i+1:] == trimmedAllowed {
		return true
	}
	return false
}

func oidcIdentity(email, preferredUsername, subject string) string {
	switch {
	case email != "":
		return email
	case preferredUsername != "":
		return preferredUsername
	default:
		return subject
	}
}
