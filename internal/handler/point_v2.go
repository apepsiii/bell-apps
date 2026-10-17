package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/repository"
)

// =====================
// ACHIEVEMENT RULES (R1-R10)
// =====================

type AchievementRule struct {
	ID           int    `json:"id"`
	Code         string `json:"code"`
	Category     string `json:"category"`
	CategoryCode string `json:"category_code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Points       int    `json:"points"`
	MinPoints    int    `json:"min_points"`
	MaxPoints    int    `json:"max_points"`
	IsActive     bool   `json:"is_active"`
}

type ViolationRule struct {
	ID           int    `json:"id"`
	Code         string `json:"code"`
	Category     string `json:"category"`
	CategoryCode string `json:"category_code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Points1      int    `json:"points_1"`
	Points2      int    `json:"points_2"`
	Points3      int    `json:"points_3"`
	IsActive     bool   `json:"is_active"`
}

type StudentDualPointProfile struct {
	Student              StudentInfo              `json:"student"`
	AcademicYear         string                   `json:"academic_year"`
	AchievementPoints    int                      `json:"achievement_points"`
	ViolationPoints      int                      `json:"violation_points"`
	AchievementGrade     string                   `json:"achievement_grade"`
	ViolationLevel       string                   `json:"violation_level"`
	AchievementHistory   []AchievementPointLog    `json:"achievement_history"`
	ViolationHistory     []ViolationPointLog      `json:"violation_history"`
	AchievementBreakdown []CategoryBreakdown      `json:"achievement_breakdown"`
	ViolationBreakdown   []CategoryBreakdown      `json:"violation_breakdown"`
	ClassRank            int                      `json:"class_rank"`
	TotalClassStudents   int                      `json:"total_class_students"`
	NIS                  string                   `json:"nis"`
	ClassName            string                   `json:"class_name"`
	Status               string                   `json:"status"`
	Name                 string                   `json:"name"`
}

type CategoryBreakdown struct {
	Category string `json:"category"`
	Points   int    `json:"points"`
	Count    int    `json:"count"`
}

type AchievementPointLog struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	RuleID       *int   `json:"rule_id"`
	RuleCode     string `json:"rule_code"`
	RuleName     string `json:"rule_name"`
	Points       int    `json:"points"`
	Description  string `json:"description"`
	RecordedBy   string `json:"recorded_by"`
	AcademicYear string `json:"academic_year"`
	CreatedAt    string `json:"created_at"`
}

type ViolationPointLog struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	RuleID       *int   `json:"rule_id"`
	RuleCode     string `json:"rule_code"`
	RuleName     string `json:"rule_name"`
	Occurrence   int    `json:"occurrence"`
	Points       int    `json:"points"`
	Description  string `json:"description"`
	RecordedBy   string `json:"recorded_by"`
	AcademicYear string `json:"academic_year"`
	CreatedAt    string `json:"created_at"`
}

func getAchievementGrade(points int) string {
	if points >= 151 {
		return "Anugerah Waluya Utama (≥151 Poin)"
	} else if points >= 126 {
		return "Sertifikat + Hadiah (126-150 Poin)"
	} else if points >= 100 {
		return "Sertifikat Berprestasi (100-125 Poin)"
	}
	return "Belum Mencapai Predikat"
}

func getViolationLevel(points int) string {
	if points >= 76 {
		return "SP3 - Surat Perjanjian Ketiga (≥76 Poin)"
	} else if points >= 51 {
		return "SP2 - Surat Perjanjian Kedua (51-75 Poin)"
	} else if points >= 25 {
		return "SP1 - Surat Perjanjian Pertama (25-50 Poin)"
	}
	return "Belum Ada Sanksi"
}

// GetAchievementRules returns all achievement rules
func GetAchievementRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		categoryCode := c.QueryParam("category")

		query := `
			SELECT id, code, category, category_code, name, description, points, min_points, max_points, is_active
			FROM achievement_rules WHERE is_active = 1`

		args := []interface{}{}
		if categoryCode != "" {
			query += " AND category_code = ?"
			args = append(args, categoryCode)
		}
		query += " ORDER BY category_code, code"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var rules []AchievementRule
		for rows.Next() {
			var r AchievementRule
			var isActive int
			if err := rows.Scan(&r.ID, &r.Code, &r.Category, &r.CategoryCode, &r.Name,
				&r.Description, &r.Points, &r.MinPoints, &r.MaxPoints, &isActive); err != nil {
				continue
			}
			r.IsActive = isActive == 1
			rules = append(rules, r)
		}
		if rules == nil {
			rules = []AchievementRule{}
		}
		return c.JSON(http.StatusOK, rules)
	}
}

// GetViolationRules returns all violation rules
func GetViolationRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		categoryCode := c.QueryParam("category")

		query := `
			SELECT id, code, category, category_code, name, description, points_1, points_2, points_3, is_active
			FROM violation_rules WHERE is_active = 1`

		args := []interface{}{}
		if categoryCode != "" {
			query += " AND category_code = ?"
			args = append(args, categoryCode)
		}
		query += " ORDER BY category_code, code"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var rules []ViolationRule
		for rows.Next() {
			var r ViolationRule
			var isActive int
			if err := rows.Scan(&r.ID, &r.Code, &r.Category, &r.CategoryCode, &r.Name,
				&r.Description, &r.Points1, &r.Points2, &r.Points3, &isActive); err != nil {
				continue
			}
			r.IsActive = isActive == 1
			rules = append(rules, r)
		}
		if rules == nil {
			rules = []ViolationRule{}
		}
		return c.JSON(http.StatusOK, rules)
	}
}

// GetStudentDualPointProfile returns full dual-track point profile for a student
func GetStudentDualPointProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Param("id")
		academicYear := c.QueryParam("year")
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}

		var profile StudentDualPointProfile
		var className, nis, photo, status sql.NullString
		var classID sql.NullInt64

		err := db.QueryRow(`
			SELECT s.id, s.name, COALESCE(s.nis,''), COALESCE(s.photo,''), COALESCE(c.name,''), COALESCE(s.status,''), s.class_id
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?
		`, studentID).Scan(&profile.Student.ID, &profile.Student.Name, &nis, &photo, &className, &status, &classID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Siswa tidak ditemukan"})
		}
		profile.Student.Class = className.String
		profile.Student.NIS = nis.String
		profile.Student.Photo = photo.String
		profile.AcademicYear = academicYear
		profile.NIS = nis.String
		profile.ClassName = className.String
		profile.Status = status.String
		profile.Name = profile.Student.Name

		// Total achievement points
		db.QueryRow(`
			SELECT COALESCE(SUM(points), 0) FROM student_achievement_points
			WHERE student_id = ? AND academic_year = ? AND deleted_at IS NULL
		`, studentID, academicYear).Scan(&profile.AchievementPoints)

		// Total violation points
		db.QueryRow(`
			SELECT COALESCE(SUM(points), 0) FROM student_violation_points
			WHERE student_id = ? AND academic_year = ? AND deleted_at IS NULL
		`, studentID, academicYear).Scan(&profile.ViolationPoints)

		profile.AchievementGrade = getAchievementGrade(profile.AchievementPoints)
		profile.ViolationLevel = getViolationLevel(profile.ViolationPoints)

		// Achievement breakdown by category
		aBreakdownRows, err := db.Query(`
			SELECT COALESCE(ar.category, 'Lainnya') as category, 
				SUM(sap.points) as total_points,
				COUNT(*) as count
			FROM student_achievement_points sap
			LEFT JOIN achievement_rules ar ON sap.rule_id = ar.id
			WHERE sap.student_id = ? AND sap.academic_year = ? AND sap.deleted_at IS NULL
			GROUP BY ar.category
		`, studentID, academicYear)
		if err == nil {
			defer aBreakdownRows.Close()
			for aBreakdownRows.Next() {
				var b CategoryBreakdown
				aBreakdownRows.Scan(&b.Category, &b.Points, &b.Count)
				profile.AchievementBreakdown = append(profile.AchievementBreakdown, b)
			}
		}
		if profile.AchievementBreakdown == nil {
			profile.AchievementBreakdown = []CategoryBreakdown{}
		}

		// Violation breakdown by category
		vBreakdownRows, err := db.Query(`
			SELECT COALESCE(vr.category, 'Lainnya') as category, 
				SUM(svp.points) as total_points,
				COUNT(*) as count
			FROM student_violation_points svp
			LEFT JOIN violation_rules vr ON svp.rule_id = vr.id
			WHERE svp.student_id = ? AND svp.academic_year = ? AND svp.deleted_at IS NULL
			GROUP BY vr.category
		`, studentID, academicYear)
		if err == nil {
			defer vBreakdownRows.Close()
			for vBreakdownRows.Next() {
				var b CategoryBreakdown
				vBreakdownRows.Scan(&b.Category, &b.Points, &b.Count)
				profile.ViolationBreakdown = append(profile.ViolationBreakdown, b)
			}
		}
		if profile.ViolationBreakdown == nil {
			profile.ViolationBreakdown = []CategoryBreakdown{}
		}

		// Get class rank
		if classID.Valid {
			var rank int
			err = db.QueryRow(`
				SELECT COUNT(*) + 1
				FROM (
					SELECT s.id, COALESCE(SUM(sap.points), 0) as total_achievement
					FROM students s
					LEFT JOIN student_achievement_points sap ON s.id = sap.student_id AND sap.academic_year = ? AND sap.deleted_at IS NULL
					WHERE s.class_id = ? AND s.status = 'active'
					GROUP BY s.id
					HAVING total_achievement > ?
				)
			`, academicYear, classID.Int64, profile.AchievementPoints).Scan(&rank)
			if err == nil {
				profile.ClassRank = rank
			}

			// Get total students in class
			db.QueryRow(`
				SELECT COUNT(*) FROM students WHERE class_id = ? AND status = 'active'
			`, classID.Int64).Scan(&profile.TotalClassStudents)
		}

		// Achievement history
		aRows, err := db.Query(`
			SELECT sap.id, sap.student_id, sap.rule_id,
				COALESCE(ar.code,''), COALESCE(ar.name,''),
				sap.points, sap.description, sap.recorded_by,
				sap.academic_year, sap.created_at
			FROM student_achievement_points sap
			LEFT JOIN achievement_rules ar ON sap.rule_id = ar.id
			WHERE sap.student_id = ? AND sap.academic_year = ? AND sap.deleted_at IS NULL
			ORDER BY sap.created_at DESC LIMIT 100
		`, studentID, academicYear)
		if err == nil {
			defer aRows.Close()
			for aRows.Next() {
				var l AchievementPointLog
				var ruleID sql.NullInt64
				var recordedBy, academicYearVal sql.NullString
				aRows.Scan(&l.ID, &l.StudentID, &ruleID, &l.RuleCode, &l.RuleName,
					&l.Points, &l.Description, &recordedBy, &academicYearVal, &l.CreatedAt)
				l.RecordedBy = recordedBy.String
				l.AcademicYear = academicYearVal.String
				if ruleID.Valid {
					val := int(ruleID.Int64)
					l.RuleID = &val
				}
				profile.AchievementHistory = append(profile.AchievementHistory, l)
			}
		}
		if profile.AchievementHistory == nil {
			profile.AchievementHistory = []AchievementPointLog{}
		}

		// Violation history
		vRows, err := db.Query(`
			SELECT svp.id, svp.student_id, svp.rule_id,
				COALESCE(vr.code,''), COALESCE(vr.name,''),
				svp.occurrence, svp.points, svp.description, svp.recorded_by,
				svp.academic_year, svp.created_at
			FROM student_violation_points svp
			LEFT JOIN violation_rules vr ON svp.rule_id = vr.id
			WHERE svp.student_id = ? AND svp.academic_year = ? AND svp.deleted_at IS NULL
			ORDER BY svp.created_at DESC LIMIT 100
		`, studentID, academicYear)
		if err == nil {
			defer vRows.Close()
			for vRows.Next() {
				var l ViolationPointLog
				var ruleID sql.NullInt64
				var recordedBy, academicYearVal sql.NullString
				vRows.Scan(&l.ID, &l.StudentID, &ruleID, &l.RuleCode, &l.RuleName,
					&l.Occurrence, &l.Points, &l.Description, &recordedBy, &academicYearVal, &l.CreatedAt)
				l.RecordedBy = recordedBy.String
				l.AcademicYear = academicYearVal.String
				if ruleID.Valid {
					val := int(ruleID.Int64)
					l.RuleID = &val
				}
				profile.ViolationHistory = append(profile.ViolationHistory, l)
			}
		}
		if profile.ViolationHistory == nil {
			profile.ViolationHistory = []ViolationPointLog{}
		}

		return c.JSON(http.StatusOK, profile)
	}
}

// Helpers for flexible JSON and FormData binding
func parsePointInput(c echo.Context) map[string]string {
	result := make(map[string]string)
	if strings.Contains(c.Request().Header.Get("Content-Type"), "application/json") {
		var jsonMap map[string]interface{}
		if err := c.Bind(&jsonMap); err == nil && jsonMap != nil {
			for k, v := range jsonMap {
				result[k] = toPointStr(v)
			}
		}
	}
	keys := []string{"student_id", "rule_id", "points", "occurrence", "description", "recorded_by", "academic_year", "semester_id", "reason"}
	for _, k := range keys {
		if val := c.FormValue(k); val != "" && result[k] == "" {
			result[k] = strings.TrimSpace(val)
		}
	}
	return result
}

func toPointStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toIntVal(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// RecalculateSingleStudentCompositeScore recalculates attendance, achievement, violation, and composite scores for a single student.
func RecalculateSingleStudentCompositeScore(db *sql.DB, studentID int, semesterID int) error {
	if semesterID <= 0 {
		var activeSem int
		if err := db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&activeSem); err == nil && activeSem > 0 {
			semesterID = activeSem
		} else {
			semesterID = 1
		}
	}

	var achPoints int
	_ = db.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM student_achievement_points
		WHERE student_id = ? AND semester_id = ? AND deleted_at IS NULL
	`, studentID, semesterID).Scan(&achPoints)

	var violPoints int
	_ = db.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM student_violation_points
		WHERE student_id = ? AND semester_id = ? AND deleted_at IS NULL
	`, studentID, semesterID).Scan(&violPoints)

	// Ensure record exists and update points in semester_scores
	_, err := db.Exec(`
		INSERT INTO semester_scores (student_id, semester_id, achievement_points, violation_points)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(student_id, semester_id) DO UPDATE SET
			achievement_points = excluded.achievement_points,
			violation_points = excluded.violation_points
	`, studentID, semesterID, achPoints, violPoints)
	if err != nil {
		return err
	}

	// Normalize achievement score
	_, _ = NormalizeAchievementPoints(db, studentID, semesterID)

	// Recalculate violation burden and composite score
	recalculateViolationBurden(db, studentID, semesterID)

	return nil
}

// AddAchievementPoint records an achievement point for a student
func AddAchievementPoint(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		input := parsePointInput(c)
		studentID := input["student_id"]
		ruleID := input["rule_id"]
		pointsStr := input["points"]
		description := input["description"]
		recordedBy := input["recorded_by"]
		academicYear := input["academic_year"]

		if studentID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id required"})
		}
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}
		if recordedBy == "" {
			if u := c.Get("user"); u != nil {
				if uname, ok := u.(string); ok && uname != "" {
					recordedBy = uname
				}
			}
			if recordedBy == "" {
				recordedBy = "Admin"
			}
		}

		semesterID := 1
		if semStr := input["semester_id"]; semStr != "" {
			if s, err := strconv.Atoi(semStr); err == nil && s > 0 {
				semesterID = s
			}
		}
		if semesterID <= 1 {
			var activeSem int
			if err := db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&activeSem); err == nil && activeSem > 0 {
				semesterID = activeSem
			}
		}

		// Get max points config
		maxPointsStr := GetConfigValue(db, "MAX_POINTS_PER_TRANSACTION", "100")
		maxPoints := 100
		if mp, err := strconv.Atoi(maxPointsStr); err == nil {
			maxPoints = mp
		}

		// Check rate limit
		rateLimitStr := GetConfigValue(db, "RATE_LIMIT_PER_DAY", "50")
		rateLimit := 50
		if rl, err := strconv.Atoi(rateLimitStr); err == nil {
			rateLimit = rl
		}

		exceeded, count, err := CheckRateLimit(db, recordedBy, rateLimit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check rate limit"})
		}
		if exceeded {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error":   "Batas input harian tercapai",
				"message": "Anda sudah melakukan " + strconv.Itoa(count) + " input hari ini. Maksimal " + strconv.Itoa(rateLimit) + " per hari.",
			})
		}

		var points int
		var ruleName string

		if ruleID != "" {
			err := db.QueryRow(`SELECT points, name FROM achievement_rules WHERE id = ?`, ruleID).Scan(&points, &ruleName)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Rule tidak ditemukan"})
			}
			// Allow override points
			if pointsStr != "" {
				if p, err := strconv.Atoi(pointsStr); err == nil && p > 0 {
					// Validate max points
					if p > maxPoints {
						return c.JSON(http.StatusBadRequest, map[string]string{
							"error": "Maksimal poin per transaksi adalah " + strconv.Itoa(maxPoints),
						})
					}
					points = p
				}
			}
			if description == "" {
				description = "Penghargaan: " + ruleName
			}
		} else {
			if pointsStr == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "points required"})
			}
			var err error
			points, err = strconv.Atoi(pointsStr)
			if err != nil || points <= 0 {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "points tidak valid"})
			}
			// Validate max points
			if points > maxPoints {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Maksimal poin per transaksi adalah " + strconv.Itoa(maxPoints),
				})
			}
		}

		var ruleIDPtr interface{}
		if ruleID != "" {
			ruleIDPtr = ruleID
		}

		result, err := db.Exec(`
			INSERT INTO student_achievement_points (student_id, rule_id, points, description, recorded_by, academic_year, semester_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, studentID, ruleIDPtr, points, description, recordedBy, academicYear, semesterID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Log audit trail
		insertID, _ := result.LastInsertId()
		newData := map[string]interface{}{
			"id":            insertID,
			"student_id":    studentID,
			"rule_id":       ruleID,
			"points":        points,
			"description":   description,
			"recorded_by":   recordedBy,
			"academic_year": academicYear,
			"semester_id":   semesterID,
		}
		LogAuditTrail(db, c, "INSERT", "student_achievement_points", int(insertID), nil, newData, "")

		// Auto-recalculate composite score
		if sidInt, err := strconv.Atoi(studentID); err == nil {
			_ = RecalculateSingleStudentCompositeScore(db, sidInt, semesterID)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Poin penghargaan berhasil dicatat",
			"points":  points,
		})
	}
}

// AddViolationPoint records a violation point for a student
func AddViolationPoint(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		input := parsePointInput(c)
		studentID := input["student_id"]
		ruleID := input["rule_id"]
		occurrenceStr := input["occurrence"]
		description := input["description"]
		recordedBy := input["recorded_by"]
		academicYear := input["academic_year"]

		if studentID == "" || ruleID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "student_id dan rule_id required"})
		}
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}
		if recordedBy == "" {
			if u := c.Get("user"); u != nil {
				if uname, ok := u.(string); ok && uname != "" {
					recordedBy = uname
				}
			}
			if recordedBy == "" {
				recordedBy = "Admin"
			}
		}

		semesterID := 1
		if semStr := input["semester_id"]; semStr != "" {
			if s, err := strconv.Atoi(semStr); err == nil && s > 0 {
				semesterID = s
			}
		}
		if semesterID <= 1 {
			var activeSem int
			if err := db.QueryRow("SELECT id FROM semesters WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&activeSem); err == nil && activeSem > 0 {
				semesterID = activeSem
			}
		}

		occurrence, _ := strconv.Atoi(occurrenceStr)
		if occurrence < 1 || occurrence > 3 {
			occurrence = 1
		}

		var p1, p2, p3 int
		var ruleName string
		err := db.QueryRow(`SELECT name, points_1, points_2, points_3 FROM violation_rules WHERE id = ?`, ruleID).
			Scan(&ruleName, &p1, &p2, &p3)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Rule tidak ditemukan"})
		}

		// Pick points based on occurrence
		points := p1
		if occurrence == 2 {
			points = p2
		} else if occurrence >= 3 {
			points = p3
		}

		if description == "" {
			description = "Pelanggaran: " + ruleName
		}

		result, err := db.Exec(`
			INSERT INTO student_violation_points (student_id, rule_id, occurrence, points, description, recorded_by, academic_year, semester_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, studentID, ruleID, occurrence, points, description, recordedBy, academicYear, semesterID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Log audit trail
		insertID, _ := result.LastInsertId()
		newData := map[string]interface{}{
			"id":            insertID,
			"student_id":    studentID,
			"rule_id":       ruleID,
			"occurrence":    occurrence,
			"points":        points,
			"description":   description,
			"recorded_by":   recordedBy,
			"academic_year": academicYear,
			"semester_id":   semesterID,
		}
		LogAuditTrail(db, c, "INSERT", "student_violation_points", int(insertID), nil, newData, "")

		// Auto-recalculate composite score
		if sidInt, err := strconv.Atoi(studentID); err == nil {
			_ = RecalculateSingleStudentCompositeScore(db, sidInt, semesterID)
		}

		// Check violation level for response
		var totalViolation int
		db.QueryRow(`SELECT COALESCE(SUM(points),0) FROM student_violation_points WHERE student_id = ? AND academic_year = ? AND deleted_at IS NULL`,
			studentID, academicYear).Scan(&totalViolation)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":         "Poin pelanggaran berhasil dicatat",
			"points":          points,
			"total_violation": totalViolation,
			"violation_level": getViolationLevel(totalViolation),
		})
	}
}

// DeleteAchievementPoint soft deletes an achievement point record
func DeleteAchievementPoint(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		input := parsePointInput(c)
		reason := input["reason"]
		if reason == "" {
			reason = c.QueryParam("reason")
		}

		// Check authorization - only kepala_sekolah or superadmin can delete
		if !CheckAdminRole(c, "kepala_sekolah") {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Hanya kepala sekolah yang dapat menghapus poin",
			})
		}

		if reason == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Alasan penghapusan wajib diisi",
			})
		}

		// Get record data before deletion for audit
		recordID, _ := strconv.Atoi(id)
		oldData, err := GetRecordBeforeDelete(db, "student_achievement_points", recordID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Record tidak ditemukan atau sudah dihapus",
			})
		}

		// Get user info
		deletedBy := "Admin"
		if user := c.Get("user"); user != nil {
			if username, ok := user.(string); ok && username != "" {
				deletedBy = username
			}
		}

		// Perform soft delete
		err = SoftDeleteRecord(db, "student_achievement_points", recordID, deletedBy, reason)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		// Log audit trail
		LogAuditTrail(db, c, "DELETE", "student_achievement_points", recordID, oldData, nil, reason)

		// Auto-recalculate composite score
		if sidRaw, ok := oldData["student_id"]; ok {
			sidInt := toIntVal(sidRaw)
			semID := toIntVal(oldData["semester_id"])
			if semID <= 0 {
				semID = 1
			}
			if sidInt > 0 {
				_ = RecalculateSingleStudentCompositeScore(db, sidInt, semID)
			}
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Poin penghargaan berhasil dihapus",
		})
	}
}

// DeleteViolationPoint soft deletes a violation point record
func DeleteViolationPoint(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		input := parsePointInput(c)
		reason := input["reason"]
		if reason == "" {
			reason = c.QueryParam("reason")
		}

		// Check authorization - only kepala_sekolah or superadmin can delete
		if !CheckAdminRole(c, "kepala_sekolah") {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Hanya kepala sekolah yang dapat menghapus poin",
			})
		}

		if reason == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Alasan penghapusan wajib diisi",
			})
		}

		// Get record data before deletion for audit
		recordID, _ := strconv.Atoi(id)
		oldData, err := GetRecordBeforeDelete(db, "student_violation_points", recordID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Record tidak ditemukan atau sudah dihapus",
			})
		}

		// Get user info
		deletedBy := "Admin"
		if user := c.Get("user"); user != nil {
			if username, ok := user.(string); ok && username != "" {
				deletedBy = username
			}
		}

		// Perform soft delete
		err = SoftDeleteRecord(db, "student_violation_points", recordID, deletedBy, reason)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		// Log audit trail
		LogAuditTrail(db, c, "DELETE", "student_violation_points", recordID, oldData, nil, reason)

		// Auto-recalculate composite score
		if sidRaw, ok := oldData["student_id"]; ok {
			sidInt := toIntVal(sidRaw)
			semID := toIntVal(oldData["semester_id"])
			if semID <= 0 {
				semID = 1
			}
			if sidInt > 0 {
				_ = RecalculateSingleStudentCompositeScore(db, sidInt, semID)
			}
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Poin pelanggaran berhasil dihapus",
		})
	}
}

// GetDualPointLeaderboard returns leaderboard sorted by achievement points
func GetDualPointLeaderboard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		academicYear := c.QueryParam("year")
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}

		rows, err := db.Query(`
			SELECT s.id, s.name, COALESCE(c.name,'') as class_name,
				COALESCE(SUM(sap.points), 0) as achievement_points,
				COALESCE((SELECT SUM(points) FROM student_violation_points WHERE student_id = s.id AND academic_year = ? AND deleted_at IS NULL), 0) as violation_points
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN student_achievement_points sap ON s.id = sap.student_id AND sap.academic_year = ? AND sap.deleted_at IS NULL
			WHERE s.status = 'active'
			GROUP BY s.id
			ORDER BY achievement_points DESC
			LIMIT 50
		`, academicYear, academicYear)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		type LeaderboardItem struct {
			ID                int    `json:"id"`
			Name              string `json:"name"`
			ClassName         string `json:"class_name"`
			AchievementPoints int    `json:"achievement_points"`
			ViolationPoints   int    `json:"violation_points"`
			AchievementGrade  string `json:"achievement_grade"`
			ViolationLevel    string `json:"violation_level"`
			Rank              int    `json:"rank"`
		}

		var items []LeaderboardItem
		rank := 1
		for rows.Next() {
			var item LeaderboardItem
			rows.Scan(&item.ID, &item.Name, &item.ClassName, &item.AchievementPoints, &item.ViolationPoints)
			item.AchievementGrade = getAchievementGrade(item.AchievementPoints)
			item.ViolationLevel = getViolationLevel(item.ViolationPoints)
			item.Rank = rank
			items = append(items, item)
			rank++
		}
		if items == nil {
			items = []LeaderboardItem{}
		}
		return c.JSON(http.StatusOK, items)
	}
}

// GetClassDualPointSummary returns summary of dual-track points per class
func GetClassDualPointSummary(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		academicYear := c.QueryParam("year")
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}

		rows, err := db.Query(`
			SELECT c.id, c.name,
				COUNT(DISTINCT s.id) as student_count,
				COALESCE(SUM(sap.points), 0) as total_achievement,
				COALESCE(SUM(svp.vpoints), 0) as total_violation
			FROM classes c
			LEFT JOIN students s ON s.class_id = c.id AND s.status = 'active'
			LEFT JOIN student_achievement_points sap ON sap.student_id = s.id AND sap.academic_year = ? AND sap.deleted_at IS NULL
			LEFT JOIN (
				SELECT student_id, SUM(points) as vpoints FROM student_violation_points WHERE academic_year = ? AND deleted_at IS NULL GROUP BY student_id
			) svp ON svp.student_id = s.id
			GROUP BY c.id
			ORDER BY c.name
		`, academicYear, academicYear)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		type ClassSummary struct {
			ID               int    `json:"id"`
			Name             string `json:"name"`
			StudentCount     int    `json:"student_count"`
			TotalAchievement int    `json:"total_achievement"`
			TotalViolation   int    `json:"total_violation"`
		}

		var items []ClassSummary
		for rows.Next() {
			var item ClassSummary
			rows.Scan(&item.ID, &item.Name, &item.StudentCount, &item.TotalAchievement, &item.TotalViolation)
			items = append(items, item)
		}
		if items == nil {
			items = []ClassSummary{}
		}
		return c.JSON(http.StatusOK, items)
	}
}

// GetDualPointRecentActivities returns the latest achievement and violation point activities
func GetDualPointRecentActivities(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		academicYear := c.QueryParam("year")
		if academicYear == "" {
			academicYear = repository.CurrentAcademicYear()
		}

		type ActivityItem struct {
			ID          int    `json:"id"`
			Type        string `json:"type"`
			StudentID   int    `json:"student_id"`
			StudentName string `json:"student_name"`
			ClassName   string `json:"class_name"`
			Points      int    `json:"points"`
			Description string `json:"description"`
			RecordedBy  string `json:"recorded_by"`
			CreatedAt   string `json:"created_at"`
		}

		var achievements []ActivityItem
		aRows, err := db.Query(`
			SELECT sap.id, sap.student_id, s.name, COALESCE(c.name, '-'), sap.points, sap.description, COALESCE(sap.recorded_by, 'Admin'), COALESCE(sap.created_at, '')
			FROM student_achievement_points sap
			JOIN students s ON sap.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE sap.academic_year = ? AND sap.deleted_at IS NULL
			ORDER BY sap.id DESC
			LIMIT 10
		`, academicYear)
		if err == nil {
			defer aRows.Close()
			for aRows.Next() {
				var it ActivityItem
				it.Type = "achievement"
				if err := aRows.Scan(&it.ID, &it.StudentID, &it.StudentName, &it.ClassName, &it.Points, &it.Description, &it.RecordedBy, &it.CreatedAt); err == nil {
					achievements = append(achievements, it)
				}
			}
		}

		var violations []ActivityItem
		vRows, err := db.Query(`
			SELECT svp.id, svp.student_id, s.name, COALESCE(c.name, '-'), svp.points, svp.description, COALESCE(svp.recorded_by, 'Admin'), COALESCE(svp.created_at, '')
			FROM student_violation_points svp
			JOIN students s ON svp.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE svp.academic_year = ? AND svp.deleted_at IS NULL
			ORDER BY svp.id DESC
			LIMIT 10
		`, academicYear)
		if err == nil {
			defer vRows.Close()
			for vRows.Next() {
				var it ActivityItem
				it.Type = "violation"
				if err := vRows.Scan(&it.ID, &it.StudentID, &it.StudentName, &it.ClassName, &it.Points, &it.Description, &it.RecordedBy, &it.CreatedAt); err == nil {
					violations = append(violations, it)
				}
			}
		}

		if achievements == nil {
			achievements = []ActivityItem{}
		}
		if violations == nil {
			violations = []ActivityItem{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"achievements": achievements,
			"violations":   violations,
		})
	}
}
