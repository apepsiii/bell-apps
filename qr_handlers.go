package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCodeHandler generates QR code from RFID UID
func (a *App) GenerateQRCodeHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID tidak boleh kosong"})
	}

	// Generate QR code
	qrData := "RFID:" + rfid
	png, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal generate QR code"})
	}

	// Return as base64 image
	base64Img := base64.StdEncoding.EncodeToString(png)
	return c.JSON(http.StatusOK, map[string]string{
		"qr_code": "data:image/png;base64," + base64Img,
	})
}

// ScanQRCodeHandler processes QR code scan for prayer attendance
func (a *App) ScanQRCodeHandler(c echo.Context) error {
	type ScanRequest struct {
		QRData     string `json:"qr_data"`
		PrayerType string `json:"prayer_type"` // Dzuhur or Ashar
	}

	var req ScanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	// Extract RFID from QR data (format: "RFID:1234567890")
	if len(req.QRData) < 6 || req.QRData[:5] != "RFID:" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format QR code tidak valid"})
	}
	rfid := req.QRData[5:]

	// Validate prayer type
	if req.PrayerType != "Dzuhur" && req.PrayerType != "Ashar" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tipe sholat tidak valid"})
	}

	// Find student
	var studentID int
	var studentName, className string
	var photo sql.NullString
	err := a.DB.QueryRow(`
		SELECT s.id, s.name, c.name, s.photo
		FROM students s
		LEFT JOIN classes c ON s.class_id = c.id
		WHERE s.rfid_uid = ?`, rfid).Scan(&studentID, &studentName, &className, &photo)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Kartu tidak terdaftar",
			"status":  "error",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Terjadi kesalahan sistem"})
	}

	// Check if already recorded today
	today := time.Now().Format("2006-01-02")
	var exists int
	a.DB.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE rfid_uid = ? AND date = ? AND prayer_type = ?",
		rfid, today, req.PrayerType).Scan(&exists)

	if exists > 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "duplicate",
			"message": "Sudah melakukan presensi " + req.PrayerType + " hari ini",
			"name":    studentName,
		})
	}

	// Record prayer attendance
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05")

	_, err = a.DB.Exec(`
		INSERT INTO prayer_logs (rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rfid, studentName, className, req.PrayerType, timestamp, today, "Hadir", "QR")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "success",
		"message":     "Presensi " + req.PrayerType + " berhasil",
		"name":        studentName,
		"class_name":  className,
		"prayer_type": req.PrayerType,
		"time":        now.Format("15:04"),
		"photo":       photo.String,
	})
}

// GetClassesHandler returns list of classes for manual attendance
func (a *App) GetClassesHandler(c echo.Context) error {
	rows, err := a.DB.Query(`
		SELECT c.id, c.name, m.name as major_name
		FROM classes c
		LEFT JOIN majors m ON c.major_id = m.id
		ORDER BY c.name ASC`)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	type ClassInfo struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		MajorName string `json:"major_name"`
	}

	var classes []ClassInfo
	for rows.Next() {
		var c ClassInfo
		var majorName sql.NullString
		rows.Scan(&c.ID, &c.Name, &majorName)
		c.MajorName = majorName.String
		classes = append(classes, c)
	}

	return c.JSON(http.StatusOK, classes)
}

// GetRecentPrayerLogsHandler returns recent prayer logs for operator dashboard
func (a *App) GetRecentPrayerLogsHandler(c echo.Context) error {
	limit := 10

	rows, err := a.DB.Query(`
		SELECT pl.name, pl.class_name, pl.prayer_type, pl.status, pl.timestamp, s.photo
		FROM prayer_logs pl
		LEFT JOIN students s ON pl.rfid_uid = s.rfid_uid
		WHERE pl.date = ?
		ORDER BY pl.timestamp DESC
		LIMIT ?`, time.Now().Format("2006-01-02"), limit)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	type LogEntry struct {
		Name       string `json:"name"`
		ClassName  string `json:"class_name"`
		PrayerType string `json:"prayer_type"`
		Status     string `json:"status"`
		Time       string `json:"time"`
		Photo      string `json:"photo"`
	}

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		var timestamp string
		var photo sql.NullString
		rows.Scan(&log.Name, &log.ClassName, &log.PrayerType, &log.Status, &timestamp, &photo)

		// Extract time from timestamp
		if len(timestamp) >= 16 {
			log.Time = timestamp[11:16]
		}
		log.Photo = photo.String
		logs = append(logs, log)
	}

	return c.JSON(http.StatusOK, logs)
}

// GetStudentsForPrayerHandler returns students for a class with their prayer attendance status
func (a *App) GetStudentsForPrayerHandler(c echo.Context) error {
	classID := c.QueryParam("class_id")
	date := c.QueryParam("date")
	prayerType := c.QueryParam("type")

	if classID == "" || date == "" || prayerType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "class_id, date, dan type harus diisi"})
	}

	type StudentAttendance struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		RFIDUID   string `json:"rfid_uid"`
		IsPresent bool   `json:"is_present"`
		Status    string `json:"status"`
	}

	rows, err := a.DB.Query(`
		SELECT s.id, s.name, s.rfid_uid,
		       CASE WHEN pl.id IS NOT NULL THEN 1 ELSE 0 END as is_present,
		       COALESCE(pl.status, 'Hadir') as status
		FROM students s
		LEFT JOIN prayer_logs pl ON s.rfid_uid = pl.rfid_uid 
		    AND pl.date = ? AND pl.prayer_type = ?
		WHERE s.class_id = ?
		ORDER BY s.name ASC`, date, prayerType, classID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	var students []StudentAttendance
	for rows.Next() {
		var s StudentAttendance
		var isPresent int
		rows.Scan(&s.ID, &s.Name, &s.RFIDUID, &isPresent, &s.Status)
		s.IsPresent = isPresent == 1
		students = append(students, s)
	}

	return c.JSON(http.StatusOK, students)
}

// SavePrayerAttendanceHandler saves prayer attendance for multiple students
func (a *App) SavePrayerAttendanceHandler(c echo.Context) error {
	type AttendanceRequest struct {
		ClassID    string `json:"class_id"`
		Date       string `json:"date"`
		PrayerType string `json:"prayer_type"`
		Students   []struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		} `json:"students"`
	}

	var req AttendanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	// Get operator name from session
	operatorID := c.Get("operator_id").(string)
	var operatorName string
	a.DB.QueryRow("SELECT name FROM operators WHERE id = ?", operatorID).Scan(&operatorName)

	// Begin transaction
	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memulai transaksi"})
	}
	defer tx.Rollback()

	// Delete existing records for this class/date/prayer
	_, err = tx.Exec("DELETE FROM prayer_logs WHERE date = ? AND prayer_type = ? AND rfid_uid IN (SELECT rfid_uid FROM students WHERE class_id = ?)",
		req.Date, req.PrayerType, req.ClassID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus data lama"})
	}

	// Insert new records
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, student := range req.Students {
		var rfidUID, name, className string
		err = tx.QueryRow("SELECT s.rfid_uid, s.name, c.name FROM students s LEFT JOIN classes c ON s.class_id = c.id WHERE s.id = ?", student.ID).
			Scan(&rfidUID, &name, &className)
		if err != nil {
			continue
		}

		_, err = tx.Exec(`INSERT INTO prayer_logs (rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			rfidUID, name, className, req.PrayerType, now, req.Date, student.Status, "MANUAL-"+operatorName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data: " + err.Error()})
		}
	}

	if err = tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Berhasil menyimpan %d presensi sholat %s", len(req.Students), req.PrayerType)})
}
