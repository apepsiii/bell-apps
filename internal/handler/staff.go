package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"belsekolah/pkg/utils"
)

func AddStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := strings.TrimSpace(c.FormValue("rfid_uid"))
		nip := strings.TrimSpace(c.FormValue("nip"))
		name := strings.TrimSpace(c.FormValue("name"))
		phone := utils.FormatPhone(c.FormValue("phone"))
		role := strings.TrimSpace(c.FormValue("role"))
		gender := strings.ToUpper(strings.TrimSpace(c.FormValue("gender")))
		if gender != "P" && gender != "L" {
			if DetermineStaffHonorific("", name) == "Ibu" {
				gender = "P"
			} else {
				gender = "L"
			}
		}

		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nama staf wajib diisi"})
		}

		var rfidVal interface{} = rfid
		if rfid == "" {
			rfidVal = nil
		} else {
			// Cek apakah RFID sudah terdaftar pada staf lain
			var existingName string
			err := db.QueryRow("SELECT name FROM staff WHERE rfid_uid = ?", rfid).Scan(&existingName)
			if err == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "RFID UID '" + rfid + "' sudah digunakan oleh staf: " + existingName,
				})
			}
		}

		_, err := db.Exec("INSERT INTO staff (rfid_uid, nip, name, phone, role, gender) VALUES (?, ?, ?, ?, ?, ?)", rfidVal, nip, name, phone, role, gender)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: staff.rfid_uid") || strings.Contains(err.Error(), "Duplicate entry") {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "RFID UID '" + rfid + "' sudah terdaftar pada staf lain",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menambahkan staf: " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staf berhasil ditambahkan"})
	}
}

func UpdateStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		rfid := strings.TrimSpace(c.FormValue("rfid_uid"))
		nip := strings.TrimSpace(c.FormValue("nip"))
		name := strings.TrimSpace(c.FormValue("name"))
		phone := utils.FormatPhone(c.FormValue("phone"))
		role := strings.TrimSpace(c.FormValue("role"))
		gender := strings.ToUpper(strings.TrimSpace(c.FormValue("gender")))
		if gender != "P" && gender != "L" {
			gender = "L"
		}

		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nama staf wajib diisi"})
		}

		var rfidVal interface{} = rfid
		if rfid == "" {
			rfidVal = nil
		} else {
			// Cek apakah RFID sudah terdaftar pada staf lain (selain id staf ini)
			var existingName string
			err := db.QueryRow("SELECT name FROM staff WHERE rfid_uid = ? AND id != ?", rfid, id).Scan(&existingName)
			if err == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "RFID UID '" + rfid + "' sudah digunakan oleh staf: " + existingName,
				})
			}
		}

		_, err := db.Exec("UPDATE staff SET rfid_uid=?, nip=?, name=?, phone=?, role=?, gender=? WHERE id=?", rfidVal, nip, name, phone, role, gender, id)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: staff.rfid_uid") || strings.Contains(err.Error(), "Duplicate entry") {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "RFID UID '" + rfid + "' sudah terdaftar pada staf lain",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memperbarui staf: " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staf berhasil diperbarui"})
	}
}

func DeleteStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM staff WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staf dihapus"})
	}
}
