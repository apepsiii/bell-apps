package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ScanPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "scan.html", nil)
	}
}

func ScanFacePage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "attendance_face.html", nil)
	}
}

func ScanPrayerPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "scan_prayer.html", nil)
	}
}

type ProfileData struct {
	Type       string
	ID         int
	Name       string
	IdentityNo string
	ExtraInfo  string
	Phone      string
	RFID       string
	Photo      string
}

func StudentProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID         int
			Name       string
			IdentityNo string
			ClassName  string
			Phone      string
			RFID       string
			Photo      string
		}

		err := db.QueryRow(`
			SELECT s.id, s.name, s.nis, c.name, s.parent_phone, s.rfid_uid, s.photo
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, id).Scan(&s.ID, &s.Name, &s.IdentityNo, &s.ClassName, &s.Phone, &s.RFID, &s.Photo)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin")
		}

		data := ProfileData{
			Type:       "Siswa",
			ID:         s.ID,
			Name:       s.Name,
			IdentityNo: s.IdentityNo,
			ExtraInfo:  s.ClassName,
			Phone:      s.Phone,
			RFID:       s.RFID,
			Photo:      s.Photo,
		}
		return c.Render(http.StatusOK, "profile.html", data)
	}
}

func StaffProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID         int
			Name       string
			IdentityNo string
			Role       string
			Phone      string
			RFID       string
		}

		err := db.QueryRow(`
			SELECT id, name, nip, role, phone, rfid_uid
			FROM staff WHERE id = ?`, id).Scan(&s.ID, &s.Name, &s.IdentityNo, &s.Role, &s.Phone, &s.RFID)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin")
		}

		data := ProfileData{
			Type:       "Guru/Staff",
			ID:         s.ID,
			Name:       s.Name,
			IdentityNo: s.IdentityNo,
			ExtraInfo:  s.Role,
			Phone:      s.Phone,
			RFID:       s.RFID,
		}
		return c.Render(http.StatusOK, "profile.html", data)
	}
}

func FaceRegisterPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID         int
			Name       string
			ExtraInfo  string
			Photo      string
			IdentityNo string
			RFID       string
			Phone      string
		}

		err := db.QueryRow(`
			SELECT s.id, s.name, c.name, s.photo, s.nis, s.rfid_uid, s.parent_phone
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, id).Scan(&s.ID, &s.Name, &s.ExtraInfo, &s.Photo, &s.IdentityNo, &s.RFID, &s.Phone)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/students")
		}

		data := map[string]interface{}{
			"ID":         s.ID,
			"Name":       s.Name,
			"ExtraInfo":  s.ExtraInfo,
			"Photo":      s.Photo,
			"IdentityNo": s.IdentityNo,
			"RFID":       s.RFID,
			"Phone":      s.Phone,
			"Type":       "Siswa",
		}

		return c.Render(http.StatusOK, "face_register.html", data)
	}
}
