package handler

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// downloadTTSAudio fetches an MP3 from Google Translate TTS and saves it.
// This replaces github.com/hegedustibor/htgo-tts which pulled in
// github.com/hajimehoshi/oto/v2 — an audio-playback dep that fails to
// cross-compile (GOOS=linux from Windows) because its platform-specific
// context files are excluded by build tags. We only need the file-download
// half of htgo-tts, so implement it inline.
func downloadTTSAudio(text, lang, folder, fileName string) (string, error) {
	if err := os.MkdirAll(folder, 0755); err != nil {
		return "", err
	}
	f := filepath.Join(folder, fileName+".mp3")

	// Skip download if already cached
	if _, err := os.Stat(f); err == nil {
		return f, nil
	}

	q := url.QueryEscape(text)
	dlURL := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=%d&client=tw-ob&q=%s&tl=%s", len([]rune(text)), q, lang)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(dlURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TTS download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(f)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return f, nil
}

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
		// Remove consecutive underscores and trailing underscores
		multiUnderscore := regexp.MustCompile(`_+`)
		baseFileName = multiUnderscore.ReplaceAllString(baseFileName, "_")
		baseFileName = strings.Trim(baseFileName, "_")
		if len(baseFileName) == 0 {
			baseFileName = "pengumuman_" + strconv.FormatInt(time.Now().Unix(), 10)
		} else if len(baseFileName) > 50 {
			baseFileName = strings.Trim(baseFileName[:50], "_")
		}
		// Final safety: use timestamp if still empty
		if baseFileName == "" {
			baseFileName = "pengumuman_" + strconv.FormatInt(time.Now().Unix(), 10)
		}

	speechFile, err := downloadTTSAudio(message, "id", "public/assets/audio", baseFileName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create audio file: " + err.Error()})
	}
	_ = speechFile

	fileName := baseFileName + ".mp3"

		var scheduledAt sql.NullTime
		if scheduledAtStr != "" {
			loc, err := time.LoadLocation("Asia/Jakarta")
			if err != nil {
				loc = time.Local
			}
			t, err := time.ParseInLocation("2006-01-02T15:04", scheduledAtStr, loc)
			if err == nil {
				scheduledAt = sql.NullTime{Time: t, Valid: true}
			}
		}

		status := "played"
		if scheduledAt.Valid {
			status = "pending"
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
