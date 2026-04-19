package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// --- AUTHENTICATION TESTS ---

func TestLoginHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// Test successful login
	form := make(url.Values)
	form.Set("username", AdminUser)
	form.Set("password", AdminPass)
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := app.LoginHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), CookieName)
	assert.Equal(t, "/admin", rec.Header().Get("Location"))

	// Test failed login
	form.Set("password", "wrongpassword")
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = app.LoginHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login?error=1", rec.Header().Get("Location"))
}

// --- ATTENDANCE TESTS ---

func setupAttendanceTests(t *testing.T, app *App) {
	// Seed a student
	_, err := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name) VALUES (?, ?, ?)", "student-rfid-123", "2001", "Test Student")
	assert.NoError(t, err)

	// Seed attendance settings
	now := time.Now()
	arrivalStart := now.Add(-1 * time.Hour).Format("15:04")
	arrivalEnd := now.Add(1 * time.Hour).Format("15:04")
	departureStart := now.Add(8 * time.Hour).Format("15:04")
	departureEnd := now.Add(9 * time.Hour).Format("15:04")

	tx, _ := app.DB.Begin()
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_start", arrivalStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_end", arrivalEnd)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_start", departureStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_end", departureEnd)
	tx.Commit()
}

func TestRecordAttendanceHandler_CheckIn(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	setupAttendanceTests(t, app)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/attendance/record?rfid=student-rfid-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := app.RecordAttendanceHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"success"`)
	assert.Contains(t, rec.Body.String(), `"message":"Absen Datang Berhasil"`)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE rfid_uid = ? AND status = 'Datang'", "student-rfid-123").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRecordAttendanceHandler_DuplicateCheckIn(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	setupAttendanceTests(t, app)

	// First check-in
	app.DB.Exec("INSERT INTO attendance_logs (rfid_uid, date, status) VALUES (?, ?, ?)", "student-rfid-123", time.Now().Format("2006-01-02"), "Datang")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/attendance/record?rfid=student-rfid-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := app.RecordAttendanceHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"duplicate"`)
}

func TestRecordAttendanceHandler_UnknownRFID(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	setupAttendanceTests(t, app)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/attendance/record?rfid=unknown-rfid-xxx", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := app.RecordAttendanceHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Kartu tidak dikenali")
}

func TestManualAttendanceHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed student
	res, _ := app.DB.Exec("INSERT INTO students (nis, name) VALUES (?, ?)", "2002", "Manual Student")
	studentID, _ := res.LastInsertId()

	e := echo.New()
	
	form := make(url.Values)
	form.Set("student_id", fmt.Sprintf("%d", studentID))
	form.Set("status", "Sakit")

	req := httptest.NewRequest(http.MethodPost, "/admin/attendance/manual", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := app.ManualAttendanceHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var status, method string
	err = app.DB.QueryRow("SELECT status, method FROM attendance_logs WHERE user_name = ?", "Manual Student").Scan(&status, &method)
	assert.NoError(t, err)
	assert.Equal(t, "Sakit", status)
	assert.Equal(t, "MANUAL", method)
}
