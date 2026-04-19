package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// --- STUDENT HANDLER TESTS ---

func TestAddStudentHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed a class for the student
	res, err := app.DB.Exec("INSERT INTO classes (name) VALUES (?)", "12-RPL")
	assert.NoError(t, err)
	classID, _ := res.LastInsertId()

	e := echo.New()
	
	form := make(url.Values)
	form.Set("rfid_uid", "123456789")
	form.Set("nis", "1001")
	form.Set("name", "John Doe")
	form.Set("parent_phone", "081234567890")
	form.Set("class_id", fmt.Sprintf("%d", classID))

	req := httptest.NewRequest(http.MethodPost, "/admin/student/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = app.AddStudentHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM students WHERE nis = ?", "1001").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Student should be created")
}

func TestUpdateStudentHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed class and student
	res, _ := app.DB.Exec("INSERT INTO classes (name) VALUES (?)", "12-RPL")
	classID, _ := res.LastInsertId()
	res, _ = app.DB.Exec("INSERT INTO students (nis, name, class_id) VALUES (?, ?, ?)", "1002", "Jane Doe", classID)
	studentID, _ := res.LastInsertId()

	e := echo.New()
	
	form := make(url.Values)
	form.Set("nis", "1002")
	form.Set("name", "Jane Smith") // Updated name
	form.Set("class_id", fmt.Sprintf("%d", classID))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", studentID))

	// Execute
	err = app.UpdateStudentHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var name string
	err = app.DB.QueryRow("SELECT name FROM students WHERE id = ?", studentID).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Smith", name, "Student name should be updated")
}

func TestDeleteStudentHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed student
	res, _ := app.DB.Exec("INSERT INTO students (nis, name) VALUES (?, ?)", "1003", "To Delete")
	studentID, _ := res.LastInsertId()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", studentID))

	// Execute
	err = app.DeleteStudentHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM students WHERE id = ?", studentID).Scan(&count)
	assert.Error(t, err) // Expect sql.ErrNoRows
	assert.Equal(t, 0, count)
}

// --- STAFF HANDLER TESTS ---

func TestAddStaffHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	
	form := make(url.Values)
	form.Set("rfid_uid", "987654321")
	form.Set("nip", "S001")
	form.Set("name", "Mr. Teacher")
	form.Set("phone", "081200001111")
	form.Set("role", "Guru")

	req := httptest.NewRequest(http.MethodPost, "/admin/staff/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = app.AddStaffHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM staff WHERE nip = ?", "S001").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Staff should be created")
}

func TestUpdateStaffHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed staff
	res, _ := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES (?, ?, ?)", "S002", "Old Staff", "Staf")
	staffID, _ := res.LastInsertId()

	e := echo.New()
	
	form := make(url.Values)
	form.Set("nip", "S002")
	form.Set("name", "Updated Staff Name") // Updated name
	form.Set("role", "Staf TU")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", staffID))

	// Execute
	err = app.UpdateStaffHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var name, role string
	err = app.DB.QueryRow("SELECT name, role FROM staff WHERE id = ?", staffID).Scan(&name, &role)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Staff Name", name)
	assert.Equal(t, "Staf TU", role)
}

func TestDeleteStaffHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed staff
	res, _ := app.DB.Exec("INSERT INTO staff (nip, name) VALUES (?, ?)", "S003", "Staff To Delete")
	staffID, _ := res.LastInsertId()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", staffID))

	// Execute
	err = app.DeleteStaffHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM staff WHERE id = ?", staffID).Scan(&count)
	assert.Error(t, err) // Expect sql.ErrNoRows
	assert.Equal(t, 0, count)
}
