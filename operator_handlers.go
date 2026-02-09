package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Operator struct
type Operator struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"` // Never send password in JSON
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Photo     string `json:"photo"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// OperatorLoginHandler handles operator login
func (a *App) OperatorLoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	c.Logger().Infof("Login attempt for username: %s", username)

	if username == "" || password == "" {
		c.Logger().Error("Empty username or password")
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Username dan password harus diisi"})
	}

	// Fetch operator from database
	var op Operator
	var photoNull, phoneNull sql.NullString
	err := a.DB.QueryRow("SELECT id, username, password, name, phone, photo, is_active FROM operators WHERE username = ? AND is_active = 1", username).
		Scan(&op.ID, &op.Username, &op.Password, &op.Name, &phoneNull, &photoNull, &op.IsActive)

	if err == sql.ErrNoRows {
		c.Logger().Errorf("User not found: %s", username)
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
	} else if err != nil {
		c.Logger().Errorf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan sistem: " + err.Error()})
	}

	// Handle NULL values
	op.Phone = phoneNull.String
	op.Photo = photoNull.String

	c.Logger().Infof("User found: %s (ID: %d)", op.Name, op.ID)

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(op.Password), []byte(password)); err != nil {
		c.Logger().Errorf("Password mismatch for user: %s", username)
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
	}

	c.Logger().Info("Password verified successfully")

	// Generate session token
	token := generateSessionToken()

	// Store session in cookie
	cookie := new(http.Cookie)
	cookie.Name = "operator_session"
	cookie.Value = token
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	c.Logger().Infof("Session cookie set for user: %s", username)

	// Store operator ID in session (in production, use Redis or similar)
	sessionKey := "session_" + token
	_, err = a.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", sessionKey, strconv.Itoa(op.ID))
	if err != nil {
		c.Logger().Errorf("Failed to save session: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan session: " + err.Error()})
	}

	c.Logger().Infof("Login successful for user: %s", username)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login berhasil",
		"operator": map[string]interface{}{
			"id":    op.ID,
			"name":  op.Name,
			"phone": op.Phone,
			"photo": op.Photo,
		},
	})
}

// OperatorLogoutHandler handles operator logout
func (a *App) OperatorLogoutHandler(c echo.Context) error {
	cookie, err := c.Cookie("operator_session")
	if err == nil {
		// Delete session from database
		sessionKey := "session_" + cookie.Value
		a.DB.Exec("DELETE FROM attendance_settings WHERE setting_key = ?", sessionKey)
	}

	// Clear cookie
	cookie = new(http.Cookie)
	cookie.Name = "operator_session"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.MaxAge = -1
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]string{"message": "Logout berhasil"})
}

// OperatorAuthMiddleware checks if operator is authenticated
func (a *App) OperatorAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("operator_session")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/operator/login")
		}

		// Verify session exists
		sessionKey := "session_" + cookie.Value
		var operatorID string
		err = a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = ?", sessionKey).Scan(&operatorID)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/operator/login")
		}

		// Store operator ID in context
		c.Set("operator_id", operatorID)
		return next(c)
	}
}

// GetOperatorProfileHandler returns operator profile
func (a *App) GetOperatorProfileHandler(c echo.Context) error {
	operatorID := c.Get("operator_id").(string)

	var op Operator
	err := a.DB.QueryRow("SELECT id, username, name, phone, photo, created_at FROM operators WHERE id = ?", operatorID).
		Scan(&op.ID, &op.Username, &op.Name, &op.Phone, &op.Photo, &op.CreatedAt)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Operator tidak ditemukan"})
	}

	return c.JSON(http.StatusOK, op)
}

// UpdateOperatorProfileHandler updates operator profile
func (a *App) UpdateOperatorProfileHandler(c echo.Context) error {
	operatorID := c.Get("operator_id").(string)

	type UpdateRequest struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	_, err := a.DB.Exec("UPDATE operators SET name = ?, phone = ? WHERE id = ?", req.Name, req.Phone, operatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update profil"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Profil berhasil diupdate"})
}

// ChangeOperatorPasswordHandler changes operator password
func (a *App) ChangeOperatorPasswordHandler(c echo.Context) error {
	operatorID := c.Get("operator_id").(string)

	type PasswordRequest struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	var req PasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	// Get current password
	var currentHash string
	err := a.DB.QueryRow("SELECT password FROM operators WHERE id = ?", operatorID).Scan(&currentHash)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan"})
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Password lama salah"})
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi password"})
	}

	// Update password
	_, err = a.DB.Exec("UPDATE operators SET password = ? WHERE id = ?", string(newHash), operatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update password"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password berhasil diubah"})
}

// OperatorPrayerStatsHandler returns prayer statistics for operator dashboard
func (a *App) OperatorPrayerStatsHandler(c echo.Context) error {
	today := time.Now().Format("2006-01-02")

	var totalStudents int
	a.DB.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)

	// Dzuhur stats
	var dzuhurHadir, dzuhurPMS int
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Dzuhur' AND status = 'Hadir'", today).Scan(&dzuhurHadir)
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Dzuhur' AND status = 'PMS'", today).Scan(&dzuhurPMS)

	// Ashar stats
	var asharHadir, asharPMS int
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Ashar' AND status = 'Hadir'", today).Scan(&asharHadir)
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Ashar' AND status = 'PMS'", today).Scan(&asharPMS)

	// Calculate belum sholat
	dzuhurBelum := totalStudents - dzuhurHadir - dzuhurPMS
	asharBelum := totalStudents - asharHadir - asharPMS

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_students": totalStudents,
		"dzuhur": map[string]int{
			"hadir": dzuhurHadir,
			"pms":   dzuhurPMS,
			"belum": dzuhurBelum,
		},
		"ashar": map[string]int{
			"hadir": asharHadir,
			"pms":   asharPMS,
			"belum": asharBelum,
		},
		"date": today,
	})
}

// Helper function to generate session token
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
