package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/cache"
	"github.com/maulanashalihin/laju-go/app/queries"
)

type Store struct {
	querier      *queries.Querier
	sessionCache *cache.SessionCache
	sessionName  string
	sessionTTL   time.Duration
	secure       bool
}

type Session struct {
	id        string
	userID    int64
	values    map[string]interface{}
	c         *fiber.Ctx
	store     *Store
	dirty     bool
	expiresAt time.Time
}

// SessionData represents the data stored in session
type SessionData struct {
	UserID        int64  `json:"user_id"`
	Name          string `json:"name,omitempty"`
	Email         string `json:"email"`
	Avatar        string `json:"avatar,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Role          string `json:"role"`
	CSRFToken     string `json:"csrf_token,omitempty"`
	CSRFExpiry    int64  `json:"csrf_expiry,omitempty"`
	IP            string `json:"ip,omitempty"`
	UserAgent     string `json:"ua,omitempty"`
}

// New creates a new session store with database backend and optional in-memory cache.
// Pass nil for sessionCache to disable caching. sessionTTL is the session lifetime.
func New(querier *queries.Querier, sessionCache *cache.SessionCache, sessionTTL time.Duration) *Store {
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}
	return &Store{
		querier:      querier,
		sessionCache: sessionCache,
		sessionName:  "session_id",
		sessionTTL:   sessionTTL,
		secure:       false,
	}
}

// Get retrieves a session
func (s *Store) Get(c *fiber.Ctx) (*Session, error) {
	// Get session from locals first (if already loaded)
	if sess := c.Locals("session"); sess != nil {
		slog.Debug("session retrieved from locals")
		return sess.(*Session), nil
	}

	session := &Session{
		id:        "",
		userID:    0,
		values:    make(map[string]interface{}),
		c:         c,
		store:     s,
		dirty:     false,
		expiresAt: time.Now().Add(s.sessionTTL),
	}

	cookieValue := c.Cookies(s.sessionName)
	if cookieValue == "" {
		c.Locals("session", session)
		return session, nil
	}

	reqIP := ClientIP(c)
	reqUA := c.Get("User-Agent")
	isPage := isPageRequest(c)

	// Try in-memory cache first (avoids DB lookup on every request)
	if s.sessionCache != nil {
		if cached, ok := s.sessionCache.Get(cookieValue); ok {
			if cached.ExpiresAt.Before(time.Now()) {
				s.sessionCache.Invalidate(cookieValue)
			} else {
			switch checkFingerprint(cached.IP, cached.UserAgent, reqIP, reqUA, isPage) {
			case fpInvalidate:
				slog.Warn("session fingerprint mismatch (cache) — invalidating",
					"session_id", cookieValue)
				s.invalidateSession(c, cookieValue)
			case fpOK:
				if cached.IP == "" {
					s.setFingerprint(c, cookieValue, cached)
				}
				session.populate(cookieValue, fieldsFromCached(cached))
				c.Locals("session", session)
				return session, nil
			}
			}
		}
	}

	dbSession, err := s.querier.GetSessionByID(context.Background(), hashSessionID(cookieValue))
	if err == nil {
		if dbSession.ExpiresAt.Before(time.Now()) {
			s.deleteSession(context.Background(), cookieValue)
			if s.sessionCache != nil {
				s.sessionCache.Invalidate(cookieValue)
			}
		} else {
			var data SessionData
			if err := json.Unmarshal([]byte(dbSession.Data), &data); err == nil {
			switch checkFingerprint(data.IP, data.UserAgent, reqIP, reqUA, isPage) {
			case fpInvalidate:
				slog.Warn("session fingerprint mismatch (db) — invalidating",
					"session_id", cookieValue)
				s.invalidateSession(c, cookieValue)
				c.Locals("session", session)
				return session, nil
			case fpOK:
				if data.IP == "" {
					data.IP = reqIP
					data.UserAgent = reqUA
					newJSON, _ := json.Marshal(data)
					dbSession.Data = string(newJSON)
					s.querier.UpdateSession(context.Background(), dbSession)
				}
			}
				session.populate(cookieValue, fieldsFromSessionData(&data, dbSession.ExpiresAt))

				// Store in cache for subsequent requests
				if s.sessionCache != nil {
					s.sessionCache.Set(cookieValue, cache.CachedSessionData{
						UserID:        data.UserID,
						Name:          data.Name,
						Email:         data.Email,
						Avatar:        data.Avatar,
						EmailVerified: data.EmailVerified,
						Role:          data.Role,
						CSRFToken:     data.CSRFToken,
						CSRFExpiry:    data.CSRFExpiry,
						IP:            data.IP,
						UserAgent:     data.UserAgent,
						ExpiresAt:     dbSession.ExpiresAt,
					})
				}
			}
		}
	} else {
		// Session not found or expired in DB — invalidate stale cache entry
		if s.sessionCache != nil {
			s.sessionCache.Invalidate(cookieValue)
		}
	}

	c.Locals("session", session)
	return session, nil
}

// sessionFields is the common set of session data fields used by both
// cache and DB paths for fingerprint validation and session population.
type sessionFields struct {
	UserID        int64
	Name          string
	Email         string
	Avatar        string
	EmailVerified bool
	Role          string
	CSRFToken     string
	CSRFExpiry    int64
	ExpiresAt     time.Time
}

func fieldsFromCached(c *cache.CachedSessionData) sessionFields {
	return sessionFields{
		UserID:        c.UserID,
		Name:          c.Name,
		Email:         c.Email,
		Avatar:        c.Avatar,
		EmailVerified: c.EmailVerified,
		Role:          c.Role,
		CSRFToken:     c.CSRFToken,
		CSRFExpiry:    c.CSRFExpiry,
		ExpiresAt:     c.ExpiresAt,
	}
}

func fieldsFromSessionData(d *SessionData, expiresAt time.Time) sessionFields {
	return sessionFields{
		UserID:        d.UserID,
		Name:          d.Name,
		Email:         d.Email,
		Avatar:        d.Avatar,
		EmailVerified: d.EmailVerified,
		Role:          d.Role,
		CSRFToken:     d.CSRFToken,
		CSRFExpiry:    d.CSRFExpiry,
		ExpiresAt:     expiresAt,
	}
}

// fingerprintAction represents the result of validating a session's
// User-Agent against the current request.
type fingerprintAction int

const (
	fpOK        fingerprintAction = iota // fingerprint matches or is empty
	fpInvalidate                         // fingerprint mismatch — invalidate session
)

// checkFingerprint validates the session's User-Agent against the current
// request. IP-based fingerprinting was removed because IP changes are normal
// for SaaS users (mobile networks, VPN, IPv6 privacy extensions, corporate
// NAT) and caused excessive false logouts.
//
// The only check: if this is a page request and the request sends no
// User-Agent at all (a strong indicator of a non-browser client using a
// stolen cookie), invalidate the session. Normal UA changes from browser
// updates or extensions are tolerated.
func checkFingerprint(storedIP, storedUA, reqIP, reqUA string, isPage bool) fingerprintAction {
	if isPage && reqUA == "" {
		return fpInvalidate
	}
	return fpOK
}

// populate sets session fields from the unified sessionFields struct.
// Used by both cache and DB paths to avoid duplicating field assignments.
func (sess *Session) populate(id string, f sessionFields) {
	sess.id = id
	sess.userID = f.UserID
	sess.expiresAt = f.ExpiresAt
	sess.values["user_id"] = f.UserID
	if f.Name != "" {
		sess.values["name"] = f.Name
	}
	sess.values["email"] = f.Email
	if f.Avatar != "" {
		sess.values["avatar"] = f.Avatar
	}
	sess.values["email_verified"] = f.EmailVerified
	sess.values["role"] = f.Role
	if f.CSRFToken != "" {
		sess.values["csrf_token"] = f.CSRFToken
	}
	if f.CSRFExpiry != 0 {
		sess.values["csrf_expiry"] = f.CSRFExpiry
	}
}

// invalidateSession removes a session from DB + cache and clears the cookie.
// Used by both cache and DB paths when a fingerprint mismatch is detected.
func (s *Store) invalidateSession(c *fiber.Ctx, sessionID string) {
	s.deleteSession(context.Background(), sessionID)
	if s.sessionCache != nil {
		s.sessionCache.Invalidate(sessionID)
	}
	c.ClearCookie(s.sessionName)
}

// generateSessionID generates a random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashSessionID returns the SHA-256 hex digest of a raw session ID.
// The DB stores only the hash — if the database leaks, the attacker
// cannot reconstruct valid cookies from the hashed session IDs.
// The in-memory cache keys on the raw ID (no hash) for speed.
func hashSessionID(rawID string) string {
	h := sha256.Sum256([]byte(rawID))
	return hex.EncodeToString(h[:])
}

// Set sets a value in the session
func (s *Session) Set(key string, value interface{}) {
	s.values[key] = value
	s.dirty = true
}

// Get gets a value from the session
func (s *Session) Get(key string) interface{} {
	return s.values[key]
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
	delete(s.values, key)
	s.dirty = true
}

// Save saves the session to database
func (s *Session) Save() error {
	// Encode session data
	sessionData := SessionData{
		UserID: 0,
		Email:  "",
		Role:   "",
	}

	if userID, ok := s.values["user_id"].(int64); ok {
		sessionData.UserID = userID
	} else if userID, ok := s.values["user_id"].(int); ok {
		sessionData.UserID = int64(userID)
	} else if userID, ok := s.values["user_id"].(float64); ok {
		sessionData.UserID = int64(userID)
	}

	if name, ok := s.values["name"].(string); ok {
		sessionData.Name = name
	}

	if email, ok := s.values["email"].(string); ok {
		sessionData.Email = email
	}

	if avatar, ok := s.values["avatar"].(string); ok {
		sessionData.Avatar = avatar
	}

	if emailVerified, ok := s.values["email_verified"].(bool); ok {
		sessionData.EmailVerified = emailVerified
	}

	if role, ok := s.values["role"].(string); ok {
		sessionData.Role = role
	}

	if csrfToken, ok := s.values["csrf_token"].(string); ok {
		sessionData.CSRFToken = csrfToken
	}

	if csrfExpiry, ok := s.values["csrf_expiry"].(int64); ok {
		sessionData.CSRFExpiry = csrfExpiry
	}

	// Capture client fingerprint
	sessionData.IP = ClientIP(s.c)
	sessionData.UserAgent = s.c.Get("User-Agent")

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		slog.Error("session marshal error", "error", err)
		return err
	}

	// Refresh expiry on every save (sliding expiration)
	s.expiresAt = time.Now().Add(s.store.sessionTTL)

	if s.id == "" {
		// Create new session
		sessionID, err := generateSessionID()
		if err != nil {
			slog.Error("session generate id error", "error", err)
			return err
		}
		s.id = sessionID

		dbSession := &queries.Session{
			ID:        hashSessionID(s.id),
			UserID:    sessionData.UserID,
			Data:      string(jsonData),
			ExpiresAt: s.expiresAt,
		}

		if err := s.store.querier.CreateSession(context.Background(), dbSession); err != nil {
			slog.Error("session create error", "error", err)
			return err
		}

		// Seed cache after creation
		if s.store.sessionCache != nil {
			s.store.sessionCache.Set(s.id, cache.CachedSessionData{
				UserID:        sessionData.UserID,
				Name:          sessionData.Name,
				Email:         sessionData.Email,
				Avatar:        sessionData.Avatar,
				EmailVerified: sessionData.EmailVerified,
				Role:          sessionData.Role,
				CSRFToken:     sessionData.CSRFToken,
				CSRFExpiry:    sessionData.CSRFExpiry,
				IP:            sessionData.IP,
				UserAgent:     sessionData.UserAgent,
				ExpiresAt:     s.expiresAt,
			})
		}
	} else {
		// Update existing session
		dbSession := &queries.Session{
			ID:        hashSessionID(s.id),
			UserID:    sessionData.UserID,
			Data:      string(jsonData),
			ExpiresAt: s.expiresAt,
		}

		if err := s.store.querier.UpdateSession(context.Background(), dbSession); err != nil {
			slog.Error("session update error", "error", err)
			return err
		}

		// Refresh cache after update
		if s.store.sessionCache != nil {
			s.store.sessionCache.Set(s.id, cache.CachedSessionData{
				UserID:        sessionData.UserID,
				Name:          sessionData.Name,
				Email:         sessionData.Email,
				Avatar:        sessionData.Avatar,
				EmailVerified: sessionData.EmailVerified,
				Role:          sessionData.Role,
				CSRFToken:     sessionData.CSRFToken,
				CSRFExpiry:    sessionData.CSRFExpiry,
				IP:            sessionData.IP,
				UserAgent:     sessionData.UserAgent,
				ExpiresAt:     s.expiresAt,
			})
		}
	}

	// Set cookie with session ID
	s.c.Cookie(&fiber.Cookie{
		Name:     s.store.sessionName,
		Value:    s.id,
		Path:     "/",
		HTTPOnly: true,
		Secure:   s.store.secure,
		SameSite: "Lax",
		MaxAge:   int(s.expiresAt.Sub(time.Now()).Seconds()),
	})
	slog.Debug("session cookie set", "name", s.store.sessionName, "value", s.id)

	return nil
}

// Destroy destroys the session
func (s *Session) Destroy() error {
	if s.id != "" {
		s.store.deleteSession(context.Background(), s.id)
		if s.store.sessionCache != nil {
			s.store.sessionCache.Invalidate(s.id)
		}
	}

	s.values = make(map[string]interface{})
	s.c.ClearCookie(s.store.sessionName)
	return nil
}

// Regenerate generates a new session ID
func (s *Session) Regenerate() error {
	if s.id == "" {
		return nil // Nothing to regenerate
	}

	newID, err := generateSessionID()
	if err != nil {
		return err
	}

	// Re-encode data with fingerprint
	sessionData := SessionData{
		UserID: 0,
		Email:  "",
		Role:   "",
	}

	if userID, ok := s.values["user_id"].(int64); ok {
		sessionData.UserID = userID
	} else if userID, ok := s.values["user_id"].(int); ok {
		sessionData.UserID = int64(userID)
	} else if userID, ok := s.values["user_id"].(float64); ok {
		sessionData.UserID = int64(userID)
	}

	if name, ok := s.values["name"].(string); ok {
		sessionData.Name = name
	}

	if email, ok := s.values["email"].(string); ok {
		sessionData.Email = email
	}

	if avatar, ok := s.values["avatar"].(string); ok {
		sessionData.Avatar = avatar
	}

	if emailVerified, ok := s.values["email_verified"].(bool); ok {
		sessionData.EmailVerified = emailVerified
	}

	if role, ok := s.values["role"].(string); ok {
		sessionData.Role = role
	}

	if csrfToken, ok := s.values["csrf_token"].(string); ok {
		sessionData.CSRFToken = csrfToken
	}

	if csrfExpiry, ok := s.values["csrf_expiry"].(int64); ok {
		sessionData.CSRFExpiry = csrfExpiry
	}

	// Capture client fingerprint
	sessionData.IP = ClientIP(s.c)
	sessionData.UserAgent = s.c.Get("User-Agent")

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	// Create new session with new ID + fingerprint
	dbSession := &queries.Session{
		ID:        hashSessionID(newID),
		UserID:    sessionData.UserID,
		Data:      string(jsonData),
		ExpiresAt: s.expiresAt,
	}

	if err := s.store.querier.CreateSession(context.Background(), dbSession); err != nil {
		return err
	}

	s.store.deleteSession(context.Background(), s.id)
	if s.store.sessionCache != nil {
		s.store.sessionCache.Invalidate(s.id)
		s.store.sessionCache.Set(newID, cache.CachedSessionData{
			UserID:        sessionData.UserID,
			Name:          sessionData.Name,
			Email:         sessionData.Email,
			Avatar:        sessionData.Avatar,
			EmailVerified: sessionData.EmailVerified,
			Role:          sessionData.Role,
			CSRFToken:     sessionData.CSRFToken,
			CSRFExpiry:    sessionData.CSRFExpiry,
			IP:            sessionData.IP,
			UserAgent:     sessionData.UserAgent,
			ExpiresAt:     s.expiresAt,
		})
	}

	s.id = newID

	// Update cookie
	s.c.Cookie(&fiber.Cookie{
		Name:     s.store.sessionName,
		Value:    s.id,
		Path:     "/",
		HTTPOnly: true,
		Secure:   s.store.secure,
		SameSite: "Lax",
		MaxAge:   int(s.expiresAt.Sub(time.Now()).Seconds()),
	})

	return nil
}

// Flash sets a flash message cookie (short-lived, one-time use)
// The flash message will be available on the next request and then cleared
func (s *Store) Flash(c *fiber.Ctx, key string, value string) {
	// Set flash cookie with short expiry (5 minutes)
	c.Cookie(&fiber.Cookie{
		Name:     "flash_" + key,
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		MaxAge:   300, // 5 minutes
	})
}

// GetFlash retrieves and clears a flash message cookie
func (s *Store) GetFlash(c *fiber.Ctx, key string) string {
	cookieName := "flash_" + key
	value := c.Cookies(cookieName)

	if value != "" {
		// Clear the flash cookie after reading (one-time use)
		c.ClearCookie(cookieName)
	}

	return value
}


// isPageRequest returns true if the request is an actual page navigation
// (Inertia XHR or initial HTML load), not an API/asset/DevTools side request.
func isPageRequest(c *fiber.Ctx) bool {
	// Inertia requests (JS-driven page transitions)
	if c.Get("X-Inertia") == "true" {
		return true
	}
	// Initial page load (Accept: text/html)
	accept := c.Get("Accept")
	return strings.Contains(accept, "text/html")
}

// ClientIP extracts the real client IP behind Cloudflare proxy or reverse proxy.
// Falls back to c.IP() and normalizes Cloudflare PROXY protocol format.
func ClientIP(c *fiber.Ctx) string {
	if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	return c.IP()
}

// setFingerprint captures and persists the client IP+UserAgent for an existing session
// that was loaded from cache but didn't have a fingerprint yet.
func (s *Store) setFingerprint(c *fiber.Ctx, sessionID string, cached *cache.CachedSessionData) {
	ip := ClientIP(c)
	ua := c.Get("User-Agent")

	// Reconstruct full session data JSON with fingerprint
	newData := SessionData{
		UserID:        cached.UserID,
		Name:          cached.Name,
		Email:         cached.Email,
		Avatar:        cached.Avatar,
		EmailVerified: cached.EmailVerified,
		Role:          cached.Role,
		CSRFToken:     cached.CSRFToken,
		CSRFExpiry:    cached.CSRFExpiry,
		IP:            ip,
		UserAgent:     ua,
	}
	newJSON, err := json.Marshal(newData)
	if err != nil {
		slog.Error("setFingerprint marshal error", "error", err)
		return
	}

	// Update DB
	dbSession := &queries.Session{
		ID:        hashSessionID(sessionID),
		UserID:    cached.UserID,
		Data:      string(newJSON),
		ExpiresAt: cached.ExpiresAt,
	}
	if err := s.querier.UpdateSession(context.Background(), dbSession); err != nil {
		slog.Error("setFingerprint db update error", "error", err)
	}

	// Update cache
	if s.sessionCache != nil {
		s.sessionCache.Set(sessionID, cache.CachedSessionData{
			UserID:        cached.UserID,
			Name:          cached.Name,
			Email:         cached.Email,
			Avatar:        cached.Avatar,
			EmailVerified: cached.EmailVerified,
			Role:          cached.Role,
			CSRFToken:     cached.CSRFToken,
			CSRFExpiry:    cached.CSRFExpiry,
			IP:            ip,
			UserAgent:     ua,
		ExpiresAt:     cached.ExpiresAt,
	})
	}

	slog.Debug("fingerprint captured for existing session",
		"session_id", sessionID, "ip", ip, "ua", ua)
}

// deleteSession removes a session from the database and logs any error.
// The raw session ID is hashed before the DB query — the DB stores only
// SHA-256(rawID), so a database leak cannot be used to hijack sessions.
func (s *Store) deleteSession(ctx context.Context, sessionID string) {
	if err := s.querier.DeleteSession(ctx, hashSessionID(sessionID)); err != nil {
		slog.Error("failed to delete session", "error", err)
	}
}

// SetSecure sets the Secure flag on session cookies.
// Should be set to true in production with HTTPS.
func (s *Store) SetSecure(secure bool) {
	s.secure = secure
}

// CreateAuthenticatedSession sets user authentication data on the session and saves it.
// Helper to avoid duplicating the Set/Save pattern across multiple handlers.
func (s *Store) CreateAuthenticatedSession(c *fiber.Ctx, userID int64, name, email, avatar, role string, emailVerified bool) error {
	sess, err := s.Get(c)
	if err != nil {
		return err
	}
	sess.Set("user_id", userID)
	sess.Set("name", name)
	sess.Set("email", email)
	sess.Set("avatar", avatar)
	sess.Set("email_verified", emailVerified)
	sess.Set("role", role)
	return sess.Save()
}
