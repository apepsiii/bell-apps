package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/pkg/utils"
)

// ChildInfo represents a child linked to a parent's phone number.
type ChildInfo struct {
	ID        int    `json:"id"`
	NIS       string `json:"nis"`
	Name      string `json:"name"`
	ClassName string `json:"class_name"`
	Photo     string `json:"photo"`
}

// GetParentInfo returns info about the logged-in parent and all linked children.
func GetParentInfo(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(ParentSessionCookie)
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"is_parent": false,
			})
		}

		var (
			parentPhone string
			parentName  string
			studentID   int
			expiresAt   time.Time
		)
		err = db.QueryRow(`
			SELECT parent_phone, parent_name, student_id, expires_at 
			FROM parent_sessions 
			WHERE token = ?
		`, cookie.Value).Scan(&parentPhone, &parentName, &studentID, &expiresAt)

		if err != nil || time.Now().After(expiresAt) {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"is_parent": false,
			})
		}

		cleanPhone := utils.FormatPhone(parentPhone)
		alt1 := "0" + strings.TrimPrefix(cleanPhone, "62")
		alt2 := "+" + cleanPhone

		rows, err := db.Query(`
			SELECT s.id, s.nis, s.name, COALESCE(c.name, '-'), COALESCE(s.photo, '')
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE (s.parent_phone = ? OR s.parent_phone = ? OR s.parent_phone = ?)
			  AND COALESCE(s.status, 'active') != 'inactive'
			ORDER BY s.id ASC
		`, cleanPhone, alt1, alt2)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database query error"})
		}
		defer rows.Close()

		children := make([]ChildInfo, 0)
		for rows.Next() {
			var ch ChildInfo
			if err := rows.Scan(&ch.ID, &ch.NIS, &ch.Name, &ch.ClassName, &ch.Photo); err == nil {
				children = append(children, ch)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"is_parent":         true,
			"parent_name":       parentName,
			"parent_phone":      parentPhone,
			"active_student_id": studentID,
			"children":          children,
		})
	}
}

// SwitchActiveChild switches the active student for a parent with multiple children.
func SwitchActiveChild(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(ParentSessionCookie)
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Sesi orang tua tidak ditemukan"})
		}

		var req struct {
			StudentID int `json:"student_id"`
		}
		if err := c.Bind(&req); err != nil || req.StudentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID siswa tidak valid"})
		}

		var (
			parentPhone string
			expiresAt   time.Time
		)
		err = db.QueryRow(`
			SELECT parent_phone, expires_at 
			FROM parent_sessions 
			WHERE token = ?
		`, cookie.Value).Scan(&parentPhone, &expiresAt)

		if err != nil || time.Now().After(expiresAt) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Sesi telah berakhir, silakan login kembali"})
		}

		cleanPhone := utils.FormatPhone(parentPhone)
		alt1 := "0" + strings.TrimPrefix(cleanPhone, "62")
		alt2 := "+" + cleanPhone

		// Security Check: Verify target student is indeed linked to this parent's phone!
		var (
			targetID   int
			targetName string
		)
		err = db.QueryRow(`
			SELECT id, name 
			FROM students 
			WHERE id = ? 
			  AND (parent_phone = ? OR parent_phone = ? OR parent_phone = ?)
			  AND COALESCE(status, 'active') != 'inactive'
		`, req.StudentID, cleanPhone, alt1, alt2).Scan(&targetID, &targetName)

		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Anda tidak memiliki akses ke data siswa ini"})
		}

		// Update parent_sessions
		_, _ = db.Exec("UPDATE parent_sessions SET student_id = ? WHERE token = ?", targetID, cookie.Value)

		// Update student_sessions if student_session cookie exists
		if sCookie, err := c.Cookie(StudentSessionCookie); err == nil && sCookie.Value != "" {
			_, _ = db.Exec("UPDATE student_sessions SET student_id = ? WHERE token = ?", targetID, sCookie.Value)
			_, _ = db.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
				studentSessionPrefix+sCookie.Value, targetID)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":            "success",
			"active_student_id": targetID,
			"student_name":      targetName,
			"message":           "Berhasil beralih ke data " + targetName,
		})
	}
}

