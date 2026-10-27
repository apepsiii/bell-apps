package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
)

type adminSession struct {
	expiresAt time.Time
	username  string
	role      string
}

var (
	adminSessions   = make(map[string]*adminSession)
	adminSessionsMu sync.RWMutex
)

func generateAdminSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// AdminSessionInfo is the public view of an authenticated admin session.
type AdminSessionInfo struct {
	Username string
	Role     string
}

// GetAdminSession returns the identity bound to a session token, if valid.
func GetAdminSession(token string) (AdminSessionInfo, bool) {
	adminSessionsMu.RLock()
	sess, ok := adminSessions[token]
	adminSessionsMu.RUnlock()
	if !ok {
		return AdminSessionInfo{}, false
	}
	if time.Now().After(sess.expiresAt) {
		adminSessionsMu.Lock()
		delete(adminSessions, token)
		adminSessionsMu.Unlock()
		return AdminSessionInfo{}, false
	}
	return AdminSessionInfo{Username: sess.username, Role: sess.role}, true
}

func ValidateAdminSession(token string) bool {
	_, ok := GetAdminSession(token)
	return ok
}

// ResolveAdminRole looks up the role for an admin username. It fails closed:
// an unknown or inactive account has no role and therefore no privileges.
func ResolveAdminRole(db *sql.DB, username string) string {
	if db == nil || username == "" {
		return ""
	}
	var role string
	var isActive bool
	err := db.QueryRow(
		"SELECT role, is_active FROM admin_roles WHERE username = ?",
		username,
	).Scan(&role, &isActive)
	if err != nil || !isActive {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(role))
}

func init() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			adminSessionsMu.Lock()
			for tok, sess := range adminSessions {
				if now.After(sess.expiresAt) {
					delete(adminSessions, tok)
				}
			}
			adminSessionsMu.Unlock()
		}
	}()
}

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(config.GetCookieName())
		if err == nil && cookie.Value != "" {
			adminSessionsMu.Lock()
			delete(adminSessions, cookie.Value)
			adminSessionsMu.Unlock()
		}

		// Comprehensive cookie expiration
		expired := &http.Cookie{
			Name:     config.GetCookieName(),
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
		}
		c.SetCookie(expired)

		// Set anti-cache headers to purge bfcache
		c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().Header().Set("Expires", "0")

		if strings.Contains(c.Request().Header.Get("Accept"), "application/json") || c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": "Berhasil keluar akun",
			})
		}
		return c.Redirect(http.StatusSeeOther, "/login")
	}
}

func Login(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username != config.GetAdminUser() {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		storedPass := config.GetAdminPass()
		var valid bool
		isBcrypt := strings.HasPrefix(storedPass, "$2a$") || strings.HasPrefix(storedPass, "$2b$")

		if isBcrypt {
			if err := bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password)); err == nil {
				valid = true
			}
		} else {
			// Legacy plaintext verification: allow transitional access but emit strong security warning
			if storedPass == password {
				valid = true
				slog.Warn("SECURITY WARNING: Admin authenticated using plaintext password. Please update config.yaml with a bcrypt hash!")
			}
		}

		if username == "admin" && (password == "admin123" || storedPass == "admin123") {
			slog.Warn("SECURITY WARNING: Admin is using the default factory password 'admin123'. Change it immediately!")
		}

		if !valid {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		// Refuse to establish a session while the account still uses factory
		// default credentials. This is the last line of defence for deployments
		// that never ran the setup wizard.
		if isDefaultAdminCredential(username, storedPass) {
			slog.Error("SECURITY: Login blocked because admin credentials are still at factory defaults. Update config.yaml before signing in.")
			return c.Redirect(http.StatusSeeOther, "/login?error=default_credentials")
		}

		role := ResolveAdminRole(db, username)
		if role == "" {
			slog.Error("SECURITY: Login blocked because no active admin role is assigned to this account.")
			return c.Redirect(http.StatusSeeOther, "/login?error=no_role")
		}

		token, err := generateAdminSessionToken()
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		adminSessionsMu.Lock()
		adminSessions[token] = &adminSession{
			expiresAt: time.Now().Add(24 * time.Hour),
			username:  username,
			role:      role,
		}
		adminSessionsMu.Unlock()

		cookie := new(http.Cookie)
		cookie.Name = config.GetCookieName()
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/admin")
	}
}

// isDefaultAdminCredential reports whether the configured admin password is
// still one of the well-known factory defaults.
func isDefaultAdminCredential(username, storedPass string) bool {
	if username == "admin" && storedPass == "admin123" {
		return true
	}
	return storedPass == "admin-secret-key-change-me" ||
		storedPass == "change-this-password" ||
		storedPass == "change-this-secret-key" ||
		storedPass == ""
}


