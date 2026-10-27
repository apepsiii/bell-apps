package tests

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	appMiddleware "belsekolah/internal/middleware"
	"belsekolah/internal/repository"
	"belsekolah/internal/router"
)

// --- SEC-01: Zip Slip Path Traversal Protection ---

func TestSEC01_ZipSlip_PathTraversal_Blocked(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Create a malicious ZIP containing a path traversal entry: uploads/photos/../../evil.txt
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	// Add normal database.db
	dbWriter, err := zw.Create("database.db")
	assert.NoError(t, err)
	dbWriter.Write([]byte("mock database content"))

	// Add malicious traversal file
	evilWriter, err := zw.Create("uploads/photos/../../evil.txt")
	assert.NoError(t, err)
	evilWriter.Write([]byte("malicious content payload"))

	err = zw.Close()
	assert.NoError(t, err)

	// 2. Prepare multipart request to /admin/restore
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("backup_file", "backup_malicious.zip")
	assert.NoError(t, err)
	_, err = io.Copy(part, &zipBuf)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	restoreHandler := handler.RestoreDatabase(app.DB)
	err = restoreHandler(c)
	assert.NoError(t, err)

	// Verify it was rejected with 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "error", resp["status"])
	assert.Contains(t, resp["message"], "path traversal ilegal")

	// Ensure evil.txt was never written anywhere outside
	evilPath := filepath.Join(filepath.Dir(config.GetPhotoPath()), "evil.txt")
	_, statErr := os.Stat(evilPath)
	assert.True(t, os.IsNotExist(statErr), "evil.txt must not exist on filesystem")
}

// --- SEC-02: CSRF Middleware Protection ---

func TestSEC02_CSRF_MissingToken_Forbidden(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.CSRF())

	// Protected POST endpoint
	e.POST("/admin/test-action", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/test-action", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should be rejected with 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "CSRF token missing", resp["error"])
}

func TestSEC02_CSRF_ValidToken_Success(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.CSRF())

	e.GET("/test-page", func(c echo.Context) error {
		return c.String(http.StatusOK, "page rendered")
	})
	e.POST("/admin/test-action", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	})

	// 1. GET request to obtain CSRF cookie
	getReq := httptest.NewRequest(http.MethodGet, "/test-page", nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusOK, getRec.Code)
	cookieHeader := getRec.Header().Get("Set-Cookie")
	assert.Contains(t, cookieHeader, "csrf_token=")

	// Extract csrf_token cookie value
	var token string
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "csrf_token=") {
			token = strings.TrimPrefix(part, "csrf_token=")
			break
		}
	}
	assert.NotEmpty(t, token)

	// 2. POST request sending both cookie AND X-CSRF-Token header
	postReq := httptest.NewRequest(http.MethodPost, "/admin/test-action", nil)
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	postReq.Header.Set("X-CSRF-Token", token)
	postRec := httptest.NewRecorder()

	e.ServeHTTP(postRec, postReq)

	assert.Equal(t, http.StatusOK, postRec.Code)
	var resp map[string]string
	json.Unmarshal(postRec.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp["status"])
}

func TestSEC02_CSRF_TokenMismatch_Forbidden(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.CSRF())

	e.GET("/test-page", func(c echo.Context) error {
		return c.String(http.StatusOK, "page rendered")
	})
	e.POST("/admin/test-action", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	})

	// 1. GET to generate valid token
	getReq := httptest.NewRequest(http.MethodGet, "/test-page", nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	cookieHeader := getRec.Header().Get("Set-Cookie")
	var token string
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "csrf_token=") {
			token = strings.TrimPrefix(part, "csrf_token=")
			break
		}
	}

	// 2. POST with cookie matching stored token, but a DIFFERENT/forged header value
	postReq := httptest.NewRequest(http.MethodPost, "/admin/test-action", nil)
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	postReq.Header.Set("X-CSRF-Token", "forged_different_token")
	postRec := httptest.NewRecorder()

	e.ServeHTTP(postRec, postReq)

	assert.Equal(t, http.StatusForbidden, postRec.Code)
}

func TestSEC02_CSRF_Exempt_IoT_And_Auth_Endpoints(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.CSRF())

	e.POST("/api/attendance/verify-face", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "face_ok"})
	})
	e.POST("/api/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "login_ok"})
	})
	e.POST("/api/guru/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "guru_login_ok"})
	})

	// 1. Face verification POST without CSRF token should succeed
	reqFace := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", nil)
	recFace := httptest.NewRecorder()
	e.ServeHTTP(recFace, reqFace)
	assert.Equal(t, http.StatusOK, recFace.Code)

	// 2. Login POST without CSRF token should succeed
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	recLogin := httptest.NewRecorder()
	e.ServeHTTP(recLogin, reqLogin)
	assert.Equal(t, http.StatusOK, recLogin.Code)

	// 3. Guru Login POST without CSRF token should succeed
	reqGuruLogin := httptest.NewRequest(http.MethodPost, "/api/guru/login", nil)
	recGuruLogin := httptest.NewRecorder()
	e.ServeHTTP(recGuruLogin, reqGuruLogin)
	assert.Equal(t, http.StatusOK, recGuruLogin.Code)
}

// --- SEC-03: Point Reset GET Route Removal ---

func TestSEC03_PointReset_GET_Disabled(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	router.RegisterRoutes(e, app.DB, os.DirFS("."), app)

	// 1. Authenticate as admin to get session cookie
	// Use a non-default password so the factory-credential guard does not block it.
	strongPass := "Str0ngAdmin#Pass2026"
	cfg := config.GetConfig()
	origPass := cfg.Auth.AdminPass
	cfg.Auth.AdminPass = strongPass
	defer func() { cfg.Auth.AdminPass = origPass }()

	loginForm := url.Values{
		"username": {config.GetAdminUser()},
		"password": {strongPass},
	}
	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == config.GetCookieName() {
			sessionCookie = c
			break
		}
	}
	assert.NotNil(t, sessionCookie, "Session cookie should be set after admin login")

	// 2. Attempt GET on /admin/points/reset/1 with admin session
	req := httptest.NewRequest(http.MethodGet, "/admin/points/reset/1", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Since GET route was removed, Echo should return 404 (Not Found) or 405 (Method Not Allowed)
	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"GET /admin/points/reset/:id should be 404 or 405, got %d", rec.Code)

	// 3. Attempt GET on /admin/points/reset-bulk with admin session
	reqBulk := httptest.NewRequest(http.MethodGet, "/admin/points/reset-bulk", nil)
	reqBulk.AddCookie(sessionCookie)
	recBulk := httptest.NewRecorder()
	e.ServeHTTP(recBulk, reqBulk)

	assert.True(t, recBulk.Code == http.StatusNotFound || recBulk.Code == http.StatusMethodNotAllowed,
		"GET /admin/points/reset-bulk should be 404 or 405, got %d", recBulk.Code)
}

// --- SEC-04: Enforced Bcrypt Authentication Hardening ---

func TestSEC04_Bcrypt_Authentication(t *testing.T) {
	e := echo.New()

	// Hash a custom password with bcrypt
	rawPass := "SecureAdminPass#2026"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPass), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Set in-memory config for this test
	cfg := config.GetConfig()
	origPass := cfg.Auth.AdminPass
	cfg.Auth.AdminPass = string(hashedBytes)
	defer func() { cfg.Auth.AdminPass = origPass }()

	// Seed an in-memory database so role resolution can succeed.
	testDB, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	defer testDB.Close()
	testDB.SetMaxOpenConns(1)
	repository.RunMigrations(testDB)

	loginHandler := handler.Login(testDB)

	// 1. Valid password with bcrypt hash
	formSuccess := url.Values{
		"username": {config.GetAdminUser()},
		"password": {rawPass},
	}
	reqSuccess := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(formSuccess.Encode()))
	reqSuccess.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recSuccess := httptest.NewRecorder()
	cSuccess := e.NewContext(reqSuccess, recSuccess)

	err = loginHandler(cSuccess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, recSuccess.Code)
	assert.Equal(t, "/admin", recSuccess.Header().Get("Location"))

	// 2. Invalid password
	formFail := url.Values{
		"username": {config.GetAdminUser()},
		"password": {"WrongPassword123"},
	}
	reqFail := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(formFail.Encode()))
	reqFail.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recFail := httptest.NewRecorder()
	cFail := e.NewContext(reqFail, recFail)

	err = loginHandler(cFail)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, recFail.Code)
	assert.Equal(t, "/login?error=1", recFail.Header().Get("Location"))
}

