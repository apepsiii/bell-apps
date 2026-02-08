package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// GetPrayerAttendanceHandler fetches students and their prayer status for a specific class, date, and prayer type.
func (a *App) GetPrayerAttendanceHandler(c echo.Context) error {
	classID := c.QueryParam("class_id")
	date := c.QueryParam("date")       // YYYY-MM-DD
	prayerType := c.QueryParam("type") // "Dzuhur" or "Ashar"

	if classID == "" || date == "" || prayerType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
	}

	// Fetch all students in the class
	rows, err := a.DB.Query("SELECT id, name, rfid_uid FROM students WHERE class_id = ? ORDER BY name", classID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	type StudentPrayerStatus struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		RFID      string `json:"rfid"`
		IsPresent bool   `json:"is_present"`
		Status    string `json:"status"` // "Hadir", "PMS"
	}

	var students []StudentPrayerStatus

	for rows.Next() {
		var s StudentPrayerStatus
		if err := rows.Scan(&s.ID, &s.Name, &s.RFID); err != nil {
			continue
		}

		// Check if this student has a prayer log for this date and type
		var status sql.NullString
		a.DB.QueryRow(`
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

// BulkPrayerAttendanceHandler handles manual submission of prayer attendance.
func (a *App) BulkPrayerAttendanceHandler(c echo.Context) error {
	type StudentInput struct {
		ID     int    `json:"id"`
		Status string `json:"status"` // "Hadir", "PMS"
	}

	type Request struct {
		Date       string         `json:"date"`        // YYYY-MM-DD
		PrayerType string         `json:"prayer_type"` // "Dzuhur" or "Ashar"
		Students   []StudentInput `json:"students"`    // List of students with status
		ClassID    string         `json:"class_id"`    // Context for fetching RFID
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	// Fetch all students in the class to map ID -> RFID
	studentMap := make(map[int]struct {
		RFID string
		Name string
	})

	rows, err := a.DB.Query(`
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

	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	// Timestamp
	now := time.Now()
	timestampStr := req.Date + " " + now.Format("15:04:05")

	// Process each student in the class
	for id, student := range studentMap {
		var inputStatus string
		present := false

		// Find in request
		for _, s := range req.Students {
			if s.ID == id {
				inputStatus = s.Status
				present = true
				break
			}
		}

		if present && (inputStatus == "Hadir" || inputStatus == "PMS") {
			// Insert/Update
			var exists int
			tx.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE rfid_uid=? AND date=? AND prayer_type=?", 
				student.RFID, req.Date, req.PrayerType).Scan(&exists)
			
			if exists == 0 {
				_, err = tx.Exec(`
					INSERT INTO prayer_logs (rfid_uid, name, class_name, prayer_type, timestamp, date, status) 
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					student.RFID, student.Name, className, req.PrayerType, timestampStr, req.Date, inputStatus)
			} else {
				// Update status if exists
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
			// Delete if exists (Unchecked or Invalid Status)
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

// PrayerReportHandler fetches statistics for reports.
func (a *App) PrayerReportHandler(c echo.Context) error {
	classID := c.QueryParam("class_id")
	month := c.QueryParam("month") // YYYY-MM
	
	if classID == "" || month == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
	}

	type DailyLog map[string][]string // Date -> [Dzuhur, Ashar, PMS...]

	type StudentPrayerReport struct {
		Name        string   `json:"name"`
		Attendance  DailyLog `json:"attendance"`
		DzuhurTotal int      `json:"dzuhur"`
		AsharTotal  int      `json:"ashar"`
		PMSTotal    int      `json:"pms"` // Added PMS Total
	}

	type Response struct {
		Students []StudentPrayerReport `json:"students"`
	}

	// 1. Get all students
	studentMap := make(map[string]*StudentPrayerReport)
	var studentsOrder []string

	rows, err := a.DB.Query("SELECT id, name, rfid_uid FROM students WHERE class_id = ? ORDER BY name", classID)
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

	// 2. Fetch logs for the month
	query := `
		SELECT rfid_uid, prayer_type, status, strftime('%Y-%m-%d', date) 
		FROM prayer_logs 
		WHERE strftime('%Y-%m', date) = ? 
		  AND rfid_uid IN (SELECT rfid_uid FROM students WHERE class_id = ?)
	`
	logRows, err := a.DB.Query(query, month, classID)
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
		if status == "" { status = "Hadir" }

		if entry, exists := studentMap[rfid]; exists {
			// Append to daily log
			// If status is PMS, we mark it as such. 
			// If Hadir, we mark the prayer type.
			marker := pType
			if status == "PMS" {
				marker = "PMS" 
			}
			
			entry.Attendance[date] = append(entry.Attendance[date], marker)
			
			// Increment totals
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

	// 3. Convert map to slice
	finalResult := Response{Students: []StudentPrayerReport{}}
	for _, rfid := range studentsOrder {
		if val, ok := studentMap[rfid]; ok {
			finalResult.Students = append(finalResult.Students, *val)
		}
	}

	return c.JSON(http.StatusOK, finalResult)
}
