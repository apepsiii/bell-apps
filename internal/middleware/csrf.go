package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "csrf_token"
	csrfHeaderName  = "X-CSRF-Token"
	csrfFormField   = "csrf_token"
	csrfTokenTTL    = 24 * time.Hour
)

// CSRFToken represents a CSRF token with expiration
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFStore stores CSRF tokens in memory
type CSRFStore struct {
	mu     sync.RWMutex
	tokens map[string]*CSRFToken
}

var csrfStore = &CSRFStore{
	tokens: make(map[string]*CSRFToken),
}

// generateToken creates a new random CSRF token
func generateToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// isCSRFExempt checks if a path is exempt from CSRF checks (e.g. IoT devices, machine APIs, login)
func isCSRFExempt(path string) bool {
	exemptPaths := []string{
		"/api/attendance/verify-face",
		"/api/attendance/record",
		"/api/attendance/prayer",
		"/admin/face/save-descriptor",
		"/admin/face/register",
		"/api/login",
		"/api/logout",
		"/api/operator/login",
		"/api/operator/logout",
		"/api/student/login",
		"/api/student/logout",
		"/api/guru/login",
		"/api/guru/logout",
		"/guru/logout",
		"/api/auth/magic-request",
		"/api/webhook/whatsapp",
		"/api/webhook/onesender",
	}

	for _, ep := range exemptPaths {
		if path == ep || strings.HasPrefix(path, ep+"/") {
			return true
		}
	}
	return false
}

// CSRF middleware protects against Cross-Site Request Forgery attacks
func CSRF() echo.MiddlewareFunc {
	// Cleanup expired tokens periodically
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			csrfStore.cleanup()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
			if c.Request().Method == http.MethodGet ||
				c.Request().Method == http.MethodHead ||
				c.Request().Method == http.MethodOptions {

				var token string
				if cookie, err := c.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
					csrfStore.mu.RLock()
					stored, ok := csrfStore.tokens[cookie.Value]
					csrfStore.mu.RUnlock()
					if ok && time.Now().Before(stored.ExpiresAt) {
						token = cookie.Value
					}
				}

				if token == "" || strings.Contains(token, "=") {
					var err error
					token, err = generateToken()
					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate CSRF token")
					}

					csrfStore.mu.Lock()
					csrfStore.tokens[token] = &CSRFToken{
						Token:     token,
						ExpiresAt: time.Now().Add(csrfTokenTTL),
					}
					csrfStore.mu.Unlock()
				}

				// Set/refresh cookie
				cookie := &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: false, // Need to be accessible by JavaScript
					Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
					SameSite: http.SameSiteLaxMode,
					MaxAge:   int(csrfTokenTTL.Seconds()),
				}
				c.SetCookie(cookie)

				// Also set in context for template rendering
				c.Set("csrf_token", token)

				return next(c)
			}

			// Check if request is exempt (IoT hardware, public login endpoints)
			if isCSRFExempt(c.Request().URL.Path) {
				return next(c)
			}

			// For POST, PUT, PATCH, DELETE - validate token
			// Strictly extract token from header or form field (NEVER fallback to cookie alone)
			tokenFromRequest := c.Request().Header.Get(csrfHeaderName)
			if tokenFromRequest == "" {
				tokenFromRequest = c.FormValue(csrfFormField)
			}

			if tokenFromRequest == "" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "CSRF token missing",
					"message": "Token CSRF tidak ditemukan. Sertakan header X-CSRF-Token atau form csrf_token.",
				})
			}

			// Verify token exists in memory store
			csrfStore.mu.RLock()
			storedToken, exists := csrfStore.tokens[tokenFromRequest]
			if !exists {
				// Fallback: check without or with '=' padding
				unpadded := strings.TrimRight(tokenFromRequest, "=")
				storedToken, exists = csrfStore.tokens[unpadded]
				if !exists {
					for _, pad := range []string{"=", "=="} {
						if st, ok := csrfStore.tokens[unpadded+pad]; ok {
							storedToken = st
							exists = true
							tokenFromRequest = unpadded + pad
							break
						}
					}
				}
			}
			csrfStore.mu.RUnlock()

			// In-memory recovery on server restart:
			// If server restarted, memory store is empty, but client sent genuine cookie that matches token
			cookie, err := c.Cookie(csrfCookieName)
			cookieVal := ""
			if err == nil {
				cookieVal = cookie.Value
			}

			tokensMatch := (cookieVal != "" && tokenFromRequest != "" &&
				(cookieVal == tokenFromRequest || strings.TrimRight(cookieVal, "=") == strings.TrimRight(tokenFromRequest, "=")))

			if !exists && tokensMatch && len(tokenFromRequest) >= 32 {
				storedToken = &CSRFToken{
					Token:     tokenFromRequest,
					ExpiresAt: time.Now().Add(csrfTokenTTL),
				}
				csrfStore.mu.Lock()
				csrfStore.tokens[tokenFromRequest] = storedToken
				if cookieVal != tokenFromRequest {
					csrfStore.tokens[cookieVal] = storedToken
				}
				csrfStore.mu.Unlock()
				exists = true
			}

			if !exists {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "Invalid CSRF token",
					"message": "Token CSRF tidak valid",
				})
			}

			// Check if token is expired
			if time.Now().After(storedToken.ExpiresAt) {
				csrfStore.mu.Lock()
				delete(csrfStore.tokens, tokenFromRequest)
				csrfStore.mu.Unlock()

				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "CSRF token expired",
					"message": "Token CSRF sudah kadaluarsa, silakan refresh halaman",
				})
			}

			// Double Submit Cookie verification:
			// The token submitted in the header/form must match the token sent in the cookie
			if err != nil || !tokensMatch {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "CSRF token mismatch",
					"message": "Token CSRF tidak cocok dengan cookie sesi",
				})
			}

			return next(c)
		}
	}
}

// cleanup removes expired tokens from the store
func (s *CSRFStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, data := range s.tokens {
		if now.After(data.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

// GetCSRFToken extracts CSRF token from context (for rendering in templates)
func GetCSRFToken(c echo.Context) string {
	if token, ok := c.Get("csrf_token").(string); ok {
		return token
	}
	return ""
}
