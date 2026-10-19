package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
	"belsekolah/pkg/utils"
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
		expiresAt := time.Now().Add(7 * 24 * time.Hour)

		// Opportunistic cleanup of expired sessions
		_, _ = db.Exec("DELETE FROM student_sessions WHERE expires_at < ?", time.Now())

		cookie := new(http.Cookie)
		cookie.Name = StudentSessionCookie
		cookie.Value = token
		cookie.Path = "/"
		cookie.Expires = expiresAt
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
		c.SetCookie(cookie)

		// Save session to student_sessions with TTL
		_, err = db.Exec("REPLACE INTO student_sessions (token, student_id, expires_at) VALUES (?, ?, ?)",
			token, id, expiresAt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan sesi"})
		}

		// Also keep legacy write for transitional safety
		_, _ = db.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
			studentSessionPrefix+token, id)

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
			db.Exec("DELETE FROM student_sessions WHERE token = ?", cookie.Value)
			db.Exec("DELETE FROM attendance_settings WHERE setting_key = ?", studentSessionPrefix+cookie.Value)
		}

		if pCookie, err := c.Cookie(ParentSessionCookie); err == nil {
			db.Exec("DELETE FROM parent_sessions WHERE token = ?", pCookie.Value)
			c.SetCookie(&http.Cookie{
				Name:   ParentSessionCookie,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
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
			var expiresAt time.Time
			err = db.QueryRow("SELECT student_id, expires_at FROM student_sessions WHERE token = ?",
				cookie.Value).Scan(&studentID, &expiresAt)
			if err == nil {
				if time.Now().After(expiresAt) {
					// Expired session
					db.Exec("DELETE FROM student_sessions WHERE token = ?", cookie.Value)
					return studentUnauthorized(c)
				}
			} else {
				// Fallback to legacy attendance_settings
				err = db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = ?",
					studentSessionPrefix+cookie.Value).Scan(&studentID)
				if err != nil {
					return studentUnauthorized(c)
				}
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
			rfid, nis, name, photo, password, birthday, parentName, parentPhone string
			className                                                            sql.NullString
		)
		err := db.QueryRow(`
			SELECT s.rfid_uid, s.nis, s.name, COALESCE(s.photo,''), COALESCE(s.password,''),
			       COALESCE(s.birthday,''), COALESCE(s.parent_name,''), COALESCE(s.parent_phone,''), c.name
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, studentID).
			Scan(&rfid, &nis, &name, &photo, &password, &birthday, &parentName, &parentPhone, &className)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		now := time.Now()
		today := now.Format("2006-01-02")

		// --- Today's attendance ---
		timeIn, timeOut, todayStatus := "", "", "Belum Presensi"
		rows, err := db.Query(
			"SELECT status, timestamp FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date = ? ORDER BY timestamp ASC",
			studentID, rfid, today)
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
				case "Sakit":
					if timeIn == "" {
						timeIn = timeOnly
					}
					todayStatus = "Sakit"
				case "Izin":
					if timeIn == "" {
						timeIn = timeOnly
					}
					todayStatus = "Izin"
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

		// --- Yesterday's attendance (for insight comparison) ---
		yesterdayDate := now.AddDate(0, 0, -1).Format("2006-01-02")
		yesterdayTimeIn, yesterdayStatus := "", "Belum Presensi"
		yRows, err := db.Query(
			"SELECT status, timestamp FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date = ? ORDER BY timestamp ASC",
			studentID, rfid, yesterdayDate)
		if err == nil {
			defer yRows.Close()
			for yRows.Next() {
				var st, ts string
				if yRows.Scan(&st, &ts) != nil {
					continue
				}
				timeOnly := ""
				if len(ts) >= 16 {
					timeOnly = ts[11:16]
				}
				switch st {
				case "Datang", "Hadir":
					if yesterdayTimeIn == "" {
						yesterdayTimeIn = timeOnly
						yesterdayStatus = "Hadir"
					}
				case "Terlambat":
					if yesterdayTimeIn == "" {
						yesterdayTimeIn = timeOnly
					}
					yesterdayStatus = "Terlambat"
				case "Sakit":
					if yesterdayTimeIn == "" {
						yesterdayTimeIn = timeOnly
					}
					yesterdayStatus = "Sakit"
				case "Izin":
					if yesterdayTimeIn == "" {
						yesterdayTimeIn = timeOnly
					}
					yesterdayStatus = "Izin"
				}
			}
		}

		// --- Total points (attendance legacy + english XP) ---
		totalPoints := 0
		db.QueryRow("SELECT COALESCE(SUM(points_change),0) FROM student_points WHERE student_id = ?",
			studentID).Scan(&totalPoints)
		// Add English Daily Quest XP so the home balance reflects everything.
		var englishXP int
		db.QueryRow("SELECT COALESCE(total_xp,0) FROM english_streaks WHERE student_id = ?", studentID).Scan(&englishXP)
		totalPoints += englishXP

		// --- Pancawaluya dual-track points ---
		var achievementPts, violationPts int
		db.QueryRow("SELECT COALESCE(SUM(points),0) FROM student_achievement_points WHERE student_id = ?", studentID).Scan(&achievementPts)
		db.QueryRow("SELECT COALESCE(SUM(points),0) FROM student_violation_points WHERE student_id = ?", studentID).Scan(&violationPts)
		pancawalaPoints := achievementPts - violationPts
		if pancawalaPoints < 0 {
			pancawalaPoints = 0
		}

		// --- Academic year (e.g. "2025/2026") ---
		nowMonth := now.Month()
		nowYear := now.Year()
		var academicYear string
		if nowMonth >= 7 {
			academicYear = fmt.Sprintf("%d/%d", nowYear, nowYear+1)
		} else {
			academicYear = fmt.Sprintf("%d/%d", nowYear-1, nowYear)
		}

		// --- This month mini stats (fix: += to accumulate Datang+Hadir both) ---
		month := now.Format("01")
		year := now.Format("2006")
		stats := map[string]int{"present": 0, "late": 0, "sick": 0, "permission": 0, "alpha": 0}
		statRows, err := db.Query(`
			SELECT status, COUNT(DISTINCT date) FROM attendance_logs
			WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND strftime('%m', date) = ? AND strftime('%Y', date) = ?
			GROUP BY status`, studentID, rfid, month, year)
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
					stats["present"] += cnt // += accumulates both statuses
				case "Terlambat":
					stats["late"] += cnt
				case "Sakit":
					stats["sick"] += cnt
				case "Izin":
					stats["permission"] += cnt
				case "Alpha":
					stats["alpha"] += cnt
				}
			}
		}

		// --- Last 5 point history entries ---
		type phItem struct {
			Description  string `json:"description"`
			PointsChange int    `json:"points_change"`
			Timestamp    string `json:"timestamp"`
		}
		pointHistory := []phItem{}
		phRows, phErr := db.Query(
			"SELECT COALESCE(description,''), points_change, timestamp FROM student_points WHERE student_id = ? ORDER BY timestamp DESC LIMIT 5",
			studentID)
		if phErr == nil {
			defer phRows.Close()
			for phRows.Next() {
				var ph phItem
				if phRows.Scan(&ph.Description, &ph.PointsChange, &ph.Timestamp) == nil {
					pointHistory = append(pointHistory, ph)
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
				"id":           studentID,
				"nis":          nis,
				"name":         name,
				"class":        className.String,
				"photo":        photoURL,
				"birthday":     birthday,
				"parent_name":  parentName,
				"parent_phone": parentPhone,
			},
			"today": map[string]interface{}{
				"date":         today,
				"status":       todayStatus,
				"time_in":      timeIn,
				"time_out":     timeOut,
				"holiday_name": holidayName,
			},
			"yesterday": map[string]interface{}{
				"date":    yesterdayDate,
				"status":  yesterdayStatus,
				"time_in": yesterdayTimeIn,
			},
			"points":             totalPoints,
			"weighted_score":     totalPoints,
			"pancawaluya_points": pancawalaPoints,
			"violation_points":   violationPts,
			"academic_year":      academicYear,
			"month":              stats,
			"next_bell":          nextBell,
			"announcements":      announcements,
			"point_history":      pointHistory,
			"using_default_pin":  password == "",
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

// UpdateStudentPortalProfile lets a student update their own birthday, parent_name, parent_phone.
func UpdateStudentPortalProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		type ProfileReq struct {
			Birthday    string `json:"birthday"`
			ParentName  string `json:"parent_name"`
			ParentPhone string `json:"parent_phone"`
		}
		var req ProfileReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		// Build dynamic update
		setClauses := []string{}
		args := []interface{}{}
		if req.Birthday != "" {
			setClauses = append(setClauses, "birthday = ?")
			args = append(args, req.Birthday)
		}
		if req.ParentName != "" {
			setClauses = append(setClauses, "parent_name = ?")
			args = append(args, req.ParentName)
		}
		if req.ParentPhone != "" {
			setClauses = append(setClauses, "parent_phone = ?")
			phone := utils.FormatPhone(req.ParentPhone)
			args = append(args, phone)
		}
		if len(setClauses) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tidak ada data yang diubah"})
		}
		args = append(args, studentID)
		query := "UPDATE students SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
		if _, err := db.Exec(query, args...); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan profil"})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Profil berhasil diperbarui"})
	}
}

// UploadStudentPhoto lets a student upload or replace their profile photo.
func UploadStudentPhoto(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		// Get student NIS for filename
		var nis string
		if err := db.QueryRow("SELECT nis FROM students WHERE id = ?", studentID).Scan(&nis); err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		var photoFile string

		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Gagal membuka file"})
			}
			defer src.Close()
			validated, verr := utils.ValidatePhotoFile(src, file.Filename, file.Size)
			if verr != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": verr.Error()})
			}
			ext := utils.GetPhotoExtension(file.Filename)
			filename := utils.BuildPhotoFilename(nis, ext)
			dstPath := filepath.Join(config.GetPhotoPath(), filename)
			if err := utils.SaveUploadedFile(validated, dstPath); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan foto"})
			}
			photoFile = filename
		} else {
			// Try base64 captured photo
			captured := c.FormValue("captured_photo")
			if captured == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tidak ada foto yang dikirim"})
			}
			filename := utils.BuildPhotoFilename(nis, ".jpg")
			dstPath := filepath.Join(config.GetPhotoPath(), filename)
			if err := utils.SaveBase64Image(captured, dstPath); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan foto"})
			}
			photoFile = filename
		}

		if _, err := db.Exec("UPDATE students SET photo = ? WHERE id = ?", photoFile, studentID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memperbarui data foto"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":   "Foto berhasil diperbarui",
			"photo_url": "/assets/photos/" + photoFile,
		})
	}
}

// GetStudentClaimRules returns the list of claimable point rules for students.
func GetStudentClaimRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ensure point_rules table exists
		_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS point_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT,
			name TEXT,
			points INTEGER,
			description TEXT
		)`)

		type Rule struct {
			ID          int    `json:"id"`
			Category    string `json:"category"`
			Name        string `json:"name"`
			Points      int    `json:"points"`
			Description string `json:"description"`
		}

		rows, err := db.Query("SELECT id, COALESCE(category,'Umum'), COALESCE(name,''), COALESCE(points,0), COALESCE(description,'') FROM point_rules ORDER BY category, name")
		if err != nil {
			return c.JSON(http.StatusOK, []Rule{})
		}
		defer rows.Close()

		var rules []Rule
		for rows.Next() {
			var r Rule
			if err := rows.Scan(&r.ID, &r.Category, &r.Name, &r.Points, &r.Description); err != nil {
				continue
			}
			rules = append(rules, r)
		}
		if rules == nil {
			rules = []Rule{}
		}
		return c.JSON(http.StatusOK, rules)
	}
}

// GetMyClaimsHistory returns the logged-in student's point claims.
func GetMyClaimsHistory(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		sid, _ := strconv.Atoi(studentID)

		// Ensure point_claims table exists
		_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS point_claims (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER NOT NULL,
			description TEXT,
			evidence TEXT,
			status TEXT DEFAULT 'pending',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by INTEGER
		)`)

		type Claim struct {
			ID             int    `json:"id"`
			RuleName       string `json:"rule_name"`
			Points         int    `json:"points"`
			Description    string `json:"description"`
			Evidence       string `json:"evidence"`
			Status         string `json:"status"`
			SubmittedAt    string `json:"submitted_at"`
			RejectionNotes string `json:"rejection_notes"`
		}

		rows, err := db.Query(`
			SELECT pc.id, COALESCE(pr.name, 'Klaim Poin'), COALESCE(pr.points, 0),
			       COALESCE(pc.description,''), COALESCE(pc.evidence,''),
			       COALESCE(pc.status,'pending'), COALESCE(pc.submitted_at,'')
			FROM point_claims pc
			LEFT JOIN point_rules pr ON pc.rule_id = pr.id
			WHERE pc.student_id = ? OR pc.student_id = ?
			ORDER BY pc.submitted_at DESC`, studentID, sid)
		if err != nil {
			return c.JSON(http.StatusOK, []Claim{})
		}
		defer rows.Close()

		var claims []Claim
		for rows.Next() {
			var cl Claim
			if err := rows.Scan(&cl.ID, &cl.RuleName, &cl.Points, &cl.Description, &cl.Evidence,
				&cl.Status, &cl.SubmittedAt); err != nil {
				continue
			}
			cl.RejectionNotes = ""
			claims = append(claims, cl)
		}
		if claims == nil {
			claims = []Claim{}
		}
		return c.JSON(http.StatusOK, claims)
	}
}

// SubmitStudentClaim lets a student submit a new point claim (with optional photo evidence).
func SubmitStudentClaim(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)

		ruleIDStr := c.FormValue("rule_id")
		description := strings.TrimSpace(c.FormValue("description"))
		caption := c.FormValue("caption")

		ruleID, err := strconv.Atoi(ruleIDStr)
		if err != nil || ruleID == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Pilih jenis kegiatan terlebih dahulu"})
		}
		if description == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Deskripsi kegiatan harus diisi"})
		}
		if caption != "" {
			description = description + " [" + caption + "]"
		}

		// Verify rule exists
		var ruleName string
		if err := db.QueryRow("SELECT name FROM point_rules WHERE id = ?", ruleID).Scan(&ruleName); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Jenis kegiatan tidak ditemukan"})
		}

		// Handle evidence upload
		evidencePath := ""
		file, err := c.FormFile("evidence")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				// Get student NIS for filename
				var nis string
				db.QueryRow("SELECT nis FROM students WHERE id = ?", studentID).Scan(&nis)
				ext := utils.GetPhotoExtension(file.Filename)
				filename := fmt.Sprintf("claim_%s_%d%s", nis, time.Now().UnixNano(), ext)
				dstPath := filepath.Join(config.GetPhotoPath(), filename)
				validated, verr := utils.ValidatePhotoFile(src, file.Filename, file.Size)
				if verr == nil {
					if saveErr := utils.SaveUploadedFile(validated, dstPath); saveErr == nil {
						evidencePath = "/assets/photos/" + filename
					}
				}
			}
		} else if b64 := c.FormValue("evidence_base64"); b64 != "" {
			var nis string
			db.QueryRow("SELECT nis FROM students WHERE id = ?", studentID).Scan(&nis)
			filename := fmt.Sprintf("claim_%s_%d.jpg", nis, time.Now().UnixNano())
			dstPath := filepath.Join(config.GetPhotoPath(), filename)
			if utils.SaveBase64Image(b64, dstPath) == nil {
				evidencePath = "/assets/photos/" + filename
			}
		}

		// Ensure point_claims table exists
		_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS point_claims (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			rule_id INTEGER NOT NULL,
			description TEXT,
			evidence TEXT,
			status TEXT DEFAULT 'pending',
			submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			reviewed_by INTEGER
		)`)

		sidInt, _ := strconv.Atoi(studentID)
		_, err = db.Exec(`
			INSERT INTO point_claims (student_id, rule_id, description, evidence, status)
			VALUES (?, ?, ?, ?, 'pending')`, sidInt, ruleID, description, evidencePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengajukan klaim: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Klaim berhasil diajukan, menunggu persetujuan admin"})
	}
}

// MarkAnnouncementRead marks a single announcement as read for this student (client-side simulation via 200 OK).
func MarkAnnouncementRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Simple acknowledgement — read state is tracked client-side in this version.
		// A future version can persist to announcement_reads table.
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}
}

// MarkAllAnnouncementsRead marks all announcements as read (client-side acknowledgement).
func MarkAllAnnouncementsRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}
}
