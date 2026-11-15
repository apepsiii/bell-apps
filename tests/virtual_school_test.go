package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"belsekolah/internal/handler"
)

func TestVirtualSchoolStateEndpoint(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Seed class & students
	_, err := app.DB.Exec("INSERT INTO classes (id, name, major_id) VALUES (1, 'X-MPLB', 1)")
	if err != nil {
		t.Fatalf("failed to insert class: %v", err)
	}

	_, err = app.DB.Exec("INSERT INTO students (id, nis, name, class_id, status) VALUES (1, '202401', 'Budi Pratama', 1, 'Aktif')")
	if err != nil {
		t.Fatalf("failed to insert student: %v", err)
	}

	// Test GET /api/virtual-school/state
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/virtual-school/state?sim_time=07:35&sim=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.GetVirtualSchoolState(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if res["status"] != "success" {
		t.Fatalf("expected status 'success', got %v", res["status"])
	}

	students, ok := res["students"].([]interface{})
	if !ok || len(students) == 0 {
		t.Fatalf("expected students in response, got %v", res["students"])
	}

	firstStudent := students[0].(map[string]interface{})
	if firstStudent["name"] != "Budi Pratama" {
		t.Fatalf("expected student name 'Budi Pratama', got %v", firstStudent["name"])
	}

	if firstStudent["current_room"] != "kelas-x-mplb" {
		t.Fatalf("expected current_room 'kelas-x-mplb', got %v", firstStudent["current_room"])
	}
}

func TestVirtualSchoolBellsEndpoint(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/virtual-school/bells", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.GetVirtualSchoolBells(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestVirtualSchoolAddPointEndpoint(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Seed student
	_, err := app.DB.Exec("INSERT INTO classes (id, name, major_id) VALUES (1, 'XI-PM', 2)")
	if err != nil {
		t.Fatalf("failed to insert class: %v", err)
	}

	_, err = app.DB.Exec("INSERT INTO students (id, nis, name, class_id, status) VALUES (121, 'AHBAHAHHA', 'ALGI DAVA KURNIAWAN', 1, 'active')")
	if err != nil {
		t.Fatalf("failed to insert student: %v", err)
	}

	e := echo.New()
	payload := `{"student_id": 121, "points_change": 15, "description": "Aktif Menjawab di KBM", "recorded_by": "Pak Guru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/virtual-school/points", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.AddVirtualSchoolPoint(app.DB)
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if res["status"] != "success" {
		t.Fatalf("expected status success, got %v", res["status"])
	}

	if res["points_change"] != float64(15) {
		t.Fatalf("expected points_change 15, got %v", res["points_change"])
	}
}
