package handler

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

func UploadAudio(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		displayName := c.FormValue("display_name")
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
		}
		if displayName == "" {
			displayName = file.Filename
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca file"})
		}
		defer src.Close()

		os.MkdirAll(config.UploadPath, 0755)
		dstPath := filepath.Join(config.UploadPath, filepath.Base(file.Filename))

		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan file"})
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyalin file"})
		}

		db.Exec("INSERT INTO audio_files (file_name, display_name) VALUES (?, ?)", file.Filename, displayName)
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
			os.Remove(filepath.Join(config.UploadPath, fileName))
		}

		_, err = db.Exec("DELETE FROM audio_files WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Audio dihapus"})
	}
}
