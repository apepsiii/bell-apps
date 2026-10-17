package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/handler"
	"belsekolah/internal/router"
)

// --- SEC-09: Rate Limiter IP Extraction Hardening (Anti-Spoofing) ---

func TestSEC09_RateLimiter_Spoofing_Blocked(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	router.RegisterRoutes(e, app.DB, os.DirFS("."), app)

	// Obtain valid CSRF token from GET /login
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

	// Auth rate limiter limit is 5 attempts per minute.
	// An attacker attempts 7 logins from RemoteAddr "192.168.1.50:12345",
	// spoofing X-Forwarded-For with unique IPs on each attempt.
	const fixedRemoteAddr = "192.168.1.50:12345"
	var lastStatusCode int

	for i := 1; i <= 7; i++ {
		form := url.Values{
			"username":   {"admin"},
			"password":   {"WrongPass"},
			"csrf_token": {csrfToken},
		}
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.Header.Set("X-Forwarded-For", "203.0.113."+string(rune('0'+i))) // Spoofed IP header
		req.RemoteAddr = fixedRemoteAddr
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		lastStatusCode = rec.Code
	}

	// Because e.IPExtractor = echo.ExtractIPDirect() is set,
	// spoofed X-Forwarded-For is ignored, and the rate limiter triggers 429 Too Many Requests.
	assert.Equal(t, http.StatusTooManyRequests, lastStatusCode,
		"Rate limiter should block spoofed requests once threshold is exceeded for the actual RemoteAddr")
}

// --- SEC-10: SSRF Protection on AI Settings ---

func TestSEC10_SSRF_Protection(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	updateHandler := handler.UpdateAISettings(app.DB)

	// 1. Disallow non-HTTP/HTTPS schemes (e.g. file://, ftp://, gopher://)
	formFile := url.Values{
		"ai_base_url": {"file:///etc/passwd"},
		"ai_api_key":  {"test-key"},
		"ai_model":    {"gpt-4o-mini"},
	}
	reqFile := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formFile.Encode()))
	reqFile.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recFile := httptest.NewRecorder()
	cFile := e.NewContext(reqFile, recFile)

	err := updateHandler(cFile)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recFile.Code)
	assert.Contains(t, recFile.Body.String(), "Hanya protokol http dan https yang diizinkan")

	// 2. Disallow Cloud Metadata IP (169.254.169.254)
	formMeta := url.Values{
		"ai_base_url": {"http://169.254.169.254/latest/meta-data/"},
		"ai_api_key":  {"test-key"},
		"ai_model":    {"gpt-4o-mini"},
	}
	reqMeta := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formMeta.Encode()))
	reqMeta.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recMeta := httptest.NewRecorder()
	cMeta := e.NewContext(reqMeta, recMeta)

	err = updateHandler(cMeta)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recMeta.Code)
	assert.Contains(t, recMeta.Body.String(), "Akses ke endpoint metadata cloud dilarang")

	// 3. Disallow Link-Local IP (e.g. 169.254.10.20)
	formLinkLocal := url.Values{
		"ai_base_url": {"http://169.254.10.20:8000/v1"},
		"ai_api_key":  {"test-key"},
		"ai_model":    {"gpt-4o-mini"},
	}
	reqLinkLocal := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formLinkLocal.Encode()))
	reqLinkLocal.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recLinkLocal := httptest.NewRecorder()
	cLinkLocal := e.NewContext(reqLinkLocal, recLinkLocal)

	err = updateHandler(cLinkLocal)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recLinkLocal.Code)
	assert.Contains(t, recLinkLocal.Body.String(), "Akses ke alamat link-local dilarang")

	// 4. Allow valid public API endpoint
	formValid := url.Values{
		"ai_base_url": {"https://api.openai.com/v1"},
		"ai_api_key":  {"sk-proj-validtestkey12345678"},
		"ai_model":    {"gpt-4o-mini"},
	}
	reqValid := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formValid.Encode()))
	reqValid.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recValid := httptest.NewRecorder()
	cValid := e.NewContext(reqValid, recValid)

	err = updateHandler(cValid)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recValid.Code)
}

// --- SEC-11: Adaptive Secure Flag on Session Cookies ---

func TestSEC11_SessionCookies_Secure_Flag(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Admin Login via HTTPS (X-Forwarded-Proto: https) -> Cookie must have Secure: true
	loginAdmin := handler.Login()
	formAdmin := url.Values{
		"username": {"admin"},
		"password": {"admin123"},
	}
	reqAdmin := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(formAdmin.Encode()))
	reqAdmin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqAdmin.Header.Set("X-Forwarded-Proto", "https")
	recAdmin := httptest.NewRecorder()
	cAdmin := e.NewContext(reqAdmin, recAdmin)

	err := loginAdmin(cAdmin)
	assert.NoError(t, err)

	var adminCookie *http.Cookie
	for _, cookie := range recAdmin.Result().Cookies() {
		if cookie.Name == "session_token" {
			adminCookie = cookie
			break
		}
	}
	assert.NotNil(t, adminCookie)
	assert.True(t, adminCookie.Secure, "Admin session cookie must have Secure=true over HTTPS/X-Forwarded-Proto")

	// 2. Operator Login via HTTPS
	hashedOpPass, _ := bcrypt.GenerateFromPassword([]byte("Operator#123"), bcrypt.DefaultCost)
	app.DB.Exec("INSERT INTO operators (username, password, name, is_active) VALUES ('sec_op', ?, 'Sec Op', 1)", string(hashedOpPass))

	loginOp := handler.OperatorLogin(app.DB)
	formOp := url.Values{
		"username": {"sec_op"},
		"password": {"Operator#123"},
	}
	reqOp := httptest.NewRequest(http.MethodPost, "/operator/login", strings.NewReader(formOp.Encode()))
	reqOp.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqOp.Header.Set("X-Forwarded-Proto", "https")
	recOp := httptest.NewRecorder()
	cOp := e.NewContext(reqOp, recOp)

	err = loginOp(cOp)
	assert.NoError(t, err)

	var opCookie *http.Cookie
	for _, cookie := range recOp.Result().Cookies() {
		if cookie.Name == "operator_session" {
			opCookie = cookie
			break
		}
	}
	assert.NotNil(t, opCookie)
	assert.True(t, opCookie.Secure, "Operator session cookie must have Secure=true over HTTPS/X-Forwarded-Proto")

	// 3. Student Login via HTTPS
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte("654321"), bcrypt.DefaultCost)
	app.DB.Exec("INSERT INTO students (nis, name, password, status) VALUES ('12345001', 'Budi Santoso', ?, 'active')", string(hashedPin))

	loginStudent := handler.StudentLogin(app.DB)
	formStudent := url.Values{
		"nis": {"12345001"},
		"pin": {"654321"},
	}
	reqStudent := httptest.NewRequest(http.MethodPost, "/student/login", strings.NewReader(formStudent.Encode()))
	reqStudent.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqStudent.Header.Set("X-Forwarded-Proto", "https")
	recStudent := httptest.NewRecorder()
	cStudent := e.NewContext(reqStudent, recStudent)

	err = loginStudent(cStudent)
	assert.NoError(t, err)

	var studentCookie *http.Cookie
	for _, cookie := range recStudent.Result().Cookies() {
		if cookie.Name == handler.StudentSessionCookie {
			studentCookie = cookie
			break
		}
	}
	assert.NotNil(t, studentCookie)
	assert.True(t, studentCookie.Secure, "Student session cookie must have Secure=true over HTTPS/X-Forwarded-Proto")

	// 4. Over plain HTTP, Secure flag is false (to avoid breaking LAN HTTP intranet)
	reqHttp := httptest.NewRequest(http.MethodPost, "/student/login", strings.NewReader(formStudent.Encode()))
	reqHttp.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recHttp := httptest.NewRecorder()
	cHttp := e.NewContext(reqHttp, recHttp)

	err = loginStudent(cHttp)
	assert.NoError(t, err)

	for _, cookie := range recHttp.Result().Cookies() {
		if cookie.Name == handler.StudentSessionCookie {
			assert.False(t, cookie.Secure, "Cookie should not force Secure=true over plain HTTP")
			break
		}
	}
}

// --- SEC-12: API Key Masking & Safe Update ---

func TestSEC12_API_Key_Masking_And_Preservation(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	getHandler := handler.GetAISettings(app.DB)
	updateHandler := handler.UpdateAISettings(app.DB)

	// 1. Save an initial secret API key
	const originalRawKey = "sk-proj-RealSecretAPIKey1234567890ABCDEF"
	formInitial := url.Values{
		"ai_base_url": {"https://api.openai.com/v1"},
		"ai_api_key":  {originalRawKey},
		"ai_model":    {"gpt-4o-mini"},
	}
	reqInit := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formInitial.Encode()))
	reqInit.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recInit := httptest.NewRecorder()
	cInit := e.NewContext(reqInit, recInit)

	err := updateHandler(cInit)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recInit.Code)

	// 2. Fetch settings via GetAISettings: API key must be masked
	reqGet := httptest.NewRequest(http.MethodGet, "/admin/ai/settings", nil)
	recGet := httptest.NewRecorder()
	cGet := e.NewContext(reqGet, recGet)

	err = getHandler(cGet)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recGet.Code)

	var settings map[string]string
	err = json.Unmarshal(recGet.Body.Bytes(), &settings)
	assert.NoError(t, err)

	maskedKey := settings["ai_api_key"]
	assert.NotEqual(t, originalRawKey, maskedKey, "Raw API key must NOT be returned in plaintext")
	assert.Contains(t, maskedKey, "****", "Masked key should contain ****")
	assert.True(t, strings.HasPrefix(maskedKey, "sk-p"), "Masked key should keep prefix")
	assert.True(t, strings.HasSuffix(maskedKey, "CDEF"), "Masked key should keep suffix")

	// 3. User saves settings again without changing the masked key (submits maskedKey)
	formUpdateUnchanged := url.Values{
		"ai_base_url": {"https://api.openai.com/v1"},
		"ai_api_key":  {maskedKey}, // Submitting the masked key back
		"ai_model":    {"gpt-4o"},  // Changed model
	}
	reqUpdateUnchanged := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formUpdateUnchanged.Encode()))
	reqUpdateUnchanged.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recUpdateUnchanged := httptest.NewRecorder()
	cUpdateUnchanged := e.NewContext(reqUpdateUnchanged, recUpdateUnchanged)

	err = updateHandler(cUpdateUnchanged)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recUpdateUnchanged.Code)

	// Verify that the original raw key in database was NOT overwritten with the masked string!
	var dbKey string
	err = app.DB.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = 'ai_api_key'").Scan(&dbKey)
	assert.NoError(t, err)
	assert.Equal(t, originalRawKey, dbKey, "Database must preserve original raw key when masked string is submitted")

	// 4. User provides a new valid API key
	const newRawKey = "sk-new-supersecretkey-987654321"
	formUpdateNew := url.Values{
		"ai_base_url": {"https://api.openai.com/v1"},
		"ai_api_key":  {newRawKey},
		"ai_model":    {"gpt-4o"},
	}
	reqUpdateNew := httptest.NewRequest(http.MethodPost, "/admin/ai/settings", strings.NewReader(formUpdateNew.Encode()))
	reqUpdateNew.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recUpdateNew := httptest.NewRecorder()
	cUpdateNew := e.NewContext(reqUpdateNew, recUpdateNew)

	err = updateHandler(cUpdateNew)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recUpdateNew.Code)

	// Verify the new raw key is now stored in database
	err = app.DB.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = 'ai_api_key'").Scan(&dbKey)
	assert.NoError(t, err)
	assert.Equal(t, newRawKey, dbKey, "Database should update to new raw key")
}

