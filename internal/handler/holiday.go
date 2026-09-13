package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Holiday struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type SchoolSetting struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetHolidays(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		year := c.QueryParam("year")
		month := c.QueryParam("month")

		query := "SELECT id, date, name, type, description FROM holidays WHERE 1=1"
		var args []interface{}

		if year != "" {
			query += " AND strftime('%Y', date) = ?"
			args = append(args, year)
		}
		if month != "" {
			query += " AND strftime('%m', date) = ?"
			args = append(args, month)
		}

		query += " ORDER BY date ASC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var holidays []Holiday
		for rows.Next() {
			var h Holiday
			var desc sql.NullString
			rows.Scan(&h.ID, &h.Date, &h.Name, &h.Type, &desc)
			h.Description = desc.String
			holidays = append(holidays, h)
		}

		return c.JSON(http.StatusOK, holidays)
	}
}

func AddHoliday(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var h Holiday
		if err := c.Bind(&h); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
		}

		if h.Date == "" || h.Name == "" || h.Type == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Date, Name, and Type are required"})
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM holidays WHERE date=?", h.Date).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"message": "Holiday for this date already exists"})
		}

		_, err := db.Exec("INSERT INTO holidays (date, name, type, description) VALUES (?, ?, ?, ?)",
			h.Date, h.Name, h.Type, h.Description)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]string{"message": "Holiday added successfully"})
	}
}

func UpdateHoliday(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var h Holiday
		if err := c.Bind(&h); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
		}

		_, err := db.Exec("UPDATE holidays SET date=?, name=?, type=?, description=? WHERE id=?",
			h.Date, h.Name, h.Type, h.Description, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Holiday updated successfully"})
	}
}

func DeleteHoliday(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM holidays WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Holiday deleted successfully"})
	}
}

func GetSchoolSettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT id, setting_key, setting_value FROM school_settings")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		settings := make(map[string]string)
		for rows.Next() {
			var id int
			var k, v string
			rows.Scan(&id, &k, &v)
			settings[k] = v
		}

		return c.JSON(http.StatusOK, settings)
	}
}

func UpdateSchoolSettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body map[string]string
		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
		}

		for k, v := range body {
			_, err := db.Exec("INSERT INTO school_settings (setting_key, setting_value) VALUES (?, ?) ON CONFLICT(setting_key) DO UPDATE SET setting_value=excluded.setting_value", k, v)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated successfully"})
	}
}

func IsWorkingDay(db *sql.DB, date time.Time) (bool, string) {
	dateStr := date.Format("2006-01-02")
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	var holidayName string
	err := db.QueryRow("SELECT name FROM holidays WHERE date=?", dateStr).Scan(&holidayName)
	if err == nil {
		return false, "Libur: " + holidayName
	}

	var workDaysStr string
	err = db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key='work_days'").Scan(&workDaysStr)
	if err != nil {
		workDaysStr = "1,2,3,4,5"
	}

	checkDay := weekday
	if date.Weekday() == time.Sunday {
		checkDay = 7
	}

	days := strings.Split(workDaysStr, ",")
	isWork := false
	for _, d := range days {
		dayChar := ""
		switch checkDay {
		case 1:
			dayChar = "1"
		case 2:
			dayChar = "2"
		case 3:
			dayChar = "3"
		case 4:
			dayChar = "4"
		case 5:
			dayChar = "5"
		case 6:
			dayChar = "6"
		case 7:
			dayChar = "7"
		}

		if d == dayChar {
			isWork = true
			break
		}
	}
	if !isWork {
		return false, "Libur Akhir Pekan/Bukan Hari Kerja"
	}

	return true, ""
}

func GetWorkingDaysInMonth(db *sql.DB, year int, month time.Month) (int, error) {
	holidays, err := GetHolidaysForMonth(db, year, month)
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

func GetHolidaysForMonth(db *sql.DB, year int, month time.Month) (map[int]string, error) {
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

func ImportNationalHolidays(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		yearStr := c.QueryParam("year")
		if yearStr == "" {
			yearStr = time.Now().Format("2006")
		}

		// Data resmi Hari Libur Nasional & Cuti Bersama Tahun 2026
		// Sumber: Keputusan Bersama Menag, Menaker, dan MenPANRB Nomor 1497/2025, 2/2025, 5/2025
		// https://jdih.kemenkoinfra.go.id/pemerintah-tetapkan-hari-libur-nasional-dan-cuti-bersama-tahun-2026
		type holidayData struct {
			date string
			name string
			typ  string // "national" atau "cuti_bersama"
		}

		var holidays []holidayData

		if yearStr == "2026" {
			holidays = []holidayData{
				// === HARI LIBUR NASIONAL (16 hari) ===
				{"2026-01-01", "Tahun Baru 2026 Masehi", "national"},
				{"2026-01-16", "Isra Mikraj Nabi Muhammad S.A.W.", "national"},
				{"2026-02-17", "Tahun Baru Imlek 2577 Kongzili", "national"},
				{"2026-03-19", "Hari Suci Nyepi (Tahun Baru Saka 1948)", "national"},
				{"2026-03-21", "Idul Fitri 1447 Hijriah", "national"},
				{"2026-03-22", "Idul Fitri 1447 Hijriah", "national"},
				{"2026-04-03", "Wafat Yesus Kristus", "national"},
				{"2026-04-05", "Kebangkitan Yesus Kristus (Paskah)", "national"},
				{"2026-05-01", "Hari Buruh Internasional", "national"},
				{"2026-05-14", "Kenaikan Yesus Kristus", "national"},
				{"2026-05-27", "Idul Adha 1447 Hijriah", "national"},
				{"2026-05-31", "Hari Raya Waisak 2570 BE", "national"},
				{"2026-06-01", "Hari Lahir Pancasila", "national"},
				{"2026-06-16", "1 Muharam Tahun Baru Islam 1448 Hijriah", "national"},
				{"2026-08-17", "Proklamasi Kemerdekaan RI", "national"},
				{"2026-08-25", "Maulid Nabi Muhammad S.A.W.", "national"},
				{"2026-12-25", "Kelahiran Yesus Kristus (Natal)", "national"},

				// === CUTI BERSAMA (6 hari) ===
				{"2026-02-16", "Cuti Bersama Tahun Baru Imlek 2577 Kongzili", "cuti_bersama"},
				{"2026-03-18", "Cuti Bersama Hari Suci Nyepi (Tahun Baru Saka 1948)", "cuti_bersama"},
				{"2026-03-20", "Cuti Bersama Idul Fitri 1447 Hijriah", "cuti_bersama"},
				{"2026-03-23", "Cuti Bersama Idul Fitri 1447 Hijriah", "cuti_bersama"},
				{"2026-03-24", "Cuti Bersama Idul Fitri 1447 Hijriah", "cuti_bersama"},
				{"2026-05-15", "Cuti Bersama Kenaikan Yesus Kristus", "cuti_bersama"},
				{"2026-05-28", "Cuti Bersama Idul Adha 1447 Hijriah", "cuti_bersama"},
				{"2026-12-24", "Cuti Bersama Kelahiran Yesus Kristus", "cuti_bersama"},
			}
		} else {
			// Untuk tahun lain, gunakan hari libur fix (tanggal sama tiap tahun)
			fixedHolidays := []holidayData{
				{yearStr + "-01-01", "Tahun Baru Masehi", "national"},
				{yearStr + "-05-01", "Hari Buruh Internasional", "national"},
				{yearStr + "-06-01", "Hari Lahir Pancasila", "national"},
				{yearStr + "-08-17", "Hari Kemerdekaan RI", "national"},
				{yearStr + "-12-25", "Hari Raya Natal", "national"},
			}
			holidays = fixedHolidays
		}

		addedCount := 0
		skippedCount := 0
		var added []string
		var skipped []string

		for _, h := range holidays {
			var count int
			db.QueryRow("SELECT COUNT(*) FROM holidays WHERE date=?", h.date).Scan(&count)
			if count == 0 {
				_, err := db.Exec(
					"INSERT INTO holidays (date, name, type, description) VALUES (?, ?, ?, ?)",
					h.date, h.name, h.typ, "Import Nasional Standar (SKB 3 Menteri 2025)",
				)
				if err == nil {
					addedCount++
					added = append(added, h.date+" "+h.name)
				}
			} else {
				skippedCount++
				skipped = append(skipped, h.date+" "+h.name)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Import selesai: %d hari libur ditambahkan, %d sudah ada", addedCount, skippedCount),
			"added":   addedCount,
			"skipped": skippedCount,
			"added_list":   added,
			"skipped_list": skipped,
			"year":    yearStr,
			"source":  "Keputusan Bersama Menag, Menaker, MenPANRB Nomor 1497/2025",
		})
	}
}
