package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AddClass(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		majorID := c.FormValue("major_id")
		waGroupID := c.FormValue("wa_group_id")
		if majorID == "" {
			majorID = "0"
		}
		_, err := db.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES (?, ?, ?)", name, majorID, waGroupID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas berhasil ditambahkan"})
	}
}

func UpdateClass(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		name := c.FormValue("name")
		majorID := c.FormValue("major_id")
		waGroupID := c.FormValue("wa_group_id")
		_, err := db.Exec("UPDATE classes SET name=?, major_id=?, wa_group_id=? WHERE id=?", name, majorID, waGroupID, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas berhasil diperbarui"})
	}
}

func DeleteClass(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM classes WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas dihapus"})
	}
}
