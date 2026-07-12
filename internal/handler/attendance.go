package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func UpdateAttendanceSettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		arrivalStart := c.FormValue("arrival_start")
		arrivalEnd := c.FormValue("arrival_end")
		departureStart := c.FormValue("departure_start")
		departureEnd := c.FormValue("departure_end")

		waURL := c.FormValue("onesender_api_url")
		waToken := c.FormValue("onesender_api_token")
		waTemplateIn := c.FormValue("wa_template_in")
		waTemplateLate := c.FormValue("wa_template_late")
		waTemplateOut := c.FormValue("wa_template_out")
		waTemplateStaffIn := c.FormValue("wa_template_staff_in")
		waTemplateStaffOut := c.FormValue("wa_template_staff_out")
		waImage := c.FormValue("wa_image_link")

		dzStart := c.FormValue("dzuhur_start")
		dzEnd := c.FormValue("dzuhur_end")
		asStart := c.FormValue("ashar_start")
		asEnd := c.FormValue("ashar_end")

		birthdayEnabled := c.FormValue("birthday_enabled")
		birthdayTemplate := c.FormValue("wa_template_birthday")
		birthdayImage := c.FormValue("wa_image_birthday")
		birthdayTime := c.FormValue("birthday_time")

		tx, _ := db.Begin()
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_start", arrivalStart)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_end", arrivalEnd)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_start", departureStart)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_end", departureEnd)

		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_url", waURL)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_token", waToken)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_in", waTemplateIn)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_late", waTemplateLate)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_out", waTemplateOut)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_in", waTemplateStaffIn)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_out", waTemplateStaffOut)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_image_link", waImage)

		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_start", dzStart)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_end", dzEnd)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_start", asStart)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_end", asEnd)

		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "birthday_enabled", birthdayEnabled)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_birthday", birthdayTemplate)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_image_birthday", birthdayImage)
		tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "birthday_time", birthdayTime)

		tx.Commit()

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Pengaturan disimpan"})
	}
}

func ExportAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		exportType := c.QueryParam("type")
		dateVal := c.QueryParam("date")
		monthVal := c.QueryParam("month")
		yearVal := c.QueryParam("year")
		startDate := c.QueryParam("start_date")
		endDate := c.QueryParam("end_date")

		var query string
		var args []interface{}
		filename := "attendance_export.csv"

		baseQuery := `
			SELECT a.date, a.timestamp, s.nis, s.name, c.name, a.status, a.method
			FROM attendance_logs a
			LEFT JOIN students s ON a.rfid_uid = s.rfid_uid
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE a.user_type = 'Siswa'`

		switch exportType {
		case "daily":
			query = baseQuery + " AND a.date = ? ORDER BY a.timestamp ASC"
			args = append(args, dateVal)
			filename = fmt.Sprintf("presensi_harian_%s.csv", dateVal)
		case "monthly":
			query = baseQuery + " AND strftime('%m', a.date) = ? AND strftime('%Y', a.date) = ? ORDER BY a.date ASC, a.timestamp ASC"
			args = append(args, monthVal, yearVal)
			filename = fmt.Sprintf("presensi_bulanan_%s_%s.csv", monthVal, yearVal)
		case "custom":
			query = baseQuery + " AND a.date BETWEEN ? AND ? ORDER BY a.date ASC, a.timestamp ASC"
			args = append(args, startDate, endDate)
			filename = fmt.Sprintf("presensi_custom_%s_sd_%s.csv", startDate, endDate)
		default:
			return c.String(http.StatusBadRequest, "Tipe export tidak valid")
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer rows.Close()

		c.Response().Header().Set("Content-Type", "text/csv")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		writer := csv.NewWriter(c.Response().Writer)
		defer writer.Flush()

		writer.Write([]string{"Tanggal", "Waktu", "NIS", "Nama", "Kelas", "Status", "Metode"})

		for rows.Next() {
			var date, timestamp, nis, name, className, status, method sql.NullString
			rows.Scan(&date, &timestamp, &nis, &name, &className, &status, &method)

			ts := timestamp.String
			if len(ts) >= 16 {
				ts = ts[11:16]
			}

			writer.Write([]string{
				date.String,
				ts,
				nis.String,
				name.String,
				className.String,
				status.String,
				method.String,
			})
		}

		return nil
	}
}

type CalendarLog struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func GetStudentCalendar(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.QueryParam("id")
		month := c.QueryParam("month")
		year := c.QueryParam("year")

		if studentID == "" || month == "" || year == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
		}

		var rfid string
		err := db.QueryRow("SELECT rfid_uid FROM students WHERE id=?", studentID).Scan(&rfid)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		query := `
			SELECT status, timestamp, date
			FROM attendance_logs
			WHERE rfid_uid = ?
			  AND strftime('%m', date) = ?
			  AND strftime('%Y', date) = ?
			ORDER BY timestamp ASC`

		rows, err := db.Query(query, rfid, month, year)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		calendarData := make(map[int][]CalendarLog)

		for rows.Next() {
			var status, ts, dateStr string
			rows.Scan(&status, &ts, &dateStr)

			t, _ := time.Parse("2006-01-02", dateStr)
			timeOnly := ""
			if len(ts) >= 16 {
				timeOnly = ts[11:16]
			}

			calendarData[t.Day()] = append(calendarData[t.Day()], CalendarLog{Status: status, Time: timeOnly})
		}

		totalPresent, totalLate, totalSick, totalPermission, totalAlpha := 0, 0, 0, 0, 0

		monthInt, _ := strconv.Atoi(month)
		monthType := time.Month(monthInt)
		yearInt, _ := strconv.Atoi(year)

		workingDays, _ := getWorkingDaysInMonth(db, yearInt, monthType)
		holidaysMap, _ := getHolidaysForMonth(db, yearInt, monthType)

		for _, logs := range calendarData {
			for _, log := range logs {
				switch log.Status {
				case "Datang", "Hadir":
					totalPresent++
				case "Terlambat":
					totalLate++
				case "Sakit":
					totalSick++
				case "Izin":
					totalPermission++
				case "Alpha":
					totalAlpha++
				}
			}
		}

		attendanceRate := 0.0
		if workingDays > 0 {
			attendanceRate = float64(totalPresent+totalLate) / float64(workingDays) * 100
		}

		response := map[string]interface{}{
			"calendar": calendarData,
			"stats": map[string]interface{}{
				"working_days":    workingDays,
				"present":         totalPresent,
				"late":            totalLate,
				"sick":            totalSick,
				"permission":      totalPermission,
				"alpha":           totalAlpha,
				"attendance_rate": attendanceRate,
			},
			"holidays": holidaysMap,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetStaffCalendar(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		staffID := c.QueryParam("id")
		month := c.QueryParam("month")
		year := c.QueryParam("year")

		if staffID == "" || month == "" || year == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
		}

		var rfid string
		err := db.QueryRow("SELECT rfid_uid FROM staff WHERE id=?", staffID).Scan(&rfid)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Staff tidak ditemukan"})
		}

		query := `
			SELECT status, timestamp, date
			FROM attendance_logs
			WHERE rfid_uid = ?
			  AND strftime('%m', date) = ?
			  AND strftime('%Y', date) = ?
			ORDER BY timestamp ASC`

		rows, err := db.Query(query, rfid, month, year)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		calendarData := make(map[int][]CalendarLog)

		for rows.Next() {
			var status, ts, dateStr string
			rows.Scan(&status, &ts, &dateStr)
			t, _ := time.Parse("2006-01-02", dateStr)
			timeOnly := ""
			if len(ts) >= 16 {
				timeOnly = ts[11:16]
			}
			calendarData[t.Day()] = append(calendarData[t.Day()], CalendarLog{Status: status, Time: timeOnly})
		}

		totalPresent, totalLate, totalSick, totalPermission, totalAlpha := 0, 0, 0, 0, 0

		monthInt, _ := strconv.Atoi(month)
		monthType := time.Month(monthInt)
		yearInt, _ := strconv.Atoi(year)

		workingDays, _ := getWorkingDaysInMonth(db, yearInt, monthType)
		holidaysMap, _ := getHolidaysForMonth(db, yearInt, monthType)

		for _, logs := range calendarData {
			for _, log := range logs {
				switch log.Status {
				case "Datang":
					totalPresent++
				case "Terlambat":
					totalLate++
				case "Sakit":
					totalSick++
				case "Izin":
					totalPermission++
				case "Alpha":
					totalAlpha++
				}
			}
		}

		attendanceRate := 0.0
		if workingDays > 0 {
			attendanceRate = float64(totalPresent+totalLate) / float64(workingDays) * 100
		}

		response := map[string]interface{}{
			"calendar": calendarData,
			"stats": map[string]interface{}{
				"working_days":    workingDays,
				"present":         totalPresent,
				"late":            totalLate,
				"sick":            totalSick,
				"permission":      totalPermission,
				"alpha":           totalAlpha,
				"attendance_rate": attendanceRate,
			},
			"holidays": holidaysMap,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func getWorkingDaysInMonth(db *sql.DB, year int, month time.Month) (int, error) {
	holidays, err := getHolidaysForMonth(db, year, month)
	if err != nil {
		return 0, err
	}

	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	workingDays := 0
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		if date.Weekday() == time.Sunday {
			continue
		}
		if _, isHoliday := holidays[day]; isHoliday {
			continue
		}
		workingDays++
	}
	return workingDays, nil
}

func getHolidaysForMonth(db *sql.DB, year int, month time.Month) (map[int]string, error) {
	query := `SELECT date, description FROM holidays
	          WHERE strftime('%Y', date) = ? AND strftime('%m', date) = ?`

	rows, err := db.Query(query, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holidays := make(map[int]string)
	for rows.Next() {
		var dateStr string
		var description string
		if err := rows.Scan(&dateStr, &description); err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		holidays[t.Day()] = description
	}
	return holidays, nil
}

func GetDailyAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		dateStr := c.QueryParam("date")
		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}

		rows, err := db.Query(`
			SELECT al.id, al.rfid_uid, al.user_name, al.user_type, al.status, al.method, al.timestamp, al.date, s.photo
			FROM attendance_logs al
			LEFT JOIN students s ON al.rfid_uid = s.rfid_uid
			WHERE al.date = ?
			ORDER BY al.timestamp DESC`, dateStr)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		type AttendanceRow struct {
			ID        int    `json:"id"`
			RFID      string `json:"rfid_uid"`
			UserName  string `json:"user_name"`
			UserType  string `json:"user_type"`
			Status    string `json:"status"`
			Method    string `json:"method"`
			Timestamp string `json:"timestamp"`
			Date      string `json:"date"`
			Photo     string `json:"photo"`
		}

		var logs []AttendanceRow
		for rows.Next() {
			var l AttendanceRow
			var photo sql.NullString
			rows.Scan(&l.ID, &l.RFID, &l.UserName, &l.UserType, &l.Status, &l.Method, &l.Timestamp, &l.Date, &photo)
			l.Photo = photo.String
			logs = append(logs, l)
		}

		return c.JSON(http.StatusOK, logs)
	}
}

func BulkAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type Request struct {
			Attendance []struct {
				StudentID int    `json:"student_id"`
				Status    string `json:"status"`
			} `json:"attendance"`
		}

		var req Request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		dateStr := time.Now().Format("2006-01-02")
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		tx, _ := db.Begin()
		successCount := 0

		for _, att := range req.Attendance {
			var rfid, name string
			err := db.QueryRow("SELECT rfid_uid, name FROM students WHERE id=?", att.StudentID).Scan(&rfid, &name)
			if err != nil {
				continue
			}

			_, err = tx.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, 'Siswa', ?, 'MANUAL', ?, ?)",
				rfid, name, att.Status, timestamp, dateStr)
			if err == nil {
				successCount++
			}
		}

		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("%d attendance recorded", successCount)})
	}
}
