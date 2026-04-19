package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username == config.AdminUser && password == config.AdminPass {
			cookie := new(http.Cookie)
			cookie.Name = config.CookieName
			cookie.Value = config.SecretKey
			cookie.Path = "/"
			cookie.Expires = time.Now().Add(24 * time.Hour)
			cookie.HttpOnly = true
			cookie.SameSite = http.SameSiteLaxMode
			c.SetCookie(cookie)
			return c.Redirect(http.StatusSeeOther, "/admin")
		}
		return c.Redirect(http.StatusSeeOther, "/login?error=1")
	}
}

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie := new(http.Cookie)
		cookie.Name = config.CookieName
		cookie.Value = ""
		cookie.Path = "/"
		cookie.MaxAge = -1
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/login")
	}
}
