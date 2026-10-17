package tests

import (
	"bytes"
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
	"belsekolah/internal/repository"
)

// --- P2-1: Face Attendance Dynamic Late Threshold & Points (LOG-10 & LOG-11) ---

func TestP2_FaceAttendance_DynamicLateThreshold_And_Points(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "FACE01", "Siswa Wajah Hadir", "", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='FACE01'").Scan(&studentID)

	// Spin up mock face recognition server
	mockFaceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"matched":    true,
			"student_id": studentID,
			"name":       "Siswa Wajah Hadir",
			"class_name": "XII RPL 1",
			"distance":   0.25,
			"message":    "Match found",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockFaceServer.Close()

	// Configure face_service_url to mock server
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('face_service_url', ?)", mockFaceServer.URL)

	// Set arrival_end far in the future today so current time is considered "Datang"
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_end', '23:59')")

	e := echo.New()

	bodyJSON := []byte(`{"image_base64": "fake_image_data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", bytes.NewReader(bodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := app.VerifyFaceAttendanceHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "matched", resp["status"])
	assert.Equal(t, "Datang", resp["attendance_status"])

	// Verify points were recorded (+20 points)
	var pointsChange int
	err = app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='Terminal-Face'", studentID).Scan(&pointsChange)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, pointsChange, 20, "Should award at least 20 points for on-time arrival")
}

func TestP2_FaceAttendance_LateThreshold_And_Points(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "FACE02", "Siswa Wajah Terlambat", "", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='FACE02'").Scan(&studentID)

	// Mock face recognition server
	mockFaceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"matched":    true,
			"student_id": studentID,
			"name":       "Siswa Wajah Terlambat",
			"class_name": "XII RPL 1",
			"distance":   0.25,
			"message":    "Match found",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockFaceServer.Close()

	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('face_service_url', ?)", mockFaceServer.URL)

	// Set arrival_end in the past (00:01) so current time is strictly "Terlambat"
	app.DB.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_end', '00:01')")

	e := echo.New()

	bodyJSON := []byte(`{"image_base64": "fake_image_data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", bytes.NewReader(bodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := app.VerifyFaceAttendanceHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "matched", resp["status"])
	assert.Equal(t, "Terlambat", resp["attendance_status"])

	// Verify points were deducted (-10 points)
	var pointsChange int
	err = app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='Terminal-Face'", studentID).Scan(&pointsChange)
	assert.NoError(t, err)
	assert.Equal(t, -10, pointsChange, "Should deduct 10 points for late arrival")
}

// --- P2-2: Manual Attendance Point Mutations (LOG-10) ---

func TestP2_ManualAttendance_PointMutations(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "MAN01", "Siswa Manual Hadir", "", true)
	seedStudent(t, app, "MAN02", "Siswa Manual Terlambat", "", true)
	seedStudent(t, app, "MAN03", "Siswa Manual Alpha", "", true)

	var sID1, sID2, sID3 int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='MAN01'").Scan(&sID1)
	app.DB.QueryRow("SELECT id FROM students WHERE nis='MAN02'").Scan(&sID2)
	app.DB.QueryRow("SELECT id FROM students WHERE nis='MAN03'").Scan(&sID3)

	e := echo.New()

	// 1. Hadir -> +20
	form1 := url.Values{}
	form1.Set("student_id", strconv.Itoa(sID1))
	form1.Set("status", "Hadir")
	req1 := httptest.NewRequest(http.MethodPost, "/api/attendance/manual", strings.NewReader(form1.Encode()))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err := app.ManualAttendanceHandler(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	var pt1 int
	app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='Guru/Admin'", sID1).Scan(&pt1)
	assert.Equal(t, 20, pt1, "Manual Hadir should award +20 points")

	// 2. Terlambat -> -10
	form2 := url.Values{}
	form2.Set("student_id", strconv.Itoa(sID2))
	form2.Set("status", "Terlambat")
	req2 := httptest.NewRequest(http.MethodPost, "/api/attendance/manual", strings.NewReader(form2.Encode()))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = app.ManualAttendanceHandler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var pt2 int
	app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='Guru/Admin'", sID2).Scan(&pt2)
	assert.Equal(t, -10, pt2, "Manual Terlambat should deduct -10 points")

	// 3. Alpha -> -20
	form3 := url.Values{}
	form3.Set("student_id", strconv.Itoa(sID3))
	form3.Set("status", "Alpha")
	req3 := httptest.NewRequest(http.MethodPost, "/api/attendance/manual", strings.NewReader(form3.Encode()))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = app.ManualAttendanceHandler(c3)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec3.Code)

	var pt3 int
	app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='Guru/Admin'", sID3).Scan(&pt3)
	assert.Equal(t, -20, pt3, "Manual Alpha should deduct -20 points")
}

// --- P2-3: Bulk Attendance Point Mutations (LOG-10) ---

func TestP2_BulkAttendance_PointMutations(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "BULK01", "Siswa Bulk Hadir", "", true)
	seedStudent(t, app, "BULK02", "Siswa Bulk Alpha", "", true)

	var sID1, sID2 int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='BULK01'").Scan(&sID1)
	app.DB.QueryRow("SELECT id FROM students WHERE nis='BULK02'").Scan(&sID2)

	e := echo.New()

	today := time.Now().Format("2006-01-02")
	bulkReq := BulkAttendanceRequest{
		ClassID: 1,
		Date:    today,
		Items: []BulkAttendanceItem{
			{StudentID: sID1, Status: "Hadir"},
			{StudentID: sID2, Status: "Alpha"},
		},
	}

	bodyJSON, _ := json.Marshal(bulkReq)
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/bulk", bytes.NewReader(bodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := app.BulkAttendanceHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pt1, pt2 int
	app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='System-Bulk'", sID1).Scan(&pt1)
	app.DB.QueryRow("SELECT points_change FROM student_points WHERE student_id=? AND recorded_by='System-Bulk'", sID2).Scan(&pt2)

	assert.Equal(t, 20, pt1, "Bulk Hadir should award +20 points")
	assert.Equal(t, -20, pt2, "Bulk Alpha should deduct -20 points")
}

// --- P2-4: Achievement Score Normalization Target Points (LOG-12) ---

func TestP2_AchievementNormalization_TargetPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	repository.MigrationSemesterScores(app.DB)

	seedStudent(t, app, "ACH01", "Siswa Prestasi", "", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='ACH01'").Scan(&studentID)

	semesterID := 1

	// Case 1: Student earns 100 achievement points. Target is 200 points.
	// Expected normalized: (100 / 200) * 100 = 50.0
	_, err := app.DB.Exec(`
		INSERT INTO semester_scores (student_id, semester_id, achievement_points)
		VALUES (?, ?, 100)
		ON CONFLICT(student_id, semester_id) DO UPDATE SET achievement_points = 100
	`, studentID, semesterID)
	assert.NoError(t, err)

	score, err := handler.NormalizeAchievementPoints(app.DB, studentID, semesterID)
	assert.NoError(t, err)
	assert.InDelta(t, 50.0, score, 0.01, "100 achievement points should normalize to 50.0 out of 100 with target 200")

	// Case 2: Student earns 250 achievement points (exceeds target).
	// Expected normalized: capped at 100.0
	_, err = app.DB.Exec(`
		UPDATE semester_scores SET achievement_points = 250 WHERE student_id = ? AND semester_id = ?
	`, studentID, semesterID)
	assert.NoError(t, err)

	scoreCap, err := handler.NormalizeAchievementPoints(app.DB, studentID, semesterID)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, scoreCap, "Achievement score exceeding target should be capped at 100.0")
}

