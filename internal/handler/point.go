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
	StudentID   int               `json:"student_id"`
	Name        string            `json:"name"`
	ClassName   string            `json:"class_name"`
	TotalPoints int               `json:"total_points"`
	History     []StudentPointLog `json:"history"`
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
		var className sql.NullString
		err := db.QueryRow(`
			SELECT s.id, s.name, c.name, COALESCE(SUM(sp.points_change), 0) as total
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN student_points sp ON s.id = sp.student_id
			WHERE s.id = ?
			GROUP BY s.id
		`, studentID).Scan(&profile.StudentID, &profile.Name, &className, &profile.TotalPoints)
		profile.ClassName = className.String

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

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
