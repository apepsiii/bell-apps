package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"belsekolah/internal/handler"
)

// ============================================================
// ADMIN FEATURE TESTS
// Tests for features documented in FEATURE.md section A
// ============================================================

// --- A2. Schedule (Bell Schedule) ---

func TestAddSchedule_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("time", "07:30")
	form.Set("label", "Upacara")
	form.Set("audio_file", "bell-upacara.mp3")

	req := httptest.NewRequest(http.MethodPost, "/admin/schedule/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddSchedule(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "success", body["status"])
}

func TestGetSchedules_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES ('07:00', 'Masuk', 'bell.mp3')")
	app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES ('15:00', 'Pulang', 'bell.mp3')")

	req := httptest.NewRequest(http.MethodGet, "/api/schedules", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := app.SyncHandler(c)
	assert.NoError(t, err)
}

func TestUpdateSchedule_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES ('07:00', 'Masuk', 'bell.mp3')")
	id, _ := res.LastInsertId()

	e := echo.New()
	form := url.Values{}
	form.Set("time", "07:15")
	form.Set("label", "Upacara Masuk")
	form.Set("audio_file", "bell-upacara.mp3")

	req := httptest.NewRequest(http.MethodPost, "/admin/schedule/update/"+itoa(id), strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.UpdateSchedule(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var name string
	app.DB.QueryRow("SELECT label FROM schedules WHERE id=?", id).Scan(&name)
	assert.Equal(t, "Upacara Masuk", name)
}

func TestDeleteSchedule_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES ('07:00', 'Masuk', 'bell.mp3')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/schedule/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeleteSchedule(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var count int
	app.DB.QueryRow("SELECT COUNT(*) FROM schedules WHERE id=?", id).Scan(&count)
	assert.Equal(t, 0, count)
}

// --- A4. Major ---

func TestAddMajor_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("name", "Teknik Komputer Jaringan")

	req := httptest.NewRequest(http.MethodPost, "/admin/major/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddMajor(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var name string
	app.DB.QueryRow("SELECT name FROM majors WHERE name='Teknik Komputer Jaringan'").Scan(&name)
	assert.Equal(t, "Teknik Komputer Jaringan", name)
}

func TestDeleteMajor_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO majors (name) VALUES ('TKR')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/major/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeleteMajor(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A5. Class ---

func TestAddClass_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO majors (name) VALUES ('TKJ')")
	majorID, _ := res.LastInsertId()

	e := echo.New()
	form := url.Values{}
	form.Set("name", "X-TKJ-1")
	form.Set("major_id", itoa(majorID))
	form.Set("wa_group_id", "group-123@s.whatsapp.net")

	req := httptest.NewRequest(http.MethodPost, "/admin/class/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddClass(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var waGroupID string
	app.DB.QueryRow("SELECT wa_group_id FROM classes WHERE name='X-TKJ-1'").Scan(&waGroupID)
	assert.Equal(t, "group-123@s.whatsapp.net", waGroupID)
}

func TestGetClassesJSON_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES ('X-TKJ-1', 1, '')")
	app.DB.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES ('XI-TKJ-2', 1, '')")

	req := httptest.NewRequest(http.MethodGet, "/admin/classes/json", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetClassesJSON(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A6. Student CRUD ---

func TestAddStudent_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("rfid_uid", "RFID-1001")
	form.Set("nis", "1001")
	form.Set("nis_siswa", "2024001")
	form.Set("name", "Andi Wijaya")
	form.Set("parent_name", "Bapak Wijaya")
	form.Set("parent_phone", "081234567890")
	form.Set("class_id", "1")
	form.Set("birthday", "2008-05-15")
	form.Set("status", "active")
	form.Set("password", "1234")

	req := httptest.NewRequest(http.MethodPost, "/admin/student/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddStudent(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var parentName string
	app.DB.QueryRow("SELECT parent_name FROM students WHERE nis='1001'").Scan(&parentName)
	assert.Equal(t, "Bapak Wijaya", parentName)
}

func TestAddStudent_DuplicateRFID_Fails(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES ('RFID-1001', '1001', 'Andi', 'active')")

	e := echo.New()
	form := url.Values{}
	form.Set("rfid_uid", "RFID-1001")
	form.Set("nis", "1002")
	form.Set("name", "Budi")
	form.Set("parent_phone", "081234567890")
	form.Set("class_id", "1")

	req := httptest.NewRequest(http.MethodPost, "/admin/student/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler.AddStudent(app.DB)(c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- A9. Attendance ---

func TestManualAttendance_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "7001", "Cici", "", true)
	// get student id
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='7001'").Scan(&studentID)

	e := echo.New()
	form := url.Values{}
	form.Set("student_id", itoa(int64(studentID)))
	form.Set("status", "Hadir")

	req := httptest.NewRequest(http.MethodPost, "/admin/attendance/manual", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := app.ManualAttendanceHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A14. Points System ---

func TestGetPointRules_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/admin/point-rules", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetPointRules(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rules []handler.PointRule
	json.Unmarshal(rec.Body.Bytes(), &rules)
	assert.Greater(t, len(rules), 0) // SeedPointRules pre-populates rules
}

func TestAddPointRule_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("category", "Prestasi")
	form.Set("name", "Juara Lomba")
	form.Set("points", "50")
	form.Set("description", "Juara lomba tingkat kota")

	req := httptest.NewRequest(http.MethodPost, "/admin/point-rules/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddPointRule(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeletePointRule_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO point_rules (category, name, points, description) VALUES ('Prestasi', 'Juara', 50, 'desc')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/point-rules/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeletePointRule(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAddPointReward_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("name", "Voucher Makan")
	form.Set("points_cost", "100")
	form.Set("stock", "50")
	form.Set("description", "Voucher makan kantin")

	req := httptest.NewRequest(http.MethodPost, "/admin/point-rewards/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddPointReward(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var stock int
	app.DB.QueryRow("SELECT stock FROM point_rewards WHERE name='Voucher Makan'").Scan(&stock)
	assert.Equal(t, 50, stock)
}

func TestGetLeaderboard_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "8001", "Andi", "", true)
	seedStudent(t, app, "8002", "Budi", "", true)

	app.DB.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (1, 50, 'Juara', datetime('now'), 'admin')")
	app.DB.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (2, 30, 'Juara 2', datetime('now'), 'admin')")

	req := httptest.NewRequest(http.MethodGet, "/admin/points/leaderboard", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetLeaderboard(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var items []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &items)
	assert.Greater(t, len(items), 0)
}

// --- A15. Holidays ---

func TestAddHoliday_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	holidayJSON := `{"date":"2026-08-17","name":"Hari Kemerdekaan","type":"Libur Nasional","description":"HUT RI"}`

	req := httptest.NewRequest(http.MethodPost, "/admin/holiday/add", strings.NewReader(holidayJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.AddHoliday(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAddHoliday_DuplicateDate_Fails(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES ('2026-08-17', 'HUT RI', 'Libur Nasional', 'desc')")

	holidayJSON := `{"date":"2026-08-17","name":"Another Holiday","type":"Libur Nasional","description":"desc"}`

	req := httptest.NewRequest(http.MethodPost, "/admin/holiday/add", strings.NewReader(holidayJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler.AddHoliday(app.DB)(c)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAddHoliday_MissingFields_Fails(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	holidayJSON := `{"date":"2026-08-17","name":""}`

	req := httptest.NewRequest(http.MethodPost, "/admin/holiday/add", strings.NewReader(holidayJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler.AddHoliday(app.DB)(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetHolidays_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES ('2026-08-17', 'HUT RI', 'Libur Nasional', 'desc')")
	app.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES ('2026-12-25', 'Natal', 'Libur Nasional', 'desc')")

	req := httptest.NewRequest(http.MethodGet, "/admin/holidays?year=2026", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetHolidays(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var holidays []handler.Holiday
	json.Unmarshal(rec.Body.Bytes(), &holidays)
	assert.Equal(t, 2, len(holidays))
}

func TestDeleteHoliday_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO holidays (date, name, type, description) VALUES ('2026-08-17', 'HUT RI', 'Libur Nasional', 'desc')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/holiday/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeleteHoliday(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A16. Devices ---

func TestAddDevice_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	form := url.Values{}
	form.Set("name", "Scanner Gerbang")
	form.Set("ip_address", "192.168.1.100")

	req := httptest.NewRequest(http.MethodPost, "/admin/device/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddDevice(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var status string
	app.DB.QueryRow("SELECT status FROM devices WHERE name='Scanner Gerbang'").Scan(&status)
	assert.Equal(t, "online", status)
}

func TestDeleteDevice_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	res, _ := app.DB.Exec("INSERT INTO devices (name, ip_address, status) VALUES ('Scanner', '192.168.1.1', 'online')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/device/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeleteDevice(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A11. Announcements ---

func TestGetAnnouncements_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO announcements (title, message, status) VALUES ('Test', 'Pesan', 'Selesai')")

	req := httptest.NewRequest(http.MethodGet, "/admin/announcements", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetAnnouncements(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteAnnouncement_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Insert with non-null audio_file to avoid NULL scan issue
	res, _ := app.DB.Exec("INSERT INTO announcements (title, message, status, audio_file) VALUES ('Test', 'Pesan', 'Selesai', '')")
	id, _ := res.LastInsertId()

	req := httptest.NewRequest(http.MethodDelete, "/admin/announcement/"+itoa(id), nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(id))

	err := handler.DeleteAnnouncement(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- A12. English Daily Quest ---

func TestGetEnglishLeaderboard_Empty_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/student/english/leaderboard", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetEnglishLeaderboard(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var items []interface{}
	json.Unmarshal(rec.Body.Bytes(), &items)
	assert.Equal(t, 0, len(items))
}

// --- A19. School Settings ---

func TestGetSchoolSettings_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	app.DB.Exec("INSERT INTO school_settings (setting_key, setting_value) VALUES ('active_days', '1,2,3,4,5')")

	req := httptest.NewRequest(http.MethodGet, "/admin/settings/school", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetSchoolSettings(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var settings map[string]string
	json.Unmarshal(rec.Body.Bytes(), &settings)
	assert.Equal(t, "1,2,3,4,5", settings["active_days"])
}

// ============================================================
// STUDENT PORTAL FEATURE TESTS (additional)
// Tests for features documented in FEATURE.md section B
// ============================================================

// --- B1/B3. Auth ---

func TestStudentAuth_NoCookie_Unauthorized(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	mw := handler.StudentAuth(app.DB)
	mw(func(c echo.Context) error { return nil })(c)

	// c.JSON returns nil error but sets the status code
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestStudentAuth_PageRedirect_NoCookie(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/student/app", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	mw := handler.StudentAuth(app.DB)
	mw(func(c echo.Context) error { return nil })(c)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- B4. Dashboard: Yesterday attendance for insight ---

func TestStudentDashboard_YesterdayData(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6001", "Dodi", "1234", true)

	// Insert yesterday's attendance (use real date format)
	yesterdayDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	yesterdayTS := yesterdayDate + " 07:15:00"
	app.DB.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, timestamp, date, method) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"rfid-6001", "Dodi", "Siswa", "Hadir", yesterdayTS, yesterdayDate, "RFID")

	// Login to get session
	form := url.Values{}
	form.Set("nis", "6001")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)

	cookies := rec.Header()["Set-Cookie"]
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
	assert.NotNil(t, body["yesterday"])
	yesterdayData := body["yesterday"].(map[string]interface{})
	assert.Equal(t, "Hadir", yesterdayData["status"])
	assert.Equal(t, "07:15", yesterdayData["time_in"])
}

// --- B4e. Dashboard: Today attendance with all statuses ---

func TestStudentDashboard_TodaySickStatus(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6002", "Eka", "1234", true)

	today := time.Now().Format("2006-01-02")
	todayTS := today + " 08:00:00"
	app.DB.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, timestamp, date, method) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"rfid-6002", "Eka", "Siswa", "Sakit", todayTS, today, "Manual")

	// Login
	form := url.Values{}
	form.Set("nis", "6002")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)

	cookies := rec.Header()["Set-Cookie"]
	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetStudentDashboard(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	mw(func(c echo.Context) error { return handlerFunc(c) })(c2)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	todayData := body["today"].(map[string]interface{})
	assert.Equal(t, "Sakit", todayData["status"])
	assert.Equal(t, "08:00", todayData["time_in"])
}

func TestStudentDashboard_TodayIzinStatus(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6003", "Fani", "1234", true)

	today := time.Now().Format("2006-01-02")
	todayTS := today + " 08:30:00"
	app.DB.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, timestamp, date, method) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"rfid-6003", "Fani", "Siswa", "Izin", todayTS, today, "Manual")

	form := url.Values{}
	form.Set("nis", "6003")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)

	cookies := rec.Header()["Set-Cookie"]
	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetStudentDashboard(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	mw(func(c echo.Context) error { return handlerFunc(c) })(c2)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	todayData := body["today"].(map[string]interface{})
	assert.Equal(t, "Izin", todayData["status"])
}

// --- B7a. Change PIN ---

func TestChangeStudentPIN_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6004", "Gita", "1234", true)

	// Login to get session
	form := url.Values{}
	form.Set("nis", "6004")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.ChangeStudentPIN(app.DB)

	// ChangeStudentPIN expects JSON body with old_pin and new_pin
	pinJSON := `{"old_pin":"1234","new_pin":"9999"}`

	req2 := httptest.NewRequest(http.MethodPut, "/api/student/pin", strings.NewReader(pinJSON))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestChangeStudentPIN_WrongOldPIN_Fails(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6005", "Hadi", "1234", true)

	form := url.Values{}
	form.Set("nis", "6005")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.ChangeStudentPIN(app.DB)

	pinJSON := `{"old_pin":"0000","new_pin":"9999"}`

	req2 := httptest.NewRequest(http.MethodPut, "/api/student/pin", strings.NewReader(pinJSON))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	mw(func(c echo.Context) error { return handlerFunc(c) })(c2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

// --- B11. English Daily Quest ---

func TestGetMyEnglishProfile_Empty_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "6006", "Ika", "1234", true)

	form := url.Values{}
	form.Set("nis", "6006")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetMyEnglishProfile(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/english/profile", nil)
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
	assert.Equal(t, false, body["today_done"])
}

// --- SearchStudent: flexible search by NIS/name/RFID ---

func TestSearchStudent_ByName(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9101", "Budi Santoso", "", true)
	seedStudent(t, app, "9102", "Budi Hartono", "", true)

	req := httptest.NewRequest(http.MethodGet, "/admin/points/search-student?q=Budi", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.SearchStudent(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var results []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &results)
	assert.Equal(t, 2, len(results))
}

func TestSearchStudent_ByNIS(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9201", "Cici", "", true)
	seedStudent(t, app, "9202", "Dedi", "", true)

	req := httptest.NewRequest(http.MethodGet, "/admin/points/search-student?q=9201", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.SearchStudent(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var results []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &results)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Cici", results[0]["name"])
}

func TestSearchStudent_EmptyQuery(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/admin/points/search-student?q=", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler.SearchStudent(app.DB)(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSearchStudent_NoResults(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9301", "Eka", "", true)

	req := httptest.NewRequest(http.MethodGet, "/admin/points/search-student?q=NonexistentName", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.SearchStudent(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var results []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &results)
	assert.Equal(t, 0, len(results))
}

// --- Helper: int to string ---

func itoa(i int64) string {
	return fmt.Sprintf("%d", i)
}

// ============================================================
// POINTS SYSTEM BUG-FIX TESTS
// Tests for the unified points system (FEATURE.md A14 + B11 merge)
// ============================================================

// --- GetPointRulesV2: must not 500 on the real schema ---

func TestGetPointRulesV2_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/point-rules", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetPointRulesV2(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rules []handler.PointClaimRule
	json.Unmarshal(rec.Body.Bytes(), &rules)
	assert.Greater(t, len(rules), 0)
	// Seeded rules carry "[CODE | TIER] desc" in description; code/tier must be parsed out.
	assert.NotEmpty(t, rules[0].Code)
	assert.NotEmpty(t, rules[0].Tier)
}

// --- GetStudentByRFID: must return current point balance ---

func TestGetStudentByRFID_ReturnsPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9001", "Joko", "", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9001'").Scan(&studentID)

	app.DB.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (?, 25, 'Juara', datetime('now'), 'admin')", studentID)
	app.DB.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (?, -5, 'Terlambat', datetime('now'), 'admin')", studentID)

	req := httptest.NewRequest(http.MethodGet, "/api/student-by-rfid?rfid=rfid-9001", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetStudentByRFID(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "Joko", body["name"])
	assert.Equal(t, float64(20), body["points"]) // 25 - 5 = 20
}

func TestGetStudentByRFID_NotFound(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/student-by-rfid?rfid=nonexistent", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler.GetStudentByRFID(app.DB)(c)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetStudentByRFID_MissingParam(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/student-by-rfid", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler.GetStudentByRFID(app.DB)(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- AddPointTransaction: must record the rule's point value ---

func TestAddPointTransaction_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9002", "Kiki", "", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9002'").Scan(&studentID)

	res, _ := app.DB.Exec("INSERT INTO point_rules (category, name, points, description) VALUES ('Prestasi', 'Juara Lomba', 50, 'desc')")
	ruleID, _ := res.LastInsertId()

	e := echo.New()
	form := url.Values{}
	form.Set("student_id", itoa(int64(studentID)))
	form.Set("rule_id", itoa(ruleID))

	req := httptest.NewRequest(http.MethodPost, "/admin/points/transaction", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddPointTransaction(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var total int
	app.DB.QueryRow("SELECT SUM(points_change) FROM student_points WHERE student_id=?", studentID).Scan(&total)
	assert.Equal(t, 50, total)
}

func TestAddPointTransaction_InvalidRule(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9003", "Lili", "", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9003'").Scan(&studentID)

	e := echo.New()
	form := url.Values{}
	form.Set("student_id", itoa(int64(studentID)))
	form.Set("rule_id", "999999") // nonexistent rule

	req := httptest.NewRequest(http.MethodPost, "/admin/points/transaction", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler.AddPointTransaction(app.DB)(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Unified Leaderboard: English XP must flow into total ---

func TestGetLeaderboard_IncludesEnglishXP(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9004", "Mimi", "", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9004'").Scan(&studentID)

	// Give the student English XP directly (simulating an approved submission).
	app.DB.Exec("INSERT INTO english_streaks (student_id, current_streak, longest_streak, total_xp, last_submit_date) VALUES (?, 1, 1, 30, date('now'))", studentID)

	req := httptest.NewRequest(http.MethodGet, "/admin/points/leaderboard", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler.GetLeaderboard(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var items []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &items)
	found := false
	for _, item := range items {
		if int(item["id"].(float64)) == studentID {
			assert.Equal(t, float64(30), item["points"])
			found = true
			break
		}
	}
	assert.True(t, found, "student with english XP should appear in leaderboard")
}

// --- English submission approval mirrors XP into student_points ---

func TestReviewEnglishSubmission_MirrorsXPToPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9005", "Nana", "", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9005'").Scan(&studentID)

	// Create a quest + a pending submission for it.
	qres, _ := app.DB.Exec("INSERT INTO english_quests (date, title, description, quest_type, xp_reward) VALUES (date('now'), 'Write a diary', 'desc', 'written', 15)")
	questID, _ := qres.LastInsertId()

	sres, _ := app.DB.Exec("INSERT INTO english_submissions (student_id, quest_id, quest_type, content, status, submitted_at) VALUES (?, ?, 'written', 'Today was good.', 'pending', datetime('now'))", studentID, questID)
	subID, _ := sres.LastInsertId()

	e := echo.New()
	form := url.Values{}
	form.Set("action", "approve")
	form.Set("feedback", "Bagus!")

	req := httptest.NewRequest(http.MethodPost, "/admin/english/submissions/"+itoa(subID)+"/review", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(itoa(subID))

	err := handler.ReviewEnglishSubmission(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// The XP should now be mirrored into student_points.
	var mirrored int
	app.DB.QueryRow("SELECT COALESCE(SUM(points_change),0) FROM student_points WHERE student_id=? AND recorded_by='english-quest'", studentID).Scan(&mirrored)
	assert.Equal(t, 15, mirrored)
}

// --- Student dashboard shows unified points (student_points + english XP) ---

func TestStudentDashboard_UnifiedPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()
	seedStudent(t, app, "9006", "Omar", "1234", true)

	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9006'").Scan(&studentID)

	// 20 points from student_points, 15 XP from english streaks.
	app.DB.Exec("INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by) VALUES (?, 20, 'Juara', datetime('now'), 'admin')", studentID)
	app.DB.Exec("INSERT INTO english_streaks (student_id, current_streak, longest_streak, total_xp, last_submit_date) VALUES (?, 1, 1, 15, date('now'))", studentID)

	form := url.Values{}
	form.Set("nis", "9006")
	form.Set("pin", "1234")
	req := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	handler.StudentLogin(app.DB)(c)
	cookies := rec.Header()["Set-Cookie"]

	mw := handler.StudentAuth(app.DB)
	handlerFunc := handler.GetStudentDashboard(app.DB)

	req2 := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	for _, ck := range cookies {
		req2.Header.Add("Cookie", strings.Split(ck, ";")[0])
	}
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	mw(func(c echo.Context) error { return handlerFunc(c) })(c2)

	var body map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &body)
	// 20 (student_points) + 15 (english XP) = 35
	assert.Equal(t, float64(35), body["points"])
}

// --- parseRuleCodeTier helper ---

func TestParseRuleCodeTier(t *testing.T) {
	code, tier := handler.ParseRuleCodeTier("[A.1.1.1 | Dasar] Shalat Dhuhur/Ashar")
	assert.Equal(t, "A.1.1.1", code)
	assert.Equal(t, "Dasar", tier)

	code2, tier2 := handler.ParseRuleCodeTier("Just a plain description")
	assert.Equal(t, "", code2)
	assert.Equal(t, "", tier2)

	code3, tier3 := handler.ParseRuleCodeTier("[OnlyCode] description")
	assert.Equal(t, "OnlyCode", code3)
	assert.Equal(t, "", tier3)
}
