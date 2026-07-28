package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

// AlreadyLoggedIn redirects to /admin if user is already logged in
func AlreadyLoggedIn() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(config.GetCookieName())
			if err == nil && cookie.Value == config.GetSecretKey() {
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
			cookie, err := c.Cookie(config.GetCookieName())
			if err != nil || cookie.Value != config.GetSecretKey() {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return next(c)
		}
	}
}
