package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// AdminGlobalSearch searches students, staff, subjects, and assets in one call.
func AdminGlobalSearch(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := strings.TrimSpace(c.QueryParam("q"))
		if len(q) < 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Minimal 2 karakter"})
		}
		like := "%" + q + "%"
		out := map[string]interface{}{"students": []interface{}{}, "staff": []interface{}{}, "subjects": []interface{}{}, "assets": []interface{}{}}

		rows, err := db.Query(`SELECT id, nis, name, class_id FROM students WHERE name LIKE ? OR nis LIKE ? LIMIT 8`, like, like)
		if err == nil {
			for rows.Next() {
				var id, classID int
				var nis, name string
				if rows.Scan(&id, &nis, &name, &classID) == nil {
					var className string
					_ = db.QueryRow("SELECT name FROM classes WHERE id=?", classID).Scan(&className)
					out["students"] = append(out["students"].([]interface{}), map[string]interface{}{"id": id, "label": name, "sub": "NIS " + nis + " • " + className, "url": "/admin/student/" + strconv.Itoa(id)})
				}
			}
			rows.Close()
		}

		rows, err = db.Query(`SELECT id, nip, name FROM staff WHERE name LIKE ? OR nip LIKE ? LIMIT 8`, like, like)
		if err == nil {
			for rows.Next() {
				var id int
				var nip, name string
				if rows.Scan(&id, &nip, &name) == nil {
					out["staff"] = append(out["staff"].([]interface{}), map[string]interface{}{"id": id, "label": name, "sub": "NIP " + nip, "url": ""})
				}
			}
			rows.Close()
		}

		rows, err = db.Query(`SELECT id, code, name FROM subjects WHERE name LIKE ? OR code LIKE ? LIMIT 8`, like, like)
		if err == nil {
			for rows.Next() {
				var id int
				var code, name string
				if rows.Scan(&id, &code, &name) == nil {
					out["subjects"] = append(out["subjects"].([]interface{}), map[string]interface{}{"id": id, "label": name, "sub": "Kode " + code, "url": "/admin/academics"})
				}
			}
			rows.Close()
		}

		rows, err = db.Query(`SELECT id, inventory_code, name FROM assets WHERE name LIKE ? OR inventory_code LIKE ? LIMIT 8`, like, like)
		if err == nil {
			for rows.Next() {
				var id int
				var code, name string
				if rows.Scan(&id, &code, &name) == nil {
					out["assets"] = append(out["assets"].([]interface{}), map[string]interface{}{"id": id, "label": name, "sub": code, "url": "/admin/sarpras?tab=assets"})
				}
			}
			rows.Close()
		}
		return c.JSON(http.StatusOK, out)
	}
}

// AdminLMSAnalytics returns aggregate LMS stats for the admin dashboard.
func AdminLMSAnalytics(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		stats := map[string]interface{}{}
		counts := []struct {
			key   string
			query string
		}{
			{"total_assignments", `SELECT COUNT(*) FROM teacher_assignments`},
			{"total_meetings", `SELECT COUNT(*) FROM lms_meetings`},
			{"total_materials", `SELECT COUNT(*) FROM lms_materials`},
			{"total_tasks", `SELECT COUNT(*) FROM lms_assignments`},
			{"total_quizzes", `SELECT COUNT(*) FROM lms_quizzes`},
			{"total_discussions", `SELECT COUNT(*) FROM lms_discussions`},
			{"journals_draft", `SELECT COUNT(*) FROM teacher_journals WHERE status='draft'`},
			{"journals_submitted", `SELECT COUNT(*) FROM teacher_journals WHERE status='submitted'`},
			{"journals_approved", `SELECT COUNT(*) FROM teacher_journals WHERE status='approved'`},
			{"journals_rejected", `SELECT COUNT(*) FROM teacher_journals WHERE status='rejected'`},
			{"submissions_graded", `SELECT COUNT(*) FROM lms_submissions WHERE score IS NOT NULL`},
			{"quiz_attempts", `SELECT COUNT(*) FROM lms_quiz_attempts WHERE completed_at IS NOT NULL`},
		}
		for _, item := range counts {
			var n int
			if db.QueryRow(item.query).Scan(&n) == nil {
				stats[item.key] = n
			}
		}
		return c.JSON(http.StatusOK, stats)
	}
}
