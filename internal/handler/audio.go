package handler

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

const (
	maxAudioSize  = 20 * 1024 * 1024 // 20 MB
)

var allowedAudioExts = map[string]bool{
	".mp3": true,
	".wav": true,
	".ogg": true,
	".aac": true,
	".m4a": true,
	".flac": true,
}

func UploadAudio(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		displayName := c.FormValue("display_name")
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
		}

		if file.Size > maxAudioSize {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Ukuran file maksimal 20 MB"})
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedAudioExts[ext] {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format file tidak didukung. Gunakan MP3, WAV, OGG, AAC, M4A, atau FLAC"})
		}

		if displayName == "" {
			displayName = file.Filename
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca file"})
		}
		defer src.Close()

		buf := make([]byte, 512)
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca file"})
		}
		mimeType := http.DetectContentType(buf[:n])
		if !strings.HasPrefix(mimeType, "audio/") && mimeType != "application/ogg" && mimeType != "video/ogg" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File bukan audio yang valid"})
		}

		os.MkdirAll(config.GetUploadPath(), 0755)
		safeFilename := filepath.Base(file.Filename)
		dstPath := filepath.Join(config.GetUploadPath(), safeFilename)

		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan file"})
		}
		defer dst.Close()

		dst.Write(buf[:n])
		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyalin file"})
		}

		db.Exec("INSERT INTO audio_files (file_name, display_name) VALUES (?, ?)", safeFilename, displayName)
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Audio berhasil diupload"})
	}
}

func RenameAudio(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		newName := c.FormValue("display_name")
		_, err := db.Exec("UPDATE audio_files SET display_name=? WHERE id=?", newName, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Nama audio diperbarui"})
	}
}

func DeleteAudio(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var fileName string
		err := db.QueryRow("SELECT file_name FROM audio_files WHERE id=?", id).Scan(&fileName)
		if err == nil {
			os.Remove(filepath.Join(config.GetUploadPath(), fileName))
		}

		_, err = db.Exec("DELETE FROM audio_files WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Audio dihapus"})
	}
}
