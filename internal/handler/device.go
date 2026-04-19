package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AddDevice(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		ipAddress := c.FormValue("ip_address")
		_, err := db.Exec("INSERT INTO devices (name, ip_address, status) VALUES (?, ?, 'online')", name, ipAddress)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Device berhasil ditambahkan"})
	}
}

func UpdateDevice(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		name := c.FormValue("name")
		ipAddress := c.FormValue("ip_address")
		status := c.FormValue("status")
		_, err := db.Exec("UPDATE devices SET name=?, ip_address=?, status=? WHERE id=?", name, ipAddress, status, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Device berhasil diperbarui"})
	}
}

func DeleteDevice(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM devices WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Device dihapus"})
	}
}
