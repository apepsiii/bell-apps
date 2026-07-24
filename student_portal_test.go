package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	"belsekolah/internal/repository"
)

func setupTestApp(t *testing.T) (*App, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	repository.RunMigrations(db)
	repository.SeedDefaultData(db)
	repository.SeedPointRules(db)

	app := &App{DB: db}
	os.MkdirAll(config.GetUploadPath(), 0755)

	teardown := func() {
		db.Close()
		os.RemoveAll("./public")
	}
	return app, teardown
}

func seedStudent(t *testing.T, app *App, nis, name, password string, active bool) {
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
	assert.NoError(t, err)
}

func TestStudentLogin_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "1001", "Andi", "4321", true)

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "1001")
	form.Set("pin", "4321")

	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.StudentLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "Login berhasil", body["message"])
	assert.Contains(t, rec.Header().Get("Set-Cookie"), handler.StudentSessionCookie)
}

func TestStudentLogin_DefaultPIN(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "2001", "Budi", "", true) // empty password = default PIN

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "2001")
	form.Set("pin", handler.StudentDefaultPIN)

	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.StudentLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "Login berhasil", body["message"])
}

func TestStudentLogin_WrongPIN(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "3001", "Citra", "9999", true)

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "3001")
	form.Set("pin", "0000")

	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.StudentLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestStudentLogin_Inactive(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "4001", "Dewi", "5555", false) // inactive

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "4001")
	form.Set("pin", "5555")

	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.StudentLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestStudentLogin_EmptyFields(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "")
	form.Set("pin", "")

	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.StudentLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestStudentDashboard(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "5001", "Eko", "1234", true)

	// login first
	form := url.Values{}
	form.Set("nis", "5001")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)

	// extract cookie
	cookies := rec.Header()["Set-Cookie"]

	// use auth middleware
	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetStudentDashboard(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	assert.NotNil(t, body["student"])
	assert.NotNil(t, body["today"])
}

func TestStudentLogout(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6001", "Fani", "1111", true)

	// login
	e := echo.New()
	form := url.Values{}
	form.Set("nis", "6001")
	form.Set("pin", "1111")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))

	cookies := rec.Header()["Set-Cookie"]

	// logout
	req2 := httptest.NewRequest(http.MethodPost, "/api/student/logout", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := handler.StudentLogout(app.DB)(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// cookie should be cleared
	assert.Contains(t, rec2.Header().Get("Set-Cookie"), "Max-Age=0")
}

func TestChangeStudentPIN(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "7001", "Gina", "oldpin", true)

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "7001")
	form.Set("pin", "oldpin")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))
	cookies := rec.Header()["Set-Cookie"]

	// change PIN
	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.ChangeStudentPIN(app.DB)

	body := `{"old_pin":"oldpin","new_pin":"654321"}`
	req2 := httptest.NewRequest(http.MethodPut, "/api/student/pin", strings.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// verify new PIN works
	form2 := url.Values{}
	form2.Set("nis", "7001")
	form2.Set("pin", "654321")
	req3 := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form2.Encode()))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = handler.StudentLogin(app.DB)(c3)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

func TestStudentAuth_MiddlewareBlocksUnauthenticated(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	mw := handler.StudentAuth(app.DB)

	// Test page redirect
	req := httptest.NewRequest(http.MethodGet, "/student/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := mw(func(c echo.Context) error { return c.String(http.StatusOK, "ok") })(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/student/login", rec.Header().Get("Location"))

	// Test API 401
	req2 := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = mw(func(c echo.Context) error { return c.String(http.StatusOK, "ok") })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

func TestStudentProfile(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	tx, _ := app.DB.Begin()
	tx.Exec("INSERT INTO classes (id, name) VALUES (?, ?)", 1, "XII RPL")
	tx.Exec("INSERT INTO students (rfid_uid, nis, name, class_id, parent_name, parent_phone, birthday, password) VALUES (?, ?, ?, ?, ?, ?, ?, '')",
		"rfid-8001", "8001", "Hani", 1, "Ibu Hani", "0812", "2006-05-10")
	tx.Commit()

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "8001")
	form.Set("pin", handler.StudentDefaultPIN)
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetStudentPortalProfile(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/profile", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	assert.Equal(t, "Hani", body["name"])
	assert.Equal(t, "XII RPL", body["class"])
	assert.Equal(t, "Ibu Hani", body["parent_name"])
	assert.Equal(t, true, body["using_default_pin"])
}

func TestGetMyPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	tx, _ := app.DB.Begin()
	tx.Exec("INSERT INTO students (rfid_uid, nis, name) VALUES (?, ?, ?)", "rfid-9001", "9001", "Indah")
	tx.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (?, ?, ?, ?, ?)", 1, 10, "Bonus awal", time.Now().Format("2006-01-02 15:04:05"), "system")
	tx.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (?, ?, ?, ?, ?)", 1, -5, "Terlambat", time.Now().Format("2006-01-02 15:04:05"), "system")
	tx.Commit()

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "9001")
	form.Set("pin", handler.StudentDefaultPIN)
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetMyPoints(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/points", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	assert.Equal(t, float64(5), body["total_points"])
}

func TestGetMyQRCard(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	tx, _ := app.DB.Begin()
	tx.Exec("INSERT INTO classes (id, name) VALUES (?, ?)", 1, "XII TKR")
	tx.Exec("INSERT INTO students (rfid_uid, nis, name, class_id) VALUES (?, ?, ?, ?)", "rfid-qr001", "qr001", "Joni", 1)
	tx.Commit()

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "qr001")
	form.Set("pin", handler.StudentDefaultPIN)
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetMyQRCard(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/qrcard", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	assert.Equal(t, "Joni", body["name"])
	assert.Equal(t, "qr001", body["nis"])
	assert.NotEmpty(t, body["qr_image"])
}

func TestGetMyCalendar(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	today := time.Now()
	dateStr := today.Format("2006-01-02")
	timestamp := today.Format("2006-01-02 15:04:05")
	monthStr := today.Format("01")
	yearStr := today.Format("2006")

	tx, _ := app.DB.Begin()
	tx.Exec("INSERT INTO students (rfid_uid, nis, name) VALUES (?, ?, ?)", "rfid-cal001", "cal001", "Kiki")
	tx.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, timestamp, date, method) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"rfid-cal001", "Kiki", "Siswa", "Datang", timestamp, dateStr, "RFID")
	tx.Commit()

	e := echo.New()
	form := url.Values{}
	form.Set("nis", "cal001")
	form.Set("pin", handler.StudentDefaultPIN)
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	handler.StudentLogin(app.DB)(e.NewContext(req, rec))
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetMyCalendar(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/calendar?month="+monthStr+"&year="+yearStr, nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	calendar, ok := body["calendar"].(map[string]interface{})
	assert.True(t, ok, "calendar should be a map")
	dayKey := fmt.Sprintf("%d", today.Day())
	assert.NotNil(t, calendar[dayKey], "today should have attendance entry")

	stats, ok := body["stats"].(map[string]interface{})
	assert.True(t, ok, "stats should be a map")
	assert.Equal(t, float64(1), stats["present"])
}
