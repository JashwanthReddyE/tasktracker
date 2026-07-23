package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookie = "tt_session"
	sessionTTL    = 30 * 24 * time.Hour
	bcryptCost    = 12

	// bcrypt ignores everything past 72 bytes. Silently truncating would mean
	// two different long passwords authenticate the same account, so this is
	// rejected explicitly rather than passed through.
	maxPasswordBytes = 72
	minPasswordChars = 8
)

type ctxKey int

const userIDKey ctxKey = 0

// userID pulls the authenticated user out of the request context. It is only
// valid inside a handler wrapped by requireAuth.
func userID(r *http.Request) int64 {
	id, _ := r.Context().Value(userIDKey).(int64)
	return id
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ── Session plumbing ──────────────────────────────────

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (a *app) startSession(w http.ResponseWriter, r *http.Request, uid int64) error {
	sid, err := newSessionID()
	if err != nil {
		return err
	}
	now := time.Now()
	if _, err := a.db.Exec(`
		INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES (?,?,?,?)`,
		sid, uid, now.Format(timeFmt), now.Add(sessionTTL).Format(timeFmt)); err != nil {
		return err
	}
	a.setSessionCookie(w, r, sid, sessionTTL)
	return nil
}

func (a *app) setSessionCookie(w http.ResponseWriter, r *http.Request, sid string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.ForceSecure || requestIsTLS(r),
		MaxAge:   int(ttl.Seconds()),
	})
}

func (a *app) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.ForceSecure || requestIsTLS(r),
		MaxAge:   -1,
	})
}

// lookupSession resolves a cookie to a user id, renewing the session when it is
// past halfway to expiry so active users are not logged out mid-use.
func (a *app) lookupSession(w http.ResponseWriter, r *http.Request) (int64, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return 0, false
	}
	var uid int64
	var expires string
	err = a.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE id=?`, c.Value).
		Scan(&uid, &expires)
	if err != nil {
		return 0, false
	}
	exp := parseTime(expires)
	if time.Now().After(exp) {
		a.db.Exec(`DELETE FROM sessions WHERE id=?`, c.Value)
		return 0, false
	}
	if time.Until(exp) < sessionTTL/2 {
		newExp := time.Now().Add(sessionTTL)
		if _, err := a.db.Exec(`UPDATE sessions SET expires_at=? WHERE id=?`,
			newExp.Format(timeFmt), c.Value); err == nil {
			a.setSessionCookie(w, r, c.Value, sessionTTL)
		}
	}
	return uid, true
}

// requireAuth rejects unauthenticated requests before the handler runs. Every
// /api route except signup and login goes through this.
func (a *app) requireAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := a.lookupSession(w, r)
		if !ok {
			writeErr(w, http.StatusUnauthorized, "not signed in")
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), userIDKey, uid)))
	}
}

// ── Handlers ──────────────────────────────────────────

func (a *app) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email, err := normalizeEmail(c.Email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := checkPassword(c.Password); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcryptCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create account")
		return
	}

	var uid int64
	err = tx(a.db, func(tr *sql.Tx) error {
		now := time.Now().Format(timeFmt)
		res, err := tr.Exec(`
			INSERT INTO users (email, display_name, created_at) VALUES (?,?,?)`,
			email, strings.Split(email, "@")[0], now)
		if err != nil {
			return err
		}
		uid, err = res.LastInsertId()
		if err != nil {
			return err
		}
		_, err = tr.Exec(`
			INSERT INTO identities (user_id, provider, subject, password_hash, created_at)
			VALUES (?,?,?,?,?)`, uid, "password", email, string(hash), now)
		return err
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeErr(w, http.StatusConflict, "an account with that email already exists")
			return
		}
		writeErr(w, http.StatusInternalServerError, "could not create account")
		return
	}

	if err := a.startSession(w, r, uid); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not start session")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"email": email, "displayName": strings.Split(email, "@")[0],
	})
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !a.loginLimiter.allow(clientIP(r)) {
		writeErr(w, http.StatusTooManyRequests, "too many attempts, try again shortly")
		return
	}
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email := strings.ToLower(strings.TrimSpace(c.Email))

	var uid int64
	var hash sql.NullString
	var display string
	err := a.db.QueryRow(`
		SELECT u.id, i.password_hash, u.display_name
		  FROM identities i JOIN users u ON u.id = i.user_id
		 WHERE i.provider='password' AND i.subject=?`, email).Scan(&uid, &hash, &display)

	// Unknown email and wrong password return the same message and both run a
	// bcrypt comparison, so response content and timing do not reveal whether
	// an account exists.
	if err != nil || !hash.Valid {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(c.Password))
		writeErr(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(c.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	a.loginLimiter.reset(clientIP(r))
	if err := a.startSession(w, r, uid); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not start session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"email": email, "displayName": display})
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		a.db.Exec(`DELETE FROM sessions WHERE id=?`, c.Value)
	}
	a.clearSessionCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) handleMe(w http.ResponseWriter, r *http.Request) {
	var email, display string
	if err := a.db.QueryRow(`SELECT email, display_name FROM users WHERE id=?`,
		userID(r)).Scan(&email, &display); err != nil {
		writeErr(w, http.StatusUnauthorized, "not signed in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"email": email, "displayName": display})
}

// ── Validation helpers ────────────────────────────────

// dummyHash is a real bcrypt digest compared against when an email is unknown,
// to keep failed-login timing uniform. Its plaintext is irrelevant.
var dummyHash = []byte("$2a$12$eImiTXuWVxfM37uY4JANjQ.uHbn0dHfmMAlQ0NrYWq3nCLNRVWTMS")

func normalizeEmail(raw string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(raw))
	if e == "" {
		return "", errors.New("email is required")
	}
	if _, err := mail.ParseAddress(e); err != nil {
		return "", errors.New("that does not look like a valid email address")
	}
	return e, nil
}

func checkPassword(p string) error {
	if len([]rune(p)) < minPasswordChars {
		return errors.New("password must be at least 8 characters")
	}
	if len(p) > maxPasswordBytes {
		return errors.New("password must be 72 bytes or fewer")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "UNIQUE CONSTRAINT FAILED")
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		first, _, _ := strings.Cut(xf, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ── Login rate limiting ───────────────────────────────

// limiter is a fixed-window counter keyed by client IP. It exists to blunt
// credential stuffing; it is per-process and resets on restart, which is an
// accepted trade for having no external dependency.
type limiter struct {
	mu     sync.Mutex
	hits   map[string]*window
	max    int
	window time.Duration
	lastGC time.Time
}

type window struct {
	count int
	start time.Time
}

func newLimiter(max int, per time.Duration) *limiter {
	return &limiter{hits: map[string]*window{}, max: max, window: per, lastGC: time.Now()}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()

	if now.Sub(l.lastGC) > l.window {
		for k, w := range l.hits {
			if now.Sub(w.start) > l.window {
				delete(l.hits, k)
			}
		}
		l.lastGC = now
	}

	w, ok := l.hits[key]
	if !ok || now.Sub(w.start) > l.window {
		l.hits[key] = &window{count: 1, start: now}
		return true
	}
	w.count++
	return w.count <= l.max
}

func (l *limiter) reset(key string) {
	l.mu.Lock()
	delete(l.hits, key)
	l.mu.Unlock()
}
