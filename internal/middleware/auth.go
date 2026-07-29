package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

// AdminAuth middleware: validates admin session cookie.
func AdminAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(config.GetCookieName())
			if err != nil || cookie.Value != config.GetSecretKey() {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return next(c)
		}
	}
}

// AlreadyLoggedIn redirects to /admin if admin session cookie is valid.
func AlreadyLoggedIn() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cookie, err := c.Cookie(config.GetCookieName()); err == nil && cookie.Value == config.GetSecretKey() {
				return c.Redirect(http.StatusSeeOther, "/admin")
			}
			return next(c)
		}
	}
}
