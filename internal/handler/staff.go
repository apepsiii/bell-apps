package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/pkg/utils"
)

func AddStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.FormValue("rfid_uid")
		nip := c.FormValue("nip")
		name := c.FormValue("name")
		phone := utils.FormatPhone(c.FormValue("phone"))
		role := c.FormValue("role")
		_, err := db.Exec("INSERT INTO staff (rfid_uid, nip, name, phone, role) VALUES (?, ?, ?, ?, ?)", rfid, nip, name, phone, role)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staf berhasil ditambahkan"})
	}
}

func UpdateStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		rfid := c.FormValue("rfid_uid")
		nip := c.FormValue("nip")
		name := c.FormValue("name")
		phone := utils.FormatPhone(c.FormValue("phone"))
		role := c.FormValue("role")
		_, err := db.Exec("UPDATE staff SET rfid_uid=?, nip=?, name=?, phone=?, role=? WHERE id=?", rfid, nip, name, phone, role, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
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
