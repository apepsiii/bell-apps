package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ViolationRecordRequest for recording a violation
type ViolationRecordRequest struct {
	StudentID   int    `json:"student_id"`
	RuleID      int    `json:"rule_id"`
	Description string `json:"description"`
	RecordedBy  string `json:"recorded_by"`
}

// RecordViolation records a student violation with automatic escalation
func RecordViolation(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ViolationRecordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.StudentID == 0 || req.RuleID == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id dan rule_id wajib diisi"})
		}

		// Get rule details
		var ruleCode, ruleName, category, categoryCode string
		var points1, points2, points3 int
		err := db.QueryRow(`
			SELECT code, name, category, category_code, points_1, points_2, points_3
			FROM violation_rules
			WHERE id = ? AND is_active = 1
		`, req.RuleID).Scan(&ruleCode, &ruleName, &category, &categoryCode, &points1, &points2, &points3)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Rule pelanggaran tidak ditemukan"})
		}

		// Get current semester
		var semesterID int
		err = db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&semesterID)
		if err != nil {
			semesterID = 1
		}

		// Count previous violations of the SAME rule for this student in this semester
		// to determine escalation level (1, 2, or 3)
		var prevCount int
		db.QueryRow(`
			SELECT COUNT(*)
			FROM student_violation_points
			WHERE student_id = ? AND rule_id = ? AND semester_id = ?
			  AND deleted_at IS NULL
		`, req.StudentID, req.RuleID, semesterID).Scan(&prevCount)

		// Escalation: 1st offense = level 1, 2nd = level 2, 3rd+ = level 3
		escalationLevel := prevCount + 1
		if escalationLevel > 3 {
			escalationLevel = 3
		}

		// Determine points based on escalation level
		var points int
		switch escalationLevel {
		case 1:
			points = points1
		case 2:
			points = points2
		case 3:
			points = points3
		}

		// Build description
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("%s (Pelanggaran ke-%d)", ruleName, escalationLevel)
		} else {
			description = fmt.Sprintf("%s — %s (Pelanggaran ke-%d)", ruleName, description, escalationLevel)
		}

		// Insert into student_violation_points
		var violationID int
		err = db.QueryRow(`
			INSERT INTO student_violation_points
			(student_id, rule_id, points, escalation_level, description, recorded_by, recorded_at, semester_id, redemption_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'none')
		`, req.StudentID, req.RuleID, points, escalationLevel, description,
			req.RecordedBy, time.Now(), semesterID).Scan(&violationID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal mencatat pelanggaran: " + err.Error()})
		}

		// Update semester_scores violation_points
		_, err = db.Exec(`
			UPDATE semester_scores
			SET violation_points = violation_points + ?
			WHERE student_id = ? AND semester_id = ?
		`, points, req.StudentID, semesterID)

		if err != nil {
			// Try insert if no row exists
			db.Exec(`
				INSERT INTO semester_scores (student_id, semester_id, violation_points)
				VALUES (?, ?, ?)
				ON CONFLICT(student_id, semester_id) DO UPDATE SET
					violation_points = violation_points + ?
			`, req.StudentID, semesterID, points, points)
		}

		// Recalculate violation_burden and composite_score
		recalculateViolationBurden(db, req.StudentID, semesterID)

		// Check SP threshold for early warning
		var totalViolationPoints int
		db.QueryRow(`
			SELECT COALESCE(violation_points, 0)
			FROM semester_scores
			WHERE student_id = ? AND semester_id = ?
		`, req.StudentID, semesterID).Scan(&totalViolationPoints)

		spWarning := checkSPThreshold(db, totalViolationPoints)

		// Log to audit trail
		db.Exec(`
			INSERT INTO point_audit_log (action, table_name, record_id, new_data, performed_by, performed_at, reason)
			VALUES ('VIOLATION_RECORDED', 'student_violation_points', ?, ?, ?, ?, ?)
		`, violationID, description, req.RecordedBy, time.Now(), fmt.Sprintf("Pelanggaran %s eskalasi %d", ruleCode, escalationLevel))

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Pelanggaran dicatat: %s (+%d poin)", ruleName, points),
			"data": map[string]interface{}{
				"violation_id":      violationID,
				"rule_code":         ruleCode,
				"rule_name":         ruleName,
				"category":          category,
				"escalation_level":  escalationLevel,
				"points":            points,
				"total_violation":   totalViolationPoints,
				"sp_warning":        spWarning,
			},
		})
	}
}

// RedemptionRequest for submitting a redemption (kerja positif)
type RedemptionRequest struct {
	ViolationID         int    `json:"violation_id"`
	StudentID           int    `json:"student_id"`
	RedemptionType     string `json:"redemption_type"`     // e.g. "piket", "mentoring", "kerja_sosial"
	RedemptionDescription string `json:"redemption_description"`
	PointsToRedeem      int    `json:"points_to_redeem"`   // how many points to redeem
	VerifiedBy          string `json:"verified_by"`
}

// SubmitRedemption submits a redemption request (kerja positif to pay off violation)
func SubmitRedemption(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RedemptionRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.ViolationID == 0 || req.StudentID == 0 || req.PointsToRedeem <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "violation_id, student_id, dan points_to_redeem wajib diisi"})
		}

		// Get violation details
		var violationPoints, semesterID int
		var redemptionStatus string
		var redemptionPoints int
		err := db.QueryRow(`
			SELECT points, semester_id, redemption_status, redemption_points
			FROM student_violation_points
			WHERE id = ? AND student_id = ? AND deleted_at IS NULL
		`, req.ViolationID, req.StudentID).Scan(&violationPoints, &semesterID, &redemptionStatus, &redemptionPoints)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Pelanggaran tidak ditemukan"})
		}

		// Check if already fully redeemed
		remaining := violationPoints - redemptionPoints
		if remaining <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Pelanggaran ini sudah dilunasi sepenuhnya"})
		}

		// Cap points to redeem at remaining
		if req.PointsToRedeem > remaining {
			req.PointsToRedeem = remaining
		}

		// Get current semester
		if semesterID == 0 {
			db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&semesterID)
		}

		// Insert redemption record
		_, err = db.Exec(`
			INSERT INTO violation_redemptions
			(student_id, semester_id, violation_id, redemption_type, redemption_description, points_redeemed, verified_by, verified_at, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
		`, req.StudentID, semesterID, req.ViolationID, req.RedemptionType, req.RedemptionDescription,
			req.PointsToRedeem, req.VerifiedBy, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan pengajuan pemulihan: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Pengajuan pemulihan %d poin berhasil dikirim untuk verifikasi", req.PointsToRedeem),
			"data": map[string]interface{}{
				"violation_id":    req.ViolationID,
				"points_to_redeem": req.PointsToRedeem,
				"remaining":       remaining - req.PointsToRedeem,
			},
		})
	}
}

// VerifyRedemption verifies and applies a redemption (approve/reject)
type VerifyRedemptionRequest struct {
	RedemptionID int    `json:"redemption_id"`
	Action       string `json:"action"` // "approve" or "reject"
	VerifiedBy   string `json:"verified_by"`
	RejectReason string `json:"reject_reason,omitempty"`
}

// VerifyRedemption verifies and applies a redemption
func VerifyRedemption(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req VerifyRedemptionRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.Action != "approve" && req.Action != "reject" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Action harus 'approve' atau 'reject'"})
		}

		// Get redemption details
		var studentID, violationID, semesterID, pointsRedeemed int
		var status string
		err := db.QueryRow(`
			SELECT student_id, violation_id, semester_id, points_redeemed, status
			FROM violation_redemptions
			WHERE id = ?
		`, req.RedemptionID).Scan(&studentID, &violationID, &semesterID, &pointsRedeemed, &status)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Pengajuan pemulihan tidak ditemukan"})
		}

		if status != "pending" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Pengajuan sudah diproses"})
		}

		if req.Action == "approve" {
			tx, err := db.Begin()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction failed"})
			}

			// Update redemption status
			_, err = tx.Exec(`
				UPDATE violation_redemptions
				SET status = 'approved', verified_by = ?, verified_at = ?
				WHERE id = ?
			`, req.VerifiedBy, time.Now(), req.RedemptionID)
			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal update status"})
			}

			// Update violation: add redemption_points
			_, err = tx.Exec(`
				UPDATE student_violation_points
				SET redemption_points = redemption_points + ?,
				    redemption_status = CASE
				        WHEN redemption_points + ? >= points THEN 'fully_redeemed'
				        ELSE 'partially_redeemed'
				    END
				WHERE id = ?
			`, pointsRedeemed, pointsRedeemed, violationID)
			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal update pelanggaran"})
			}

			// Update semester_scores: add redemption_points
			_, err = tx.Exec(`
				UPDATE semester_scores
				SET redemption_points = redemption_points + ?
				WHERE student_id = ? AND semester_id = ?
			`, pointsRedeemed, studentID, semesterID)
			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal update skor"})
			}

			tx.Commit()

			// Recalculate violation_burden
			recalculateViolationBurden(db, studentID, semesterID)

			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": fmt.Sprintf("Pemulihan disetujui, %d poin dilunasi", pointsRedeemed),
			})
		} else {
			// Reject
			_, err = db.Exec(`
				UPDATE violation_redemptions
				SET status = 'rejected', verified_by = ?, verified_at = ?
				WHERE id = ?
			`, req.VerifiedBy, time.Now(), req.RedemptionID)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal reject"})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": "Pengajuan pemulihan ditolak",
			})
		}
	}
}

// GetStudentViolations returns violation history for a student
func GetStudentViolations(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.QueryParam("student_id")
		semesterID := c.QueryParam("semester_id")

		if studentID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id wajib diisi"})
		}

		if semesterID == "" {
			semesterID = "1"
		}

		query := `
			SELECT svp.id, svp.student_id, svp.rule_id, svp.points, svp.escalation_level,
			       svp.description, svp.recorded_by, svp.recorded_at,
			       svp.redemption_status, svp.redemption_points,
			       vr.code, vr.name as rule_name, vr.category
			FROM student_violation_points svp
			JOIN violation_rules vr ON svp.rule_id = vr.id
			WHERE svp.student_id = ? AND svp.semester_id = ?
			  AND svp.deleted_at IS NULL
			ORDER BY svp.recorded_at DESC
		`

		rows, err := db.Query(query, studentID, semesterID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var id, studentIDVal, ruleID, points, escalationLevel, redemptionPoints int
			var description, recordedBy, recordedAt, redemptionStatus, ruleCode, ruleName, category string

			rows.Scan(&id, &studentIDVal, &ruleID, &points, &escalationLevel,
				&description, &recordedBy, &recordedAt,
				&redemptionStatus, &redemptionPoints,
				&ruleCode, &ruleName, &category)

			remaining := points - redemptionPoints

			results = append(results, map[string]interface{}{
				"id":                id,
				"student_id":        studentIDVal,
				"rule_id":           ruleID,
				"rule_code":         ruleCode,
				"rule_name":         ruleName,
				"category":          category,
				"points":            points,
				"escalation_level":  escalationLevel,
				"description":       description,
				"recorded_by":        recordedBy,
				"recorded_at":        recordedAt,
				"redemption_status": redemptionStatus,
				"redemption_points":  redemptionPoints,
				"remaining":         remaining,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"total":   len(results),
			"results": results,
		})
	}
}

// GetPendingRedemptions returns list of pending redemptions
func GetPendingRedemptions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status")
		if status == "" {
			status = "pending"
		}

		query := `
			SELECT vr.id, vr.student_id, s.name as student_name, s.nis,
			       vr.violation_id, vr.redemption_type, vr.redemption_description,
			       vr.points_redeemed, vr.verified_by, vr.verified_at, vr.status, vr.created_at
			FROM violation_redemptions vr
			JOIN students s ON vr.student_id = s.id
			WHERE vr.status = ?
			ORDER BY vr.created_at DESC
		`

		rows, err := db.Query(query, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var id, studentID, violationID, pointsRedeemed int
			var studentName, nis, redemptionType, redemptionDesc, verifiedBy, verifiedAt, statusVal, createdAt string

			rows.Scan(&id, &studentID, &studentName, &nis,
				&violationID, &redemptionType, &redemptionDesc,
				&pointsRedeemed, &verifiedBy, &verifiedAt, &statusVal, &createdAt)

			results = append(results, map[string]interface{}{
				"id":                    id,
				"student_id":            studentID,
				"student_name":         studentName,
				"nis":                   nis,
				"violation_id":          violationID,
				"redemption_type":       redemptionType,
				"redemption_description": redemptionDesc,
				"points_redeemed":       pointsRedeemed,
				"verified_by":           verifiedBy,
				"verified_at":           verifiedAt,
				"status":                statusVal,
				"created_at":            createdAt,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"total":   len(results),
			"results": results,
		})
	}
}

// recalculateViolationBurden recalculates violation_burden (0-100) and updates composite_score
// Formula: violation_burden = (violation_points - redemption_points) / max_violation_threshold * 100
// max_violation_threshold = SP3_THRESHOLD (76)
func recalculateViolationBurden(db *sql.DB, studentID, semesterID int) {
	var violationPoints, redemptionPoints int
	db.QueryRow(`
		SELECT COALESCE(violation_points, 0), COALESCE(redemption_points, 0)
		FROM semester_scores
		WHERE student_id = ? AND semester_id = ?
	`, studentID, semesterID).Scan(&violationPoints, &redemptionPoints)

	// Net violation = violation - redemption
	netViolation := violationPoints - redemptionPoints
	if netViolation < 0 {
		netViolation = 0
	}

	// Get SP3 threshold as max burden
	sp3Threshold := getConfigInt(db, "SP3_THRESHOLD", 76)

	// Burden: 0 if no violations, 100 if at SP3 threshold
	var burden float64
	if sp3Threshold > 0 {
		burden = float64(netViolation) / float64(sp3Threshold) * 100
	}
	if burden > 100 {
		burden = 100
	}

	// Get weights
	attendanceWeight := getConfigFloat(db, "ATTENDANCE_WEIGHT", 30)
	achievementWeight := getConfigFloat(db, "ACHIEVEMENT_WEIGHT", 50)
	violationWeight := getConfigFloat(db, "VIOLATION_WEIGHT", 20)

	// Get other components
	var attIndex, achNorm float64
	db.QueryRow(`
		SELECT COALESCE(attendance_index, 0), COALESCE(achievement_normalized, 0)
		FROM semester_scores
		WHERE student_id = ? AND semester_id = ?
	`, studentID, semesterID).Scan(&attIndex, &achNorm)

	// Calculate composite score
	composite := (attIndex * attendanceWeight / 100) +
		(achNorm * achievementWeight / 100) +
		((100 - burden) * violationWeight / 100)

	// Update
	db.Exec(`
		UPDATE semester_scores
		SET violation_burden = ?, composite_score = ?
		WHERE student_id = ? AND semester_id = ?
	`, burden, composite, studentID, semesterID)
}

// checkSPThreshold checks if student has reached SP1/SP2/SP3 threshold
// Returns SP level and counseling recommendation
func checkSPThreshold(db *sql.DB, totalPoints int) map[string]interface{} {
	sp1 := getConfigInt(db, "SP1_THRESHOLD", 25)
	sp2 := getConfigInt(db, "SP2_THRESHOLD", 51)
	sp3 := getConfigInt(db, "SP3_THRESHOLD", 76)

	if totalPoints >= sp3 {
		return map[string]interface{}{
			"level":            "SP3",
			"threshold":        sp3,
			"total_points":     totalPoints,
			"recommendation":   "Pemanggilan orang tua untuk pembicaraan konseling",
			"action":           "NOTIFY_PARENTS",
		}
	} else if totalPoints >= sp2 {
		return map[string]interface{}{
			"level":          "SP2",
			"threshold":      sp2,
			"total_points":   totalPoints,
			"recommendation": "Surat peringatan kedua + sesi konseling individual",
			"action":         "SCHEDULE_COUNSELING",
		}
	} else if totalPoints >= sp1 {
		return map[string]interface{}{
			"level":          "SP1",
			"threshold":      sp1,
			"total_points":   totalPoints,
			"recommendation": "Surat peringatan pertama + nasihat dari wali kelas",
			"action":         "NOTIFY_HOMEROOM",
		}
	}
	return nil
}
