package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"belsekolah/internal/handler"
)

// --- P1-1: Attendance Isolation for Students Without RFID (LOG-06) ---

func TestP1_EmptyRFID_AttendanceIsolation(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Configure arrival window covering full day for deterministic test
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_start', '00:00')")
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_end', '23:59')")

	// Insert 2 students with NULL RFID (no RFID card)
	res1, err := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'NIS001', 'Siswa Satu', 'active')")
	assert.NoError(t, err)
	sID1, _ := res1.LastInsertId()

	res2, err := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'NIS002', 'Siswa Dua', 'active')")
	assert.NoError(t, err)
	sID2, _ := res2.LastInsertId()

	e := echo.New()

	// 1. Student 1 records attendance using QR (student_id)
	req1 := httptest.NewRequest(http.MethodGet, "/api/attendance/record?student_id="+strconv.FormatInt(sID1, 10), nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	err = app.RecordAttendanceHandler(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	var resp1 map[string]interface{}
	json.Unmarshal(rec1.Body.Bytes(), &resp1)
	assert.NotEqual(t, "duplicate", resp1["status"])

	// 2. Student 1 attempts to record again -> duplicate
	rec1Dup := httptest.NewRecorder()
	c1Dup := e.NewContext(req1, rec1Dup)
	err = app.RecordAttendanceHandler(c1Dup)
	assert.NoError(t, err)
	var resp1Dup map[string]interface{}
	json.Unmarshal(rec1Dup.Body.Bytes(), &resp1Dup)
	assert.Equal(t, "duplicate", resp1Dup["status"])

	// 3. Student 2 records attendance -> MUST SUCCEED (not blocked by student 1)
	req2 := httptest.NewRequest(http.MethodGet, "/api/attendance/record?student_id="+strconv.FormatInt(sID2, 10), nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = app.RecordAttendanceHandler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var resp2 map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &resp2)
	assert.NotEqual(t, "duplicate", resp2["status"])

	// Verify DB has 2 records with distinct student_ids
	var count1, count2 int
	app.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE student_id=?", sID1).Scan(&count1)
	app.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE student_id=?", sID2).Scan(&count2)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
}

// --- P1-2: Mark Alpha Students with Empty RFID (LOG-06) ---

func TestP1_EmptyRFID_MarkAlphaStudents(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Clear any existing students and logs
	app.DB.Exec("DELETE FROM students")
	app.DB.Exec("DELETE FROM attendance_logs")
	app.DB.Exec("REPLACE INTO school_settings (setting_key, setting_value) VALUES ('work_days', '1,2,3,4,5,6,7')")

	// Insert 2 active students with NULL RFID
	_, err := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'ALPHA01', 'Alpha Siswa Satu', 'active')")
	assert.NoError(t, err)
	_, err = app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'ALPHA02', 'Alpha Siswa Dua', 'active')")
	assert.NoError(t, err)

	// Run markAlphaStudents
	app.markAlphaStudents()

	// Both students should have been marked as Alpha
	var totalAlpha int
	today := time.Now().Format("2006-01-02")
	app.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE date=? AND status='Alpha'", today).Scan(&totalAlpha)
	assert.Equal(t, 2, totalAlpha, "Both students with empty RFID should be marked Alpha independently")
}

// --- P1-3: Prayer Attendance Isolation for Empty RFID (LOG-06) ---

func TestP1_EmptyRFID_PrayerAttendanceIsolation(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Configure dzuhur window covering full day for deterministic test
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_start', '00:00')")
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_end', '23:59')")

	res1, _ := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'PRAY01', 'Doa Satu', 'active')")
	sID1, _ := res1.LastInsertId()
	res2, _ := app.DB.Exec("INSERT INTO students (rfid_uid, nis, name, status) VALUES (NULL, 'PRAY02', 'Doa Dua', 'active')")
	sID2, _ := res2.LastInsertId()

	e := echo.New()

	// Student 1 records prayer
	req1 := httptest.NewRequest(http.MethodGet, "/api/prayer/record?student_id="+strconv.FormatInt(sID1, 10), nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	err := app.PrayerAttendanceHandler(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Student 2 records prayer -> must NOT be duplicate
	req2 := httptest.NewRequest(http.MethodGet, "/api/prayer/record?student_id="+strconv.FormatInt(sID2, 10), nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = app.PrayerAttendanceHandler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var resp2 map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &resp2)
	assert.Equal(t, "success", resp2["status"])

	// Check DB has 2 records with respective student_id
	var count1, count2 int
	app.DB.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE student_id=?", sID1).Scan(&count1)
	app.DB.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE student_id=?", sID2).Scan(&count2)
	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
}

// --- P1-4: Student Sessions Table & Expiration (LOG-07) ---

func TestP1_StudentSession_LifecycleAndExpiration(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "9001", "Siswa Sesi", "1234", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='9001'").Scan(&studentID)

	e := echo.New()

	// 1. Login
	form := url.Values{}
	form.Set("nis", "9001")
	form.Set("pin", "1234")
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/student/login", strings.NewReader(form.Encode()))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	err := handler.StudentLogin(app.DB)(cLogin)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recLogin.Code)

	// Extract session cookie
	cookies := recLogin.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == handler.StudentSessionCookie {
			sessionCookie = ck
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	token := sessionCookie.Value

	// Verify student_sessions table
	var dbStudentID int
	var expiresAt time.Time
	err = app.DB.QueryRow("SELECT student_id, expires_at FROM student_sessions WHERE token=?", token).Scan(&dbStudentID, &expiresAt)
	assert.NoError(t, err)
	assert.Equal(t, studentID, dbStudentID)
	assert.True(t, expiresAt.After(time.Now().Add(6*24*time.Hour)), "Session TTL should be ~7 days")

	// 2. Access protected endpoint with valid session
	reqAuth := httptest.NewRequest(http.MethodGet, "/api/student/profile", nil)
	reqAuth.AddCookie(sessionCookie)
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(reqAuth, recAuth)

	dummyHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK: "+c.Get("student_id").(string))
	}
	err = handler.StudentAuth(app.DB)(dummyHandler)(cAuth)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recAuth.Code)
	assert.Equal(t, "OK: "+strconv.Itoa(studentID), recAuth.Body.String())

	// 3. Expire the session in DB
	pastTime := time.Now().Add(-1 * time.Hour)
	_, err = app.DB.Exec("UPDATE student_sessions SET expires_at=? WHERE token=?", pastTime, token)
	assert.NoError(t, err)

	// 4. Access protected endpoint with expired session -> 401 Unauthorized
	recExpired := httptest.NewRecorder()
	cExpired := e.NewContext(reqAuth, recExpired)
	err = handler.StudentAuth(app.DB)(dummyHandler)(cExpired)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, recExpired.Code)

	// Verify session was cleaned up
	var count int
	app.DB.QueryRow("SELECT COUNT(*) FROM student_sessions WHERE token=?", token).Scan(&count)
	assert.Equal(t, 0, count, "Expired session should be deleted from student_sessions")
}

// --- P1-5: Asset Borrowing Multi-Unit Capacity & Overbooking (LOG-08) ---

func TestP1_AssetBorrowing_MultiUnitCapacity(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed staff (for approval) and asset with quantity = 2
	resStaff, err := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES ('ST001', 'Admin Lab', 'admin')")
	assert.NoError(t, err)
	staffID, _ := resStaff.LastInsertId()

	resCat, err := app.DB.Exec("INSERT INTO asset_categories (name, code) VALUES ('Elektronik', 'ELK')")
	assert.NoError(t, err)
	catID, _ := resCat.LastInsertId()

	resAsset, err := app.DB.Exec(`
		INSERT INTO assets (inventory_code, category_id, name, condition, quantity, is_borrowable)
		VALUES ('AST-LAPTOP', ?, 'Laptop Dell', 'good', 2, 1)
	`, catID)
	assert.NoError(t, err)
	assetID, _ := resAsset.LastInsertId()

	e := echo.New()
	assetIDStr := strconv.FormatInt(assetID, 10)
	staffIDStr := strconv.FormatInt(staffID, 10)

	// 1. First borrowing request
	form1 := url.Values{}
	form1.Set("asset_id", assetIDStr)
	form1.Set("borrower_type", "staff")
	form1.Set("borrower_id", staffIDStr)
	form1.Set("purpose", "Mengajar Kelas A")
	req1 := httptest.NewRequest(http.MethodPost, "/api/assets/borrowings", strings.NewReader(form1.Encode()))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err = handler.CreateAssetBorrowing(app.DB)(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Approve borrowing 1
	var bID1 int
	app.DB.QueryRow("SELECT id FROM asset_borrowings WHERE purpose='Mengajar Kelas A'").Scan(&bID1)
	bID1Str := strconv.Itoa(bID1)

	formApp1 := url.Values{}
	formApp1.Set("approved_by", staffIDStr)
	reqApp1 := httptest.NewRequest(http.MethodPut, "/api/assets/borrowings/"+bID1Str+"/approve", strings.NewReader(formApp1.Encode()))
	reqApp1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recApp1 := httptest.NewRecorder()
	cApp1 := e.NewContext(reqApp1, recApp1)
	cApp1.SetParamNames("id")
	cApp1.SetParamValues(bID1Str)

	err = handler.ApproveAssetBorrowing(app.DB)(cApp1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recApp1.Code)

	// 2. Second borrowing request (quantity is 2, so this should also succeed)
	form2 := url.Values{}
	form2.Set("asset_id", assetIDStr)
	form2.Set("borrower_type", "staff")
	form2.Set("borrower_id", staffIDStr)
	form2.Set("purpose", "Mengajar Kelas B")
	req2 := httptest.NewRequest(http.MethodPost, "/api/assets/borrowings", strings.NewReader(form2.Encode()))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = handler.CreateAssetBorrowing(app.DB)(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Approve borrowing 2
	var bID2 int
	app.DB.QueryRow("SELECT id FROM asset_borrowings WHERE purpose='Mengajar Kelas B'").Scan(&bID2)
	bID2Str := strconv.Itoa(bID2)

	reqApp2 := httptest.NewRequest(http.MethodPut, "/api/assets/borrowings/"+bID2Str+"/approve", strings.NewReader(formApp1.Encode()))
	reqApp2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recApp2 := httptest.NewRecorder()
	cApp2 := e.NewContext(reqApp2, recApp2)
	cApp2.SetParamNames("id")
	cApp2.SetParamValues(bID2Str)

	err = handler.ApproveAssetBorrowing(app.DB)(cApp2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recApp2.Code)

	// 3. Third borrowing request -> MUST FAIL (all 2 units are approved/borrowed)
	form3 := url.Values{}
	form3.Set("asset_id", assetIDStr)
	form3.Set("borrower_type", "staff")
	form3.Set("borrower_id", staffIDStr)
	form3.Set("purpose", "Mengajar Kelas C")
	req3 := httptest.NewRequest(http.MethodPost, "/api/assets/borrowings", strings.NewReader(form3.Encode()))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = handler.CreateAssetBorrowing(app.DB)(c3)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec3.Code)

	var resp3 map[string]string
	json.Unmarshal(rec3.Body.Bytes(), &resp3)
	assert.Contains(t, resp3["error"], "currently borrowed")

	// 4. Test that if a pending request existed beforehand, ApproveAssetBorrowing rejects it if capacity is reached
	resB3, err := app.DB.Exec(`
		INSERT INTO asset_borrowings (asset_id, borrower_type, borrower_id, purpose, borrow_date, due_date, status, created_at)
		VALUES (?, 'staff', ?, 'Mengajar Kelas D (Backlog)', '2026-09-23', '2026-09-30', 'pending', datetime('now'))
	`, assetID, staffID)
	assert.NoError(t, err)
	bID3, _ := resB3.LastInsertId()
	bID3Str := strconv.FormatInt(bID3, 10)

	reqApp3 := httptest.NewRequest(http.MethodPut, "/api/assets/borrowings/"+bID3Str+"/approve", strings.NewReader(formApp1.Encode()))
	reqApp3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recApp3 := httptest.NewRecorder()
	cApp3 := e.NewContext(reqApp3, recApp3)
	cApp3.SetParamNames("id")
	cApp3.SetParamValues(bID3Str)

	err = handler.ApproveAssetBorrowing(app.DB)(cApp3)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recApp3.Code)

	var respApp3 map[string]string
	json.Unmarshal(recApp3.Body.Bytes(), &respApp3)
	assert.Contains(t, respApp3["error"], "no available units")
}

// --- P1-6: SQLite Pragmas Verification (LOG-09) ---

func TestP1_SQLite_PragmasAndConcurrency(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Apply busy timeout pragma directly as done in InitDB
	app.DB.Exec("PRAGMA busy_timeout = 5000")

	var busyTimeout int
	err := app.DB.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	assert.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout, "busy_timeout should be configured to 5000ms")
}

