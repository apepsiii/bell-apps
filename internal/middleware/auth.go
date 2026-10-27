package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
)

// AlreadyLoggedIn redirects to /admin if user is already logged in
func AlreadyLoggedIn() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(config.GetCookieName())
			if err == nil && cookie.Value != "" && handler.ValidateAdminSession(cookie.Value) {
				return c.Redirect(http.StatusSeeOther, "/admin")
			}
			return next(c)
		}
	}
}

// AdminAuth checks if user is authenticated as admin
func AdminAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent browser caching of protected admin content
			c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Response().Header().Set("Pragma", "no-cache")
			c.Response().Header().Set("Expires", "0")

			cookie, err := c.Cookie(config.GetCookieName())
			if err != nil || cookie.Value == "" {
				return adminUnauthorized(c)
			}

			session, ok := handler.GetAdminSession(cookie.Value)
			if !ok {
				// Clear any invalid or expired cookie
				expired := &http.Cookie{
					Name:     config.GetCookieName(),
					Value:    "",
					Path:     "/",
					Expires:  time.Unix(0, 0),
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				}
				c.SetCookie(expired)
				return adminUnauthorized(c)
			}

			// Bind identity to the request so downstream authorization and
			// audit logging can rely on it. Never trust a client-supplied role.
			c.Set("user", session.Username)
			c.Set("role", session.Role)
			return next(c)
		}
	}
}

func adminUnauthorized(c echo.Context) error {
	if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		c.Request().Header.Get("Accept") == "application/json" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Sesi telah berakhir. Silakan login kembali.",
		})
	}
	return c.Redirect(http.StatusSeeOther, "/login")
}
