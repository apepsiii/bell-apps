package handler

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hegedustibor/htgo-tts"
	"github.com/labstack/echo/v4"
)

type Announcement struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Message     string       `json:"message"`
	AudioFile   string       `json:"audio_file"`
	ScheduledAt sql.NullTime `json:"scheduled_at"`
	PlayedAt    sql.NullTime `json:"played_at"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

func GetAnnouncements(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT id, title, message, audio_file, scheduled_at, played_at, status, created_at FROM announcements ORDER BY created_at DESC")
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
}

func CreateAnnouncement(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		title := c.FormValue("title")
		message := c.FormValue("message")
		scheduledAtStr := c.FormValue("scheduled_at")

		if title == "" || message == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Title and message are required"})
		}

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

		res, err := db.Exec("INSERT INTO announcements (title, message, audio_file, scheduled_at, status) VALUES (?, ?, ?, ?, ?)",
			title, message, fileName, scheduledAt, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error: " + err.Error()})
		}

		id, _ := res.LastInsertId()

		if status == "played" {
			db.Exec("UPDATE announcements SET played_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		}

		return c.JSON(http.StatusCreated, map[string]string{"message": "Announcement created successfully"})
	}
}

func DeleteAnnouncement(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID is required"})
		}

		var audioFile string
		err := db.QueryRow("SELECT audio_file FROM announcements WHERE id = ?", id).Scan(&audioFile)
		if err != nil && err != sql.ErrNoRows {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to find announcement: " + err.Error()})
		}

		if audioFile != "" {
			filePath := filepath.Join("public/assets/audio", audioFile)
			os.Remove(filePath)
		}

		_, err = db.Exec("DELETE FROM announcements WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete announcement: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Announcement deleted successfully"})
	}
}

func PlayAnnouncement(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID is required"})
		}

		_, err := db.Exec("UPDATE announcements SET status = 'played', played_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to play announcement: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Announcement played successfully"})
	}
}
