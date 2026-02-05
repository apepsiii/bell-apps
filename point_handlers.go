package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// --- STRUCTS ---

type PointRule struct {
	ID          int    `json:"id"`
	Category    string `json:"category"` // 'achievement' or 'violation'
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
	StudentID   int               `json:"student_id"`
	Name        string            `json:"name"`
	ClassName   string            `json:"class_name"`
	TotalPoints int               `json:"total_points"`
	History     []StudentPointLog `json:"history"`
}

// --- HANDLERS ---

// 1. Get All Rules
func (a *App) GetPointRulesHandler(c echo.Context) error {
	rows, err := a.DB.Query("SELECT id, category, name, points, description FROM point_rules ORDER BY category, name")
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
	// Return empty array if nil
	if rules == nil {
		rules = []PointRule{}
	}
	return c.JSON(http.StatusOK, rules)
}

// 2. Add New Rule
func (a *App) AddPointRuleHandler(c echo.Context) error {
	category := c.FormValue("category")
	name := c.FormValue("name")
	pointsStr := c.FormValue("points")
	desc := c.FormValue("description")

	points, _ := strconv.Atoi(pointsStr)

	_, err := a.DB.Exec("INSERT INTO point_rules (category, name, points, description) VALUES (?, ?, ?, ?)", category, name, points, desc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Aturan poin berhasil ditambahkan"})
}

// 3. Delete Rule
func (a *App) DeletePointRuleHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM point_rules WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Aturan dihapus"})
}

// 4. Get Student Points Profile (with Total & History)
func (a *App) GetStudentPointProfileHandler(c echo.Context) error {
	studentID := c.Param("id") // or query param? let's use param /api/points/student/:id

	// Get Student Info
	var profile StudentPointProfile
	err := a.DB.QueryRow(`
		SELECT s.id, s.name, c.name, COALESCE(SUM(sp.points_change), 0) as total
		FROM students s
		LEFT JOIN classes c ON s.class_id = c.id
		LEFT JOIN student_points sp ON s.id = sp.student_id
		WHERE s.id = ?
		GROUP BY s.id
	`, studentID).Scan(&profile.StudentID, &profile.Name, &profile.ClassName, &profile.TotalPoints)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}

	// Get History
	rows, err := a.DB.Query(`
		SELECT id, student_id, rule_id, reward_id, points_change, description, timestamp, recorded_by
		FROM student_points 
		WHERE student_id = ? 
		ORDER BY id DESC LIMIT 50`, studentID)
	
	if err == nil {
		defer rows.Close()
		for rows.Next() { // FIX: Use rows.Next() instead of checking err
			var l StudentPointLog
			// Handle nullable fields
			var ruleID, rewardID sql.NullInt64
			
			rows.Scan(&l.ID, &l.StudentID, &ruleID, &rewardID, &l.PointsChange, &l.Description, &l.Timestamp, &l.RecordedBy)
			
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

// 5. Transaction: Give Points (Achievement/Violation)
func (a *App) AddPointTransactionHandler(c echo.Context) error {
	studentID := c.FormValue("student_id")
	ruleID := c.FormValue("rule_id")
	// If custom/manual input not via rule
	// customPoints := c.FormValue("custom_points") 
	// description := c.FormValue("description")
	
	// For now assume strictly rule-based
	var points int
	var desc string
	
	// Get Rule Info
	err := a.DB.QueryRow("SELECT points, name FROM point_rules WHERE id=?", ruleID).Scan(&points, &desc)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Aturan tidak valid"})
	}

	// Insert Transaction
	_, err = a.DB.Exec(`
		INSERT INTO student_points (student_id, rule_id, points_change, description, recorded_by)
		VALUES (?, ?, ?, ?, ?)`, studentID, ruleID, points, desc, "Admin") // Simplified "Admin" for now
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Poin berhasil dicatat"})
}

// 6. Leaderboard
func (a *App) GetLeaderboardHandler(c echo.Context) error {
	rows, err := a.DB.Query(`
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
		ID int `json:"id"`
		Name string `json:"name"`
		ClassName string `json:"class_name"`
		Points int `json:"points"`
	}

	var items []LeaderboardItem
	for rows.Next() {
		var i LeaderboardItem
		rows.Scan(&i.ID, &i.Name, &i.ClassName, &i.Points)
		items = append(items, i)
	}

	if items == nil {
		items = []LeaderboardItem{}
	}
	return c.JSON(http.StatusOK, items)
}

// --- REWARD HANDLERS ---

// 7. Get All Rewards
func (a *App) GetPointRewardsHandler(c echo.Context) error {
	rows, err := a.DB.Query("SELECT id, name, points_cost, stock, description FROM point_rewards ORDER BY points_cost")
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

// 8. Add Reward
func (a *App) AddPointRewardHandler(c echo.Context) error {
	name := c.FormValue("name")
	pointsStr := c.FormValue("points_cost")
	stockStr := c.FormValue("stock")
	desc := c.FormValue("description")

	points, _ := strconv.Atoi(pointsStr)
	stock, _ := strconv.Atoi(stockStr)

	_, err := a.DB.Exec("INSERT INTO point_rewards (name, points_cost, stock, description) VALUES (?, ?, ?, ?)", name, points, stock, desc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Hadiah berhasil ditambahkan"})
}

// 9. Delete Reward
func (a *App) DeletePointRewardHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM point_rewards WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Hadiah dihapus"})
}

// 10. Redeem Reward
func (a *App) RedeemRewardHandler(c echo.Context) error {
	studentID := c.FormValue("student_id")
	rewardID := c.FormValue("reward_id")

	// 1. Get Reward Info & Stock
	var pointsCost, stock int
	var rewardName string
	err := a.DB.QueryRow("SELECT name, points_cost, stock FROM point_rewards WHERE id=?", rewardID).Scan(&rewardName, &pointsCost, &stock)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Hadiah tidak valid"})
	}

	if stock <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Stok hadiah habis"})
	}

	// 2. Check Student Balance
	var currentPoints int
	err = a.DB.QueryRow(`
		SELECT COALESCE(SUM(points_change), 0) 
		FROM student_points 
		WHERE student_id=?`, studentID).Scan(&currentPoints)
	
	if currentPoints < pointsCost {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Poin tidak mencukupi"})
	}

	// 3. Process Transaction (Use Transaction for safety)
	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database Error"})
	}

	// Deduct Stock
	_, err = tx.Exec("UPDATE point_rewards SET stock = stock - 1 WHERE id=?", rewardID)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal update stok"})
	}

	// Record Points Deduction
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

// 11. Get Student by RFID (Lookup for Frontend)
func (a *App) GetStudentByRFIDHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID required"})
	}

	var student struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		ClassName string `json:"class_name"`
	}

	err := a.DB.QueryRow(`
		SELECT s.id, s.name, c.name 
		FROM students s 
		LEFT JOIN classes c ON s.class_id = c.id 
		WHERE s.rfid_uid = ?`, rfid).Scan(&student.ID, &student.Name, &student.ClassName)
	
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}

	return c.JSON(http.StatusOK, student)
}
