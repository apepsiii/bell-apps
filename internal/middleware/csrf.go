package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
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
	return base64.URLEncoding.EncodeToString(b), nil
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
				// Generate and set token for GET requests
				token, err := generateToken()
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate CSRF token")
				}

				// Store token
				csrfStore.mu.Lock()
				csrfStore.tokens[token] = &CSRFToken{
					Token:     token,
					ExpiresAt: time.Now().Add(csrfTokenTTL),
				}
				csrfStore.mu.Unlock()

				// Set cookie
				cookie := &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: false, // Need to be accessible by JavaScript
					Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
					SameSite: http.SameSiteStrictMode,
					MaxAge:   int(csrfTokenTTL.Seconds()),
				}
				c.SetCookie(cookie)

				// Also set in context for template rendering
				c.Set("csrf_token", token)

				return next(c)
			}

			// For POST, PUT, PATCH, DELETE - validate token
			var tokenFromRequest string

			// Check header first
			tokenFromRequest = c.Request().Header.Get(csrfHeaderName)

			// If not in header, check form field
			if tokenFromRequest == "" {
				tokenFromRequest = c.FormValue(csrfFormField)
			}

			// If still not found, check cookie (for single-page apps)
			if tokenFromRequest == "" {
				cookie, err := c.Cookie(csrfCookieName)
				if err == nil {
					tokenFromRequest = cookie.Value
				}
			}

			// Validate token
			if tokenFromRequest == "" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "CSRF token missing",
					"message": "Token CSRF tidak ditemukan",
				})
			}

			// Check if token exists and is valid
			csrfStore.mu.RLock()
			storedToken, exists := csrfStore.tokens[tokenFromRequest]
			csrfStore.mu.RUnlock()

			if !exists {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "Invalid CSRF token",
					"message": "Token CSRF tidak valid",
				})
			}

			// Check if token is expired
			if time.Now().After(storedToken.ExpiresAt) {
				// Remove expired token
				csrfStore.mu.Lock()
				delete(csrfStore.tokens, tokenFromRequest)
				csrfStore.mu.Unlock()

				return c.JSON(http.StatusForbidden, map[string]string{
					"error":   "CSRF token expired",
					"message": "Token CSRF sudah kadaluarsa, silakan refresh halaman",
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
