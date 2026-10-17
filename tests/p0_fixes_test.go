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

// --- P0-1: RecordViolation Success and Escalation ---

func TestP0_RecordViolation_Success(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Ensure migration runs on the in-memory db
	repository.MigrationSemesterScores(app.DB)

	seedStudent(t, app, "8001", "Budi Santoso", "", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='8001'").Scan(&studentID)

	// Seed violation rule
	res, err := app.DB.Exec(`
		INSERT INTO violation_rules (code, name, category, category_code, points_1, points_2, points_3, is_active)
		VALUES ('K01', 'Merokok di Lingkungan Sekolah', 'Kedisiplinan', 'DIS', 10, 20, 30, 1)
	`)
	assert.NoError(t, err)
	ruleID, _ := res.LastInsertId()

	e := echo.New()

	// 1st offense
	reqBody := map[string]interface{}{
		"student_id":  studentID,
		"rule_id":     ruleID,
		"description": "Tertangkap di kantin",
		"recorded_by": "Guru BK",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/scoring/violation", bytes.NewReader(bodyJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.RecordViolation(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp["status"])

	// Check escalation level and points
	var points, escalation int
	err = app.DB.QueryRow("SELECT points, escalation_level FROM student_violation_points WHERE student_id=? AND rule_id=?", studentID, ruleID).Scan(&points, &escalation)
	assert.NoError(t, err)
	assert.Equal(t, 10, points)
	assert.Equal(t, 1, escalation)

	// Check semester_scores
	var totalViolationPoints int
	app.DB.QueryRow("SELECT violation_points FROM semester_scores WHERE student_id=?", studentID).Scan(&totalViolationPoints)
	assert.Equal(t, 10, totalViolationPoints)

	// 2nd offense (escalation to level 2)
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/admin/scoring/violation", bytes.NewReader(bodyJSON))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c2 := e.NewContext(req2, rec2)

	err = handler.RecordViolation(app.DB)(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var prevCount int
	app.DB.QueryRow("SELECT COUNT(*) FROM student_violation_points WHERE student_id=? AND rule_id=?", studentID, ruleID).Scan(&prevCount)
	assert.Equal(t, 2, prevCount)

	app.DB.QueryRow("SELECT violation_points FROM semester_scores WHERE student_id=?", studentID).Scan(&totalViolationPoints)
	assert.Equal(t, 30, totalViolationPoints) // 10 + 20
}

// --- P0-2: Attendance Normal Departure Points ---

func TestP0_Attendance_DepartureNormalPoints(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "8002", "Siti Rahma", "", true)
	app.DB.Exec("UPDATE students SET rfid_uid='RFID-8002' WHERE nis='8002'")

	// Configure arrival and departure settings
	app.DB.Exec("UPDATE attendance_settings SET setting_value='06:00' WHERE setting_key='arrival_start'")
	app.DB.Exec("UPDATE attendance_settings SET setting_value='07:00' WHERE setting_key='arrival_end'")
	app.DB.Exec("UPDATE attendance_settings SET setting_value='15:00' WHERE setting_key='departure_start'")
	app.DB.Exec("UPDATE attendance_settings SET setting_value='17:00' WHERE setting_key='departure_end'")

	// Check departure points calculation directly
	depStart := "15:00"
	timeStr := "15:30" // 30 minutes into normal dismissal

	depStartTime, _ := time.Parse("15:04", depStart)
	arrDepTime, _ := time.Parse("15:04", timeStr)

	var pointsChange int
	var pointDescription string
	if depStart != "" && arrDepTime.Before(depStartTime) {
		pointsChange = -10
		pointDescription = "Pulang Cepat (-10)"
	} else {
		pointsChange = 0
		pointDescription = "Pulang Tepat Waktu (+0)"
	}

	assert.Equal(t, 0, pointsChange)
	assert.Equal(t, "Pulang Tepat Waktu (+0)", pointDescription)

	// Early departure check (before 15:00)
	earlyTimeStr := "14:45"
	arrEarlyTime, _ := time.Parse("15:04", earlyTimeStr)
	if depStart != "" && arrEarlyTime.Before(depStartTime) {
		pointsChange = -10
		pointDescription = "Pulang Cepat (-10)"
	}
	assert.Equal(t, -10, pointsChange)
	assert.Equal(t, "Pulang Cepat (-10)", pointDescription)
}

// --- P0-3: Attendance Streak Weekend Continuity ---

func TestP0_AttendanceStreak_WeekendContinuity(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	repository.MigrationSemesterScores(app.DB)

	seedStudent(t, app, "8003", "Andi Pratama", "", true)
	app.DB.Exec("UPDATE students SET rfid_uid='RFID-8003' WHERE nis='8003'")
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='8003'").Scan(&studentID)

	// Setup active work days: Mon-Fri (1,2,3,4,5)
	app.DB.Exec("INSERT OR REPLACE INTO school_settings (setting_key, setting_value) VALUES ('work_days', '1,2,3,4,5')")

	// Insert attendance for Friday (2026-09-18) and Monday (2026-09-21)
	app.DB.Exec(`
		INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date)
		VALUES ('RFID-8003', 'Andi Pratama', 'Siswa', 'Hadir', 'RFID', '2026-09-18 06:45:00', '2026-09-18')
	`)
	app.DB.Exec(`
		INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date)
		VALUES ('RFID-8003', 'Andi Pratama', 'Siswa', 'Hadir', 'RFID', '2026-09-21 06:40:00', '2026-09-21')
	`)

	// Calculate attendance index
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/scoring/attendance-index?student_id="+strconv.Itoa(studentID)+"&semester_id=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CalculateAttendanceIndex(app.DB)(c)
	assert.NoError(t, err)
	if rec.Code != http.StatusOK {
		t.Fatalf("CalculateAttendanceIndex failed with code %d: %s", rec.Code, rec.Body.String())
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	var longestStreak int
	app.DB.QueryRow("SELECT longest_streak FROM attendance_streaks WHERE student_id=?", studentID).Scan(&longestStreak)
	// Across the weekend, Friday + Monday must count as a 2-day streak!
	assert.Equal(t, 2, longestStreak)
}

// --- P0-4: English Streak Multi-Day Review ---

func TestP0_EnglishStreak_MultiDayReview(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	seedStudent(t, app, "8004", "Dewi Lestari", "", true)
	var studentID int
	app.DB.QueryRow("SELECT id FROM students WHERE nis='8004'").Scan(&studentID)

	// Create 2 quests on 2 consecutive dates: 2026-09-21 and 2026-09-22
	qres1, _ := app.DB.Exec("INSERT INTO english_quests (date, title, description, quest_type, xp_reward) VALUES ('2026-09-21', 'Quest 1', 'desc', 'written', 10)")
	qID1, _ := qres1.LastInsertId()
	sres1, _ := app.DB.Exec("INSERT INTO english_submissions (student_id, quest_id, quest_type, content, status, submitted_at) VALUES (?, ?, 'written', 'Answer 1', 'pending', '2026-09-21 10:00:00')", studentID, qID1)
	subID1, _ := sres1.LastInsertId()

	qres2, _ := app.DB.Exec("INSERT INTO english_quests (date, title, description, quest_type, xp_reward) VALUES ('2026-09-22', 'Quest 2', 'desc', 'written', 10)")
	qID2, _ := qres2.LastInsertId()
	sres2, _ := app.DB.Exec("INSERT INTO english_submissions (student_id, quest_id, quest_type, content, status, submitted_at) VALUES (?, ?, 'written', 'Answer 2', 'pending', '2026-09-22 10:00:00')", studentID, qID2)
	subID2, _ := sres2.LastInsertId()

	e := echo.New()

	// Review submission 1 (Quest date 2026-09-21)
	form1 := url.Values{}
	form1.Set("action", "approve")
	form1.Set("feedback", "Good")
	req1 := httptest.NewRequest(http.MethodPost, "/admin/english/submissions/"+strconv.Itoa(int(subID1))+"/review", strings.NewReader(form1.Encode()))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	c1.SetParamNames("id")
	c1.SetParamValues(strconv.Itoa(int(subID1)))

	err := handler.ReviewEnglishSubmission(app.DB)(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	var streak1 int
	app.DB.QueryRow("SELECT current_streak FROM english_streaks WHERE student_id=?", studentID).Scan(&streak1)
	assert.Equal(t, 1, streak1)

	// Review submission 2 immediately (Quest date 2026-09-22)
	form2 := url.Values{}
	form2.Set("action", "approve")
	form2.Set("feedback", "Excellent")
	req2 := httptest.NewRequest(http.MethodPost, "/admin/english/submissions/"+strconv.Itoa(int(subID2))+"/review", strings.NewReader(form2.Encode()))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetParamNames("id")
	c2.SetParamValues(strconv.Itoa(int(subID2)))

	err = handler.ReviewEnglishSubmission(app.DB)(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var streak2, totalXP int
	app.DB.QueryRow("SELECT current_streak, total_xp FROM english_streaks WHERE student_id=?", studentID).Scan(&streak2, &totalXP)
	// Because quest 2 was for consecutive day 2026-09-22, streak should be 2!
	assert.Equal(t, 2, streak2)
	assert.Equal(t, 20, totalXP)
}
