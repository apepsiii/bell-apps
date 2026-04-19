package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AddSchedule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		timeVal := c.FormValue("time")
		label := c.FormValue("label")
		audio := c.FormValue("audio_file")
		_, err := db.Exec("INSERT INTO schedules (time, label, audio_file) VALUES (?, ?, ?)", timeVal, label, audio)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal database: " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal berhasil ditambahkan"})
	}
}

func UpdateSchedule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		timeVal := c.FormValue("time")
		label := c.FormValue("label")
		audio := c.FormValue("audio_file")
		_, err := db.Exec("UPDATE schedules SET time=?, label=?, audio_file=? WHERE id=?", timeVal, label, audio, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal berhasil diperbarui"})
	}
}

func DeleteSchedule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM schedules WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal dihapus"})
	}
}
