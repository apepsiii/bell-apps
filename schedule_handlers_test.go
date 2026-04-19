package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAddScheduleHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	
	// Prepare form data
	form := make(url.Values)
	form.Set("time", "07:00")
	form.Set("label", "Test Bell")
	form.Set("audio_file", "test.mp3")

	req := httptest.NewRequest(http.MethodPost, "/admin/schedule/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := app.AddScheduleHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM schedules WHERE label = ?", "Test Bell").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Schedule should be created in the database")
}

func TestUpdateScheduleHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed DB
	res, err := app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES (?, ?, ?)", "08:00", "Old Label", "old.mp3")
	assert.NoError(t, err)
	id, _ := res.LastInsertId()

	e := echo.New()
	
	// Prepare form data
	form := make(url.Values)
	form.Set("time", "08:05")
	form.Set("label", "Updated Label")
	form.Set("audio_file", "updated.mp3")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", id))

	// Execute
	err = app.UpdateScheduleHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var label, audioFile string
	err = app.DB.QueryRow("SELECT label, audio_file FROM schedules WHERE id = ?", id).Scan(&label, &audioFile)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Label", label)
	assert.Equal(t, "updated.mp3", audioFile)
}

func TestDeleteScheduleHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed DB
	res, err := app.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES (?, ?, ?)", "09:00", "To Delete", "delete.mp3")
	assert.NoError(t, err)
	id, _ := res.LastInsertId()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", id))

	// Execute
	err = app.DeleteScheduleHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM schedules WHERE id = ?", id).Scan(&count)
	assert.Error(t, err, "Should be no rows") // sql.ErrNoRows is expected
	assert.Equal(t, 0, count)
}
