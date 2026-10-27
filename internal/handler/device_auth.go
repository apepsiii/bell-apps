package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

// deviceTokenHeaders lists the headers a device may use to present its token.
var deviceTokenHeaders = []string{"X-Device-Token", "X-Api-Token"}

// ValidateDeviceToken reports whether the request carries the shared device
// token configured in attendance_settings. It fails closed: an unconfigured
// token, a missing header, or a mismatch all deny access.
//
// The token is read from the database on every call so an operator can rotate
// it from the settings page without redeploying.
func ValidateDeviceToken(db *sql.DB, c echo.Context) bool {
	if db == nil {
		return false
	}

	var configured string
	err := db.QueryRow(
		"SELECT setting_value FROM attendance_settings WHERE setting_key = 'device_token'",
	).Scan(&configured)
	if err != nil || configured == "" {
		return false
	}

	presented := ""
	for _, h := range deviceTokenHeaders {
		if v := c.Request().Header.Get(h); v != "" {
			presented = v
			break
		}
	}
	if presented == "" {
		// IoT clients that cannot set headers may pass it as a query param.
		presented = c.QueryParam("device_token")
	}
	if presented == "" {
		presented = c.FormValue("device_token")
	}
	if presented == "" {
		return false
	}

	return subtleEqual(presented, configured)
}

// DeviceAuth is Echo middleware that rejects requests without a valid device
// token. Use it for machine-to-machine endpoints such as RFID attendance and
// biometric descriptor delivery.
func DeviceAuth(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !ValidateDeviceToken(db, c) {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "Token perangkat tidak valid",
				})
			}
			return next(c)
		}
	}
}

// subtleEqual compares two strings without leaking timing information.
func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
