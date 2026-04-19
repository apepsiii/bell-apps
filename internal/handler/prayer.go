package handler

import (
	"database/sql"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

func GenerateQR(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.QueryParam("rfid")
		if rfid == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID tidak boleh kosong"})
		}

		qrData := "RFID:" + rfid
		png, err := qrcode.Encode(qrData, qrcode.Medium, 256)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal generate QR code"})
		}

		base64Img := base64.StdEncoding.EncodeToString(png)
		return c.JSON(http.StatusOK, map[string]string{
			"qr_code": "data:image/png;base64," + base64Img,
		})
	}
}

func ScanQR(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type ScanRequest struct {
			QRData     string `json:"qr_data"`
			PrayerType string `json:"prayer_type"`
		}

		var req ScanRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		if len(req.QRData) < 6 || req.QRData[:5] != "RFID:" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format QR code tidak valid"})
		}
		rfid := req.QRData[5:]

		if req.PrayerType != "Dzuhur" && req.PrayerType != "Ashar" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tipe sholat tidak valid"})
		}

		var studentID int
		var studentName, className string
		var photo sql.NullString
		err := db.QueryRow(`
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

		today := time.Now().Format("2006-01-02")
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE rfid_uid = ? AND date = ? AND prayer_type = ?",
			rfid, today, req.PrayerType).Scan(&exists)

		if exists > 0 {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "duplicate",
				"message": "Sudah melakukan presensi " + req.PrayerType + " hari ini",
				"name":    studentName,
			})
		}

		now := time.Now()
		timestamp := now.Format("2006-01-02 15:04:05")

		_, err = db.Exec(`
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
}

func GetClasses(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
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
			var cl ClassInfo
			var majorName sql.NullString
			rows.Scan(&cl.ID, &cl.Name, &majorName)
			cl.MajorName = majorName.String
			classes = append(classes, cl)
		}

		return c.JSON(http.StatusOK, classes)
	}
}

func GetRecentPrayerLogs(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := 10

		rows, err := db.Query(`
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

			if len(timestamp) >= 16 {
				log.Time = timestamp[11:16]
			}
			log.Photo = photo.String
			logs = append(logs, log)
		}

		return c.JSON(http.StatusOK, logs)
	}
}

func GetPrayerAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		classID := c.QueryParam("class_id")
		date := c.QueryParam("date")
		prayerType := c.QueryParam("type")

		if classID == "" || date == "" || prayerType == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
		}

		rows, err := db.Query("SELECT id, name, rfid_uid FROM students WHERE class_id = ? ORDER BY name", classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		type StudentPrayerStatus struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			RFID      string `json:"rfid"`
			IsPresent bool   `json:"is_present"`
			Status    string `json:"status"`
		}

		var students []StudentPrayerStatus

		for rows.Next() {
			var s StudentPrayerStatus
			if err := rows.Scan(&s.ID, &s.Name, &s.RFID); err != nil {
				continue
			}

			var status sql.NullString
			db.QueryRow(`
				SELECT status FROM prayer_logs 
				WHERE rfid_uid = ? AND date = ? AND prayer_type = ?`,
				s.RFID, date, prayerType).Scan(&status)

			if status.Valid {
				s.IsPresent = true
				s.Status = status.String
			} else {
				s.IsPresent = false
				s.Status = ""
			}
			students = append(students, s)
		}

		return c.JSON(http.StatusOK, students)
	}
}

func BulkPrayerAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type StudentInput struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		}

		type Request struct {
			Date       string         `json:"date"`
			PrayerType string         `json:"prayer_type"`
			Students   []StudentInput `json:"students"`
			ClassID    string         `json:"class_id"`
		}

		var req Request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}

		studentMap := make(map[int]struct {
			RFID string
			Name string
		})

		rows, err := db.Query(`
			SELECT s.id, s.rfid_uid, s.name, c.name 
			FROM students s 
			JOIN classes c ON s.class_id = c.id 
			WHERE s.class_id = ?`, req.ClassID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		className := ""
		for rows.Next() {
			var id int
			var rfid, name, cName string
			rows.Scan(&id, &rfid, &name, &cName)
			studentMap[id] = struct {
				RFID string
				Name string
			}{rfid, name}
			className = cName
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		now := time.Now()
		timestampStr := req.Date + " " + now.Format("15:04:05")

		for id, student := range studentMap {
			var inputStatus string
			present := false

			for _, s := range req.Students {
				if s.ID == id {
					inputStatus = s.Status
					present = true
					break
				}
			}

			if present && (inputStatus == "Hadir" || inputStatus == "PMS") {
				var exists int
				tx.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE rfid_uid=? AND date=? AND prayer_type=?",
					student.RFID, req.Date, req.PrayerType).Scan(&exists)

				if exists == 0 {
					_, err = tx.Exec(`
						INSERT INTO prayer_logs (rfid_uid, name, class_name, prayer_type, timestamp, date, status) 
						VALUES (?, ?, ?, ?, ?, ?, ?)`,
						student.RFID, student.Name, className, req.PrayerType, timestampStr, req.Date, inputStatus)
				} else {
					_, err = tx.Exec(`
						UPDATE prayer_logs SET status = ? 
						WHERE rfid_uid=? AND date=? AND prayer_type=?`,
						inputStatus, student.RFID, req.Date, req.PrayerType)
				}
				if err != nil {
					tx.Rollback()
					return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data: " + err.Error()})
				}
			} else {
				_, err = tx.Exec("DELETE FROM prayer_logs WHERE rfid_uid=? AND date=? AND prayer_type=?",
					student.RFID, req.Date, req.PrayerType)
				if err != nil {
					tx.Rollback()
					return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus data: " + err.Error()})
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal commit database"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Data presensi sholat berhasil disimpan"})
	}
}

func PrayerReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		classID := c.QueryParam("class_id")
		month := c.QueryParam("month")

		if classID == "" || month == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
		}

		type DailyLog map[string][]string

		type StudentPrayerReport struct {
			Name        string   `json:"name"`
			Attendance  DailyLog `json:"attendance"`
			DzuhurTotal int      `json:"dzuhur"`
			AsharTotal  int      `json:"ashar"`
			PMSTotal    int      `json:"pms"`
		}

		type Response struct {
			Students []StudentPrayerReport `json:"students"`
		}

		studentMap := make(map[string]*StudentPrayerReport)
		var studentsOrder []string

		rows, err := db.Query("SELECT id, name, rfid_uid FROM students WHERE class_id = ? ORDER BY name", classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var name, rfid string
			if err := rows.Scan(&id, &name, &rfid); err != nil {
				continue
			}
			studentMap[rfid] = &StudentPrayerReport{
				Name:       name,
				Attendance: make(DailyLog),
			}
			studentsOrder = append(studentsOrder, rfid)
		}

		query := `
			SELECT rfid_uid, prayer_type, status, strftime('%Y-%m-%d', date) 
			FROM prayer_logs 
			WHERE strftime('%Y-%m', date) = ? 
			  AND rfid_uid IN (SELECT rfid_uid FROM students WHERE class_id = ?)
		`
		logRows, err := db.Query(query, month, classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer logRows.Close()

		for logRows.Next() {
			var rfid, pType, status, date string
			var statusNull sql.NullString
			if err := logRows.Scan(&rfid, &pType, &statusNull, &date); err != nil {
				continue
			}
			status = statusNull.String
			if status == "" {
				status = "Hadir"
			}

			if entry, exists := studentMap[rfid]; exists {
				marker := pType
				if status == "PMS" {
					marker = "PMS"
				}

				entry.Attendance[date] = append(entry.Attendance[date], marker)

				if status == "PMS" {
					entry.PMSTotal++
				} else {
					if pType == "Dzuhur" {
						entry.DzuhurTotal++
					} else if pType == "Ashar" {
						entry.AsharTotal++
					}
				}
			}
		}

		finalResult := Response{Students: []StudentPrayerReport{}}
		for _, rfid := range studentsOrder {
			if val, ok := studentMap[rfid]; ok {
				finalResult.Students = append(finalResult.Students, *val)
			}
		}

		return c.JSON(http.StatusOK, finalResult)
	}
}
