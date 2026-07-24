package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	// StudentSessionCookie is the cookie name for student portal sessions.
	StudentSessionCookie = "student_session"
	// StudentDefaultPIN is the fallback PIN when a student has not set one yet.
	StudentDefaultPIN = "123456"
	// studentSessionPrefix prefixes session keys stored in attendance_settings.
	studentSessionPrefix = "student_session_"
)

// StudentLogin authenticates a student via NIS (or nis_siswa) + PIN.
// Students with an empty password column use StudentDefaultPIN.
func StudentLogin(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		nis := strings.TrimSpace(c.FormValue("nis"))
		pin := strings.TrimSpace(c.FormValue("pin"))

		if nis == "" || pin == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "NIS dan PIN harus diisi"})
		}

		var (
			id       int
			storedN  string
			name     string
			password string
			status   string
		)
		err := db.QueryRow(
			"SELECT id, nis, name, COALESCE(password,''), COALESCE(status,'active') FROM students WHERE nis = ? OR nis_siswa = ? LIMIT 1",
			nis, nis,
		).Scan(&id, &storedN, &name, &password, &status)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "NIS atau PIN salah"})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan sistem"})
		}

		if status != "active" {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akun siswa nonaktif. Hubungi admin sekolah."})
		}

		if password == "" {
			if pin != StudentDefaultPIN {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "NIS atau PIN salah"})
			}
		} else if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(pin)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "NIS atau PIN salah"})
		}

		token := GenerateSessionToken()

		cookie := new(http.Cookie)
		cookie.Name = StudentSessionCookie
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		c.SetCookie(cookie)

		// REPLACE INTO works on both SQLite and MySQL (setting_key is PK).
		_, err = db.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
			studentSessionPrefix+token, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan sesi"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Login berhasil",
			"student": map[string]interface{}{
				"id":   id,
				"name": name,
			},
		})
	}
}

// StudentLogout clears the student session.
func StudentLogout(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if cookie, err := c.Cookie(StudentSessionCookie); err == nil {
			db.Exec("DELETE FROM attendance_settings WHERE setting_key = ?", studentSessionPrefix+cookie.Value)
		}

		cookie := new(http.Cookie)
		cookie.Name = StudentSessionCookie
		cookie.Value = ""
		cookie.Path = "/"
		cookie.MaxAge = -1
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{"message": "Logout berhasil"})
	}
}

// StudentAuth guards student portal pages (redirect) and API (401 JSON).
func StudentAuth(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(StudentSessionCookie)
			if err != nil || cookie.Value == "" {
				return studentUnauthorized(c)
			}

			var studentID string
			err = db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = ?",
				studentSessionPrefix+cookie.Value).Scan(&studentID)
			if err != nil {
				return studentUnauthorized(c)
			}

			c.Set("student_id", studentID)
			return next(c)
		}
	}
}

func studentUnauthorized(c echo.Context) error {
	if strings.HasPrefix(c.Request().URL.Path, "/api/") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Sesi berakhir, silakan login kembali"})
	}
	return c.Redirect(http.StatusSeeOther, "/student/login")
}

func studentIDFromCtx(c echo.Context) string {
	if v, ok := c.Get("student_id").(string); ok {
		return v
	}
	return ""
}

// GetStudentDashboard returns everything the student home page needs in one call.
func GetStudentDashboard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		var (
			rfid, nis, name, photo, password string
			className                        sql.NullString
		)
		err := db.QueryRow(`
			SELECT s.rfid_uid, s.nis, s.name, COALESCE(s.photo,''), COALESCE(s.password,''), c.name
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, studentID).
			Scan(&rfid, &nis, &name, &photo, &password, &className)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		now := time.Now()
		today := now.Format("2006-01-02")

		// --- Today's attendance ---
		timeIn, timeOut, todayStatus := "", "", "Belum Presensi"
		rows, err := db.Query(
			"SELECT status, timestamp FROM attendance_logs WHERE rfid_uid = ? AND date = ? ORDER BY timestamp ASC",
			rfid, today)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var st, ts string
				if rows.Scan(&st, &ts) != nil {
					continue
				}
				timeOnly := ""
				if len(ts) >= 16 {
					timeOnly = ts[11:16]
				}
				switch st {
				case "Datang", "Hadir":
					if timeIn == "" {
						timeIn = timeOnly
						todayStatus = "Hadir"
					}
				case "Terlambat":
					if timeIn == "" {
						timeIn = timeOnly
					}
					todayStatus = "Terlambat"
				case "Pulang":
					timeOut = timeOnly
				}
			}
		}

		// Holiday check (only relevant when no attendance yet)
		holidayName := ""
		if todayStatus == "Belum Presensi" {
			db.QueryRow("SELECT name FROM holidays WHERE date = ?", today).Scan(&holidayName)
		}

		// --- Total points ---
		totalPoints := 0
		db.QueryRow("SELECT COALESCE(SUM(points_change),0) FROM student_points WHERE student_id = ?",
			studentID).Scan(&totalPoints)

		// --- This month mini stats ---
		month := now.Format("01")
		year := now.Format("2006")
		stats := map[string]int{"present": 0, "late": 0, "sick": 0, "permission": 0, "alpha": 0}
		statRows, err := db.Query(`
			SELECT status, COUNT(DISTINCT date) FROM attendance_logs
			WHERE rfid_uid = ? AND strftime('%m', date) = ? AND strftime('%Y', date) = ?
			GROUP BY status`, rfid, month, year)
		if err == nil {
			defer statRows.Close()
			for statRows.Next() {
				var st string
				var cnt int
				if statRows.Scan(&st, &cnt) != nil {
					continue
				}
				switch st {
				case "Datang", "Hadir":
					stats["present"] = cnt
				case "Terlambat":
					stats["late"] = cnt
				case "Sakit":
					stats["sick"] = cnt
				case "Izin":
					stats["permission"] = cnt
				case "Alpha":
					stats["alpha"] = cnt
				}
			}
		}

		// --- Next bell ---
		var nextBell interface{}
		var bellTime, bellLabel string
		nowHM := now.Format("15:04")
		err = db.QueryRow("SELECT time, label FROM schedules WHERE time > ? ORDER BY time ASC LIMIT 1", nowHM).
			Scan(&bellTime, &bellLabel)
		if err != nil {
			err = db.QueryRow("SELECT time, label FROM schedules ORDER BY time ASC LIMIT 1").
				Scan(&bellTime, &bellLabel)
		}
		if err == nil {
			nextBell = map[string]string{"time": bellTime, "label": bellLabel}
		}

		// --- Latest announcements ---
		announcements := []map[string]string{}
		annRows, err := db.Query(
			"SELECT title, message, created_at FROM announcements ORDER BY created_at DESC LIMIT 3")
		if err == nil {
			defer annRows.Close()
			for annRows.Next() {
				var title, message, createdAt string
				if annRows.Scan(&title, &message, &createdAt) != nil {
					continue
				}
				announcements = append(announcements, map[string]string{
					"title": title, "message": message, "created_at": createdAt,
				})
			}
		}

		photoURL := ""
		if photo != "" {
			photoURL = "/assets/photos/" + photo
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"student": map[string]interface{}{
				"id":    studentID,
				"nis":   nis,
				"name":  name,
				"class": className.String,
				"photo": photoURL,
			},
			"today": map[string]interface{}{
				"date":         today,
				"status":       todayStatus,
				"time_in":      timeIn,
				"time_out":     timeOut,
				"holiday_name": holidayName,
			},
			"points":           totalPoints,
			"month":            stats,
			"next_bell":        nextBell,
			"announcements":    announcements,
			"using_default_pin": password == "",
		})
	}
}

// GetStudentPortalProfile returns the logged-in student's full profile.
func GetStudentPortalProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		var (
			nis, nisSiswa, name, photo, birthday, parentName, parentPhone, password string
			className                                                                sql.NullString
		)
		err := db.QueryRow(`
			SELECT s.nis, COALESCE(s.nis_siswa,''), s.name, COALESCE(s.photo,''),
			       COALESCE(s.birthday,''), COALESCE(s.parent_name,''), COALESCE(s.parent_phone,''),
			       COALESCE(s.password,''), c.name
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, studentID).
			Scan(&nis, &nisSiswa, &name, &photo, &birthday, &parentName, &parentPhone, &password, &className)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		photoURL := ""
		if photo != "" {
			photoURL = "/assets/photos/" + photo
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"nis":               nis,
			"nis_siswa":         nisSiswa,
			"name":              name,
			"class":             className.String,
			"photo":             photoURL,
			"birthday":          birthday,
			"parent_name":       parentName,
			"parent_phone":      parentPhone,
			"using_default_pin": password == "",
		})
	}
}

// ChangeStudentPIN lets a student replace their PIN (old PIN + new PIN).
func ChangeStudentPIN(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		type PINRequest struct {
			OldPIN string `json:"old_pin"`
			NewPIN string `json:"new_pin"`
		}
		var req PINRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		if len(req.NewPIN) < 4 || len(req.NewPIN) > 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "PIN baru harus 4-8 digit angka"})
		}
		for _, ch := range req.NewPIN {
			if ch < '0' || ch > '9' {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "PIN baru harus berupa angka"})
			}
		}

		var currentHash string
		err := db.QueryRow("SELECT COALESCE(password,'') FROM students WHERE id = ?", studentID).Scan(&currentHash)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan"})
		}

		if currentHash == "" {
			if req.OldPIN != StudentDefaultPIN {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "PIN lama salah"})
			}
		} else if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPIN)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "PIN lama salah"})
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPIN), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi PIN"})
		}

		if _, err := db.Exec("UPDATE students SET password = ? WHERE id = ?", string(newHash), studentID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan PIN baru"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "PIN berhasil diubah"})
	}
}

// GetMyQRCard delegates to GetStudentIDCard with the session student's ID.
func GetMyQRCard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.SetParamNames("id")
		c.SetParamValues(studentIDFromCtx(c))
		return GetStudentIDCard(db)(c)
	}
}

// GetMyCalendar delegates to GetStudentCalendar with the session student's ID.
func GetMyCalendar(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := c.QueryParams()
		q.Set("id", studentIDFromCtx(c))
		c.Request().URL.RawQuery = q.Encode()
		return GetStudentCalendar(db)(c)
	}
}

// GetMyPoints delegates to GetStudentPointProfile with the session student's ID.
func GetMyPoints(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.SetParamNames("id")
		c.SetParamValues(studentIDFromCtx(c))
		return GetStudentPointProfile(db)(c)
	}
}
