package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"belsekolah/internal/handler"
	appMiddleware "belsekolah/internal/middleware"
)

func TestAdminModularLayout_StislaMeridian(t *testing.T) {
	// 1. Verify layout_head.html
	headBytes, err := os.ReadFile("views/admin/layout_head.html")
	assert.NoError(t, err, "views/admin/layout_head.html must exist")
	headContent := string(headBytes)
	assert.Contains(t, headContent, "stisla-theme", "Must include stisla-theme localStorage key")
	assert.Contains(t, headContent, `data-theme="dark"`, "Must include dark mode data-theme support")
	assert.Contains(t, headContent, "#6777ef", "Must configure Stisla primary color #6777ef")
	assert.Contains(t, headContent, "X-CSRF-Token", "Must include CSRF fetch interceptor")

	// 2. Verify layout_sidebar.html
	sidebarBytes, err := os.ReadFile("views/admin/layout_sidebar.html")
	assert.NoError(t, err, "views/admin/layout_sidebar.html must exist")
	sidebarContent := string(sidebarBytes)
	assert.Contains(t, sidebarContent, `id="main-sidebar"`, "Must feature #main-sidebar")
	assert.Contains(t, sidebarContent, `id="sidebar-backdrop"`, "Must feature mobile #sidebar-backdrop")
	assert.Contains(t, sidebarContent, "/admin/dashboard", "Must link to /admin/dashboard")
	assert.Contains(t, sidebarContent, "/admin/schedules", "Must link to /admin/schedules")
	assert.Contains(t, sidebarContent, "/admin/audio", "Must link to /admin/audio")
	assert.Contains(t, sidebarContent, "/admin/attendance", "Must link to /admin/attendance")
	assert.Contains(t, sidebarContent, "/admin/academics", "Must link to /admin/academics")
	assert.Contains(t, sidebarContent, "/admin/points", "Must link to /admin/points")
	assert.Contains(t, sidebarContent, "/admin/settings", "Must link to /admin/settings")
	assert.Contains(t, sidebarContent, "/admin/reports", "Must link to /admin/reports")

	// 3. Verify layout_topbar.html
	topbarBytes, err := os.ReadFile("views/admin/layout_topbar.html")
	assert.NoError(t, err, "views/admin/layout_topbar.html must exist")
	topbarContent := string(topbarBytes)
	assert.Contains(t, topbarContent, `id="global-search-input"`, "Must feature #global-search-input")
	assert.Contains(t, topbarContent, "Ctrl+K", "Must feature Ctrl+K shortcut badge")
	assert.Contains(t, topbarContent, `id="server-clock"`, "Must feature #server-clock")
	assert.Contains(t, topbarContent, "data-stisla-theme-toggle", "Must feature theme toggle button")

	// 4. Verify layout_footer.html
	footerBytes, err := os.ReadFile("views/admin/layout_footer.html")
	assert.NoError(t, err, "views/admin/layout_footer.html must exist")
	footerContent := string(footerBytes)
	assert.Contains(t, footerContent, "toggleStislaTheme()", "Must implement toggleStislaTheme()")
	assert.Contains(t, footerContent, "toggleMobileSidebar()", "Must implement toggleMobileSidebar()")
	assert.Contains(t, footerContent, "openModal", "Must implement openModal()")
	assert.Contains(t, footerContent, "deleteItem", "Must implement deleteItem()")
	assert.Contains(t, footerContent, "previewAudio", "Must implement previewAudio()")
}

func TestAdminModularDomainPages_Content(t *testing.T) {
	pages := []struct {
		file     string
		keywords []string
	}{
		{
			file:     "views/admin/dashboard.html",
			keywords: []string{"Dashboard Operasional Sekolah", "PresentStudents", "AbsentStudents", "chartWeeklyClass", "chartStatus", "leaderboard-body"},
		},
		{
			file:     "views/admin/schedules.html",
			keywords: []string{"Manajemen Jadwal Bel", "modal-add-schedule", "previewAudio", "deleteItem"},
		},
		{
			file:     "views/admin/audio.html",
			keywords: []string{"Audio Library", "uploadAudioFile", "upload-progress-container", "renameAudio", "spotify-player-bar", "player-seek-slider", "player-volume-slider", "playSpotifyTrack"},
		},
		{
			file:     "views/admin/attendance.html",
			keywords: []string{"Kontrol Presensi Siswa", "table-absent", "table-present", "Presensi Sholat", "modal-att-settings"},
		},
		{
			file:     "views/admin/academics.html",
			keywords: []string{"Data Master Akademik", "switchAcademicTab", "modal-student", "modal-staff", "modal-class", "modal-major"},
		},
		{
			file:     "views/admin/points.html",
			keywords: []string{"Sistem Poin & Karakter Siswa", "komposit-leaderboard-body", "prestasi-rule-select", "violation-rule-select", "modal-add-point-rule"},
		},
		{
			file:     "views/admin/settings.html",
			keywords: []string{"Pengaturan Sistem & Integrasi", "onesender_api_url", "modal-add-device", "modal-add-holiday", "downloadBackup"},
		},
		{
			file:     "views/admin/reports.html",
			keywords: []string{"Pusat Laporan & English Quest", "reportPeriodType", "generateReport", "english-submissions-table"},
		},
	}

	for _, p := range pages {
		t.Run(p.file, func(t *testing.T) {
			bytes, err := os.ReadFile(p.file)
			assert.NoError(t, err, "File %s must exist", p.file)
			content := string(bytes)
			for _, kw := range p.keywords {
				assert.Contains(t, content, kw, "File %s must contain keyword %s", p.file, kw)
			}
		})
	}
}

func TestAdminModularTemplates_ParseAndRender(t *testing.T) {
	// Parse all views including views/admin/*.html using os.DirFS(".")
	tmpl, err := template.ParseFS(os.DirFS("."), "views/*.html", "views/admin/*.html")
	assert.NoError(t, err, "All views including views/admin/*.html must parse without template errors")

	// Dummy mock data for template execution
	mockData := map[string]interface{}{
		"Username": "TestAdmin",
		"Stats": struct {
			TotalSchedules   int
			NextBell         string
			OnlineDevices    int
			TotalDevices     int
			TotalStudents    int
			TotalStaff       int
			ActiveStudents   int
			InactiveStudents int
		}{
			TotalSchedules: 5,
			NextBell:       "07:00 Masuk",
			TotalStudents:  100,
			TotalStaff:     20,
		},
		"ChartWeeklyClass": `{"labels":[],"datasets":[]}`,
		"ChartStatus":      `{"labels":[],"datasets":[]}`,
		"ChartArrival":     `{"labels":[],"datasets":[]}`,
		"PresentStudents":     []interface{}{},
		"AbsentStudents":      []interface{}{},
		"StaffAttendanceList": []interface{}{},
		"StaffAttendanceStats": struct {
			TotalStaff   int
			PresentStaff int
			OnTimeStaff  int
			LateStaff    int
			AbsentStaff  int
		}{
			TotalStaff:   20,
			PresentStaff: 18,
			OnTimeStaff:  15,
			LateStaff:    3,
			AbsentStaff:  2,
		},
		"Students":  []interface{}{},
		"StaffList": []interface{}{},
		"Classes":         []interface{}{},
		"Majors":          []interface{}{},
		"Devices":         []interface{}{},
		"Holidays":        []interface{}{},
		"AttendanceSettings": map[string]string{
			"arrival_start":   "06:30",
			"arrival_end":     "07:15",
			"departure_start": "15:00",
			"departure_end":   "17:00",
			"dzuhur_start":    "12:00",
			"dzuhur_end":      "12:45",
			"ashar_start":     "15:30",
			"ashar_end":       "16:00",
		},
	}

	templatesToTest := []string{
		"dashboard.html",
		"schedules.html",
		"audio.html",
		"attendance.html",
		"academics.html",
		"points.html",
		"settings.html",
		"reports.html",
	}

	for _, tmplName := range templatesToTest {
		t.Run(tmplName, func(t *testing.T) {
			var buf bytes.Buffer
			err := tmpl.ExecuteTemplate(&buf, tmplName, mockData)
			assert.NoError(t, err, "Template %s must execute without error", tmplName)
			assert.NotEmpty(t, buf.String(), "Rendered output of %s must not be empty", tmplName)
		})
	}
}

func TestAdminCompositeScoresAPI(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed test student
	seedStudent(t, app, "1001", "Budi Santoso", "secret", true)

	e := echo.New()

	// 1. Test RecalculateAllCompositeScores
	{
		req := httptest.NewRequest(http.MethodPost, "/api/composite-scores/recalculate-all", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.RecalculateAllCompositeScores(app.DB)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
	}

	// 2. Test GetCompositeScores (Leaderboard)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/composite-scores/leaderboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.GetCompositeScores(app.DB)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
		assert.NotNil(t, resp["results"])
	}
}

func TestAdminPointsView_ScriptHandling(t *testing.T) {
	bytes, err := os.ReadFile("views/admin/points.html")
	assert.NoError(t, err)
	content := string(bytes)

	// Ensure points.html properly handles both raw array and object with results field
	assert.Contains(t, content, "data.results", "views/admin/points.html must safely unpack data.results")
	assert.Contains(t, content, "/api/composite-scores/leaderboard", "views/admin/points.html must query leaderboard API")
	assert.Contains(t, content, "/api/composite-scores/recalculate-all", "views/admin/points.html must call recalculate API")
}

func TestAdminDualPointDashboard_TemplateRender(t *testing.T) {
	dualPointViews := []struct {
		file     string
		expected string
	}{
		{"views/admin_dual_point_dashboard.html", "Sistem Poin Karakter"},
		{"views/admin_dual_point_input.html", "Input Poin Siswa"},
		{"views/admin_dual_point_leaderboard.html", "Papan Peringkat Siswa"},
		{"views/admin_dual_point_reports.html", "Laporan & Analitik Poin"},
		{"views/admin_dual_point_student_detail.html", "Detail & Rapor Poin Siswa"},
		{"views/admin_dual_point_rules.html", "Master Aturan Poin"},
	}

	for _, v := range dualPointViews {
		t.Run(v.file, func(t *testing.T) {
			tmpl, err := template.ParseFiles(v.file)
			assert.NoError(t, err, "%s should parse without html/template syntax or context errors", v.file)
			var buf bytes.Buffer
			err = tmpl.Execute(&buf, map[string]interface{}{})
			assert.NoError(t, err, "%s should execute without errors", v.file)
			assert.Contains(t, buf.String(), v.expected)
		})
	}
}

func TestAllViewsParse_HtmlTemplate(t *testing.T) {
	tmpl, err := template.ParseFS(os.DirFS("."), "views/*.html", "views/admin/*.html")
	assert.NoError(t, err, "All templates must parse cleanly")
	assert.NotNil(t, tmpl)

	viewsToExecute := []struct {
		name     string
		expected string
	}{
		{"admin_dual_point_dashboard.html", "Sistem Poin Karakter"},
		{"admin_dual_point_input.html", "Input Poin Siswa"},
		{"admin_dual_point_leaderboard.html", "Papan Peringkat Siswa"},
		{"admin_dual_point_reports.html", "Laporan & Analitik Poin"},
		{"admin_dual_point_student_detail.html", "Detail & Rapor Poin Siswa"},
		{"admin_dual_point_rules.html", "Master Aturan Poin"},
	}

	for _, v := range viewsToExecute {
		var buf bytes.Buffer
		err = tmpl.ExecuteTemplate(&buf, v.name, nil)
		assert.NoError(t, err, "Executing %s must not throw any context errors", v.name)
		assert.Contains(t, buf.String(), v.expected)
	}
}

func TestAdminCSRFFix_DeleteAndForms(t *testing.T) {
	e := echo.New()
	e.Use(appMiddleware.CSRF())

	// Route simulating delete schedule
	e.DELETE("/admin/schedule/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "deleted", "id": c.Param("id")})
	})

	// 1. Get valid CSRF token from safe GET method
	getReq := httptest.NewRequest(http.MethodGet, "/any-page", nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	var token string
	for _, cookie := range getRec.Result().Cookies() {
		if cookie.Name == "csrf_token" {
			token = cookie.Value
			break
		}
	}
	assert.NotEmpty(t, token, "GET request must issue a csrf_token cookie")
	assert.NotContains(t, token, "=", "Token generated with RawURLEncoding should not contain '=' padding")

	// 2. Perform DELETE request with cookie and header matching
	delReq := httptest.NewRequest(http.MethodDelete, "/admin/schedule/2", nil)
	delReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	delReq.Header.Set("X-CSRF-Token", token)
	delRec := httptest.NewRecorder()

	e.ServeHTTP(delRec, delReq)
	assert.Equal(t, http.StatusOK, delRec.Code, "DELETE /admin/schedule/2 with valid CSRF must succeed with 200 OK")

	// 3. Verify restart recovery / tolerance: if client sends a genuine 32+ char token matching cookie
	restartToken := "test-token-from-before-server-restart-12345678"
	recReq := httptest.NewRequest(http.MethodDelete, "/admin/schedule/3", nil)
	recReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: restartToken})
	recReq.Header.Set("X-CSRF-Token", restartToken)
	recRec := httptest.NewRecorder()

	e.ServeHTTP(recRec, recReq)
	assert.Equal(t, http.StatusOK, recRec.Code, "DELETE with matching cookie+header after restart must succeed with 200 OK")

	// 4. Verify templates use window.getCsrfToken without split('=')[1]
	headBytes, err := os.ReadFile("views/admin/layout_head.html")
	assert.NoError(t, err)
	headContent := string(headBytes)
	assert.Contains(t, headContent, "window.getCsrfToken", "layout_head.html must expose window.getCsrfToken")
	assert.NotContains(t, headContent, "?.split('=')[1]", "layout_head.html must NOT use buggy split('=')[1]")

	footerBytes, err := os.ReadFile("views/admin/layout_footer.html")
	assert.NoError(t, err)
	footerContent := string(footerBytes)
	assert.Contains(t, footerContent, "window.getCsrfToken", "layout_footer.html must use window.getCsrfToken")
}

func TestAdminSpotifyAudioPlayer_UI(t *testing.T) {
	bytes, err := os.ReadFile("views/admin/audio.html")
	assert.NoError(t, err, "views/admin/audio.html must exist")
	content := string(bytes)

	// 1. Verify Player Bar Container & Controls
	assert.Contains(t, content, `id="spotify-player-bar"`, "Must contain #spotify-player-bar floating container")
	assert.Contains(t, content, `id="player-track-title"`, "Must contain #player-track-title")
	assert.Contains(t, content, `id="player-track-subtitle"`, "Must contain #player-track-subtitle")
	assert.Contains(t, content, `id="player-equalizer"`, "Must contain animated soundwave equalizer")
	assert.Contains(t, content, `id="btn-play-pause"`, "Must contain play/pause toggle button")
	assert.Contains(t, content, `id="player-seek-slider"`, "Must contain timeline seek slider")
	assert.Contains(t, content, `id="player-volume-slider"`, "Must contain volume control slider")
	assert.Contains(t, content, `id="btn-playback-rate"`, "Must contain playback rate control button")

	// 2. Verify Table Track Integration
	assert.Contains(t, content, "data-track-index", "Table rows must include data-track-index")
	assert.Contains(t, content, "data-track-url", "Table rows must include data-track-url")
	assert.Contains(t, content, "playSpotifyTrack", "Play buttons must invoke playSpotifyTrack")

	// 3. Verify JavaScript Controller Methods
	assert.Contains(t, content, "playNextTrack", "Controller must support playNextTrack")
	assert.Contains(t, content, "playPrevTrack", "Controller must support playPrevTrack")
	assert.Contains(t, content, "toggleRepeat", "Controller must support toggleRepeat")
	assert.Contains(t, content, "toggleShuffle", "Controller must support toggleShuffle")
	assert.Contains(t, content, "toggleMute", "Controller must support toggleMute")
	assert.Contains(t, content, "closeSpotifyPlayer", "Controller must support closeSpotifyPlayer")
}

func TestAdminStudentProfileRoute(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO classes (id, name) VALUES (1, 'X RPL 1')")
	app.DB.Exec("INSERT INTO students (id, nis, name, class_id, parent_phone, rfid_uid, photo, status) VALUES (10, '1001', 'Budi Santoso', 1, '08123456789', 'rfid-1001', 'photo.jpg', 'active')")

	e := echo.New()
	tmpl, err := template.ParseFS(os.DirFS("."), "views/*.html", "views/admin/*.html")
	assert.NoError(t, err)

	e.Renderer = &testTemplateRenderer{templates: tmpl}

	// 1. Full data test
	req := httptest.NewRequest(http.MethodGet, "/admin/student/10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("10")

	h := handler.StudentProfile(app.DB)
	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Budi Santoso")
	assert.Contains(t, body, "X RPL 1")
	assert.Contains(t, body, "Cetak Kartu Siswa")

	// 2. Nullable data test (student with NULL fields)
	app.DB.Exec("INSERT INTO students (id, nis, name) VALUES (11, '1002', 'Siti Aminah')")
	reqNull := httptest.NewRequest(http.MethodGet, "/admin/student/11", nil)
	recNull := httptest.NewRecorder()
	cNull := e.NewContext(reqNull, recNull)
	cNull.SetParamNames("id")
	cNull.SetParamValues("11")

	err = h(cNull)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recNull.Code, "Must not redirect on NULL fields")
	assert.Contains(t, recNull.Body.String(), "Siti Aminah")
}

func TestAdminStudentIDCardRoute(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO classes (id, name) VALUES (1, 'X RPL 1')")
	app.DB.Exec("INSERT INTO students (id, nis, name, class_id, rfid_uid) VALUES (20, '2001', 'Joko Susilo', 1, 'rfid-2001')")

	e := echo.New()
	tmpl, err := template.ParseFS(os.DirFS("."), "views/*.html", "views/admin/*.html")
	assert.NoError(t, err)
	e.Renderer = &testTemplateRenderer{templates: tmpl}

	h := handler.GetStudentIDCard(app.DB)

	// 1. Browser request (Accept: text/html) -> returns HTML ID card page
	reqHTML := httptest.NewRequest(http.MethodGet, "/admin/student/idcard/20", nil)
	reqHTML.Header.Set("Accept", "text/html,application/xhtml+xml")
	recHTML := httptest.NewRecorder()
	cHTML := e.NewContext(reqHTML, recHTML)
	cHTML.SetParamNames("id")
	cHTML.SetParamValues("20")

	err = h(cHTML)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recHTML.Code)
	assert.Contains(t, recHTML.Header().Get("Content-Type"), "text/html")
	htmlBody := recHTML.Body.String()
	assert.Contains(t, htmlBody, "Kartu Siswa (ID Card)")
	assert.Contains(t, htmlBody, "Joko Susilo")
	assert.Contains(t, htmlBody, "Cetak Kartu")

	// 2. API request (format=json) -> returns raw JSON
	reqJSON := httptest.NewRequest(http.MethodGet, "/admin/student/idcard/20?format=json", nil)
	recJSON := httptest.NewRecorder()
	cJSON := e.NewContext(reqJSON, recJSON)
	cJSON.SetParamNames("id")
	cJSON.SetParamValues("20")

	err = h(cJSON)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recJSON.Code)
	assert.Contains(t, recJSON.Header().Get("Content-Type"), "application/json")
	var jsonMap map[string]interface{}
	err = json.Unmarshal(recJSON.Body.Bytes(), &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, "Joko Susilo", jsonMap["name"])
	assert.NotEmpty(t, jsonMap["qr_image"])
}

type testTemplateRenderer struct {
	templates *template.Template
}

func (t *testTemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// TestStudentDashboardAPIFields verifies the dashboard API returns all enriched fields the frontend needs.
func TestStudentDashboardAPIFields(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Insert a test student
	_, err := db.Exec(`INSERT INTO students (id, rfid_uid, nis, name, password, status) VALUES (999, 'RFID999', '9990001', 'Audit Siswa', '', 'active')`)
	if err != nil {
		// Student might already exist from a previous test run — ignore
		t.Logf("Student insert skipped: %v", err)
	}

	// Create a session for student 999
	token := "test-dashboard-token-999"
	db.Exec(`REPLACE INTO student_sessions (token, student_id, expires_at) VALUES (?, 999, datetime('now','+7 days'))`, token)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	req.Header.Set("Cookie", handler.StudentSessionCookie+"="+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("student_id", "999")

	h := handler.GetStudentDashboard(db)
	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	// Verify all fields the frontend expects are present
	assert.Contains(t, resp, "student", "dashboard must have student object")
	assert.Contains(t, resp, "today", "dashboard must have today object")
	assert.Contains(t, resp, "yesterday", "dashboard must have yesterday object")
	assert.Contains(t, resp, "points", "dashboard must have points")
	assert.Contains(t, resp, "weighted_score", "dashboard must have weighted_score")
	assert.Contains(t, resp, "pancawaluya_points", "dashboard must have pancawaluya_points")
	assert.Contains(t, resp, "violation_points", "dashboard must have violation_points")
	assert.Contains(t, resp, "academic_year", "dashboard must have academic_year")
	assert.Contains(t, resp, "month", "dashboard must have month stats")
	assert.Contains(t, resp, "point_history", "dashboard must have point_history array")
	assert.Contains(t, resp, "announcements", "dashboard must have announcements array")
	assert.Contains(t, resp, "using_default_pin", "dashboard must have using_default_pin")

	// academic_year must be non-empty string
	acy, _ := resp["academic_year"].(string)
	assert.NotEmpty(t, acy, "academic_year must not be empty")

	// Clean up
	db.Exec("DELETE FROM student_sessions WHERE token = ?", token)
	db.Exec("DELETE FROM students WHERE id = 999")
}

// TestStudentProfileUpdate verifies PUT /api/student/profile works correctly.
func TestStudentProfileUpdate(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Insert test student
	db.Exec(`INSERT OR REPLACE INTO students (id, rfid_uid, nis, name, password, status) VALUES (998, 'RFID998', '9980001', 'Profil Siswa', '', 'active')`)

	// Session
	token := "test-profile-update-998"
	db.Exec(`REPLACE INTO student_sessions (token, student_id, expires_at) VALUES (?, 998, datetime('now','+7 days'))`, token)

	e := echo.New()
	body := bytes.NewBufferString(`{"birthday":"2005-06-15","parent_name":"Budi","parent_phone":"081234567890"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/student/profile", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", handler.StudentSessionCookie+"="+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("student_id", "998")

	h := handler.UpdateStudentPortalProfile(db)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Contains(t, resp["message"], "berhasil")

	// Verify DB was updated
	var bday string
	db.QueryRow("SELECT COALESCE(birthday,'') FROM students WHERE id = 998").Scan(&bday)
	assert.Equal(t, "2005-06-15", bday)

	// Clean up
	db.Exec("DELETE FROM student_sessions WHERE token = ?", token)
	db.Exec("DELETE FROM students WHERE id = 998")
}

// TestStudentClaimsHistory verifies GET /api/student/claims returns 200 and an array without SQL errors.
func TestStudentClaimsHistory(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Insert test student & rule
	db.Exec(`INSERT OR REPLACE INTO students (id, rfid_uid, nis, name, password, status) VALUES (997, 'RFID997', '9970001', 'Siswa Klaim', '', 'active')`)
	db.Exec(`INSERT OR REPLACE INTO point_rules (id, category, name, points, description) VALUES (101, 'Prestasi', 'Juara 1 Lomba', 50, 'Deskripsi lomba')`)

	token := "test-claims-token-997"
	db.Exec(`REPLACE INTO student_sessions (token, student_id, expires_at) VALUES (?, 997, datetime('now','+7 days'))`, token)

	e := echo.New()

	// 1. Initial get: should return 200 and []
	req := httptest.NewRequest(http.MethodGet, "/api/student/claims", nil)
	req.Header.Set("Cookie", handler.StudentSessionCookie+"="+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("student_id", "997")

	h := handler.GetMyClaimsHistory(db)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var claims []map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &claims))
	assert.Empty(t, claims)

	// 2. Insert a claim into point_claims
	db.Exec(`INSERT INTO point_claims (student_id, rule_id, description, status) VALUES (997, 101, 'Mengikuti Olimpiade Matematika', 'pending')`)

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)
	c2.Set("student_id", "997")

	err = h(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var claims2 []map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &claims2))
	assert.Len(t, claims2, 1)
	assert.Equal(t, "Juara 1 Lomba", claims2[0]["rule_name"])
	assert.Equal(t, "pending", claims2[0]["status"])

	// Clean up
	db.Exec("DELETE FROM student_sessions WHERE token = ?", token)
	db.Exec("DELETE FROM students WHERE id = 997")
	db.Exec("DELETE FROM point_rules WHERE id = 101")
	db.Exec("DELETE FROM point_claims WHERE student_id = 997")
}

// TestResetStudentPointsComplete tests that resetting points removes all point records and English XP.
func TestResetStudentPointsComplete(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Seed student 995 with various points
	db.Exec(`INSERT OR REPLACE INTO students (id, nis, name, class_id, status) VALUES (995, '9950001', 'Siswa Reset', 1, 'active')`)
	db.Exec(`INSERT INTO student_points (student_id, points_change, description) VALUES (995, 10, 'Presensi')`)
	db.Exec(`INSERT INTO student_points (student_id, points_change, description) VALUES (995, 5, 'Bonus')`)
	db.Exec(`INSERT INTO student_achievement_points (student_id, points, description) VALUES (995, 20, 'Juara 1')`)
	db.Exec(`INSERT INTO student_violation_points (student_id, points, description) VALUES (995, 15, 'Terlambat')`)
	db.Exec(`INSERT INTO point_claims (student_id, rule_id, description, status) VALUES (995, 1, 'Klaim Mandiri', 'approved')`)
	db.Exec(`INSERT OR REPLACE INTO english_streaks (student_id, total_xp, current_streak) VALUES (995, 75, 5)`)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/points/reset/995", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("995")

	h := handler.ResetStudentPoints(db)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "success", resp["status"])

	// Verify all points are 0 or empty
	var spCount, achCount, violCount, claimCount, engXP int
	db.QueryRow("SELECT COUNT(*) FROM student_points WHERE student_id = 995").Scan(&spCount)
	db.QueryRow("SELECT COUNT(*) FROM student_achievement_points WHERE student_id = 995").Scan(&achCount)
	db.QueryRow("SELECT COUNT(*) FROM student_violation_points WHERE student_id = 995").Scan(&violCount)
	db.QueryRow("SELECT COUNT(*) FROM point_claims WHERE student_id = 995").Scan(&claimCount)
	db.QueryRow("SELECT COALESCE(total_xp, 0) FROM english_streaks WHERE student_id = 995").Scan(&engXP)

	assert.Equal(t, 0, spCount, "student_points should be 0")
	assert.Equal(t, 0, achCount, "achievement points should be 0")
	assert.Equal(t, 0, violCount, "violation points should be 0")
	assert.Equal(t, 0, claimCount, "claims should be 0")
	assert.Equal(t, 0, engXP, "English XP should be reset to 0")

	// Clean up
	db.Exec("DELETE FROM students WHERE id = 995")
}

// TestBulkResetPoints tests resetting all students in a class.
func TestBulkResetPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Seed class 99 and 2 students
	db.Exec(`INSERT OR REPLACE INTO classes (id, name) VALUES (99, 'Kelas Reset')`)
	db.Exec(`INSERT OR REPLACE INTO students (id, nis, name, class_id, status) VALUES (991, '9910001', 'Siswa 1', 99, 'active')`)
	db.Exec(`INSERT OR REPLACE INTO students (id, nis, name, class_id, status) VALUES (992, '9920001', 'Siswa 2', 99, 'active')`)
	db.Exec(`INSERT INTO student_points (student_id, points_change, description) VALUES (991, 25, 'Test 1')`)
	db.Exec(`INSERT INTO student_points (student_id, points_change, description) VALUES (992, 30, 'Test 2')`)

	e := echo.New()
	body := bytes.NewBufferString(`{"class_id": 99, "reset_by": "TestAdmin"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/points/reset-bulk", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.BulkResetPoints(db)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "success", resp["status"])

	var totalPointsLeft int
	db.QueryRow("SELECT COUNT(*) FROM student_points WHERE student_id IN (991, 992)").Scan(&totalPointsLeft)
	assert.Equal(t, 0, totalPointsLeft, "All points for class 99 should be deleted")

	// Clean up
	db.Exec("DELETE FROM classes WHERE id = 99")
	db.Exec("DELETE FROM students WHERE id IN (991, 992)")
}

// TestDualPointSystem_JSONAndFormDataBinding tests JSON & Form binding, composite auto-calculation, and deletion authorization.
func TestDualPointSystem_JSONAndFormDataBinding(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	db := app.DB

	// Seed test rules and student
	db.Exec(`INSERT OR REPLACE INTO achievement_rules (id, code, name, category, category_code, points, is_active)
		VALUES (901, 'ACH-TEST', 'Juara 1 Lomba Test', 'Akademik', 'AKAD', 25, 1)`)
	db.Exec(`INSERT OR REPLACE INTO violation_rules (id, code, name, category, category_code, points_1, points_2, points_3, is_active)
		VALUES (902, 'VIO-TEST', 'Terlambat Masuk Test', 'Disiplin', 'DIS', 10, 20, 30, 1)`)
	db.Exec(`INSERT OR REPLACE INTO students (id, nis, name, status) VALUES (998, '9980001', 'Siswa Dual Point', 'active')`)

	e := echo.New()

	// 1. Test AddAchievementPoint via JSON payload
	{
		jsonBody := `{"student_id": 998, "rule_id": 901, "description": "Lomba Sains Nasional"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/v2/student/achievement", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.AddAchievementPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(25), resp["points"])

		// Verify semester_scores was automatically updated
		var achPoints int
		db.QueryRow("SELECT achievement_points FROM semester_scores WHERE student_id = 998").Scan(&achPoints)
		assert.Equal(t, 25, achPoints, "semester_scores.achievement_points must be 25 after JSON insert")
	}

	// 2. Test AddAchievementPoint via FormValue
	{
		form := make(url.Values)
		form.Set("student_id", "998")
		form.Set("rule_id", "901")
		form.Set("description", "Lomba Kedua")
		req := httptest.NewRequest(http.MethodPost, "/admin/v2/student/achievement", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.AddAchievementPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify semester_scores updated to 50
		var achPoints int
		db.QueryRow("SELECT achievement_points FROM semester_scores WHERE student_id = 998").Scan(&achPoints)
		assert.Equal(t, 50, achPoints, "semester_scores.achievement_points must be 50 after second insert")
	}

	// 3. Test AddViolationPoint via JSON payload
	{
		jsonBody := `{"student_id": 998, "rule_id": 902, "occurrence": 1, "description": "Terlambat Upacara"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/v2/student/violation", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.AddViolationPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(10), resp["points"])

		// Verify semester_scores violation_points updated
		var violPoints int
		db.QueryRow("SELECT violation_points FROM semester_scores WHERE student_id = 998").Scan(&violPoints)
		assert.Equal(t, 10, violPoints, "semester_scores.violation_points must be 10 after JSON insert")
	}

	// 4. Test AddViolationPoint via FormValue
	{
		form := make(url.Values)
		form.Set("student_id", "998")
		form.Set("rule_id", "902")
		form.Set("occurrence", "2")
		req := httptest.NewRequest(http.MethodPost, "/admin/v2/student/violation", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler.AddViolationPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Occurrence 2 is 20 points, so total violation is 10 + 20 = 30
		var violPoints int
		db.QueryRow("SELECT violation_points FROM semester_scores WHERE student_id = 998").Scan(&violPoints)
		assert.Equal(t, 30, violPoints, "semester_scores.violation_points must be 30 after second insert")
	}

	// 5. Test DeleteAchievementPoint with Role check & Auto-recalculation
	{
		var achID int
		db.QueryRow("SELECT id FROM student_achievement_points WHERE student_id = 998 ORDER BY id DESC LIMIT 1").Scan(&achID)
		assert.True(t, achID > 0)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/v2/achievement/%d?reason=Salah+input", achID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(strconv.Itoa(achID))

		h := handler.DeleteAchievementPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify achievement points in semester_scores decreased from 50 to 25
		var achPoints int
		db.QueryRow("SELECT achievement_points FROM semester_scores WHERE student_id = 998").Scan(&achPoints)
		assert.Equal(t, 25, achPoints, "semester_scores.achievement_points must be 25 after deleting 1 record")
	}

	// 6. Test DeleteViolationPoint with Role check & Auto-recalculation
	{
		var vioID int
		db.QueryRow("SELECT id FROM student_violation_points WHERE student_id = 998 ORDER BY id DESC LIMIT 1").Scan(&vioID)
		assert.True(t, vioID > 0)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/v2/violation/%d?reason=Koreksi+pembina", vioID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(strconv.Itoa(vioID))

		h := handler.DeleteViolationPoint(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify violation points in semester_scores decreased from 30 to 10
		var violPoints int
		db.QueryRow("SELECT violation_points FROM semester_scores WHERE student_id = 998").Scan(&violPoints)
		assert.Equal(t, 10, violPoints, "semester_scores.violation_points must be 10 after deleting second violation")
	}

	// Clean up
	db.Exec("DELETE FROM students WHERE id = 998")
	db.Exec("DELETE FROM achievement_rules WHERE id = 901")
	db.Exec("DELETE FROM violation_rules WHERE id = 902")
	db.Exec("DELETE FROM student_achievement_points WHERE student_id = 998")
	db.Exec("DELETE FROM student_violation_points WHERE student_id = 998")
	db.Exec("DELETE FROM semester_scores WHERE student_id = 998")
}

func TestDualPointStudentDetail_Endpoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	db := app.DB

	// Seed test student
	db.Exec("INSERT INTO students (id, name, nis, status) VALUES (999, 'Test Student Detail', 'NIS999', 'active')")
	defer db.Exec("DELETE FROM students WHERE id = 999")

	e := echo.New()

	// Test 1: GetStudentDualPointProfile via path parameter /v2/student/:id/points
	{
		req := httptest.NewRequest(http.MethodGet, "/admin/v2/student/999/points", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")

		h := handler.GetStudentDualPointProfile(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Test Student Detail", resp["name"])
	}

	// Test 2: GetStudentDualPointProfile via query param /v2/student-detail?id=999
	{
		req := httptest.NewRequest(http.MethodGet, "/admin/v2/student-detail?id=999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		id := c.QueryParam("id")
		if id == "" {
			id = c.Param("id")
		}
		c.SetParamNames("id")
		c.SetParamValues(id)

		h := handler.GetStudentDualPointProfile(db)
		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Test Student Detail", resp["name"])
	}
}


