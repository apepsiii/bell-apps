package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type PointRule struct {
	ID          int    `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

type PointReward struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PointsCost  int    `json:"points_cost"`
	Stock       int    `json:"stock"`
	Description string `json:"description"`
}

type StudentPointLog struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	RuleID       *int   `json:"rule_id"`
	RewardID     *int   `json:"reward_id"`
	PointsChange int    `json:"points_change"`
	Description  string `json:"description"`
	Timestamp    string `json:"timestamp"`
	RecordedBy   string `json:"recorded_by"`
}

type StudentPointProfile struct {
	Student     StudentInfo       `json:"student"`
	TotalPoints int               `json:"total_points"`
	History     []StudentPointLog `json:"history"`
}

type StudentInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Class    string `json:"class"`
	NIS      string `json:"nis"`
	Photo    string `json:"photo"`
}

func GetPointRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT id, category, name, points, description FROM point_rules ORDER BY category, name")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var rules []PointRule
		for rows.Next() {
			var r PointRule
			if err := rows.Scan(&r.ID, &r.Category, &r.Name, &r.Points, &r.Description); err != nil {
				continue
			}
			rules = append(rules, r)
		}
		if rules == nil {
			rules = []PointRule{}
		}
		return c.JSON(http.StatusOK, rules)
	}
}

func AddPointRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		category := c.FormValue("category")
		name := c.FormValue("name")
		pointsStr := c.FormValue("points")
		desc := c.FormValue("description")

		points, _ := strconv.Atoi(pointsStr)

		_, err := db.Exec("INSERT INTO point_rules (category, name, points, description) VALUES (?, ?, ?, ?)", category, name, points, desc)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Aturan poin berhasil ditambahkan"})
	}
}

func DeletePointRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM point_rules WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Aturan dihapus"})
	}
}

func GetStudentPointProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Param("id")

		var profile StudentPointProfile
		var className, nis, photo sql.NullString

		err := db.QueryRow(`
			SELECT s.id, s.name, s.nis, s.photo, c.name, COALESCE(SUM(sp.points_change), 0) as total
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN student_points sp ON s.id = sp.student_id
			WHERE s.id = ?
			GROUP BY s.id
		`, studentID).Scan(&profile.Student.ID, &profile.Student.Name, &nis, &photo, &className, &profile.TotalPoints)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Siswa tidak ditemukan"})
		}

		profile.Student.Class = className.String
		profile.Student.NIS = nis.String
		profile.Student.Photo = photo.String

		rows, err := db.Query(`
			SELECT id, student_id, rule_id, reward_id, points_change, description, timestamp, recorded_by
			FROM student_points
			WHERE student_id = ?
			ORDER BY id DESC LIMIT 50`, studentID)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var l StudentPointLog
				var ruleID, rewardID sql.NullInt64
				var recordedBy sql.NullString

				rows.Scan(&l.ID, &l.StudentID, &ruleID, &rewardID, &l.PointsChange, &l.Description, &l.Timestamp, &recordedBy)

				l.RecordedBy = recordedBy.String

				if ruleID.Valid {
					val := int(ruleID.Int64)
					l.RuleID = &val
				}
				if rewardID.Valid {
					val := int(rewardID.Int64)
					l.RewardID = &val
				}
				profile.History = append(profile.History, l)
			}
		}

		if profile.History == nil {
			profile.History = []StudentPointLog{}
		}

		return c.JSON(http.StatusOK, profile)
	}
}

func AddPointTransaction(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.FormValue("student_id")
		ruleID := c.FormValue("rule_id")

		var points int
		var desc string

		err := db.QueryRow("SELECT points, name FROM point_rules WHERE id=?", ruleID).Scan(&points, &desc)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Aturan tidak valid"})
		}

		_, err = db.Exec(`
			INSERT INTO student_points (student_id, rule_id, points_change, description, recorded_by)
			VALUES (?, ?, ?, ?, ?)`, studentID, ruleID, points, desc, "Admin")

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Poin berhasil dicatat"})
	}
}

func GetLeaderboard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT s.id, s.name, c.name, COALESCE(SUM(sp.points_change), 0) as total_points
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN student_points sp ON s.id = sp.student_id
			GROUP BY s.id
			ORDER BY total_points DESC
			LIMIT 20
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		type LeaderboardItem struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			ClassName string `json:"class_name"`
			Points    int    `json:"points"`
		}

		var items []LeaderboardItem
		for rows.Next() {
			var i LeaderboardItem
			var className sql.NullString
			rows.Scan(&i.ID, &i.Name, &className, &i.Points)
			i.ClassName = className.String
			items = append(items, i)
		}

		if items == nil {
			items = []LeaderboardItem{}
		}
		return c.JSON(http.StatusOK, items)
	}
}

func GetPointRewards(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT id, name, points_cost, stock, description FROM point_rewards ORDER BY points_cost")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var rewards []PointReward
		for rows.Next() {
			var r PointReward
			rows.Scan(&r.ID, &r.Name, &r.PointsCost, &r.Stock, &r.Description)
			rewards = append(rewards, r)
		}
		if rewards == nil {
			rewards = []PointReward{}
		}
		return c.JSON(http.StatusOK, rewards)
	}
}

func AddPointReward(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		pointsStr := c.FormValue("points_cost")
		stockStr := c.FormValue("stock")
		desc := c.FormValue("description")

		points, _ := strconv.Atoi(pointsStr)
		stock, _ := strconv.Atoi(stockStr)

		_, err := db.Exec("INSERT INTO point_rewards (name, points_cost, stock, description) VALUES (?, ?, ?, ?)", name, points, stock, desc)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Hadiah berhasil ditambahkan"})
	}
}

func DeletePointReward(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM point_rewards WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Hadiah dihapus"})
	}
}

func RedeemReward(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.FormValue("student_id")
		rewardID := c.FormValue("reward_id")

		var pointsCost, stock int
		var rewardName string
		err := db.QueryRow("SELECT name, points_cost, stock FROM point_rewards WHERE id=?", rewardID).Scan(&rewardName, &pointsCost, &stock)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Hadiah tidak valid"})
		}

		if stock <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Stok hadiah habis"})
		}

		var currentPoints int
		err = db.QueryRow(`
			SELECT COALESCE(SUM(points_change), 0) 
			FROM student_points 
			WHERE student_id=?`, studentID).Scan(&currentPoints)

		if currentPoints < pointsCost {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Poin tidak mencukupi"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database Error"})
		}

		_, err = tx.Exec("UPDATE point_rewards SET stock = stock - 1 WHERE id=?", rewardID)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update stok"})
		}

		_, err = tx.Exec(`
			INSERT INTO student_points (student_id, reward_id, points_change, description, recorded_by)
			VALUES (?, ?, ?, ?, ?)`,
			studentID, rewardID, -pointsCost, "Penukaran Poin: "+rewardName, "Admin")

		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mencatat transaksi"})
		}

		tx.Commit()

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Penukaran berhasil"})
	}
}

func GetStudentByRFID(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.QueryParam("rfid")
		if rfid == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID required"})
		}

		var student struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			ClassName string `json:"class_name"`
		}

		var className sql.NullString
		err := db.QueryRow(`
			SELECT s.id, s.name, c.name 
			FROM students s 
			LEFT JOIN classes c ON s.class_id = c.id 
			WHERE s.rfid_uid = ?`, rfid).Scan(&student.ID, &student.Name, &className)
		student.ClassName = className.String

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		return c.JSON(http.StatusOK, student)
	}
}

// --- POINT CLAIM TYPES ---
type PointClaimRule struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Points      int    `json:"points"`
	Tier        string `json:"tier"`
}

type PointClaim struct {
	ID          int    `json:"id"`
	StudentID   int    `json:"student_id"`
	StudentName string `json:"student_name"`
	ClassName   string `json:"class_name"`
	RuleID      int    `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	RuleCode    string `json:"rule_code"`
	Points      int    `json:"points"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at"`
	ReviewedAt  string `json:"reviewed_at"`
}

// GET /api/point-rules - Get all active point rules
func GetPointRulesV2(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT id, code, category, name, description, points, tier
			FROM point_rules
			WHERE is_active = 1
			ORDER BY category, code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var rules []PointClaimRule
		for rows.Next() {
			var r PointClaimRule
			if err := rows.Scan(&r.ID, &r.Code, &r.Category, &r.Name, &r.Description, &r.Points, &r.Tier); err != nil {
				continue
			}
			rules = append(rules, r)
		}

		if rules == nil {
			rules = []PointClaimRule{}
		}
		return c.JSON(http.StatusOK, rules)
	}
}

// POST /api/point-claims - Submit a point claim
func SubmitPointClaim(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type ClaimRequest struct {
			StudentID   int    `json:"student_id"`
			RuleID      int    `json:"rule_id"`
			Description string `json:"description"`
			Evidence    string `json:"evidence"`
		}

		var req ClaimRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		if req.StudentID == 0 || req.RuleID == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Student ID and Rule ID required"})
		}

		_, err := db.Exec(`
			INSERT INTO point_claims (student_id, rule_id, description, evidence, status)
			VALUES (?, ?, ?, ?, 'pending')
		`, req.StudentID, req.RuleID, req.Description, req.Evidence)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to submit claim: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Klaim poin berhasil disubmit untuk divalidasi"})
	}
}

// GET /api/point-claims - Get pending claims (for admin)
func GetPointClaims(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status") // pending, approved, rejected, all

		query := `
			SELECT pc.id, pc.student_id, s.name, COALESCE(c.name, '-') as class_name,
				   pc.rule_id, pr.name, pr.code, pr.points,
				   pc.description, pc.evidence, pc.status, pc.submitted_at, pc.reviewed_at
			FROM point_claims pc
			JOIN students s ON pc.student_id = s.id
			JOIN point_rules pr ON pc.rule_id = pr.id
		`

		if status != "" && status != "all" {
			query += " WHERE pc.status = ? ORDER BY pc.submitted_at DESC"
			rows, err := db.Query(query, status)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}
			defer rows.Close()
			return scanPointClaims(rows)
		}

		query += " ORDER BY pc.submitted_at DESC"
		rows, err := db.Query(query)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()
		return scanPointClaims(rows)
	}
}

func scanPointClaims(rows *sql.Rows) error {
	var claims []PointClaim
	for rows.Next() {
		var pc PointClaim
		var reviewedAt sql.NullString
		if err := rows.Scan(&pc.ID, &pc.StudentID, &pc.StudentName, &pc.ClassName,
			&pc.RuleID, &pc.RuleName, &pc.RuleCode, &pc.Points,
			&pc.Description, &pc.Evidence, &pc.Status, &pc.SubmittedAt, &reviewedAt); err != nil {
			continue
		}
		pc.ReviewedAt = reviewedAt.String
		claims = append(claims, pc)
	}
	if claims == nil {
		claims = []PointClaim{}
	}
	return nil
}

// POST /api/point-claims/:id/approve - Approve a claim
func ApprovePointClaim(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Get claim info first
		var studentID, ruleID, points int
		var ruleName string
		err := db.QueryRow(`
			SELECT pc.student_id, pc.rule_id, pr.points, pr.name
			FROM point_claims pc
			JOIN point_rules pr ON pc.rule_id = pr.id
			WHERE pc.id = ?
		`, id).Scan(&studentID, &ruleID, &points, &ruleName)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Klaim tidak ditemukan"})
		}

		// Update claim status
		_, err = db.Exec(`
			UPDATE point_claims
			SET status = 'approved', reviewed_at = datetime('now')
			WHERE id = ?
		`, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to approve claim"})
		}

		// Add points to student
		_, err = db.Exec(`
			INSERT INTO student_points (student_id, rule_id, points_change, description, recorded_by)
			VALUES (?, ?, ?, ?, 'Admin-Claim')
		`, studentID, ruleID, points, "Klaim Poin: "+ruleName)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Claim approved but failed to add points: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Klaim poin disetujui danpoin ditambahkan"})
	}
}

// POST /api/point-claims/:id/reject - Reject a claim
func RejectPointClaim(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		_, err := db.Exec(`
			UPDATE point_claims
			SET status = 'rejected', reviewed_at = datetime('now')
			WHERE id = ?
		`, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to reject claim"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Klaim poin ditolak"})
	}
}
