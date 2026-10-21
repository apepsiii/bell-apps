package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	"belsekolah/internal/repository"
	"belsekolah/internal/router"
)

// TestMain sets the working directory to the repository root so relative paths
// like "views/admin.html" and os.DirFS(".") work consistently across all tests.
func TestMain(m *testing.M) {
	if _, err := os.Stat("views"); os.IsNotExist(err) {
		// If "views" doesn't exist in current dir, we are running from tests/
		if err := os.Chdir(".."); err != nil {
			log.Fatalf("Failed to chdir to repository root: %v", err)
		}
	}
	os.Exit(m.Run())
}

// BulkAttendance structs for test payloads
type BulkAttendanceItem struct {
	StudentID int    `json:"student_id"`
	Status    string `json:"status"`
}

type BulkAttendanceRequest struct {
	ClassID int                  `json:"class_id"`
	Date    string               `json:"date"`
	Items   []BulkAttendanceItem `json:"items"`
}

// TestApp provides an in-memory test harness satisfying router.AppHandlers.
type TestApp struct {
	DB *sql.DB
}

// App type alias so test signatures expecting *App compile without modification.
type App = TestApp

// setupTestApp initializes an in-memory SQLite database, runs migrations and seeds.
func setupTestApp(t *testing.T) (*TestApp, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)

	repository.RunMigrations(db)
	repository.SeedDefaultData(db)
	repository.SeedPointRules(db)
	repository.MigrationAuditTrail(db)
	_ = repository.MigrationSemesterScores(db)

	app := &TestApp{DB: db}
	os.MkdirAll(config.GetUploadPath(), 0755)

	teardown := func() {
		db.Close()
		os.RemoveAll(config.GetUploadPath())
	}
	return app, teardown
}

// seedStudent creates a test student record in the in-memory database.
func seedStudent(t *testing.T, app *TestApp, nis, name, password string, active bool) {
	status := "active"
	if !active {
		status = "inactive"
	}
	hash := ""
	if password != "" {
		h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		hash = string(h)
	}
	_, err := app.DB.Exec(
		"INSERT INTO students (rfid_uid, nis, name, password, status) VALUES (?, ?, ?, ?, ?)",
		"rfid-"+nis, nis, name, hash, status,
	)
	if err != nil {
		t.Fatalf("failed to seed student %s: %v", nis, err)
	}
}

// --- Specific Handlers Tested in Suite ---

func (a *TestApp) RecordAttendanceHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		rfid = c.FormValue("rfid")
	}
	studentIDParam := c.QueryParam("student_id")
	if studentIDParam == "" {
		studentIDParam = c.FormValue("student_id")
	}

	// Device Authentication
	var expectedToken string
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='device_token'").Scan(&expectedToken)
	if expectedToken != "" {
		providedToken := c.Request().Header.Get("X-Device-Token")
		if providedToken == "" {
			providedToken = c.QueryParam("token")
		}
		if providedToken == "" {
			providedToken = c.FormValue("token")
		}
		if providedToken != expectedToken {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token perangkat tidak valid"})
		}
	}

	if rfid == "" && studentIDParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID kosong"})
	}

	var studentID int
	var name, userType, photo, identityNo, className string
	var photoNull, classNameNull sql.NullString
	var studentStatus string
	var err error

	if studentIDParam != "" {
		var rfidNull sql.NullString
		err = a.DB.QueryRow(`
			SELECT s.id, s.rfid_uid, s.name, s.photo, s.nis, c.name, s.status
			FROM students s 
			LEFT JOIN classes c ON s.class_id = c.id 
			WHERE s.id=?`, studentIDParam).Scan(&studentID, &rfidNull, &name, &photoNull, &identityNo, &classNameNull, &studentStatus)
		if err == nil {
			rfid = rfidNull.String
		}
	} else {
		err = a.DB.QueryRow(`
			SELECT s.id, s.name, s.photo, s.nis, c.name, s.status
			FROM students s 
			LEFT JOIN classes c ON s.class_id = c.id 
			WHERE s.rfid_uid=?`, rfid).Scan(&studentID, &name, &photoNull, &identityNo, &classNameNull, &studentStatus)
	}

	if err == nil {
		if studentStatus == "inactive" {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akun tidak aktif. Hubungi admin.", "name": name})
		}
		userType = "Siswa"
		photo = photoNull.String
		className = classNameNull.String
	} else {
		err = a.DB.QueryRow("SELECT name, nip FROM staff WHERE rfid_uid=?", rfid).Scan(&name, &identityNo)
		if err == nil {
			userType = "Staff"
			className = ""
		} else {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Kartu tidak dikenali"})
		}
	}

	now := time.Now()
	timeStr := now.Format("15:04")
	dateStr := now.Format("2006-01-02")
	timestamp := now.Format("2006-01-02 15:04:05")

	var arrStart, arrEnd, depStart, depEnd string
	rows, _ := a.DB.Query("SELECT setting_key, setting_value FROM attendance_settings")
	settings := make(map[string]string)
	if rows != nil {
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			settings[k] = v
		}
		rows.Close()
	}

	arrStart, arrEnd = settings["arrival_start"], settings["arrival_end"]
	depStart, depEnd = settings["departure_start"], settings["departure_end"]

	status := "Unknown"
	if timeStr >= arrStart && timeStr <= arrEnd {
		status = "Datang"
		var count int
		a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date=? AND status='Datang'", studentID, rfid, dateStr).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen datang hari ini", "name": name})
		}
	} else if timeStr >= depStart && timeStr <= depEnd {
		status = "Pulang"
		var count int
		a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date=? AND status='Pulang'", studentID, rfid, dateStr).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen pulang hari ini", "name": name})
		}
	} else {
		if timeStr > arrEnd && timeStr < depStart {
			status = "Terlambat"
			var count int
			a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date=? AND (status='Datang' OR status='Terlambat')", studentID, rfid, dateStr).Scan(&count)
			if count > 0 {
				return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen hari ini", "name": name})
			}
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Di luar jam absen", "name": name})
		}
	}

	var pointsChange int
	var pointDescription string
	if userType == "Siswa" {
		switch status {
		case "Datang":
			pointsChange = 20
			pointDescription = "Hadir Tepat Waktu (+20)"
		case "Terlambat":
			pointsChange = -10
			pointDescription = "Terlambat (-10)"
		case "Pulang":
			pointsChange = 0
			pointDescription = "Pulang Tepat Waktu (+0)"
		}
	}

	var sID interface{}
	if studentID > 0 {
		sID = studentID
	}
	_, err = a.DB.Exec("INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, ?, ?, ?, 'RFID', ?, ?)",
		sID, rfid, name, userType, status, timestamp, dateStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	if userType == "Siswa" && pointsChange != 0 && studentID > 0 {
		a.DB.Exec("INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, ?, ?, ?)",
			studentID, pointsChange, pointDescription, "System-Auto")
	}

	response := map[string]interface{}{
		"status":      "success",
		"message":     "Absen " + status + " Berhasil",
		"name":        name,
		"type":        userType,
		"time":        timeStr,
		"photo":       photo,
		"identity_no": identityNo,
		"class_name":  className,
	}
	if userType == "Siswa" && pointsChange != 0 {
		response["points"] = pointsChange
		response["point_description"] = pointDescription
	}

	return c.JSON(http.StatusOK, response)
}

func (a *TestApp) PrayerAttendanceHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		rfid = c.FormValue("rfid")
	}
	studentIDParam := c.QueryParam("student_id")
	if studentIDParam == "" {
		studentIDParam = c.FormValue("student_id")
	}

	var expectedToken string
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='device_token'").Scan(&expectedToken)
	if expectedToken != "" {
		providedToken := c.Request().Header.Get("X-Device-Token")
		if providedToken == "" {
			providedToken = c.QueryParam("token")
		}
		if providedToken == "" {
			providedToken = c.FormValue("token")
		}
		if providedToken != expectedToken {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token perangkat tidak valid"})
		}
	}

	if rfid == "" && studentIDParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID kosong"})
	}

	var studentID int
	var name, studentStatus string
	var classNameNull sql.NullString
	var err error

	if studentIDParam != "" {
		var rfidNull sql.NullString
		err = a.DB.QueryRow(`
			SELECT s.id, s.rfid_uid, s.name, c.name, s.status
			FROM students s 
			LEFT JOIN classes c ON s.class_id = c.id 
			WHERE s.id=?`, studentIDParam).Scan(&studentID, &rfidNull, &name, &classNameNull, &studentStatus)
		if err == nil {
			rfid = rfidNull.String
		}
	} else {
		err = a.DB.QueryRow(`
			SELECT s.id, s.name, c.name, s.status
			FROM students s 
			LEFT JOIN classes c ON s.class_id = c.id 
			WHERE s.rfid_uid=?`, rfid).Scan(&studentID, &name, &classNameNull, &studentStatus)
	}

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Kartu tidak dikenali"})
	}
	if studentStatus == "inactive" {
		return c.JSON(http.StatusForbidden, map[string]string{"message": "Akun tidak aktif. Hubungi admin.", "name": name})
	}

	now := time.Now()
	timeStr := now.Format("15:04")
	dateStr := now.Format("2006-01-02")
	timestamp := now.Format("2006-01-02 15:04:05")

	var dzStart, dzEnd, asStart, asEnd string
	rows, _ := a.DB.Query("SELECT setting_key, setting_value FROM attendance_settings WHERE setting_key IN ('dzuhur_start', 'dzuhur_end', 'ashar_start', 'ashar_end')")
	if rows != nil {
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			switch k {
			case "dzuhur_start":
				dzStart = v
			case "dzuhur_end":
				dzEnd = v
			case "ashar_start":
				asStart = v
			case "ashar_end":
				asEnd = v
			}
		}
		rows.Close()
	}

	prayerType := "Unknown"
	if dzStart != "" && dzEnd != "" && timeStr >= dzStart && timeStr <= dzEnd {
		prayerType = "Dzuhur"
	} else if asStart != "" && asEnd != "" && timeStr >= asStart && timeStr <= asEnd {
		prayerType = "Ashar"
	}

	if prayerType == "Unknown" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Di luar waktu sholat berjamaah", "name": name})
	}

	var count int
	a.DB.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date=? AND prayer_type=?", studentID, rfid, dateStr, prayerType).Scan(&count)
	if count > 0 {
		return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen sholat " + prayerType + " hari ini", "name": name})
	}

	var sID interface{}
	if studentID > 0 {
		sID = studentID
	}
	_, err = a.DB.Exec("INSERT INTO prayer_logs (student_id, rfid_uid, name, class_name, prayer_type, timestamp, date) VALUES (?, ?, ?, '', ?, ?, ?)",
		sID, rfid, name, prayerType, timestamp, dateStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	if studentID > 0 {
		a.DB.Exec("INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, 10, ?, 'System-Auto')",
			studentID, "Sholat Berjamaah "+prayerType+" (+10)")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "success",
		"message":     "Presensi Sholat " + prayerType + " Berhasil",
		"name":        name,
		"prayer_type": prayerType,
		"time":        timeStr,
	})
}

func (a *TestApp) BulkAttendanceHandler(c echo.Context) error {
	var req BulkAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}
	if req.Date == "" || len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Date and Students are required"})
	}

	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Transaction failed"})
	}

	for _, item := range req.Items {
		var rfid, name string
		err := tx.QueryRow("SELECT rfid_uid, name FROM students WHERE id=?", item.StudentID).Scan(&rfid, &name)
		if err != nil {
			continue
		}

		var logID int
		err = tx.QueryRow("SELECT id FROM attendance_logs WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date=?", item.StudentID, rfid, req.Date).Scan(&logID)
		timestamp := fmt.Sprintf("%s 07:00:00", req.Date)

		if item.Status == "Unknown" {
			if logID != 0 {
				tx.Exec("DELETE FROM attendance_logs WHERE id=?", logID)
			}
			continue
		}

		if err == sql.ErrNoRows {
			_, err = tx.Exec(`
				INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date)
				VALUES (?, ?, ?, 'Siswa', ?, 'Manual', ?, ?)
			`, item.StudentID, rfid, name, item.Status, timestamp, req.Date)

			var pointsChange int
			var pointDesc string
			switch item.Status {
			case "Hadir":
				pointsChange = 20
				pointDesc = "Hadir (Bulk) (+20)"
			case "Terlambat":
				pointsChange = -10
				pointDesc = "Terlambat (Bulk) (-10)"
			case "Alpha":
				pointsChange = -20
				pointDesc = "Alpha (Bulk) (-20)"
			}
			if pointsChange != 0 && item.StudentID > 0 {
				tx.Exec("INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, ?, ?, ?)",
					item.StudentID, pointsChange, pointDesc, "System-Bulk")
			}
		} else {
			tx.Exec(`
				UPDATE attendance_logs 
				SET student_id=?, status=?, method='Manual' 
				WHERE id=?
			`, item.StudentID, item.Status, logID)
		}
	}

	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Commit failed: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Presensi berhasil disimpan"})
}

func (a *TestApp) markAlphaStudents() {
	now := time.Now()
	isWorking, _ := handler.IsWorkingDay(a.DB, now)
	if !isWorking {
		return
	}
	today := now.Format("2006-01-02")
	rows, err := a.DB.Query(`
		SELECT s.id, s.name, s.rfid_uid, c.name, s.parent_phone
		FROM students s
		LEFT JOIN classes c ON s.class_id = c.id
		WHERE s.status = 'active'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	type studentInfo struct {
		ID          int
		Name        string
		RFID        string
		ClassName   string
		ParentPhone string
	}
	var allStudents []studentInfo

	for rows.Next() {
		var s studentInfo
		var rfidNull, classNull, phoneNull sql.NullString
		if err := rows.Scan(&s.ID, &s.Name, &rfidNull, &classNull, &phoneNull); err != nil {
			continue
		}
		s.RFID = rfidNull.String
		s.ClassName = classNull.String
		s.ParentPhone = phoneNull.String
		allStudents = append(allStudents, s)
	}
	rows.Close()

	var absentStudents []studentInfo
	for _, s := range allStudents {
		var count int
		a.DB.QueryRow(`
			SELECT COUNT(*) FROM attendance_logs
			WHERE ((student_id > 0 AND student_id = ?) OR (rfid_uid = ? AND rfid_uid != '')) AND date = ? AND user_type = 'Siswa'
		`, s.ID, s.RFID, today).Scan(&count)
		if count == 0 {
			absentStudents = append(absentStudents, s)
		}
	}

	timestamp := today + " 12:00:00"
	for _, s := range absentStudents {
		var sID interface{}
		if s.ID > 0 {
			sID = s.ID
		}
		a.DB.Exec(`
			INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date)
			VALUES (?, ?, ?, 'Siswa', 'Alpha', 'System', ?, ?)
		`, sID, s.RFID, s.Name, timestamp, today)
	}
}

// --- Dummy AppHandlers Interface Implementation ---

func (a *TestApp) ManualAttendanceHandler(c echo.Context) error {
	studentID := c.FormValue("student_id")
	status := c.FormValue("status")
	timeStr := c.FormValue("time")
	note := c.FormValue("note")
	letterStatus := c.FormValue("letter_status")

	var rfid, name string
	err := a.DB.QueryRow("SELECT rfid_uid, name FROM students WHERE id=?", studentID).Scan(&rfid, &name)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}

	now := time.Now()
	var timestamp string
	dateStr := now.Format("2006-01-02")

	if (status == "Hadir" || status == "Terlambat") && timeStr != "" {
		if _, err := time.Parse("15:04", timeStr); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format jam tidak valid"})
		}
		timestamp = fmt.Sprintf("%s %s:00", dateStr, timeStr)
	} else {
		timestamp = now.Format("2006-01-02 15:04:05")
	}

	finalStatus := status
	if status == "Sakit" {
		if letterStatus == "with" {
			finalStatus = "Sakit (Dengan Surat)"
		} else if letterStatus == "without" {
			finalStatus = "Sakit (Tanpa Surat)"
		}
	}

	_, err = a.DB.Exec(`INSERT INTO attendance_logs 
		(student_id, rfid_uid, user_name, user_type, status, method, timestamp, date, note) 
		VALUES (?, ?, ?, 'Siswa', ?, 'MANUAL', ?, ?, ?)`,
		studentID, rfid, name, finalStatus, timestamp, dateStr, note)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	var pointsChange int
	var pointDesc string
	switch status {
	case "Hadir":
		pointsChange = 20
		pointDesc = "Hadir (Manual) (+20)"
	case "Terlambat":
		pointsChange = -10
		pointDesc = "Terlambat (Manual) (-10)"
	case "Alpha":
		pointsChange = -20
		pointDesc = "Alpha (Manual) (-20)"
	}

	sidInt, _ := strconv.Atoi(studentID)
	if pointsChange != 0 && sidInt > 0 {
		a.DB.Exec("INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, ?, ?, ?)",
			sidInt, pointsChange, pointDesc, "Guru/Admin")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Presensi manual berhasil disimpan"})
}

func (a *TestApp) VerifyFaceAttendanceHandler(c echo.Context) error {
	type VerifyRequest struct {
		StudentID   interface{} `json:"student_id"`
		StaffID     interface{} `json:"staff_id"`
		UserType    string      `json:"user_type"`
		Distance    float64     `json:"distance"`
		ImageBase64 string      `json:"image_base64"`
	}

	var req VerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	var studentID int
	switch v := req.StudentID.(type) {
	case float64:
		studentID = int(v)
	case int:
		studentID = v
	case string:
		studentID, _ = strconv.Atoi(v)
	}

	var staffID int
	switch v := req.StaffID.(type) {
	case float64:
		staffID = int(v)
	case int:
		staffID = v
	case string:
		staffID, _ = strconv.Atoi(v)
	}

	userType := req.UserType

	// 1. Staff Recognition Branch
	if userType == "Staff" || staffID > 0 {
		var name, nip, phone, role, gender string
		var rfidNull sql.NullString
		err := a.DB.QueryRow(`
			SELECT name, nip, COALESCE(phone, ''), role, rfid_uid, COALESCE(gender, 'L')
			FROM staff
			WHERE id = ?`, staffID).Scan(&name, &nip, &phone, &role, &rfidNull, &gender)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Guru/Staff tidak ditemukan"})
		}
		rfid := rfidNull.String
		honorific := handler.DetermineStaffHonorific(gender, name)

		now := time.Now()
		timeStr := now.Format("15:04")
		dateStr := now.Format("2006-01-02")
		timestamp := now.Format("2006-01-02 15:04:05")

		var checkInCount, checkOutCount int
		var lastTimestamp string
		rowsStaffLogs, err := a.DB.Query(`
			SELECT status, timestamp 
			FROM attendance_logs
			WHERE user_type = 'Staff' AND date = ?
			  AND (student_id = ? OR (rfid_uid = ? AND rfid_uid != '') OR user_name = ?)
			ORDER BY timestamp ASC`,
			dateStr, staffID, rfid, name)
		if err == nil {
			for rowsStaffLogs.Next() {
				var st, ts string
				rowsStaffLogs.Scan(&st, &ts)
				lastTimestamp = ts
				if st == "Pulang" {
					checkOutCount++
				} else {
					checkInCount++
				}
			}
			rowsStaffLogs.Close()
		}

		var status, speechMessage string
		if checkOutCount > 0 {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":            "duplicate",
				"attendance_status": "Pulang",
				"user_type":         "Staff",
				"staff_id":          staffID,
				"name":              name,
				"honorific":         honorific,
				"message":           fmt.Sprintf("%s %s sudah presensi pulang hari ini", honorific, name),
				"speech_message":    fmt.Sprintf("%s %s sudah presensi pulang hari ini", honorific, name),
				"time":              timeStr,
			})
		} else if checkInCount > 0 {
			cleaned := strings.TrimSuffix(strings.Replace(lastTimestamp, "T", " ", 1), "Z")
			var lastTime time.Time
			var errParse error
			for _, f := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02T15:04:05", time.RFC3339} {
				lastTime, errParse = time.ParseInLocation(f, cleaned, time.Local)
				if errParse == nil {
					break
				}
			}
			if errParse == nil {
				diff := now.Sub(lastTime)
				if diff >= 0 && diff < 2*time.Minute {
					return c.JSON(http.StatusOK, map[string]interface{}{
						"status":            "duplicate",
						"attendance_status": "Hadir",
						"user_type":         "Staff",
						"staff_id":          staffID,
						"name":              name,
						"honorific":         honorific,
						"message":           fmt.Sprintf("%s %s baru saja presensi Hadir", honorific, name),
						"speech_message":    fmt.Sprintf("%s %s sudah presensi Hadir tadi", honorific, name),
						"time":              timeStr,
					})
				}
			}
			status = "Pulang"
			speechMessage = fmt.Sprintf("Terima kasih %s %s", honorific, name)
		} else {
			status = "Hadir"
			speechMessage = fmt.Sprintf("Selamat Datang %s %s", honorific, name)
		}

		_, err = a.DB.Exec("INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, ?, 'Staff', ?, 'FACE', ?, ?)",
			staffID, rfid, name, status, timestamp, dateStr)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":            "matched",
			"attendance_status": status,
			"user_type":         "Staff",
			"staff_id":          staffID,
			"name":              name,
			"nip":               nip,
			"role":              role,
			"gender":            gender,
			"honorific":         honorific,
			"speech_message":    speechMessage,
			"message":           "Presensi " + status + " Berhasil",
			"time":              timeStr,
		})
	}

	// 2. Student Recognition Branch
	var name, className, rfid string
	if studentID > 0 {
		var photoNull, rfidNull, classNull sql.NullString
		err := a.DB.QueryRow(`
			SELECT s.name, COALESCE(c.name, ''), s.rfid_uid, s.photo
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, studentID).Scan(&name, &classNull, &rfidNull, &photoNull)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Student not found"})
		}
		className = classNull.String
		rfid = rfidNull.String
	} else {
		// Call legacy Face Service via URL if only image_base64 provided
		faceURL := "http://localhost:8001/verify"
		var customFaceURL string
		a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='face_service_url'").Scan(&customFaceURL)
		if customFaceURL != "" {
			faceURL = customFaceURL
		}
		payload, _ := json.Marshal(map[string]string{"image_base64": req.ImageBase64})

		resp, err := http.Post(faceURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Face service returned %d", resp.StatusCode)})
		}

		var faceResult struct {
			Matched   bool    `json:"matched"`
			StudentID int     `json:"student_id"`
			Name      string  `json:"name"`
			ClassName string  `json:"class_name"`
			Distance  float64 `json:"distance"`
			Message   string  `json:"message"`
		}

		if err := json.Unmarshal(body, &faceResult); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid response from face service"})
		}

		if !faceResult.Matched {
			return c.JSON(http.StatusOK, map[string]string{
				"status":  "not_matched",
				"message": faceResult.Message,
			})
		}

		studentID = faceResult.StudentID
		name = faceResult.Name
		className = faceResult.ClassName
		var photoNull sql.NullString
		_ = a.DB.QueryRow("SELECT rfid_uid, photo FROM students WHERE id=?", studentID).Scan(&rfid, &photoNull)
	}

	now := time.Now()
	timeStr := now.Format("15:04")
	dateStr := now.Format("2006-01-02")
	timestamp := now.Format("2006-01-02 15:04:05")

	var arrEnd string
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='arrival_end'").Scan(&arrEnd)
	if arrEnd == "" {
		arrEnd = "07:15"
	}

	status := "Datang"
	if timeStr > arrEnd {
		status = "Terlambat"
	}

	var pointsChange int
	var pointDescription string
	switch status {
	case "Datang":
		pointsChange = 20
		pointDescription = "Hadir Wajah Tepat Waktu (+20)"
	case "Terlambat":
		pointsChange = -10
		pointDescription = "Terlambat Wajah (-10)"
	}

	_, err := a.DB.Exec("INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, ?, 'Siswa', ?, 'FACE', ?, ?)",
		studentID, rfid, name, status, timestamp, dateStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	if pointsChange != 0 && studentID > 0 {
		a.DB.Exec("INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, ?, ?, ?)",
			studentID, pointsChange, pointDescription, "Terminal-Face")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":            "matched",
		"attendance_status": status,
		"user_type":         "Siswa",
		"student_id":        studentID,
		"name":              name,
		"class_name":        className,
		"points":            pointsChange,
	})
}

// --- Dummy AppHandlers Interface Implementation ---

func (a *TestApp) DashboardHandler(c echo.Context) error         { return nil }
func (a *TestApp) SchedulesPageHandler(c echo.Context) error     { return nil }
func (a *TestApp) AudioPageHandler(c echo.Context) error         { return nil }
func (a *TestApp) AttendancePageHandler(c echo.Context) error    { return nil }
func (a *TestApp) AcademicsPageHandler(c echo.Context) error     { return nil }
func (a *TestApp) PointsPageHandler(c echo.Context) error        { return nil }
func (a *TestApp) SettingsPageHandler(c echo.Context) error      { return nil }
func (a *TestApp) ReportsPageHandler(c echo.Context) error       { return nil }
func (a *TestApp) TodayStatsHandler(c echo.Context) error        { return nil }
func (a *TestApp) RecentLogsHandler(c echo.Context) error        { return nil }
func (a *TestApp) SyncHandler(c echo.Context) error              { return nil }
func (a *TestApp) PublicLeaderboardHandler(c echo.Context) error { return nil }
func (a *TestApp) VersionHandler(c echo.Context) error           { return nil }

var _ router.AppHandlers = (*TestApp)(nil)
