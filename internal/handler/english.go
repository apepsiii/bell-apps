package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// GetMyEnglishQuest returns today's English quest for the logged-in student
func GetMyEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Get("student_id")
		if studentID == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}

		today := time.Now().Format("2006-01-02")
		var quest struct {
			ID       int    `json:"id"`
			Question string `json:"question"`
			Date     string `json:"date"`
		}

		err := db.QueryRow(`
			SELECT id, question, date 
			FROM english_quests 
			WHERE date = ?
		`, today).Scan(&quest.ID, &quest.Question, &quest.Date)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, map[string]interface{}{"quest": nil})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"quest": quest})
	}
}

// SubmitEnglishQuest submits a student's answer to today's quest
func SubmitEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Get("student_id")
		if studentID == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}

		questID := c.FormValue("quest_id")
		answer := c.FormValue("answer")

		if questID == "" || answer == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
		}

		_, err := db.Exec(`
			INSERT INTO english_submissions (quest_id, student_id, answer, submitted_at, status)
			VALUES (?, ?, ?, ?, 'pending')
		`, questID, studentID, answer, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Submission successful"})
	}
}

// GetMyEnglishProfile returns the English quest profile for the logged-in student
func GetMyEnglishProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Get("student_id")
		if studentID == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}

		var profile struct {
			TotalSubmissions int `json:"total_submissions"`
			ApprovedCount    int `json:"approved_count"`
			PendingCount     int `json:"pending_count"`
			RejectedCount    int `json:"rejected_count"`
			TotalPoints      int `json:"total_points"`
		}

		db.QueryRow(`
			SELECT 
				COUNT(*) as total,
				SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) as approved,
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
				SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected,
				COALESCE(SUM(CASE WHEN status = 'approved' THEN points_earned ELSE 0 END), 0) as points
			FROM english_submissions
			WHERE student_id = ?
		`, studentID).Scan(&profile.TotalSubmissions, &profile.ApprovedCount, &profile.PendingCount, &profile.RejectedCount, &profile.TotalPoints)

		return c.JSON(http.StatusOK, profile)
	}
}

// GetEnglishLeaderboard returns the English quest leaderboard
func GetEnglishLeaderboard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT 
				s.id, s.name, s.nis, c.name as class_name,
				COALESCE(SUM(es.points_earned), 0) as total_points,
				COUNT(CASE WHEN es.status = 'approved' THEN 1 END) as approved_count
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN english_submissions es ON s.id = es.student_id
			WHERE s.status = 'active'
			GROUP BY s.id
			HAVING total_points > 0
			ORDER BY total_points DESC, approved_count DESC
			LIMIT 50
		`)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var leaderboard []map[string]interface{}
		rank := 1
		for rows.Next() {
			var id, approvedCount int
			var name, nis, className string
			var totalPoints int

			if err := rows.Scan(&id, &name, &nis, &className, &totalPoints, &approvedCount); err != nil {
				continue
			}

			leaderboard = append(leaderboard, map[string]interface{}{
				"rank":           rank,
				"id":             id,
				"name":           name,
				"nis":            nis,
				"class_name":     className,
				"total_points":   totalPoints,
				"approved_count": approvedCount,
			})
			rank++
		}

		return c.JSON(http.StatusOK, leaderboard)
	}
}

// GetTodayEnglishQuest returns today's English quest (admin)
func GetTodayEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		today := time.Now().Format("2006-01-02")
		var quest struct {
			ID       int    `json:"id"`
			Question string `json:"question"`
			Date     string `json:"date"`
		}

		err := db.QueryRow(`
			SELECT id, question, date 
			FROM english_quests 
			WHERE date = ?
		`, today).Scan(&quest.ID, &quest.Question, &quest.Date)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, map[string]interface{}{"quest": nil})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"quest": quest})
	}
}

// CreateEnglishQuest creates a new English quest (admin)
func CreateEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		question := c.FormValue("question")
		date := c.FormValue("date")

		if question == "" || date == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
		}

		_, err := db.Exec(`
			INSERT INTO english_quests (question, date, created_at)
			VALUES (?, ?, ?)
		`, question, date, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Quest created successfully"})
	}
}

// GetEnglishSubmissions returns all submissions for review (admin)
func GetEnglishSubmissions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status")
		if status == "" {
			status = "pending"
		}

		rows, err := db.Query(`
			SELECT 
				es.id, es.quest_id, es.student_id, es.answer, es.submitted_at, es.status,
				COALESCE(es.review_feedback, '') as feedback, COALESCE(es.points_earned, 0) as points,
				eq.question, s.name as student_name, s.nis, c.name as class_name
			FROM english_submissions es
			JOIN english_quests eq ON es.quest_id = eq.id
			JOIN students s ON es.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE es.status = ?
			ORDER BY es.submitted_at DESC
		`, status)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var submissions []map[string]interface{}
		for rows.Next() {
			var id, questID, studentID, points int
			var answer, submittedAt, statusVal, feedback, question, studentName, nis, className string

			if err := rows.Scan(&id, &questID, &studentID, &answer, &submittedAt, &statusVal, &feedback, &points, &question, &studentName, &nis, &className); err != nil {
				continue
			}

			submissions = append(submissions, map[string]interface{}{
				"id":           id,
				"quest_id":     questID,
				"student_id":   studentID,
				"answer":       answer,
				"submitted_at": submittedAt,
				"status":       statusVal,
				"feedback":     feedback,
				"points":       points,
				"question":     question,
				"student_name": studentName,
				"nis":          nis,
				"class_name":   className,
			})
		}

		return c.JSON(http.StatusOK, submissions)
	}
}

// ReviewEnglishSubmission reviews and grades a submission (admin)
func ReviewEnglishSubmission(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		status := c.FormValue("status")
		feedback := c.FormValue("feedback")
		pointsStr := c.FormValue("points")

		if status == "" || (status != "approved" && status != "rejected") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid status"})
		}

		points := 0
		if status == "approved" && pointsStr != "" {
			var err error
			points, err = strconv.Atoi(pointsStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid points value"})
			}
		}

		_, err := db.Exec(`
			UPDATE english_submissions
			SET status = ?, review_feedback = ?, points_earned = ?, reviewed_at = ?
			WHERE id = ?
		`, status, feedback, points, time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Review submitted successfully"})
	}
}

// GetEnglishProgressReport returns progress report for all students (admin)
func GetEnglishProgressReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT 
				s.id, s.name, s.nis, c.name as class_name,
				COUNT(es.id) as total_submissions,
				SUM(CASE WHEN es.status = 'approved' THEN 1 ELSE 0 END) as approved,
				SUM(CASE WHEN es.status = 'pending' THEN 1 ELSE 0 END) as pending,
				SUM(CASE WHEN es.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
				COALESCE(SUM(CASE WHEN es.status = 'approved' THEN es.points_earned ELSE 0 END), 0) as total_points
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN english_submissions es ON s.id = es.student_id
			WHERE s.status = 'active'
			GROUP BY s.id
			ORDER BY total_points DESC, approved DESC
		`)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var report []map[string]interface{}
		for rows.Next() {
			var id, totalSubmissions, approved, pending, rejected, totalPoints int
			var name, nis, className string

			if err := rows.Scan(&id, &name, &nis, &className, &totalSubmissions, &approved, &pending, &rejected, &totalPoints); err != nil {
				continue
			}

			report = append(report, map[string]interface{}{
				"id":                id,
				"name":              name,
				"nis":               nis,
				"class_name":        className,
				"total_submissions": totalSubmissions,
				"approved":          approved,
				"pending":           pending,
				"rejected":          rejected,
				"total_points":      totalPoints,
			})
		}

		return c.JSON(http.StatusOK, report)
	}
}
