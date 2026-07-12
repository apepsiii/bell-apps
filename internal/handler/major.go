package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AddMajor(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		_, err := db.Exec("INSERT INTO majors (name) VALUES (?)", name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan berhasil ditambahkan"})
	}
}

func UpdateMajor(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		name := c.FormValue("name")
		_, err := db.Exec("UPDATE majors SET name=? WHERE id=?", name, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan berhasil diperbarui"})
	}
}

func DeleteMajor(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM majors WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan dihapus"})
	}
}
