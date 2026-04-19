package handler

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

type Operator struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Photo     string `json:"photo"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func OperatorLogin(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		c.Logger().Infof("Login attempt for username: %s", username)

		if username == "" || password == "" {
			c.Logger().Error("Empty username or password")
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Username dan password harus diisi"})
		}

		var op Operator
		var photoNull, phoneNull sql.NullString
		err := db.QueryRow("SELECT id, username, password, name, phone, photo, is_active FROM operators WHERE username = ? AND is_active = 1", username).
			Scan(&op.ID, &op.Username, &op.Password, &op.Name, &phoneNull, &photoNull, &op.IsActive)

		if err == sql.ErrNoRows {
			c.Logger().Errorf("User not found: %s", username)
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
		} else if err != nil {
			c.Logger().Errorf("Database error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan sistem: " + err.Error()})
		}

		op.Phone = phoneNull.String
		op.Photo = photoNull.String

		if err := bcrypt.CompareHashAndPassword([]byte(op.Password), []byte(password)); err != nil {
			c.Logger().Errorf("Password mismatch for user: %s", username)
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
		}

		token := GenerateSessionToken()

		cookie := new(http.Cookie)
		cookie.Name = "operator_session"
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		c.SetCookie(cookie)

		sessionKey := "session_" + token
		_, err = db.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", sessionKey, strconv.Itoa(op.ID))
		if err != nil {
			c.Logger().Errorf("Failed to save session: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan session: " + err.Error()})
		}

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
}

func OperatorLogout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("operator_session")
		if err == nil {
			sessionKey := "session_" + cookie.Value
			db.Exec("DELETE FROM attendance_settings WHERE setting_key = ?", sessionKey)
		}

		cookie = new(http.Cookie)
		cookie.Name = "operator_session"
		cookie.Value = ""
		cookie.Path = "/"
		cookie.MaxAge = -1
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{"message": "Logout berhasil"})
	}
}

func OperatorAuth(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("operator_session")
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/operator/login")
			}

			sessionKey := "session_" + cookie.Value
			var operatorID string
			err = db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = ?", sessionKey).Scan(&operatorID)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/operator/login")
			}

			c.Set("operator_id", operatorID)
			return next(c)
		}
	}
}

func GetOperatorProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		operatorID := c.Get("operator_id").(string)

		var op Operator
		err := db.QueryRow("SELECT id, username, name, phone, photo, created_at FROM operators WHERE id = ?", operatorID).
			Scan(&op.ID, &op.Username, &op.Name, &op.Phone, &op.Photo, &op.CreatedAt)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Operator tidak ditemukan"})
		}

		return c.JSON(http.StatusOK, op)
	}
}

func UpdateOperatorProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		operatorID := c.Get("operator_id").(string)

		type UpdateRequest struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
		}

		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		_, err := db.Exec("UPDATE operators SET name = ?, phone = ? WHERE id = ?", req.Name, req.Phone, operatorID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update profil"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Profil berhasil diupdate"})
	}
}

func ChangeOperatorPassword(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		operatorID := c.Get("operator_id").(string)

		type PasswordRequest struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		var req PasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		var currentHash string
		err := db.QueryRow("SELECT password FROM operators WHERE id = ?", operatorID).Scan(&currentHash)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Password lama salah"})
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi password"})
		}

		_, err = db.Exec("UPDATE operators SET password = ? WHERE id = ?", string(newHash), operatorID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update password"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password berhasil diubah"})
	}
}

func GetOperatorPrayerStats(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		today := time.Now().Format("2006-01-02")

		var totalStudents int
		db.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)

		var dzuhurHadir, dzuhurPMS int
		db.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Dzuhur' AND status = 'Hadir'", today).Scan(&dzuhurHadir)
		db.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Dzuhur' AND status = 'PMS'", today).Scan(&dzuhurPMS)

		var asharHadir, asharPMS int
		db.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Ashar' AND status = 'Hadir'", today).Scan(&asharHadir)
		db.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM prayer_logs WHERE date = ? AND prayer_type = 'Ashar' AND status = 'PMS'", today).Scan(&asharPMS)

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
}

func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
