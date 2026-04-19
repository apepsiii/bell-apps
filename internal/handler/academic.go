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

func AddStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.FormValue("rfid_uid")
		nip := c.FormValue("nip")
		name := c.FormValue("name")
		phone := c.FormValue("phone")
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
		phone := c.FormValue("phone")
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
