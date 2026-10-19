package tests

import (
	"encoding/json"
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

func TestStaffAttendanceAndFaceRecognition(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Seed Staff Members (Male and Female)
	resMale, err := app.DB.Exec("INSERT INTO staff (rfid_uid, nip, name, role, gender) VALUES (?, ?, ?, ?, ?)",
		"RFID_MALE_01", "19800101", "Syamsul Budiman", "Guru PJOK", "L")
	assert.NoError(t, err)
	maleID, _ := resMale.LastInsertId()

	resFemale, err := app.DB.Exec("INSERT INTO staff (rfid_uid, nip, name, role, gender) VALUES (?, ?, ?, ?, ?)",
		"RFID_FEMALE_01", "19850202", "Neneng Ida Rosida", "Guru Bahasa Indonesia", "P")
	assert.NoError(t, err)
	femaleID, _ := resFemale.LastInsertId()

	// 2. Set attendance windows to include current time for Check-In
	now := time.Now()
	startTime := now.Add(-1 * time.Hour).Format("15:04")
	endTime := now.Add(1 * time.Hour).Format("15:04")
	depStart := now.Add(2 * time.Hour).Format("15:04")
	depEnd := now.Add(4 * time.Hour).Format("15:04")

	_, err = app.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_start', ?)", startTime)
	assert.NoError(t, err)
	_, err = app.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('arrival_end', ?)", endTime)
	assert.NoError(t, err)
	_, err = app.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('departure_start', ?)", depStart)
	assert.NoError(t, err)
	_, err = app.DB.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES ('departure_end', ?)", depEnd)
	assert.NoError(t, err)

	// 3. Test Face Attendance Verification for Male Staff (Check-In)
	{
		reqBody := map[string]interface{}{
			"staff_id":  maleID,
			"user_type": "Staff",
			"distance":  0.25,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", strings.NewReader(string(bodyBytes)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = app.VerifyFaceAttendanceHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "matched", resp["status"])
		assert.Equal(t, "Staff", resp["user_type"])
		assert.Equal(t, "Bapak", resp["honorific"])
		assert.Equal(t, "Selamat Datang Bapak Syamsul Budiman", resp["speech_message"])
		assert.Equal(t, "Hadir", resp["attendance_status"])
	}

	// 4. Test Duplicate Prevention for Male Staff
	{
		reqBody := map[string]interface{}{
			"staff_id":  maleID,
			"user_type": "Staff",
			"distance":  0.22,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", strings.NewReader(string(bodyBytes)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = app.VerifyFaceAttendanceHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "duplicate", resp["status"])
		assert.Contains(t, resp["message"], "presensi Hadir")
	}

	// 5. Test Departure (Check-Out) with Female Staff
	{
		// Seed earlier check-in for female staff (2 hours ago) so this scan records Pulang
		twoHoursAgo := now.Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
		app.DB.Exec("INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, 'RFID_FEMALE_01', 'Neneng Ida Rosida', 'Staff', 'Hadir', 'FACE', ?, ?)",
			femaleID, twoHoursAgo, now.Format("2006-01-02"))

		reqBody := map[string]interface{}{
			"staff_id":  femaleID,
			"user_type": "Staff",
			"distance":  0.30,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/attendance/verify-face", strings.NewReader(string(bodyBytes)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = app.VerifyFaceAttendanceHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "matched", resp["status"])
		assert.Equal(t, "Staff", resp["user_type"])
		assert.Equal(t, "Ibu", resp["honorific"])
		assert.Equal(t, "Terima kasih Ibu Neneng Ida Rosida", resp["speech_message"])
		assert.Equal(t, "Pulang", resp["attendance_status"])
	}

	// 6. Test Manual Staff Attendance
	{
		form := url.Values{}
		form.Set("staff_id", "1")
		form.Set("status", "Hadir")
		form.Set("time", "06:45")
		form.Set("note", "Hadir apel pagi")

		req := httptest.NewRequest(http.MethodPost, "/admin/attendance/staff-manual", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		manualHandler := handler.ManualStaffAttendance(app.DB)
		err = manualHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "success", resp["status"])
	}

	// 7. Test Export Staff Attendance (CSV)
	{
		req := httptest.NewRequest(http.MethodGet, "/admin/attendance/staff-export?date="+now.Format("2006-01-02"), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		exportHandler := handler.ExportStaffAttendance(app.DB)
		err = exportHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, rec.Body.String(), "Syamsul Budiman")
		assert.Contains(t, rec.Body.String(), "Neneng Ida Rosida")
	}

	// 8. Test Staff Face Deletion API
	{
		// Save face descriptor first
		saveReq := handler.SaveFaceDescriptorRequest{
			StaffID:    int(maleID),
			UserType:   "Staff",
			Descriptor: make([]float64, 128),
		}
		saveBody, _ := json.Marshal(saveReq)
		req := httptest.NewRequest(http.MethodPost, "/admin/face/save-descriptor", strings.NewReader(string(saveBody)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		saveHandler := handler.SaveFaceDescriptor(app.DB)
		err = saveHandler(c)
		assert.NoError(t, err)

		// Delete staff face
		reqDel := httptest.NewRequest(http.MethodDelete, "/admin/face/staff/1", nil)
		recDel := httptest.NewRecorder()
		cDel := e.NewContext(reqDel, recDel)
		cDel.SetParamNames("staff_id")
		cDel.SetParamValues(string(rune(maleID + '0')))

		delHandler := handler.DeleteStaffFace(app.DB)
		err = delHandler(cDel)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, recDel.Code)
	}
}
