package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/handler"
)

func TestRequestMagicLogin(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/magic-request", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.RequestMagicLogin(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("RequestMagicLogin returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res["status"] != "success" {
		t.Errorf("expected status 'success', got %v", res["status"])
	}

	code, ok := res["request_code"].(string)
	if !ok || len(code) != 6 {
		t.Errorf("expected 6-char request_code, got %v", res["request_code"])
	}

	// Verify token exists in database
	var count int
	err := app.DB.QueryRow("SELECT COUNT(*) FROM magic_login_tokens WHERE request_code = ? AND status = 'pending'", code).Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("expected token in db, count=%d, err=%v", count, err)
	}
}

func TestCheckMagicStatus(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// 1. Insert a pending token
	token := "test-token-check-123456"
	code := "CHK001"
	_, err := app.DB.Exec(`
		INSERT INTO magic_login_tokens (token, request_code, status, expires_at)
		VALUES (?, ?, 'pending', ?)
	`, token, code, time.Now().Add(10*time.Minute))
	if err != nil {
		t.Fatalf("failed to insert token: %v", err)
	}

	e := echo.New()

	// Check pending
	req := httptest.NewRequest(http.MethodGet, "/api/auth/magic-status?code="+code, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.CheckMagicStatus(app.DB)
	_ = h(c)

	var res map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res["status"] != "pending" {
		t.Errorf("expected status pending, got %v", res["status"])
	}

	// 2. Mark token verified
	_, _ = app.DB.Exec("UPDATE magic_login_tokens SET status = 'verified', user_type = 'parent' WHERE request_code = ?", code)

	req2 := httptest.NewRequest(http.MethodGet, "/api/auth/magic-status?code="+code, nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	_ = h(c2)

	var res2 map[string]interface{}
	_ = json.Unmarshal(rec2.Body.Bytes(), &res2)
	if res2["status"] != "verified" {
		t.Errorf("expected status verified, got %v", res2["status"])
	}
	if !strings.Contains(res2["redirect_url"].(string), token) {
		t.Errorf("expected redirect_url to contain token, got %v", res2["redirect_url"])
	}

	// 3. Check not found code
	req3 := httptest.NewRequest(http.MethodGet, "/api/auth/magic-status?code=NONEXIST", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	_ = h(c3)
	if rec3.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown code, got %d", rec3.Code)
	}
}

func TestHandleWhatsAppWebhook_ParentMatching(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed class and student with parent_phone
	_, _ = app.DB.Exec("INSERT INTO classes (id, name) VALUES (1, 'XII RPL 1')")
	_, err := app.DB.Exec(`
		INSERT INTO students (id, nis, name, parent_phone, parent_name, class_id, status)
		VALUES (101, '2026001', 'Ahmad Dani', '081234567890', 'Bapak Hendra', 1, 'active')
	`)
	if err != nil {
		t.Fatalf("failed to seed student: %v", err)
	}

	// Create pending login token
	code := "WA9901"
	token := "wa-secret-token-parent"
	_, _ = app.DB.Exec(`
		INSERT INTO magic_login_tokens (token, request_code, status, expires_at)
		VALUES (?, ?, 'pending', ?)
	`, token, code, time.Now().Add(10*time.Minute))

	e := echo.New()

	// OneSender format payload: data.from & data.body
	payload := map[string]interface{}{
		"event": "message",
		"data": map[string]interface{}{
			"from":    "6281234567890@s.whatsapp.net",
			"body":    "LOGIN " + code,
			"message": "LOGIN " + code,
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/whatsapp", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.HandleWhatsAppWebhook(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("HandleWhatsAppWebhook returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res["status"] != "success" {
		t.Errorf("expected success, got %v", res)
	}
	if res["user_type"] != "parent" {
		t.Errorf("expected parent user_type, got %v", res["user_type"])
	}

	// Verify token status in DB became 'verified'
	var dbStatus, dbUserType string
	var dbUserID int
	_ = app.DB.QueryRow("SELECT status, user_type, user_id FROM magic_login_tokens WHERE request_code = ?", code).Scan(&dbStatus, &dbUserType, &dbUserID)
	if dbStatus != "verified" || dbUserType != "parent" || dbUserID != 101 {
		t.Errorf("unexpected token state in DB: status=%s, type=%s, id=%d", dbStatus, dbUserType, dbUserID)
	}
}

func TestHandleWhatsAppWebhook_StudentAndStaff(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed student with student phone
	_, _ = app.DB.Exec(`
		INSERT INTO students (id, nis, name, phone, status)
		VALUES (201, '2026002', 'Budi Santoso', '08111222333', 'active')
	`)
	// Seed staff with phone
	_, _ = app.DB.Exec(`
		INSERT INTO staff (id, nip, name, phone, role)
		VALUES (301, '19800101', 'Dra. Siti Aminah', '08555666777', 'Guru')
	`)

	e := echo.New()
	h := handler.HandleWhatsAppWebhook(app.DB)

	// 1. Test student phone match
	codeStudent := "STU001"
	_, _ = app.DB.Exec("INSERT INTO magic_login_tokens (token, request_code, status, expires_at) VALUES ('tok-stu', ?, 'pending', ?)",
		codeStudent, time.Now().Add(10*time.Minute))

	payloadStudent := map[string]interface{}{
		"from":    "628111222333",
		"message": "LOGIN " + codeStudent,
	}
	b1, _ := json.Marshal(payloadStudent)
	req1 := httptest.NewRequest(http.MethodPost, "/api/webhook/whatsapp", bytes.NewBuffer(b1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	_ = h(e.NewContext(req1, rec1))

	var res1 map[string]interface{}
	_ = json.Unmarshal(rec1.Body.Bytes(), &res1)
	if res1["user_type"] != "student" {
		t.Errorf("expected student, got %v", res1)
	}

	// 2. Test staff phone match
	codeStaff := "STF001"
	_, _ = app.DB.Exec("INSERT INTO magic_login_tokens (token, request_code, status, expires_at) VALUES ('tok-stf', ?, 'pending', ?)",
		codeStaff, time.Now().Add(10*time.Minute))

	payloadStaff := map[string]interface{}{
		"from":    "628555666777",
		"message": "LOGIN " + codeStaff,
	}
	b2, _ := json.Marshal(payloadStaff)
	req2 := httptest.NewRequest(http.MethodPost, "/api/webhook/whatsapp", bytes.NewBuffer(b2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	_ = h(e.NewContext(req2, rec2))

	var res2 map[string]interface{}
	_ = json.Unmarshal(rec2.Body.Bytes(), &res2)
	if res2["user_type"] != "staff" {
		t.Errorf("expected staff, got %v", res2)
	}

	// 3. Test unregistered phone
	codeUnreg := "UNR001"
	_, _ = app.DB.Exec("INSERT INTO magic_login_tokens (token, request_code, status, expires_at) VALUES ('tok-unr', ?, 'pending', ?)",
		codeUnreg, time.Now().Add(10*time.Minute))

	payloadUnreg := map[string]interface{}{
		"from":    "628999999999",
		"message": "LOGIN " + codeUnreg,
	}
	b3, _ := json.Marshal(payloadUnreg)
	req3 := httptest.NewRequest(http.MethodPost, "/api/webhook/whatsapp", bytes.NewBuffer(b3))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()
	_ = h(e.NewContext(req3, rec3))

	var res3 map[string]interface{}
	_ = json.Unmarshal(rec3.Body.Bytes(), &res3)
	if res3["status"] != "unregistered_phone" {
		t.Errorf("expected unregistered_phone, got %v", res3)
	}
}

func TestConsumeMagicLink(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	_, _ = app.DB.Exec(`
		INSERT INTO students (id, nis, name, parent_phone, parent_name, status)
		VALUES (501, '2026501', 'Fajar Pratama', '081234567890', 'Bapak Suryo', 'active')
	`)

	token := "valid-secret-magic-token-xyz"
	_, err := app.DB.Exec(`
		INSERT INTO magic_login_tokens (token, request_code, phone, user_type, user_id, status, expires_at)
		VALUES (?, 'XYZ999', '6281234567890', 'parent', 501, 'verified', ?)
	`, token, time.Now().Add(10*time.Minute))
	if err != nil {
		t.Fatalf("failed to insert verified token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/magic?token="+token, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.ConsumeMagicLink(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("ConsumeMagicLink error: %v", err)
	}

	// Should redirect to /student/app
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 SeeOther, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location != "/student/app" {
		t.Errorf("expected redirect to /student/app, got %s", location)
	}

	// Verify cookies: both parent_session and student_session should be set!
	cookies := rec.Result().Cookies()
	var hasParentCookie, hasStudentCookie bool
	for _, ck := range cookies {
		if ck.Name == handler.ParentSessionCookie && ck.Value != "" {
			hasParentCookie = true
		}
		if ck.Name == handler.StudentSessionCookie && ck.Value != "" {
			hasStudentCookie = true
		}
	}
	if !hasParentCookie {
		t.Errorf("expected ParentSessionCookie to be set")
	}
	if !hasStudentCookie {
		t.Errorf("expected StudentSessionCookie to be set")
	}

	// Token should now be marked as used
	var status string
	_ = app.DB.QueryRow("SELECT status FROM magic_login_tokens WHERE token = ?", token).Scan(&status)
	if status != "used" {
		t.Errorf("expected token status 'used', got %s", status)
	}
}

func TestParentMultiChildSwitching(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	parentPhone := "081299998888"

	// Seed 2 children for same parent
	_, _ = app.DB.Exec("INSERT INTO classes (id, name) VALUES (1, 'X RPL 1'), (2, 'XII TKJ 2')")
	_, _ = app.DB.Exec(`
		INSERT INTO students (id, nis, name, parent_phone, parent_name, class_id, status)
		VALUES 
		(601, '2026601', 'Anak Pertama', ?, 'Bapak Wahyu', 1, 'active'),
		(602, '2024602', 'Anak Kedua', ?, 'Bapak Wahyu', 2, 'active'),
		(603, '2024603', 'Anak Orang Lain', '081111111111', 'Bapak Lain', 1, 'active')
	`, parentPhone, parentPhone)

	// Create parent session starting with child 601
	parentSessionToken := "parent-session-wahyu"
	studentSessionToken := "student-session-wahyu"
	expiry := time.Now().Add(7 * 24 * time.Hour)

	_, _ = app.DB.Exec(`
		INSERT INTO parent_sessions (token, parent_phone, parent_name, student_id, expires_at)
		VALUES (?, ?, 'Bapak Wahyu', 601, ?)
	`, parentSessionToken, parentPhone, expiry)

	_, _ = app.DB.Exec(`
		INSERT INTO student_sessions (token, student_id, expires_at)
		VALUES (?, 601, ?)
	`, studentSessionToken, expiry)

	e := echo.New()

	// 1. Test GetParentInfo
	req1 := httptest.NewRequest(http.MethodGet, "/api/parent/info", nil)
	req1.AddCookie(&http.Cookie{Name: handler.ParentSessionCookie, Value: parentSessionToken})
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	hInfo := handler.GetParentInfo(app.DB)
	_ = hInfo(c1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}

	var infoRes struct {
		IsParent        bool                `json:"is_parent"`
		ParentName      string              `json:"parent_name"`
		ActiveStudentID int                 `json:"active_student_id"`
		Children        []handler.ChildInfo `json:"children"`
	}
	_ = json.Unmarshal(rec1.Body.Bytes(), &infoRes)

	if !infoRes.IsParent {
		t.Errorf("expected is_parent true")
	}
	if infoRes.ActiveStudentID != 601 {
		t.Errorf("expected active student 601, got %d", infoRes.ActiveStudentID)
	}
	if len(infoRes.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(infoRes.Children))
	}
	if infoRes.Children[0].Name != "Anak Pertama" || infoRes.Children[1].Name != "Anak Kedua" {
		t.Errorf("unexpected children list: %+v", infoRes.Children)
	}

	// 2. Test SwitchActiveChild to second child (602)
	switchPayload := map[string]interface{}{"student_id": 602}
	bSwitch, _ := json.Marshal(switchPayload)
	req2 := httptest.NewRequest(http.MethodPost, "/api/parent/switch-child", bytes.NewBuffer(bSwitch))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: handler.ParentSessionCookie, Value: parentSessionToken})
	req2.AddCookie(&http.Cookie{Name: handler.StudentSessionCookie, Value: studentSessionToken})
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	hSwitch := handler.SwitchActiveChild(app.DB)
	_ = hSwitch(c2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on switch, got %d: %s", rec2.Code, rec2.Body.String())
	}

	// Verify DB sessions updated to child 602
	var currentChildInParentSession int
	_ = app.DB.QueryRow("SELECT student_id FROM parent_sessions WHERE token = ?", parentSessionToken).Scan(&currentChildInParentSession)
	if currentChildInParentSession != 602 {
		t.Errorf("expected parent_sessions student_id=602, got %d", currentChildInParentSession)
	}

	var currentChildInStudentSession int
	_ = app.DB.QueryRow("SELECT student_id FROM student_sessions WHERE token = ?", studentSessionToken).Scan(&currentChildInStudentSession)
	if currentChildInStudentSession != 602 {
		t.Errorf("expected student_sessions student_id=602, got %d", currentChildInStudentSession)
	}

	// 3. Test Unauthorized Switch (Child 603 belongs to someone else)
	unauthPayload := map[string]interface{}{"student_id": 603}
	bUnauth, _ := json.Marshal(unauthPayload)
	req3 := httptest.NewRequest(http.MethodPost, "/api/parent/switch-child", bytes.NewBuffer(bUnauth))
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: handler.ParentSessionCookie, Value: parentSessionToken})
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	_ = hSwitch(c3)
	if rec3.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for unauthorized child switch, got %d", rec3.Code)
	}
}

