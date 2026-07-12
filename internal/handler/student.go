package handler

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
	"belsekolah/pkg/utils"
)

func AddStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.FormValue("rfid_uid")
		nis := c.FormValue("nis")
		name := c.FormValue("name")
		phone := utils.FormatPhone(c.FormValue("parent_phone"))
		classID := c.FormValue("class_id")
		birthday := c.FormValue("birthday")

		photoFile := ""

		// Check for file upload first
		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				os.MkdirAll(config.PhotoPath, 0755)
				ext := filepath.Ext(file.Filename)
				if ext == "" {
					ext = ".jpg"
				}
				newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext)
				dstPath := filepath.Join(config.PhotoPath, newFilename)
				dst, err := os.Create(dstPath)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					photoFile = newFilename
				}
			}
		} else {
			// Check for captured photo (base64)
			capturedPhoto := c.FormValue("captured_photo")
			if capturedPhoto != "" {
				os.MkdirAll(config.PhotoPath, 0755)
				newFilename := fmt.Sprintf("%s_%s.jpg", nis, time.Now().Format("20060102150405"))
				dstPath := filepath.Join(config.PhotoPath, newFilename)

				// Remove data URL prefix if present
				dataStr := capturedPhoto
				if strings.HasPrefix(dataStr, "data:image/jpeg;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/jpeg;base64,")
				} else if strings.HasPrefix(dataStr, "data:image/png;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/png;base64,")
				} else if strings.HasPrefix(dataStr, "data:image/jpg;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/jpg;base64,")
				}

				// Decode base64 and save
				decoded, err := base64.StdEncoding.DecodeString(dataStr)
				if err == nil {
					os.WriteFile(dstPath, decoded, 0644)
					photoFile = newFilename
				}
			}
		}

		_, err = db.Exec("INSERT INTO students (rfid_uid, nis, name, parent_phone, class_id, photo, birthday, status) VALUES (?, ?, ?, ?, ?, ?, ?, 'active')", rfid, nis, name, phone, classID, photoFile, birthday)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal (Mungkin RFID/NIS duplikat): " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa ditambahkan"})
	}
}

func UpdateStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		rfid := c.FormValue("rfid_uid")
		nis := c.FormValue("nis")
		name := c.FormValue("name")
		phone := utils.FormatPhone(c.FormValue("parent_phone"))
		classID := c.FormValue("class_id")
		birthday := c.FormValue("birthday")

		photoFile := ""

		// Check for file upload first
		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				os.MkdirAll(config.PhotoPath, 0755)
				ext := filepath.Ext(file.Filename)
				if ext == "" {
					ext = ".jpg"
				}
				newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext)
				dstPath := filepath.Join(config.PhotoPath, newFilename)
				dst, err := os.Create(dstPath)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					photoFile = newFilename
					db.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=?, photo=?, birthday=? WHERE id=?", rfid, nis, name, phone, classID, photoFile, birthday, id)
					return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
				}
			}
		} else {
			// Check for captured photo (base64)
			capturedPhoto := c.FormValue("captured_photo")
			if capturedPhoto != "" {
				os.MkdirAll(config.PhotoPath, 0755)
				newFilename := fmt.Sprintf("%s_%s.jpg", nis, time.Now().Format("20060102150405"))
				dstPath := filepath.Join(config.PhotoPath, newFilename)

				// Remove data URL prefix if present
				dataStr := capturedPhoto
				if strings.HasPrefix(dataStr, "data:image/jpeg;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/jpeg;base64,")
				} else if strings.HasPrefix(dataStr, "data:image/png;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/png;base64,")
				} else if strings.HasPrefix(dataStr, "data:image/jpg;base64,") {
					dataStr = strings.TrimPrefix(dataStr, "data:image/jpg;base64,")
				}

				// Decode base64 and save
				decoded, err := base64.StdEncoding.DecodeString(dataStr)
				if err == nil {
					os.WriteFile(dstPath, decoded, 0644)
					photoFile = newFilename
					db.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=?, photo=?, birthday=? WHERE id=?", rfid, nis, name, phone, classID, photoFile, birthday, id)
					return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
				}
			}
		}

		_, err = db.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=?, birthday=? WHERE id=?", rfid, nis, name, phone, classID, birthday, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
	}
}

type StatusUpdateRequest struct {
	Status string `json:"status"`
}

func UpdateStudentStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var req StatusUpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
		}

		if req.Status != "active" && req.Status != "inactive" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Status must be 'active' or 'inactive'"})
		}

		_, err := db.Exec("UPDATE students SET status = ? WHERE id = ?", req.Status, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Status siswa diperbarui"})
	}
}

func DeleteStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id != "" && id != "0" {
			_, err := db.Exec("DELETE FROM students WHERE id=?", id)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa dihapus"})
	}
}
