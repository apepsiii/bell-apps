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
		       sem.name as semester_name
		FROM semester_scores ss
		JOIN students s ON ss.student_id = s.id
		LEFT JOIN classes c ON s.class_id = c.id
		JOIN semesters sem ON ss.semester_id = sem.id
		WHERE ss.student_id = ? AND ss.semester_id = ?
	`
	var result CompositeScoreResult
	var attIndex, streakBonus, achNorm, violBurden, composite float64
	var semName, className string

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
	result.ClassName = className
	result.SemesterName = semName

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
	query := `
		SELECT ss.student_id, s.name, s.nis, c.name as class_name,
		       ss.attendance_index, ss.attendance_days, ss.late_days, ss.absent_days,
		       ss.streak_days, ss.streak_bonus,
		       COALESCE(ss.achievement_points, 0), ss.achievement_normalized,
		       COALESCE(ss.violation_points, 0), ss.violation_burden,
		       COALESCE(ss.redemption_points, 0), ss.composite_score,
		       sem.name as semester_name
		FROM semester_scores ss
		JOIN students s ON ss.student_id = s.id
		LEFT JOIN classes c ON s.class_id = c.id
		JOIN semesters sem ON ss.semester_id = sem.id
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
		var semName, className string

		rows.Scan(
			&r.StudentID, &r.StudentName, &r.NIS, &className,
			&attIndex, &r.AttendanceDays, &r.LateDays, &r.AbsentDays,
			&r.StreakDays, &streakBonus,
			&r.AchievementPoints, &achNorm,
			&r.ViolationPoints, &violBurden,
			&r.RedemptionPoints, &composite,
			&semName,
		)

		r.AttendanceIndex = attIndex
		r.StreakBonus = streakBonus
		r.AchievementNormalized = achNorm
		r.ViolationBurden = violBurden
		r.CompositeScore = composite
		r.ClassName = className
		r.SemesterName = semName
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
// Formula: (total_achievement_points / max_possible_points) * 100
// max_possible_points = 10 categories × max_per_category (200) = 2000
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

	maxPerCategory := getConfigInt(db, "MAX_POINTS_PER_CATEGORY", 200)
	numCategories := 10 // R1-R10
	maxPossible := maxPerCategory * numCategories

	if maxPossible == 0 {
		return 0, fmt.Errorf("max possible points is 0")
	}

	normalized := float64(totalPoints) / float64(maxPossible) * 100
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
