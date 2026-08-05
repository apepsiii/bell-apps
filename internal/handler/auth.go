package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
)

type adminSession struct {
	expiresAt time.Time
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

func ValidateAdminSession(token string) bool {
	adminSessionsMu.RLock()
	sess, ok := adminSessions[token]
	adminSessionsMu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(sess.expiresAt) {
		adminSessionsMu.Lock()
		delete(adminSessions, token)
		adminSessionsMu.Unlock()
		return false
	}
	return true
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
		expired := new(http.Cookie)
		expired.Name = config.GetCookieName()
		expired.Value = ""
		expired.Path = "/"
		expired.MaxAge = -1
		c.SetCookie(expired)
		return c.Redirect(http.StatusSeeOther, "/login")
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username != config.GetAdminUser() {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		storedPass := config.GetAdminPass()
		var valid bool
		if err := bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(password)); err == nil {
			valid = true
		} else if storedPass == password {
			valid = true
		}

		if !valid {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		token, err := generateAdminSessionToken()
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=1")
		}

		adminSessionsMu.Lock()
		adminSessions[token] = &adminSession{expiresAt: time.Now().Add(24 * time.Hour)}
		adminSessionsMu.Unlock()

		cookie := new(http.Cookie)
		cookie.Name = config.GetCookieName()
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/admin")
	}
}


