package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// --- TEST SETUP ---

// setupTestApp initializes a new App with an in-memory SQLite database for testing.
func setupTestApp(t *testing.T) (*App, func()) {
	// Use in-memory SQLite database for tests
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Run all migrations
	app := &App{DB: db}
	InitDB() 

	// Create necessary directories
	os.MkdirAll(UploadPath, 0755)

	// Teardown function to close the database and clean up files
	teardown := func() {
		db.Close()
		os.RemoveAll("./public")
	}

	return app, teardown
}

// --- ANNOUNCEMENT HANDLER TESTS ---

func TestCreateAnnouncementHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	
	// Prepare request body
	formData := "title=Test+Announcement&message=This+is+a+test&scheduled_at="
	req := httptest.NewRequest(http.MethodPost, "/admin/announcement/add", bytes.NewReader([]byte(formData)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute the handler
	err := app.CreateAnnouncementHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusSeeOther, rec.Code) // Expecting a redirect

	// Verify the announcement was created in the database
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM announcements WHERE title = ?", "Test Announcement").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Announcement should be created in the database")

	// Verify an audio file was created
	var audioFile string
	err = app.DB.QueryRow("SELECT audio_file FROM announcements WHERE title = ?", "Test Announcement").Scan(&audioFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, audioFile, "Audio file should not be empty")
	
	// Check if the file exists on disk
	_, err = os.Stat("./public/assets/audio/" + audioFile)
	assert.NoError(t, err, "Audio file should exist on disk")
}

func TestGetAnnouncementsHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed the database with an announcement
	_, err := app.DB.Exec("INSERT INTO announcements (title, message, audio_file, status) VALUES (?, ?, ?, ?)",
		"Test Get", "Message Get", "test_get.mp3", "pending")
	assert.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/announcements", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = app.GetAnnouncementsHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var announcements []Announcement
	err = json.Unmarshal(rec.Body.Bytes(), &announcements)
	assert.NoError(t, err)
	assert.Len(t, announcements, 1, "Should return one announcement")
	assert.Equal(t, "Test Get", announcements[0].Title)
}

func TestDeleteAnnouncementHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Create a dummy audio file
	dummyAudioPath := "./public/assets/audio/delete_me.mp3"
	os.WriteFile(dummyAudioPath, []byte("dummy"), 0644)

	// Seed the database
	res, err := app.DB.Exec("INSERT INTO announcements (title, message, audio_file, status) VALUES (?, ?, ?, ?)",
		"To Delete", "Delete me", "delete_me.mp3", "pending")
	assert.NoError(t, err)
	id, _ := res.LastInsertId()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", id))

	// Execute
	err = app.DeleteAnnouncementHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify it's deleted from DB
	var count int
	err = app.DB.QueryRow("SELECT COUNT(*) FROM announcements WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Announcement should be deleted from the database")

	// Verify the audio file is also deleted
	_, err = os.Stat(dummyAudioPath)
	assert.True(t, os.IsNotExist(err), "Audio file should be deleted from disk")
}

func TestPlayAnnouncementHandler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// Seed the database
	res, err := app.DB.Exec("INSERT INTO announcements (title, message, audio_file, status) VALUES (?, ?, ?, ?)",
		"To Play", "Play me", "play_me.mp3", "pending")
	assert.NoError(t, err)
	id, _ := res.LastInsertId()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", id))

	// Execute
	err = app.PlayAnnouncementHandler(c)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify the status is updated in the DB
	var status string
	var playedAt sql.NullTime
	err = app.DB.QueryRow("SELECT status, played_at FROM announcements WHERE id = ?", id).Scan(&status, &playedAt)
	assert.NoError(t, err)
	assert.Equal(t, "played", status, "Status should be updated to 'played'")
	assert.True(t, playedAt.Valid, "PlayedAt timestamp should be set")
}

func TestAnnouncementScheduler(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	// 1. Schedule an announcement for the past
	pastTime := time.Now().Add(-1 * time.Minute)
	_, err := app.DB.Exec("INSERT INTO announcements (title, message, audio_file, status, scheduled_at) VALUES (?, ?, ?, ?, ?)",
		"Scheduled Play", "Play me now", "scheduled.mp3", "pending", pastTime)
	assert.NoError(t, err)

	// 2. Run the scheduler check
	app.checkForPendingAnnouncements()

	// 3. Verify it's set as active
	app.mu.Lock()
	assert.NotNil(t, app.activeAnnouncment, "An announcement should be active")
	assert.Equal(t, "Scheduled Play", app.activeAnnouncment.Title)
	app.mu.Unlock()

	// 4. Verify its status is 'playing' in DB
	var status string
	err = app.DB.QueryRow("SELECT status FROM announcements WHERE title = ?", "Scheduled Play").Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "playing", status)

	// 5. Simulate a device sync
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/sync", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err = app.SyncHandler(c)
	assert.NoError(t, err)

	// 6. Verify the sync response contains the announcement
	var schedules []map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &schedules)
	assert.NoError(t, err)
	assert.Greater(t, len(schedules), 0)
	assert.Equal(t, "announcement", schedules[0]["type"])
	assert.Equal(t, "Scheduled Play", schedules[0]["label"])

	// 7. Verify the active announcement is now cleared
	app.mu.Lock()
	assert.Nil(t, app.activeAnnouncment, "Active announcement should be cleared after sync")
	app.mu.Unlock()

	// 8. Verify the status is 'played' in DB (wait a bit for goroutine)
	time.Sleep(100 * time.Millisecond)
	err = app.DB.QueryRow("SELECT status FROM announcements WHERE title = ?", "Scheduled Play").Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "played", status)
}
