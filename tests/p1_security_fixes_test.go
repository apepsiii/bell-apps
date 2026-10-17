package tests

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	appMiddleware "belsekolah/internal/middleware"
	"belsekolah/internal/router"
)

// --- SEC-05: Attendance & WA Test Endpoint Security ---

func TestSEC05_TestWA_RequiresAdminAuth(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	router.RegisterRoutes(e, app.DB, os.DirFS("."), app)

	// Obtain a valid CSRF token from a GET request to /login
	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	var csrfToken string
	for _, cookie := range getRec.Result().Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
			break
		}
	}
	assert.NotEmpty(t, csrfToken)
	csrfCookie := &http.Cookie{Name: "csrf_token", Value: csrfToken}

	// 1. Unauthenticated request to /admin/attendance/test-wa must be blocked (redirected to /login)
	req := httptest.NewRequest(http.MethodPost, "/admin/attendance/test-wa", nil)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// AdminAuth middleware redirects unauthenticated users to /login (303 See Other)
	assert.Equal(t, http.StatusSeeOther, rec.Code, "Unauthenticated call to /admin/attendance/test-wa must redirect to /login")
	assert.Equal(t, "/login", rec.Header().Get("Location"))

	// 2. Legacy /api/attendance/test-wa is also protected with AdminAuth() (must redirect to /login)
	reqOld := httptest.NewRequest(http.MethodPost, "/api/attendance/test-wa", nil)
	reqOld.AddCookie(csrfCookie)
	reqOld.Header.Set("X-CSRF-Token", csrfToken)
	recOld := httptest.NewRecorder()
	e.ServeHTTP(recOld, reqOld)
	assert.Equal(t, http.StatusSeeOther, recOld.Code, "Unauthenticated /api/attendance/test-wa must redirect to /login")
	assert.Equal(t, "/login", recOld.Header().Get("Location"))
}

func TestSEC05_Attendance_DeviceToken_Enforcement(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Configure device_token in attendance_settings
	const secretToken = "my-rfid-iot-secret-key-2026"
	_, err := app.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('device_token', ?)", secretToken)
	assert.NoError(t, err)

	// Subtest A: Record Attendance without token -> 401 Unauthorized
	reqNoToken := httptest.NewRequest(http.MethodPost, "/api/attendance/record?rfid=123456", nil)
	recNoToken := httptest.NewRecorder()
	cNoToken := e.NewContext(reqNoToken, recNoToken)
	err = app.RecordAttendanceHandler(cNoToken)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, recNoToken.Code)
	assert.Contains(t, recNoToken.Body.String(), "Token perangkat tidak valid")

	// Subtest B: Record Attendance with invalid token -> 401 Unauthorized
	reqBadToken := httptest.NewRequest(http.MethodPost, "/api/attendance/record?rfid=123456", nil)
	reqBadToken.Header.Set("X-Device-Token", "wrong-token")
	recBadToken := httptest.NewRecorder()
	cBadToken := e.NewContext(reqBadToken, recBadToken)
	err = app.RecordAttendanceHandler(cBadToken)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, recBadToken.Code)

	// Subtest C: Record Attendance with valid token via X-Device-Token -> Authenticated (passes device auth)
	reqGoodToken := httptest.NewRequest(http.MethodPost, "/api/attendance/record?rfid=UNKNOWN_RFID_123", nil)
	reqGoodToken.Header.Set("X-Device-Token", secretToken)
	recGoodToken := httptest.NewRecorder()
	cGoodToken := e.NewContext(reqGoodToken, recGoodToken)
	err = app.RecordAttendanceHandler(cGoodToken)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusUnauthorized, recGoodToken.Code, "Valid X-Device-Token must pass device authentication")

	// Subtest D: Prayer Attendance without token -> 401 Unauthorized
	reqPrayerNoToken := httptest.NewRequest(http.MethodPost, "/api/attendance/prayer?rfid=123456", nil)
	recPrayerNoToken := httptest.NewRecorder()
	cPrayerNoToken := e.NewContext(reqPrayerNoToken, recPrayerNoToken)
	err = app.PrayerAttendanceHandler(cPrayerNoToken)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, recPrayerNoToken.Code)

	// Subtest E: Prayer Attendance with valid query param token -> Authenticated
	reqPrayerGoodToken := httptest.NewRequest(http.MethodPost, "/api/attendance/prayer?rfid=UNKNOWN_RFID_123&token="+secretToken, nil)
	recPrayerGoodToken := httptest.NewRecorder()
	cPrayerGoodToken := e.NewContext(reqPrayerGoodToken, recPrayerGoodToken)
	err = app.PrayerAttendanceHandler(cPrayerGoodToken)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusUnauthorized, recPrayerGoodToken.Code, "Valid token in query must pass device authentication")
}

// --- SEC-06: Audio Upload Security & Collision Handling ---

func TestSEC06_AudioUpload_Extension_And_MIME_Validation(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	uploadFn := handler.UploadAudio(app.DB)

	// 1. Attempt to upload a dangerous PHP script disguised or named shell.php
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "shell.php")
	part.Write([]byte("<?php echo 'malicious code'; ?>"))
	writer.WriteField("display_name", "Malicious PHP")
	writer.Close()

	reqPhp := httptest.NewRequest(http.MethodPost, "/api/audio/upload", body)
	reqPhp.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	recPhp := httptest.NewRecorder()
	cPhp := e.NewContext(reqPhp, recPhp)

	err := uploadFn(cPhp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recPhp.Code)
	assert.Contains(t, recPhp.Body.String(), "Format file tidak didukung")

	// 2. Attempt to upload a text file with a spoofed .mp3 extension
	bodySpoof := &bytes.Buffer{}
	writerSpoof := multipart.NewWriter(bodySpoof)
	partSpoof, _ := writerSpoof.CreateFormFile("file", "fake_audio.mp3")
	partSpoof.Write([]byte("This is plain text and definitely not valid audio data!"))
	writerSpoof.WriteField("display_name", "Fake MP3")
	writerSpoof.Close()

	reqSpoof := httptest.NewRequest(http.MethodPost, "/api/audio/upload", bodySpoof)
	reqSpoof.Header.Set(echo.HeaderContentType, writerSpoof.FormDataContentType())
	recSpoof := httptest.NewRecorder()
	cSpoof := e.NewContext(reqSpoof, recSpoof)

	err = uploadFn(cSpoof)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recSpoof.Code)
	assert.Contains(t, recSpoof.Body.String(), "File bukan audio yang valid")
}

func TestSEC06_AudioUpload_Collision_Avoidance(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	uploadFn := handler.UploadAudio(app.DB)

	// Create a dummy audio file in uploads dir
	uploadDir := config.GetUploadPath()
	os.MkdirAll(uploadDir, 0755)
	existingFilename := "bell_ding.ogg"
	existingFilePath := filepath.Join(uploadDir, existingFilename)

	// Valid OGG Vorbis header bytes
	oggMagic := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	err := os.WriteFile(existingFilePath, oggMagic, 0644)
	assert.NoError(t, err)
	defer os.Remove(existingFilePath)

	// Upload a file with identical filename
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", existingFilename)
	part.Write(oggMagic)
	writer.WriteField("display_name", "New Bell Ding")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/audio/upload", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = uploadFn(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check DB to verify filename has timestamp suffix to prevent overwriting
	var savedFilename string
	err = app.DB.QueryRow("SELECT file_name FROM audio_files WHERE display_name = 'New Bell Ding'").Scan(&savedFilename)
	assert.NoError(t, err)
	assert.NotEqual(t, existingFilename, savedFilename, "Duplicate file must be saved with unique timestamped filename")
	assert.True(t, strings.HasPrefix(savedFilename, "bell_ding_"), "Filename should start with bell_ding_")
	assert.True(t, strings.HasSuffix(savedFilename, ".ogg"), "Filename should end with .ogg")

	// Cleanup saved collision file
	os.Remove(filepath.Join(uploadDir, savedFilename))
}

// --- SEC-07: Operator Session Standardization with Database TTL ---

func TestSEC07_Operator_Session_Database_TTL(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// Ensure operator_sessions table exists in test in-memory DB
	_, err := app.DB.Exec(`CREATE TABLE IF NOT EXISTS operator_sessions (
		token TEXT PRIMARY KEY,
		operator_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	assert.NoError(t, err)

	// Create test operator
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("OperatorSecret#2026"), bcrypt.DefaultCost)
	assert.NoError(t, err)
	res, err := app.DB.Exec("INSERT INTO operators (username, password, name, is_active) VALUES ('test_op', ?, 'Test Operator', 1)", string(hashedPass))
	assert.NoError(t, err)
	opID, err := res.LastInsertId()
	assert.NoError(t, err)

	loginFn := handler.OperatorLogin(app.DB)

	// 1. Perform successful Operator Login
	form := url.Values{
		"username": {"test_op"},
		"password": {"OperatorSecret#2026"},
	}
	reqLogin := httptest.NewRequest(http.MethodPost, "/operator/login", strings.NewReader(form.Encode()))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	err = loginFn(cLogin)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recLogin.Code)

	// Extract operator_session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range recLogin.Result().Cookies() {
		if cookie.Name == "operator_session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie, "operator_session cookie must be set upon login")
	assert.NotEmpty(t, sessionCookie.Value)

	// Verify session exists in operator_sessions table with future expiration
	var dbOpID int
	var expiresAt time.Time
	err = app.DB.QueryRow("SELECT operator_id, expires_at FROM operator_sessions WHERE token = ?", sessionCookie.Value).Scan(&dbOpID, &expiresAt)
	assert.NoError(t, err)
	assert.Equal(t, int(opID), dbOpID)
	assert.True(t, expiresAt.After(time.Now()), "Session must expire in future")

	// 2. Verify OperatorAuth Middleware succeeds with valid session
	authMiddleware := handler.OperatorAuth(app.DB)
	protectedHandler := authMiddleware(func(c echo.Context) error {
		opVal := c.Get("operator_id")
		return c.String(http.StatusOK, fmt.Sprintf("OK: %v", opVal))
	})

	reqAuth := httptest.NewRequest(http.MethodGet, "/operator/dashboard", nil)
	reqAuth.AddCookie(sessionCookie)
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(reqAuth, recAuth)

	err = protectedHandler(cAuth)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recAuth.Code)
	assert.Equal(t, fmt.Sprintf("OK: %d", opID), recAuth.Body.String())

	// 3. Test Expired Session handling
	// Manually set expires_at in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	_, err = app.DB.Exec("UPDATE operator_sessions SET expires_at = ? WHERE token = ?", pastTime, sessionCookie.Value)
	assert.NoError(t, err)

	reqExpired := httptest.NewRequest(http.MethodGet, "/operator/dashboard", nil)
	reqExpired.AddCookie(sessionCookie)
	recExpired := httptest.NewRecorder()
	cExpired := e.NewContext(reqExpired, recExpired)

	err = protectedHandler(cExpired)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, recExpired.Code, "Expired operator session must redirect to /operator/login")
	assert.Equal(t, "/operator/login", recExpired.Header().Get("Location"))

	// Verify expired session was purged from operator_sessions
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM operator_sessions WHERE token = ?", sessionCookie.Value).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Expired session must be deleted from operator_sessions")
}

// --- SEC-08: Content Security Policy & DOM XSS Hardening ---

func TestSEC08_CSP_Hardening_NoUnsafeEval(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.SecurityHeaders())
	e.GET("/test-csp", func(c echo.Context) error {
		return c.String(http.StatusOK, "CSP Test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test-csp", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	csp := rec.Header().Get("Content-Security-Policy")
	assert.NotEmpty(t, csp, "Content-Security-Policy header must be present")
	assert.NotContains(t, csp, "'unsafe-eval'", "CSP script-src must NOT contain 'unsafe-eval'")
	assert.Contains(t, csp, "default-src 'self'", "CSP must restrict default-src")
}

