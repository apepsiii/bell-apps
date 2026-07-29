package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func GetTodayEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		date := c.QueryParam("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		var q struct {
			ID              int    `json:"id"`
			Date            string `json:"date"`
			Title           string `json:"title"`
			Description     string `json:"description"`
			QuestType       string `json:"quest_type"`
			Topic           string `json:"topic"`
			VocabularyWords string `json:"vocabulary_words"`
			QuizQuestion    string `json:"quiz_question"`
			QuizChoices     string `json:"quiz_choices"`
			QuizAnswer      string `json:"quiz_answer"`
			XPReward        int    `json:"xp_reward"`
		}
		err := db.QueryRow("SELECT id, date, title, description, quest_type, topic, vocabulary_words, quiz_question, quiz_choices, quiz_answer, xp_reward FROM english_quests WHERE date = ?", date).
			Scan(&q.ID, &q.Date, &q.Title, &q.Description, &q.QuestType, &q.Topic, &q.VocabularyWords, &q.QuizQuestion, &q.QuizChoices, &q.QuizAnswer, &q.XPReward)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Quest belum dibuat untuk hari ini"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, q)
	}
}

func CreateEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		date := c.FormValue("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		xpReward, _ := strconv.Atoi(c.FormValue("xp_reward"))
		if xpReward == 0 {
			xpReward = 10
		}
		_, err := db.Exec("INSERT OR REPLACE INTO english_quests (date, title, description, quest_type, topic, vocabulary_words, quiz_question, quiz_choices, quiz_answer, xp_reward) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			date, c.FormValue("title"), c.FormValue("description"), c.FormValue("quest_type"),
			c.FormValue("topic"), c.FormValue("vocabulary_words"), c.FormValue("quiz_question"),
			c.FormValue("quiz_choices"), c.FormValue("quiz_answer"), xpReward)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Quest berhasil disimpan"})
	}
}

func GetEnglishSubmissions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		date := c.QueryParam("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		query := "SELECT es.id, es.student_id, s.name, COALESCE(cl.name,''), es.quest_type, es.content, es.audio_file, es.vocab_words, es.quiz_answer, es.xp_earned, es.status, es.feedback, es.submitted_at, COALESCE(es.reviewed_by,''), COALESCE(es.reviewed_at,'') FROM english_submissions es JOIN students s ON es.student_id = s.id LEFT JOIN classes cl ON s.class_id = cl.id JOIN english_quests eq ON es.quest_id = eq.id WHERE eq.date = ?"
		args := []interface{}{date}
		if st := c.QueryParam("status"); st != "" {
			query += " AND es.status = ?"
			args = append(args, st)
		}
		if cid := c.QueryParam("class_id"); cid != "" {
			query += " AND s.class_id = ?"
			args = append(args, cid)
		}
		query += " ORDER BY es.submitted_at DESC"
		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()
		type Sub struct {
			ID          int    `json:"id"`
			StudentID   int    `json:"student_id"`
			StudentName string `json:"student_name"`
			ClassName   string `json:"class_name"`
			QuestType   string `json:"quest_type"`
			Content     string `json:"content"`
			AudioFile   string `json:"audio_file"`
			VocabWords  string `json:"vocab_words"`
			QuizAnswer  string `json:"quiz_answer"`
			XPEarned    int    `json:"xp_earned"`
			Status      string `json:"status"`
			Feedback    string `json:"feedback"`
			SubmittedAt string `json:"submitted_at"`
			ReviewedBy  string `json:"reviewed_by"`
			ReviewedAt  string `json:"reviewed_at"`
		}
		var subs []Sub
		for rows.Next() {
			var s Sub
			rows.Scan(&s.ID, &s.StudentID, &s.StudentName, &s.ClassName, &s.QuestType, &s.Content, &s.AudioFile, &s.VocabWords, &s.QuizAnswer, &s.XPEarned, &s.Status, &s.Feedback, &s.SubmittedAt, &s.ReviewedBy, &s.ReviewedAt)
			subs = append(subs, s)
		}
		if subs == nil {
			subs = []Sub{}
		}
		return c.JSON(http.StatusOK, subs)
	}
}

func ReviewEnglishSubmission(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		action := c.FormValue("action")
		feedback := c.FormValue("feedback")
		if action != "approve" && action != "reject" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "action harus approve atau reject"})
		}
		var studentID, questID, xpReward int
		var questTitle string
		err := db.QueryRow("SELECT es.student_id, es.quest_id, eq.xp_reward, eq.title FROM english_submissions es JOIN english_quests eq ON es.quest_id = eq.id WHERE es.id = ? AND es.status = 'pending'", id).Scan(&studentID, &questID, &xpReward, &questTitle)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Submission tidak ditemukan atau sudah direview"})
		}
		status := "approved"
		xpEarned := xpReward
		if action == "reject" {
			status = "rejected"
			xpEarned = 0
		}
		tx, _ := db.Begin()
		tx.Exec("UPDATE english_submissions SET status=?, feedback=?, xp_earned=?, reviewed_at=CURRENT_TIMESTAMP, reviewed_by='admin' WHERE id=?", status, feedback, xpEarned, id)
		if action == "approve" {
			englishUpdateStreak(tx, studentID, xpEarned)
			englishCheckBadges(tx, studentID)
			// Mirror the XP into the main student_points system so it shows up
			// in the unified leaderboard and student point profile.
			tx.Exec(`INSERT INTO student_points (student_id, points_change, description, recorded_by) VALUES (?, ?, ?, ?)`,
				studentID, xpEarned, "English Daily Quest: "+questTitle, "english-quest")
		}
		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Submission berhasil di-review"})
	}
}

func GetEnglishLeaderboard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT es.student_id, s.name, COALESCE(c.name,''), es.current_streak, es.longest_streak, es.total_xp, COUNT(eb.id) FROM english_streaks es JOIN students s ON es.student_id = s.id LEFT JOIN classes c ON s.class_id = c.id LEFT JOIN english_student_badges eb ON es.student_id = eb.student_id GROUP BY es.student_id ORDER BY es.total_xp DESC LIMIT 50")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()
		type Item struct {
			StudentID     int    `json:"student_id"`
			Name          string `json:"name"`
			ClassName     string `json:"class_name"`
			CurrentStreak int    `json:"current_streak"`
			LongestStreak int    `json:"longest_streak"`
			TotalXP       int    `json:"total_xp"`
			BadgeCount    int    `json:"badge_count"`
		}
		var items []Item
		for rows.Next() {
			var i Item
			rows.Scan(&i.StudentID, &i.Name, &i.ClassName, &i.CurrentStreak, &i.LongestStreak, &i.TotalXP, &i.BadgeCount)
			items = append(items, i)
		}
		if items == nil { items = []Item{} }
		return c.JSON(http.StatusOK, items)
	}
}

func GetEnglishProgressReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		month := c.QueryParam("month")
		year := c.QueryParam("year")
		if month == "" { month = time.Now().Format("01") }
		if year == "" { year = time.Now().Format("2006") }
		query := "SELECT s.id, s.name, COALESCE(c.name,''), COALESCE(st.current_streak,0), COALESCE(st.total_xp,0), COUNT(DISTINCT CASE WHEN es.status='approved' THEN es.id END), COUNT(DISTINCT es.id) FROM students s LEFT JOIN classes c ON s.class_id = c.id LEFT JOIN english_streaks st ON s.id = st.student_id LEFT JOIN english_submissions es ON s.id = es.student_id LEFT JOIN english_quests eq ON es.quest_id = eq.id AND strftime('%m', eq.date) = ? AND strftime('%Y', eq.date) = ? WHERE s.status = 'active'"
		args := []interface{}{month, year}
		if cid := c.QueryParam("class_id"); cid != "" {
			query += " AND s.class_id = ?"
			args = append(args, cid)
		}
		query += " GROUP BY s.id ORDER BY COUNT(DISTINCT CASE WHEN es.status='approved' THEN es.id END) DESC"
		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()
		type Item struct {
			StudentID int    `json:"student_id"`
			Name      string `json:"name"`
			ClassName string `json:"class_name"`
			Streak    int    `json:"current_streak"`
			TotalXP   int    `json:"total_xp"`
			Approved  int    `json:"approved"`
			TotalSub  int    `json:"total_sub"`
		}
		var items []Item
		for rows.Next() {
			var i Item
			rows.Scan(&i.StudentID, &i.Name, &i.ClassName, &i.Streak, &i.TotalXP, &i.Approved, &i.TotalSub)
			items = append(items, i)
		}
		if items == nil { items = []Item{} }
		return c.JSON(http.StatusOK, items)
	}
}

func GetMyEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		date := time.Now().Format("2006-01-02")
		var q struct {
			ID              int    `json:"id"`
			Date            string `json:"date"`
			Title           string `json:"title"`
			Description     string `json:"description"`
			QuestType       string `json:"quest_type"`
			Topic           string `json:"topic"`
			VocabularyWords string `json:"vocabulary_words"`
			QuizQuestion    string `json:"quiz_question"`
			QuizChoices     string `json:"quiz_choices"`
			XPReward        int    `json:"xp_reward"`
			Submitted       bool   `json:"submitted"`
			SubStatus       string `json:"sub_status"`
		}
		err := db.QueryRow("SELECT id, date, title, description, quest_type, topic, vocabulary_words, quiz_question, quiz_choices, xp_reward FROM english_quests WHERE date = ?", date).
			Scan(&q.ID, &q.Date, &q.Title, &q.Description, &q.QuestType, &q.Topic, &q.VocabularyWords, &q.QuizQuestion, &q.QuizChoices, &q.XPReward)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Quest belum tersedia hari ini"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		var subStatus string
		if db.QueryRow("SELECT status FROM english_submissions WHERE student_id = ? AND quest_id = ?", studentID, q.ID).Scan(&subStatus) == nil {
			q.Submitted = true
			q.SubStatus = subStatus
		}
		return c.JSON(http.StatusOK, q)
	}
}

func SubmitEnglishQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		questID, _ := strconv.Atoi(c.FormValue("quest_id"))
		var count int
		db.QueryRow("SELECT COUNT(*) FROM english_submissions WHERE student_id = ? AND quest_id = ?", studentID, questID).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kamu sudah submit quest ini"})
		}
		_, err := db.Exec("INSERT INTO english_submissions (student_id, quest_id, quest_type, content, vocab_words, quiz_answer, status) VALUES (?, ?, ?, ?, ?, ?, 'pending')",
			studentID, questID, c.FormValue("quest_type"), c.FormValue("content"), c.FormValue("vocab_words"), c.FormValue("quiz_answer"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Setoran berhasil dikirim! Tunggu review dari guru."})
	}
}

func GetMyEnglishProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		var streak struct {
			Current  int    `json:"current_streak"`
			Longest  int    `json:"longest_streak"`
			TotalXP  int    `json:"total_xp"`
			LastDate string `json:"last_submit_date"`
		}
		// Use COALESCE to handle null values if row doesn't exist
		err := db.QueryRow("SELECT COALESCE(current_streak, 0), COALESCE(longest_streak, 0), COALESCE(total_xp, 0), COALESCE(last_submit_date, '') FROM english_streaks WHERE student_id = ?", studentID).
			Scan(&streak.Current, &streak.Longest, &streak.TotalXP, &streak.LastDate)
		if err != nil {
			// If no record exists, initialize with zeros
			streak.Current = 0
			streak.Longest = 0
			streak.TotalXP = 0
			streak.LastDate = ""
		}
		
		type Badge struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			Desc     string `json:"description"`
			Icon     string `json:"icon"`
			EarnedAt string `json:"earned_at"`
		}
		var badges []Badge
		brows, _ := db.Query("SELECT eb.code, eb.name, eb.description, eb.icon, esb.earned_at FROM english_student_badges esb JOIN english_badges eb ON esb.badge_id = eb.id WHERE esb.student_id = ? ORDER BY esb.earned_at DESC", studentID)
		if brows != nil {
			defer brows.Close()
			for brows.Next() {
				var b Badge
				brows.Scan(&b.Code, &b.Name, &b.Desc, &b.Icon, &b.EarnedAt)
				badges = append(badges, b)
			}
		}
		if badges == nil { badges = []Badge{} }
		type HistItem struct {
			QuestType   string `json:"quest_type"`
			Title       string `json:"title"`
			Status      string `json:"status"`
			XPEarned    int    `json:"xp_earned"`
			SubmittedAt string `json:"submitted_at"`
			Feedback    string `json:"feedback"`
		}
		var history []HistItem
		hrows, _ := db.Query("SELECT es.quest_type, eq.title, es.status, es.xp_earned, es.submitted_at, COALESCE(es.feedback, '') FROM english_submissions es JOIN english_quests eq ON es.quest_id = eq.id WHERE es.student_id = ? ORDER BY es.submitted_at DESC LIMIT 30", studentID)
		if hrows != nil {
			defer hrows.Close()
			for hrows.Next() {
				var h HistItem
				hrows.Scan(&h.QuestType, &h.Title, &h.Status, &h.XPEarned, &h.SubmittedAt, &h.Feedback)
				history = append(history, h)
			}
		}
		if history == nil { history = []HistItem{} }
		today := time.Now().Format("2006-01-02")
		var todayCount int
		db.QueryRow("SELECT COUNT(*) FROM english_submissions es JOIN english_quests eq ON es.quest_id = eq.id WHERE es.student_id = ? AND eq.date = ?", studentID, today).Scan(&todayCount)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"streak":     streak,
			"badges":     badges,
			"history":    history,
			"today_done": todayCount > 0,
		})
	}
}

func englishUpdateStreak(tx *sql.Tx, studentID, xpEarned int) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var current, longest int
	var lastDate string
	tx.QueryRow("SELECT current_streak, longest_streak, last_submit_date FROM english_streaks WHERE student_id = ?", studentID).
		Scan(&current, &longest, &lastDate)
	if lastDate == today {
		tx.Exec("UPDATE english_streaks SET total_xp = total_xp + ? WHERE student_id = ?", xpEarned, studentID)
		return
	}
	if lastDate == yesterday {
		current++
	} else {
		current = 1
	}
	if current > longest {
		longest = current
	}
	tx.Exec("INSERT INTO english_streaks (student_id, current_streak, longest_streak, total_xp, last_submit_date) VALUES (?, ?, ?, ?, ?) ON CONFLICT(student_id) DO UPDATE SET current_streak=?, longest_streak=?, total_xp=total_xp+?, last_submit_date=?",
		studentID, current, longest, xpEarned, today, current, longest, xpEarned, today)
}

func englishCheckBadges(tx *sql.Tx, studentID int) {
	var curStreak, longestStreak, totalXP, totalSubs, vocabCount, quizCount, writingCount int
	tx.QueryRow("SELECT current_streak, longest_streak, total_xp FROM english_streaks WHERE student_id = ?", studentID).
		Scan(&curStreak, &longestStreak, &totalXP)
	tx.QueryRow("SELECT COUNT(*) FROM english_submissions WHERE student_id = ? AND status='approved'", studentID).Scan(&totalSubs)
	tx.QueryRow("SELECT COUNT(*) FROM english_submissions WHERE student_id = ? AND status='approved' AND quest_type='vocabulary'", studentID).Scan(&vocabCount)
	tx.QueryRow("SELECT COUNT(*) FROM english_submissions WHERE student_id = ? AND status='approved' AND quest_type='quiz'", studentID).Scan(&quizCount)
	tx.QueryRow("SELECT COUNT(*) FROM english_submissions WHERE student_id = ? AND status='approved' AND quest_type='written'", studentID).Scan(&writingCount)
	rows, err := tx.Query("SELECT id, condition_type, condition_value FROM english_badges")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var badgeID, condVal int
		var condType string
		rows.Scan(&badgeID, &condType, &condVal)
		met := false
		switch condType {
		case "streak":               met = curStreak >= condVal || longestStreak >= condVal
		case "total_xp":             met = totalXP >= condVal
		case "total_submissions":    met = totalSubs >= condVal
		case "vocab_submissions":    met = vocabCount >= condVal
		case "quiz_correct":         met = quizCount >= condVal
		case "writing_submissions":  met = writingCount >= condVal
		}
		if met {
			tx.Exec("INSERT OR IGNORE INTO english_student_badges (student_id, badge_id) VALUES (?, ?)", studentID, badgeID)
		}
	}
}
