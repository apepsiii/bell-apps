package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CompositeScoreResult holds the full composite score breakdown
type CompositeScoreResult struct {
	StudentID            int     `json:"student_id"`
	StudentName          string  `json:"student_name"`
	NIS                  string  `json:"nis"`
	ClassName            string  `json:"class_name"`
	SemesterID           int     `json:"semester_id"`
	SemesterName         string  `json:"semester_name"`
	AttendanceIndex      float64 `json:"attendance_index"`
	AttendanceDays       int     `json:"attendance_days"`
	LateDays             int     `json:"late_days"`
	AbsentDays           int     `json:"absent_days"`
	StreakDays           int     `json:"streak_days"`
	StreakBonus          float64 `json:"streak_bonus"`
	AchievementPoints    int     `json:"achievement_points"`
	AchievementNormalized float64 `json:"achievement_normalized"`
	ViolationPoints      int     `json:"violation_points"`
	ViolationBurden      float64 `json:"violation_burden"`
	RedemptionPoints     int     `json:"redemption_points"`
	CompositeScore       float64 `json:"composite_score"`
	Rank                 int     `json:"rank"`
}

// GetCompositeScores returns composite scores for leaderboard
func GetCompositeScores(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		semesterID := c.QueryParam("semester_id")
		classID := c.QueryParam("class_id")
		studentID := c.QueryParam("student_id")

		if semesterID == "" {
			semesterID = "1"
		}

		// If student_id provided, return single student detail
		if studentID != "" {
			return getSingleCompositeScore(c, db, studentID, semesterID)
		}

		// Otherwise return leaderboard
		return getLeaderboard(c, db, semesterID, classID)
	}
}

func getSingleCompositeScore(c echo.Context, db *sql.DB, studentID, semesterID string) error {
	query := `
		SELECT ss.student_id, s.name, s.nis, c.name as class_name,
		       ss.attendance_index, ss.attendance_days, ss.late_days, ss.absent_days,
		       ss.streak_days, ss.streak_bonus,
		       COALESCE(ss.achievement_points, 0), ss.achievement_normalized,
		       COALESCE(ss.violation_points, 0), ss.violation_burden,
		       COALESCE(ss.redemption_points, 0), ss.composite_score,
		       COALESCE(sem.name, '') as semester_name
		FROM semester_scores ss
		JOIN students s ON ss.student_id = s.id
		LEFT JOIN classes c ON s.class_id = c.id
		LEFT JOIN semesters sem ON ss.semester_id = sem.id
		WHERE ss.student_id = ? AND ss.semester_id = ?
	`
	var result CompositeScoreResult
	var attIndex, streakBonus, achNorm, violBurden, composite float64
	var semName, className sql.NullString

	err := db.QueryRow(query, studentID, semesterID).Scan(
		&result.StudentID, &result.StudentName, &result.NIS, &className,
		&attIndex, &result.AttendanceDays, &result.LateDays, &result.AbsentDays,
		&result.StreakDays, &streakBonus,
		&result.AchievementPoints, &achNorm,
		&result.ViolationPoints, &violBurden,
		&result.RedemptionPoints, &composite,
		&semName,
	)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Skor belum dihitung untuk siswa ini"})
	}

	result.AttendanceIndex = attIndex
	result.StreakBonus = streakBonus
	result.AchievementNormalized = achNorm
	result.ViolationBurden = violBurden
	result.CompositeScore = composite
	result.ClassName = className.String
	result.SemesterName = semName.String

	// Calculate breakdown
	attendanceWeight := getConfigFloat(db, "ATTENDANCE_WEIGHT", 30)
	achievementWeight := getConfigFloat(db, "ACHIEVEMENT_WEIGHT", 50)
	violationWeight := getConfigFloat(db, "VIOLATION_WEIGHT", 20)

	breakdown := map[string]interface{}{
		"attendance": map[string]interface{}{
			"index":  attIndex,
			"weight": attendanceWeight,
			"score":  attIndex * attendanceWeight / 100,
		},
		"achievement": map[string]interface{}{
			"normalized": achNorm,
			"weight":     achievementWeight,
			"score":      achNorm * achievementWeight / 100,
		},
		"violation": map[string]interface{}{
			"burden": violBurden,
			"weight": violationWeight,
			"score":  (100 - violBurden) * violationWeight / 100,
		},
		"total": composite,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "success",
		"result":   result,
		"breakdown": breakdown,
	})
}

func getLeaderboard(c echo.Context, db *sql.DB, semesterID, classID string) error {
	var semInt int
	fmt.Sscanf(semesterID, "%d", &semInt)
	if semInt <= 0 {
		semInt = 1
	}

	// Auto-populate if no records exist for this semester
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM semester_scores WHERE semester_id = ?", semInt).Scan(&count)
	if count == 0 {
		_ = RecalculateAllScoresInternal(db, semInt)
	}

	query := `
		SELECT ss.student_id, s.name, s.nis, c.name as class_name,
		       ss.attendance_index, ss.attendance_days, ss.late_days, ss.absent_days,
		       ss.streak_days, ss.streak_bonus,
		       COALESCE(ss.achievement_points, 0), ss.achievement_normalized,
		       COALESCE(ss.violation_points, 0), ss.violation_burden,
		       COALESCE(ss.redemption_points, 0), ss.composite_score,
		       COALESCE(sem.name, '') as semester_name
		FROM semester_scores ss
		JOIN students s ON ss.student_id = s.id
		LEFT JOIN classes c ON s.class_id = c.id
		LEFT JOIN semesters sem ON ss.semester_id = sem.id
		WHERE ss.semester_id = ? AND s.status = 'active'
	`
	args := []interface{}{semesterID}

	if classID != "" {
		query += " AND s.class_id = ?"
		args = append(args, classID)
	}

	query += " ORDER BY ss.composite_score DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	var results []CompositeScoreResult
	rank := 1
	for rows.Next() {
		var r CompositeScoreResult
		var attIndex, streakBonus, achNorm, violBurden, composite float64
		var semName, className sql.NullString

		if err := rows.Scan(
			&r.StudentID, &r.StudentName, &r.NIS, &className,
			&attIndex, &r.AttendanceDays, &r.LateDays, &r.AbsentDays,
			&r.StreakDays, &streakBonus,
			&r.AchievementPoints, &achNorm,
			&r.ViolationPoints, &violBurden,
			&r.RedemptionPoints, &composite,
			&semName,
		); err != nil {
			continue
		}

		r.AttendanceIndex = attIndex
		r.StreakBonus = streakBonus
		r.AchievementNormalized = achNorm
		r.ViolationBurden = violBurden
		r.CompositeScore = composite
		r.ClassName = className.String
		r.SemesterName = semName.String
		r.Rank = rank
		rank++

		results = append(results, r)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"total":   len(results),
		"results": results,
	})
}

// NormalizeAchievementPoints calculates normalized achievement score (0-100)
// Formula: min(100, (total_achievement_points / target_points) * 100)
// target_points is configurable via ACHIEVEMENT_TARGET_POINTS (default: 200 points)
func NormalizeAchievementPoints(db *sql.DB, studentID, semesterID int) (float64, error) {
	var totalPoints int
	err := db.QueryRow(`
		SELECT COALESCE(achievement_points, 0)
		FROM semester_scores
		WHERE student_id = ? AND semester_id = ?
	`, studentID, semesterID).Scan(&totalPoints)
	if err != nil {
		return 0, err
	}

	targetPoints := getConfigInt(db, "ACHIEVEMENT_TARGET_POINTS", 200)
	if targetPoints <= 0 {
		targetPoints = 200
	}

	normalized := float64(totalPoints) / float64(targetPoints) * 100
	if normalized > 100 {
		normalized = 100
	}

	// Update in database
	db.Exec(`
		UPDATE semester_scores 
		SET achievement_normalized = ?,
		    composite_score = (
				(attendance_index * (SELECT CAST(config_value AS REAL) FROM point_config WHERE config_key = 'ATTENDANCE_WEIGHT') / 100) +
				(? * (SELECT CAST(config_value AS REAL) FROM point_config WHERE config_key = 'ACHIEVEMENT_WEIGHT') / 100) +
				((100 - violation_burden) * (SELECT CAST(config_value AS REAL) FROM point_config WHERE config_key = 'VIOLATION_WEIGHT') / 100)
			)
		WHERE student_id = ? AND semester_id = ?
	`, normalized, normalized, studentID, semesterID)

	return normalized, nil
}

// Helper functions
func getConfigFloat(db *sql.DB, key string, defaultVal float64) float64 {
	var val string
	err := db.QueryRow("SELECT config_value FROM point_config WHERE config_key = ?", key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	var f float64
	fmt.Sscanf(val, "%f", &f)
	if f == 0 {
		return defaultVal
	}
	return f
}

func getConfigInt(db *sql.DB, key string, defaultVal int) int {
	var val string
	err := db.QueryRow("SELECT config_value FROM point_config WHERE config_key = ?", key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	var i int
	fmt.Sscanf(val, "%d", &i)
	if i == 0 {
		return defaultVal
	}
	return i
}

// RecalculateAllScoresInternal calculates attendance, achievement, violation, and composite scores for all active students in a semester.
func RecalculateAllScoresInternal(db *sql.DB, semesterID int) error {
	semesterIDStr := fmt.Sprintf("%d", semesterID)

	var startDate, endDate string
	err := db.QueryRow("SELECT start_date, end_date FROM semesters WHERE id = ?", semesterID).Scan(&startDate, &endDate)
	if err != nil {
		startDate = "2026-07-01"
		endDate = "2026-12-31"
	}

	// 1. Calculate and save attendance index for all active students
	calc := &AttendanceIndexCalculator{db: db}
	attResults, err := calc.calculateForAllStudents(semesterIDStr, startDate, endDate)
	if err == nil {
		for _, res := range attResults {
			_ = calc.saveResult(res)
		}
	}

	// 2. Fetch all active student IDs
	rows, err := db.Query("SELECT id FROM students WHERE status = 'active'")
	if err != nil {
		return err
	}
	defer rows.Close()

	var studentIDs []int
	for rows.Next() {
		var sid int
		if err := rows.Scan(&sid); err == nil {
			studentIDs = append(studentIDs, sid)
		}
	}
	rows.Close()

	// 3. For each active student, aggregate achievement and violation points, then calculate composite score
	for _, sid := range studentIDs {
		var achPoints int
		_ = db.QueryRow(`
			SELECT COALESCE(SUM(points), 0)
			FROM student_achievement_points
			WHERE student_id = ? AND semester_id = ? AND deleted_at IS NULL
		`, sid, semesterID).Scan(&achPoints)

		var violPoints int
		_ = db.QueryRow(`
			SELECT COALESCE(SUM(points), 0)
			FROM student_violation_points
			WHERE student_id = ? AND semester_id = ? AND deleted_at IS NULL
		`, sid, semesterID).Scan(&violPoints)

		// Ensure record exists in semester_scores
		_, _ = db.Exec(`
			INSERT INTO semester_scores (student_id, semester_id, achievement_points, violation_points)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(student_id, semester_id) DO UPDATE SET
				achievement_points = excluded.achievement_points,
				violation_points = excluded.violation_points
		`, sid, semesterID, achPoints, violPoints)

		// Normalize achievement score
		_, _ = NormalizeAchievementPoints(db, sid, semesterID)

		// Recalculate violation burden and composite score
		recalculateViolationBurden(db, sid, semesterID)
	}

	return nil
}

// RecalculateAllCompositeScores handles HTTP POST /api/composite-scores/recalculate-all
func RecalculateAllCompositeScores(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		semesterID := 1
		var activeSem int
		if err := db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&activeSem); err == nil {
			semesterID = activeSem
		}

		err := RecalculateAllScoresInternal(db, semesterID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal menghitung ulang skor: " + err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Seluruh skor komposit siswa berhasil dihitung ulang",
		})
	}
}

