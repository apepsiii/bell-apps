package handler

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

		photoFile := ""
		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				os.MkdirAll(config.PhotoPath, 0755)
				ext := filepath.Ext(file.Filename)
				newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext)
				dstPath := filepath.Join(config.PhotoPath, newFilename)
				dst, err := os.Create(dstPath)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					photoFile = newFilename
				}
			}
		}

		_, err = db.Exec("INSERT INTO students (rfid_uid, nis, name, parent_phone, class_id, photo) VALUES (?, ?, ?, ?, ?, ?)", rfid, nis, name, phone, classID, photoFile)
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

		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				os.MkdirAll(config.PhotoPath, 0755)
				ext := filepath.Ext(file.Filename)
				newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext)
				dstPath := filepath.Join(config.PhotoPath, newFilename)
				dst, err := os.Create(dstPath)
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					db.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=?, photo=? WHERE id=?", rfid, nis, name, phone, classID, newFilename, id)
					return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
				}
			}
		}
		_, err = db.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=? WHERE id=?", rfid, nis, name, phone, classID, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
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
