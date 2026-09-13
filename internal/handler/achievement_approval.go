package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// AchievementSubmitRequest for submitting achievement points
type AchievementSubmitRequest struct {
	StudentID   int    `json:"student_id"`
	RuleID      int    `json:"rule_id"`
	Description string `json:"description"`
	ProofURL    string `json:"proof_url"` // URL foto/sertifikat bukti
	RequestedBy string `json:"requested_by"`
}

// SubmitAchievement submits achievement to pending_achievement_points
func SubmitAchievement(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AchievementSubmitRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.StudentID == 0 || req.RuleID == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id dan rule_id wajib diisi"})
		}

		// Get rule details and points
		var ruleCode, ruleName, category string
		var points int
		err := db.QueryRow(`
			SELECT code, name, category, points 
			FROM achievement_rules 
			WHERE id = ? AND is_active = 1
		`, req.RuleID).Scan(&ruleCode, &ruleName, &category, &points)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Rule tidak ditemukan atau tidak aktif"})
		}

		// Check APPROVAL_THRESHOLD
		var approvalThreshold int
		db.QueryRow("SELECT config_value FROM point_config WHERE config_key = 'APPROVAL_THRESHOLD'").Scan(&approvalThreshold)
		if approvalThreshold == 0 {
			approvalThreshold = 50 // default
		}

		// If points >= threshold and no proof, reject
		if points >= approvalThreshold && req.ProofURL == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Prestasi dengan poin >= %d wajib melampirkan bukti (foto/sertifikat)", approvalThreshold),
			})
		}

		// Get current semester
		var semesterID int
		err = db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&semesterID)
		if err != nil {
			semesterID = 1 // fallback
		}

		// Check category cap: max 200 points per category per semester
		var categoryPoints int
		db.QueryRow(`
			SELECT COALESCE(SUM(sp.points_change), 0)
			FROM student_points sp
			JOIN achievement_rules ar ON sp.rule_id = ar.id
			WHERE sp.student_id = ? AND ar.category = ?
			  AND strftime('%Y', sp.timestamp) = strftime('%Y', 'now')
		`, req.StudentID, category).Scan(&categoryPoints)

		var maxPerCategory int
		db.QueryRow("SELECT config_value FROM point_config WHERE config_key = 'MAX_POINTS_PER_CATEGORY'").Scan(&maxPerCategory)
		if maxPerCategory == 0 {
			maxPerCategory = 200
		}

		if categoryPoints+points > maxPerCategory {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":          fmt.Sprintf("Kategori '%s' sudah mencapai batas maksimal %d poin per semester", category, maxPerCategory),
				"current_points": fmt.Sprintf("%d", categoryPoints),
				"max_points":     fmt.Sprintf("%d", maxPerCategory),
			})
		}

		// Build description
		description := req.Description
		if description == "" {
			description = ruleName
		}

		// Insert into pending_achievement_points
		_, err = db.Exec(`
			INSERT INTO pending_achievement_points 
			(student_id, rule_id, points, description, proof_url, requested_by, requested_at, status, semester_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?)
		`, req.StudentID, req.RuleID, points, description, req.ProofURL, req.RequestedBy, time.Now(), semesterID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan pengajuan: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Pengajuan prestasi berhasil dikirim untuk approval",
			"data": map[string]interface{}{
				"rule_code":   ruleCode,
				"rule_name":   ruleName,
				"category":    category,
				"points":      points,
				"description": description,
			},
		})
	}
}

// ApproveAchievementRequest for approving/rejecting achievements
type ApproveAchievementRequest struct {
	PendingID      int    `json:"pending_id"`
	Action         string `json:"action"` // "approve" or "reject"
	ApprovedBy     string `json:"approved_by"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// ApproveAchievement approves or rejects pending achievement
func ApproveAchievement(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ApproveAchievementRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.Action != "approve" && req.Action != "reject" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Action harus 'approve' atau 'reject'"})
		}

		// Get pending data
		var studentID, ruleID, points, semesterID int
		var description, status, proofURL string
		err := db.QueryRow(`
			SELECT student_id, rule_id, points, description, status, proof_url, semester_id
			FROM pending_achievement_points 
			WHERE id = ?
		`, req.PendingID).Scan(&studentID, &ruleID, &points, &description, &status, &proofURL, &semesterID)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Pengajuan tidak ditemukan"})
		}

		if status != "pending" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Pengajuan sudah diproses sebelumnya"})
		}

		if req.Action == "approve" {
			// Approve: move to student_points
			tx, err := db.Begin()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction failed"})
			}

			// Insert to student_points
			_, err = tx.Exec(`
				INSERT INTO student_points 
				(student_id, rule_id, points_change, description, timestamp, recorded_by)
				VALUES (?, ?, ?, ?, ?, ?)
			`, studentID, ruleID, points, description, time.Now(), req.ApprovedBy)

			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan poin: " + err.Error()})
			}

			// Update pending status
			_, err = tx.Exec(`
				UPDATE pending_achievement_points 
				SET status = 'approved', approved_by = ?, approved_at = ?
				WHERE id = ?
			`, req.ApprovedBy, time.Now(), req.PendingID)

			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal update status"})
			}

			// Update semester_scores achievement_points
			_, err = tx.Exec(`
				UPDATE semester_scores 
				SET achievement_points = achievement_points + ?
				WHERE student_id = ? AND semester_id = ?
			`, points, studentID, semesterID)

			// If no row exists, insert
			if err != nil || tx.Commit() != nil {
				// Try insert
				db.Exec(`
					INSERT INTO semester_scores (student_id, semester_id, achievement_points)
					VALUES (?, ?, ?)
					ON CONFLICT(student_id, semester_id) DO UPDATE SET
						achievement_points = achievement_points + ?
				`, studentID, semesterID, points, points)
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": fmt.Sprintf("Prestasi disetujui, +%d poin ditambahkan", points),
			})

		} else {
			// Reject
			_, err = db.Exec(`
				UPDATE pending_achievement_points 
				SET status = 'rejected', rejected_by = ?, rejected_at = ?, rejection_reason = ?
				WHERE id = ?
			`, req.ApprovedBy, time.Now(), req.RejectionReason, req.PendingID)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal reject"})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": "Pengajuan ditolak",
			})
		}
	}
}

// GetPendingAchievements returns list of pending achievements
func GetPendingAchievements(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status") // pending, approved, rejected
		if status == "" {
			status = "pending"
		}

		query := `
			SELECT p.id, p.student_id, s.name as student_name, s.nis,
			       ar.code, ar.name as rule_name, ar.category, p.points,
			       p.description, p.proof_url, p.requested_by, p.requested_at,
			       p.status, COALESCE(p.approved_by, '') as approved_by,
			       COALESCE(p.rejected_by, '') as rejected_by,
			       COALESCE(p.rejection_reason, '') as rejection_reason
			FROM pending_achievement_points p
			JOIN students s ON p.student_id = s.id
			JOIN achievement_rules ar ON p.rule_id = ar.id
			WHERE p.status = ?
			ORDER BY p.requested_at DESC
		`
		rows, err := db.Query(query, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var id, studentID, points int
			var studentName, nis, ruleCode, ruleName, category, description, proofURL, requestedBy, requestedAt, statusVal, approvedBy, rejectedBy, rejectionReason string

			rows.Scan(&id, &studentID, &studentName, &nis, &ruleCode, &ruleName, &category, &points,
				&description, &proofURL, &requestedBy, &requestedAt, &statusVal, &approvedBy, &rejectedBy, &rejectionReason)

			results = append(results, map[string]interface{}{
				"id":               id,
				"student_id":       studentID,
				"student_name":     studentName,
				"nis":              nis,
				"rule_code":        ruleCode,
				"rule_name":        ruleName,
				"category":         category,
				"points":           points,
				"description":      description,
				"proof_url":        proofURL,
				"requested_by":     requestedBy,
				"requested_at":     requestedAt,
				"status":           statusVal,
				"approved_by":      approvedBy,
				"rejected_by":      rejectedBy,
				"rejection_reason": rejectionReason,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"total":   len(results),
			"results": results,
		})
	}
}
