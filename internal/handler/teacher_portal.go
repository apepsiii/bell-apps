package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	TeacherSessionCookie   = "teacher_session"
	TeacherDefaultPassword = "guru123"
)

// TeacherLogin handles authentication for teachers using NIP and password.
func TeacherLogin(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		nip := strings.TrimSpace(c.FormValue("nip"))
		password := strings.TrimSpace(c.FormValue("password"))

		if nip == "" || password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "NIP dan password wajib diisi",
			})
		}

		var (
			id             int
			storedNIP      string
			name           string
			role           string
			storedPassword string
		)

		err := db.QueryRow(
			`SELECT id, COALESCE(nip, ''), name, COALESCE(role, 'Guru'), COALESCE(password, '')
			 FROM staff
			 WHERE nip = ?
			 LIMIT 1`,
			nip,
		).Scan(&id, &storedNIP, &name, &role, &storedPassword)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "NIP atau password salah",
			})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Terjadi kesalahan internal server",
			})
		}

		// Password validation
		if storedPassword == "" {
			// If not set yet, accept either the default password or NIP as initial password
			if password != TeacherDefaultPassword && password != nip {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "NIP atau password salah",
				})
			}
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "NIP atau password salah",
				})
			}
		}

		token := GenerateSessionToken()
		expiresAt := time.Now().Add(7 * 24 * time.Hour)

		// Clean up expired sessions
		_, _ = db.Exec("DELETE FROM teacher_sessions WHERE expires_at < ?", time.Now())

		// Set HTTP-only Cookie
		cookie := new(http.Cookie)
		cookie.Name = TeacherSessionCookie
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = expiresAt
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
		c.SetCookie(cookie)

		// Save session into DB
		_, err = db.Exec(
			`INSERT OR REPLACE INTO teacher_sessions (token, staff_id, expires_at)
			 VALUES (?, ?, ?)`,
			token, id, expiresAt,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Gagal menyimpan sesi login",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Login berhasil",
			"teacher": map[string]interface{}{
				"id":   id,
				"name": name,
				"nip":  storedNIP,
				"role": role,
			},
		})
	}
}

// TeacherLogout terminates the teacher session
func TeacherLogout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if cookie, err := c.Cookie(TeacherSessionCookie); err == nil {
			_, _ = db.Exec("DELETE FROM teacher_sessions WHERE token = ?", cookie.Value)
		}

		expired := &http.Cookie{
			Name:     TeacherSessionCookie,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(expired)

		if strings.HasPrefix(c.Request().URL.Path, "/api/") {
			return c.JSON(http.StatusOK, map[string]string{"message": "Logout berhasil"})
		}
		return c.Redirect(http.StatusSeeOther, "/guru/login")
	}
}

// TeacherAuth middleware ensures that requests are authenticated as a teacher
func TeacherAuth(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(TeacherSessionCookie)
			if err != nil || cookie.Value == "" {
				return teacherUnauthorized(c)
			}

			var (
				staffID   int
				expiresAt time.Time
				name      string
				nip       string
				role      string
			)

			err = db.QueryRow(
				`SELECT ts.staff_id, ts.expires_at, s.name, COALESCE(s.nip, ''), COALESCE(s.role, 'Guru')
				 FROM teacher_sessions ts
				 JOIN staff s ON s.id = ts.staff_id
				 WHERE ts.token = ?`,
				cookie.Value,
			).Scan(&staffID, &expiresAt, &name, &nip, &role)

			if err != nil {
				return teacherUnauthorized(c)
			}

			if time.Now().After(expiresAt) {
				_, _ = db.Exec("DELETE FROM teacher_sessions WHERE token = ?", cookie.Value)
				return teacherUnauthorized(c)
			}

			c.Set("teacher_id", staffID)
			c.Set("teacher_name", name)
			c.Set("teacher_nip", nip)
			c.Set("teacher_role", role)

			return next(c)
		}
	}
}

func teacherUnauthorized(c echo.Context) error {
	if strings.HasPrefix(c.Request().URL.Path, "/api/") {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Sesi guru berakhir, silakan login kembali",
		})
	}
	return c.Redirect(http.StatusSeeOther, "/guru/login")
}

// TeacherIDFromCtx retrieves the logged in teacher's ID
func TeacherIDFromCtx(c echo.Context) int {
	if val, ok := c.Get("teacher_id").(int); ok {
		return val
	}
	if valStr, ok := c.Get("teacher_id").(string); ok {
		id, _ := strconv.Atoi(valStr)
		return id
	}
	return 0
}

// TeacherChangePassword handles updating a teacher's password
func TeacherChangePassword(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		if teacherID == 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Tidak terautentikasi"})
		}

		oldPass := strings.TrimSpace(c.FormValue("old_password"))
		newPass := strings.TrimSpace(c.FormValue("new_password"))

		if len(newPass) < 6 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password baru minimal 6 karakter"})
		}

		var (
			storedPass string
			nip        string
		)
		err := db.QueryRow("SELECT COALESCE(password, ''), COALESCE(nip, '') FROM staff WHERE id = ?", teacherID).Scan(&storedPass, &nip)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca data akun"})
		}

		// Verify old password
		if storedPass == "" {
			if oldPass != TeacherDefaultPassword && oldPass != nip {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password lama tidak cocok"})
			}
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(storedPass), []byte(oldPass)); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password lama tidak cocok"})
			}
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memproses password"})
		}

		_, err = db.Exec("UPDATE staff SET password = ? WHERE id = ?", string(hashed), teacherID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan password baru"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password berhasil diperbarui"})
	}
}

// GetTeacherProfile returns profile & list of assigned teaching classes
func GetTeacherProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		if teacherID == 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Tidak terautentikasi"})
		}

		var (
			name, nip, phone, role, gender string
		)
		err := db.QueryRow(
			`SELECT name, COALESCE(nip, ''), COALESCE(phone, ''), COALESCE(role, 'Guru'), COALESCE(gender, 'L')
			 FROM staff WHERE id = ?`,
			teacherID,
		).Scan(&name, &nip, &phone, &role, &gender)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Data guru tidak ditemukan"})
		}

		// Fetch teaching assignments
		rows, err := db.Query(`
			SELECT ta.id, ta.subject_id, sub.name, sub.code, COALESCE(sub.category, 'muatan_kejuruan'),
			       ta.class_id, cls.name, ta.academic_year, ta.semester,
			       COALESCE(NULLIF(ta.banner_url, ''), COALESCE(sub.banner_url, ''))
			FROM teacher_assignments ta
			JOIN subjects sub ON sub.id = ta.subject_id
			JOIN classes cls ON cls.id = ta.class_id
			WHERE ta.staff_id = ?
			ORDER BY cls.name ASC, sub.name ASC
		`, teacherID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca penugasan mengajar"})
		}
		defer rows.Close()

		assignments := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, subID, clsID, semester int
				subName, subCode, subCat, clsName, academicYear, bannerURL string
			)
			if err := rows.Scan(&id, &subID, &subName, &subCode, &subCat, &clsID, &clsName, &academicYear, &semester, &bannerURL); err == nil {
				if bannerURL == "" {
					bannerURL = DefaultBannerForSubject(subCat, subCode, subName)
				}
				assignments = append(assignments, map[string]interface{}{
					"id":            id,
					"subject_id":    subID,
					"subject_name":  subName,
					"subject_code":  subCode,
					"category":      subCat,
					"class_id":      clsID,
					"class_name":    clsName,
					"academic_year": academicYear,
					"semester":      semester,
					"banner_url":    bannerURL,
				})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"profile": map[string]interface{}{
				"id":     teacherID,
				"name":   name,
				"nip":    nip,
				"phone":  phone,
				"role":   role,
				"gender": gender,
			},
			"assignments": assignments,
		})
	}
}

