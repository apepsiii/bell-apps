package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ResetStudentPoints resets ALL points for a single student
// Deletes all student_points records, achievement/violation points, claims, and resets english_streaks & semester_scores
func ResetStudentPoints(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Param("id")
		if studentID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id wajib diisi"})
		}

		// Get student name for audit
		var studentName string
		err := db.QueryRow("SELECT name FROM students WHERE id = ?", studentID).Scan(&studentName)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Siswa tidak ditemukan"})
		}

		adminName := c.Get("admin_name")
		if adminName == nil {
			adminName = "Admin"
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction failed"})
		}

		// 1. Log the current total before reset (for audit trail)
		var currentTotal int
		tx.QueryRow("SELECT COALESCE(SUM(points_change), 0) FROM student_points WHERE student_id = ?", studentID).Scan(&currentTotal)
		var englishXP int
		tx.QueryRow("SELECT COALESCE(total_xp, 0) FROM english_streaks WHERE student_id = ?", studentID).Scan(&englishXP)
		currentTotal += englishXP

		// 2. Delete all student_points records
		_, _ = tx.Exec("DELETE FROM student_points WHERE student_id = ?", studentID)

		// 3. Delete all student_achievement_points
		_, _ = tx.Exec("DELETE FROM student_achievement_points WHERE student_id = ?", studentID)

		// 4. Delete all student_violation_points
		_, _ = tx.Exec("DELETE FROM student_violation_points WHERE student_id = ?", studentID)

		// 5. Reset english_streaks XP and streaks
		_, _ = tx.Exec("UPDATE english_streaks SET total_xp = 0, current_streak = 0, longest_streak = 0 WHERE student_id = ?", studentID)

		// 6. Delete all point_claims
		_, _ = tx.Exec("DELETE FROM point_claims WHERE student_id = ?", studentID)

		// 7. Delete all pending_achievement_points
		_, _ = tx.Exec("DELETE FROM pending_achievement_points WHERE student_id = ?", studentID)

		// 8. Delete all violation_redemptions
		_, _ = tx.Exec("DELETE FROM violation_redemptions WHERE student_id = ?", studentID)

		// 9. Reset semester_scores
		_, _ = tx.Exec(`
			UPDATE semester_scores 
			SET achievement_points = 0, achievement_normalized = 0,
			    violation_points = 0, violation_burden = 0,
			    redemption_points = 0, composite_score = 0
			WHERE student_id = ?
		`, studentID)

		// 10. Reset attendance_streaks
		_, _ = tx.Exec(`
			UPDATE attendance_streaks 
			SET current_streak = 0, longest_streak = 0, last_attendance_date = NULL
			WHERE student_id = ?
		`, studentID)

		// 11. Log to audit trail (if point_audit_log exists)
		_, _ = tx.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, old_data, new_data, performed_by, performed_at, reason)
			VALUES ('RESET_POINTS', 'student_points', ?, ?, ?, ?, ?, ?)
		`, studentID,
			fmt.Sprintf(`{"total_points": %d}`, currentTotal),
			`{"total_points": 0}`,
			adminName, time.Now(),
			fmt.Sprintf("Reset poin siswa: %s", studentName))

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan perubahan reset"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Poin siswa %s berhasil direset ke 0 (sebelumnya: %d poin)", studentName, currentTotal),
		})
	}
}

// BulkResetPointsRequest for bulk reset
type BulkResetPointsRequest struct {
	ClassID    int    `json:"class_id"`
	StudentIDs []int  `json:"student_ids"` // optional: if empty, reset all in class
	ResetBy    string `json:"reset_by"`
}

// BulkResetPoints resets points for multiple students (by class or by IDs)
func BulkResetPoints(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkResetPointsRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.ClassID == 0 && len(req.StudentIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "class_id atau student_ids wajib diisi"})
		}

		// Build student list
		var studentIDs []int
		if len(req.StudentIDs) > 0 {
			studentIDs = req.StudentIDs
		} else {
			rows, err := db.Query("SELECT id FROM students WHERE class_id = ? AND status = 'active'", req.ClassID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			defer rows.Close()
			for rows.Next() {
				var id int
				rows.Scan(&id)
				studentIDs = append(studentIDs, id)
			}
		}

		if len(studentIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tidak ada siswa ditemukan"})
		}

		if req.ResetBy == "" {
			req.ResetBy = "Admin"
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction failed"})
		}

		successCount := 0
		for _, sid := range studentIDs {
			// Delete student_points
			_, _ = tx.Exec("DELETE FROM student_points WHERE student_id = ?", sid)

			// Delete student_achievement_points
			_, _ = tx.Exec("DELETE FROM student_achievement_points WHERE student_id = ?", sid)

			// Delete student_violation_points
			_, _ = tx.Exec("DELETE FROM student_violation_points WHERE student_id = ?", sid)

			// Reset english_streaks
			_, _ = tx.Exec("UPDATE english_streaks SET total_xp = 0, current_streak = 0, longest_streak = 0 WHERE student_id = ?", sid)

			// Delete point_claims
			_, _ = tx.Exec("DELETE FROM point_claims WHERE student_id = ?", sid)

			// Delete pending_achievement_points
			_, _ = tx.Exec("DELETE FROM pending_achievement_points WHERE student_id = ?", sid)

			// Delete violation_redemptions
			_, _ = tx.Exec("DELETE FROM violation_redemptions WHERE student_id = ?", sid)

			// Reset semester_scores
			_, _ = tx.Exec(`
				UPDATE semester_scores 
				SET achievement_points = 0, achievement_normalized = 0,
				    violation_points = 0, violation_burden = 0,
				    redemption_points = 0, composite_score = 0
				WHERE student_id = ?
			`, sid)

			// Reset attendance_streaks
			_, _ = tx.Exec(`
				UPDATE attendance_streaks 
				SET current_streak = 0, longest_streak = 0, last_attendance_date = NULL
				WHERE student_id = ?
			`, sid)

			successCount++
		}

		// Log to audit trail
		_, _ = tx.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, old_data, new_data, performed_by, performed_at, reason)
			VALUES ('BULK_RESET_POINTS', 'student_points', 0, ?, ?, ?, ?, ?)
		`,
			fmt.Sprintf(`{"student_count": %d}`, len(studentIDs)),
			`{"status": "reset_complete"}`,
			req.ResetBy, time.Now(),
			fmt.Sprintf("Bulk reset %d siswa", successCount))

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan perubahan bulk reset"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("%d siswa berhasil direset poinnya ke 0", successCount),
			"count":   successCount,
		})
	}
}

// AdjustPointsRequest for adjusting (adding/subtracting) points
type AdjustPointsRequest struct {
	StudentID   int    `json:"student_id"`
	Points      int    `json:"points"`    // negative = subtract, positive = add
	Description string `json:"description"`
	AdjustedBy  string `json:"adjusted_by"`
}

// AdjustStudentPoints adjusts points for a single student (add or subtract)
func AdjustStudentPoints(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AdjustPointsRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.StudentID == 0 || req.Points == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id dan points wajib diisi (points != 0)"})
		}

		// Get student name
		var studentName string
		err := db.QueryRow("SELECT name FROM students WHERE id = ?", req.StudentID).Scan(&studentName)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Siswa tidak ditemukan"})
		}

		if req.AdjustedBy == "" {
			req.AdjustedBy = "Admin"
		}

		if req.Description == "" {
			if req.Points > 0 {
				req.Description = fmt.Sprintf("Penambahan poin manual (+%d)", req.Points)
			} else {
				req.Description = fmt.Sprintf("Pengurangan poin manual (%d)", req.Points)
			}
		}

		// Insert adjustment as a student_points record
		_, err = db.Exec(`
			INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by)
			VALUES (?, ?, ?, ?, ?)
		`, req.StudentID, req.Points, req.Description, time.Now(), req.AdjustedBy)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan: " + err.Error()})
		}

		// Log to audit trail
		db.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, old_data, new_data, performed_by, performed_at, reason)
			VALUES ('ADJUST_POINTS', 'student_points', ?, ?, ?, ?, ?, ?)
		`, req.StudentID,
			`{}`,
			fmt.Sprintf(`{"points_change": %d, "description": "%s"}`, req.Points, req.Description),
			req.AdjustedBy, time.Now(),
			fmt.Sprintf("Adjust poin siswa: %s", studentName))

		action := "ditambah"
		if req.Points < 0 {
			action = "dikurangi"
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Poin %s %s %d poin", studentName, action, abs(req.Points)),
		})
	}
}

// BulkAdjustPointsRequest for bulk adjusting points
type BulkAdjustPointsRequest struct {
	ClassID    int    `json:"class_id"`
	StudentIDs []int  `json:"student_ids"` // optional
	Points     int    `json:"points"`      // negative = subtract, positive = add
	Description string `json:"description"`
	AdjustedBy string `json:"adjusted_by"`
}

// BulkAdjustPoints adjusts points for multiple students
func BulkAdjustPoints(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkAdjustPointsRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.Points == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "points tidak boleh 0"})
		}

		if req.ClassID == 0 && len(req.StudentIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "class_id atau student_ids wajib diisi"})
		}

		// Build student list
		var studentIDs []int
		if len(req.StudentIDs) > 0 {
			studentIDs = req.StudentIDs
		} else {
			rows, err := db.Query("SELECT id FROM students WHERE class_id = ? AND status = 'active'", req.ClassID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			defer rows.Close()
			for rows.Next() {
				var id int
				rows.Scan(&id)
				studentIDs = append(studentIDs, id)
			}
		}

		if len(studentIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tidak ada siswa ditemukan"})
		}

		if req.AdjustedBy == "" {
			req.AdjustedBy = "Admin"
		}

		if req.Description == "" {
			if req.Points > 0 {
				req.Description = fmt.Sprintf("Penambahan poin bulk (+%d)", req.Points)
			} else {
				req.Description = fmt.Sprintf("Pengurangan poin bulk (%d)", req.Points)
			}
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction failed"})
		}

		successCount := 0
		for _, sid := range studentIDs {
			_, err := tx.Exec(`
				INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by)
				VALUES (?, ?, ?, ?, ?)
			`, sid, req.Points, req.Description, time.Now(), req.AdjustedBy)
			if err == nil {
				successCount++
			}
		}

		// Log to audit trail
		tx.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, old_data, new_data, performed_by, performed_at, reason)
			VALUES ('BULK_ADJUST_POINTS', 'student_points', 0, ?, ?, ?, ?, ?)
		`,
			`{}`,
			fmt.Sprintf(`{"student_count": %d, "points": %d, "description": "%s"}`, successCount, req.Points, req.Description),
			req.AdjustedBy, time.Now(),
			fmt.Sprintf("Bulk adjust %d siswa", successCount))

		tx.Commit()

		action := "ditambah"
		if req.Points < 0 {
			action = "dikurangi"
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("%d siswa %s %d poin", successCount, action, abs(req.Points)),
			"count":   successCount,
		})
	}
}

// DeletePointLog deletes a single point transaction log
func DeletePointLog(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		logID := c.Param("id")
		if logID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "log_id wajib diisi"})
		}

		adminName := c.Get("admin_name")
		if adminName == nil {
			adminName = "Admin"
		}

		// Get log details before delete
		var studentID, pointsChange int
		var description string
		err := db.QueryRow("SELECT student_id, points_change, description FROM student_points WHERE id = ?", logID).
			Scan(&studentID, &pointsChange, &description)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Log tidak ditemukan"})
		}

		_, err = db.Exec("DELETE FROM student_points WHERE id = ?", logID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menghapus: " + err.Error()})
		}

		// Audit trail
		db.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, old_data, new_data, performed_by, performed_at, reason)
			VALUES ('DELETE_POINT_LOG', 'student_points', ?, ?, ?, ?, ?, ?)
		`, logID,
			fmt.Sprintf(`{"student_id": %d, "points_change": %d, "description": "%s"}`, studentID, pointsChange, description),
			`{"status": "deleted"}`,
			adminName, time.Now(),
			"Hapus log poin individual")

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Log poin berhasil dihapus",
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
