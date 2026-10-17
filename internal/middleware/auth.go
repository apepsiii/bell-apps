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
			if err != nil || !handler.ValidateAdminSession(cookie.Value) {
				// Clear any invalid or expired cookie
				if err == nil && cookie != nil && cookie.Value != "" {
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
				}

				if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" ||
					c.Request().Header.Get("Accept") == "application/json" {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Sesi telah berakhir. Silakan login kembali.",
					})
				}
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return next(c)
		}
	}
}
