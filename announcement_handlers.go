package main

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	htgotts "github.com/hegedustibor/htgo-tts"
	"github.com/labstack/echo/v4"
)

// --- STRUCTS ---

type Announcement struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
		Message     string       `json:"message"`
	AudioFile   string       `json:"audio_file"`
	ScheduledAt sql.NullTime `json:"scheduled_at"`
	PlayedAt    sql.NullTime `json:"played_at"`
	Status      string       `json:"status"` // 'scheduled', 'played', 'cancelled'
	CreatedAt   time.Time    `json:"created_at"`
}

// --- HANDLERS ---

// GetAnnouncementsHandler retrieves all announcements
func (a *App) GetAnnouncementsHandler(c echo.Context) error {
	rows, err := a.DB.Query("SELECT id, title, message, audio_file, scheduled_at, played_at, status, created_at FROM announcements ORDER BY created_at DESC")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	var announcements []Announcement
	for rows.Next() {
		var ann Announcement
		err := rows.Scan(&ann.ID, &ann.Title, &ann.Message, &ann.AudioFile, &ann.ScheduledAt, &ann.PlayedAt, &ann.Status, &ann.CreatedAt)
		if err != nil {
			continue
		}
		announcements = append(announcements, ann)
	}

	return c.JSON(http.StatusOK, announcements)
}

// CreateAnnouncementHandler creates a new announcement
func (a *App) CreateAnnouncementHandler(c echo.Context) error {
	title := c.FormValue("title")
	message := c.FormValue("message")
	scheduledAtStr := c.FormValue("scheduled_at") // Expected format: "YYYY-MM-DDTHH:MM"

	if title == "" || message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Title and message are required"})
	}

	// --- Text-to-Speech ---
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	safeMessage := reg.ReplaceAllString(message, "")
	baseFileName := strings.ReplaceAll(strings.TrimSpace(strings.ToLower(safeMessage)), " ", "_")
	if len(baseFileName) == 0 {
		baseFileName = "pengumuman_" + strconv.FormatInt(time.Now().Unix(), 10)
	} else if len(baseFileName) > 50 {
		baseFileName = baseFileName[:50]
	}

	speech := htgotts.Speech{Folder: "public/assets/audio", Language: "id"}
	_, err := speech.CreateSpeechFile(message, baseFileName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create audio file: " + err.Error()})
	}
	
	fileName := baseFileName + ".mp3"
	// --- End Text-to-Speech ---

	var scheduledAt sql.NullTime
	if scheduledAtStr != "" {
		t, err := time.Parse("2006-01-02T15:04", scheduledAtStr)
		if err == nil {
			scheduledAt = sql.NullTime{Time: t, Valid: true}
		}
	}

	status := "played"
	if scheduledAt.Valid {
		status = "scheduled"
	}

	res, err := a.DB.Exec("INSERT INTO announcements (title, message, audio_file, scheduled_at, status) VALUES (?, ?, ?, ?, ?)",
		title, message, fileName, scheduledAt, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error: " + err.Error()})
	}

	id, _ := res.LastInsertId()

	// If not scheduled, play it now
	if status == "played" {
		// This is a placeholder for the actual play logic
		// In a real scenario, you'd trigger a websocket event or similar
		// to the clients (bell devices) to play the audio.
		a.DB.Exec("UPDATE announcements SET played_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Announcement created successfully"})
}

// DeleteAnnouncementHandler deletes an announcement
func (a *App) DeleteAnnouncementHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID is required"})
	}

	// First, get the audio file name to delete it from the filesystem
	var audioFile string
	err := a.DB.QueryRow("SELECT audio_file FROM announcements WHERE id = ?", id).Scan(&audioFile)
	if err != nil {
		// It might be already deleted or not exist, but we should still try to delete from DB
		if err != sql.ErrNoRows {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to find announcement: " + err.Error()})
		}
	}

	if audioFile != "" {
		filePath := filepath.Join("public/assets/audio", audioFile)
		os.Remove(filePath)
	}

	// Now, delete the record from the database
	_, err = a.DB.Exec("DELETE FROM announcements WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete announcement: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Announcement deleted successfully"})
}

// PlayAnnouncementHandler immediately plays an announcement
func (a *App) PlayAnnouncementHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID is required"})
	}

	// In a real-world scenario, this would trigger a WebSocket event to all connected bell devices.
	// For now, we'll just update the status in the database.
	_, err := a.DB.Exec("UPDATE announcements SET status = 'played', played_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to play announcement: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Announcement played successfully"})
}
