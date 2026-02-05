package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// --- STRUCTS ---

type Holiday struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Type        string `json:"type"` // 'national' or 'internal'
	Description string `json:"description"`
}

type SchoolSetting struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// --- HANDLERS ---

// GetHolidaysHandler - List all holidays
func (a *App) GetHolidaysHandler(c echo.Context) error {
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

	rows, err := a.DB.Query(query, args...)
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

// AddHolidayHandler - Add new holiday
func (a *App) AddHolidayHandler(c echo.Context) error {
	var h Holiday
	if err := c.Bind(&h); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}

	if h.Date == "" || h.Name == "" || h.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Date, Name, and Type are required"})
	}

	// Check if exists
	var count int
	a.DB.QueryRow("SELECT COUNT(*) FROM holidays WHERE date=?", h.Date).Scan(&count)
	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]string{"message": "Holiday for this date already exists"})
	}

	_, err := a.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES (?, ?, ?, ?)",
		h.Date, h.Name, h.Type, h.Description)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Holiday added successfully"})
}

// UpdateHolidayHandler - Update holiday
func (a *App) UpdateHolidayHandler(c echo.Context) error {
	id := c.Param("id")
	var h Holiday
	if err := c.Bind(&h); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}

	_, err := a.DB.Exec("UPDATE holidays SET date=?, name=?, type=?, description=? WHERE id=?",
		h.Date, h.Name, h.Type, h.Description, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Holiday updated successfully"})
}

// DeleteHolidayHandler - Delete holiday
func (a *App) DeleteHolidayHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM holidays WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Holiday deleted successfully"})
}

// GetSchoolSettingsHandler - Get school settings
func (a *App) GetSchoolSettingsHandler(c echo.Context) error {
	rows, err := a.DB.Query("SELECT id, setting_key, setting_value FROM school_settings")
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

// UpdateSchoolSettingsHandler - Update school settings
func (a *App) UpdateSchoolSettingsHandler(c echo.Context) error {
	var body map[string]string
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}

	for k, v := range body {
		// Use INSERT OR REPLACE or UPSERT logic
		// Since we use sqlite, INSERT OR REPLACE is handy if uniqueness is enforced, but setting_key is unique.
		_, err := a.DB.Exec("INSERT INTO school_settings (setting_key, setting_value) VALUES (?, ?) ON CONFLICT(setting_key) DO UPDATE SET setting_value=excluded.setting_value", k, v)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated successfully"})
}

// --- HELPER FUNCTIONS ---

// IsWorkingDay checks if a specific date is a working day
func (a *App) IsWorkingDay(date time.Time) (bool, string) { // returns (isWorking, reason)
	dateStr := date.Format("2006-01-02")
	weekday := int(date.Weekday()) // Sunday=0, Monday=1...
	if weekday == 0 {
		weekday = 7 // Adjust Sunday to 7 to match common 1-7 (Mon-Sun) or just handle 0 in settings
	}

	// 1. Check Holidays
	var holidayName string
	err := a.DB.QueryRow("SELECT name FROM holidays WHERE date=?", dateStr).Scan(&holidayName)
	if err == nil {
		return false, "Libur: " + holidayName
	}

	// 2. Check Work Days Setting
	var workDaysStr string
	err = a.DB.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key='work_days'").Scan(&workDaysStr)
	if err != nil {
		workDaysStr = "1,2,3,4,5" // Default Mon-Fri
	}

	// weekday is 0-6 in Go (Sun-Sat). Let's standardize to SQLite/Setting format.
	// If setting uses 1=Mon ... 7=Sun.
	// Go: Sun=0, Mon=1...
	checkDay := weekday
	if date.Weekday() == time.Sunday {
		checkDay = 7
	}

	// Format: "1,2,3,4,5"
	if !strings.Contains(workDaysStr, string(rune('0'+checkDay))) { // Simple check, careful with 10? No, days are single digit.
		// Better splitting
		days := strings.Split(workDaysStr, ",")
		isWork := false
		for _, d := range days {
			// Handle '0' as Sunday if user configured that way, but let's stick to 1-7 standard or Go's standard.
			// Let's assume input is "1,..." where 1=Mon.
			// Go: 1=Mon. So match directly.
			// Sunday in Go is 0. If config uses 7 for Sunday.
			// Let's convert current day to string.
			dayChar := ""
			switch date.Weekday() {
			case time.Monday:
				dayChar = "1"
			case time.Tuesday:
				dayChar = "2"
			case time.Wednesday:
				dayChar = "3"
			case time.Thursday:
				dayChar = "4"
			case time.Friday:
				dayChar = "5"
			case time.Saturday:
				dayChar = "6"
			case time.Sunday:
				dayChar = "7" // Standardize Sunday as 7 for settings
			}

			if d == dayChar {
				isWork = true
				break
			}
		}
		if !isWork {
			return false, "Libur Akhir Pekan/Bukan Hari Kerja"
		}
	}

	return true, ""
}

// GetWorkingDaysInMonth returns the number of working days in a specific month
func (a *App) GetWorkingDaysInMonth(year int, month time.Month) int {
	// 1. Get total days in month
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	daysInMonth := end.Day()

	workingDays := 0
	for i := 1; i <= daysInMonth; i++ {
		date := time.Date(year, month, i, 0, 0, 0, 0, time.Local)
		isWork, _ := a.IsWorkingDay(date)
		if isWork {
			workingDays++
		}
	}
	return workingDays
}

// ImportNationalHolidaysHandler - Quick add standard holidays
func (a *App) ImportNationalHolidaysHandler(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = time.Now().Format("2006")
	}

	// Fixed holidays map (MM-DD -> Name)
	fixedHolidays := map[string]string{
		"01-01": "Tahun Baru Masehi",
		"05-01": "Hari Buruh Internasional",
		"06-01": "Hari Lahir Pancasila",
		"08-17": "Hari Kemerdekaan RI",
		"12-25": "Hari Raya Natal",
	}

	addedCount := 0
	for dateSuffix, name := range fixedHolidays {
		date := yearStr + "-" + dateSuffix
		
		// Check overlap
		var count int
		a.DB.QueryRow("SELECT COUNT(*) FROM holidays WHERE date=?", date).Scan(&count)
		if count == 0 {
			_, err := a.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES (?, ?, 'national', 'Libur Nasional Standar')", date, name)
			if err == nil {
				addedCount++
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Import selesai",
		"added":   addedCount,
	})
}
