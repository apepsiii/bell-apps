package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// --- TEACHER QUIZ / CBT HANDLERS ---

// CreateMeetingQuiz creates a new quiz/CBT under a meeting
func CreateMeetingQuiz(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		meetingID, err := strconv.Atoi(c.FormValue("meeting_id"))
		if err != nil || meetingID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pertemuan tidak valid"})
		}

		// Verify teacher ownership
		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_meetings m
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE m.id = ? AND ta.staff_id = ?
		`, meetingID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		title := strings.TrimSpace(c.FormValue("title"))
		description := strings.TrimSpace(c.FormValue("description"))
		timeLimit, _ := strconv.Atoi(c.FormValue("time_limit_minutes"))
		passingScore, _ := strconv.ParseFloat(c.FormValue("passing_score"), 64)

		if title == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Judul kuis wajib diisi"})
		}
		if timeLimit <= 0 {
			timeLimit = 30
		}
		if passingScore <= 0 {
			passingScore = 75.0
		}

		res, err := db.Exec(`
			INSERT INTO lms_quizzes (meeting_id, title, description, time_limit_minutes, passing_score, is_active)
			VALUES (?, ?, ?, ?, ?, 1)
		`, meetingID, title, description, timeLimit, passingScore)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat kuis CBT"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Kuis CBT berhasil dibuat",
			"id":      id,
		})
	}
}

// AddQuizQuestionRequest defines the payload for creating a question with multiple options
type AddQuizQuestionRequest struct {
	QuizID        int    `json:"quiz_id"`
	QuestionText  string `json:"question_text"`
	Points        int    `json:"points"`
	CorrectIndex  int    `json:"correct_index"` // 0-based index of correct option
	OptionTexts   []string `json:"option_texts"`
}

// AddQuizQuestion creates a question with multiple choice options
func AddQuizQuestion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		var req AddQuizQuestionRequest
		if err := c.Bind(&req); err != nil || req.QuizID <= 0 || strings.TrimSpace(req.QuestionText) == "" || len(req.OptionTexts) < 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Pertanyaan dan minimal 2 opsi pilihan wajib diisi",
			})
		}

		// Verify teacher ownership
		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND ta.staff_id = ?
		`, req.QuizID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		if req.Points <= 0 {
			req.Points = 10
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memulai transaksi"})
		}
		defer tx.Rollback()

		resQ, err := tx.Exec(`
			INSERT INTO lms_quiz_questions (quiz_id, question_text, points)
			VALUES (?, ?, ?)
		`, req.QuizID, req.QuestionText, req.Points)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan soal"})
		}

		qID, _ := resQ.LastInsertId()

		for idx, optText := range req.OptionTexts {
			isCorrect := 0
			if idx == req.CorrectIndex {
				isCorrect = 1
			}
			_, err = tx.Exec(`
				INSERT INTO lms_quiz_options (question_id, option_text, is_correct)
				VALUES (?, ?, ?)
			`, qID, optText, isCorrect)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan opsi pilihan"})
			}
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memproses transaksi"})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message":     "Soal kuis berhasil ditambahkan",
			"question_id": qID,
		})
	}
}

// GetQuizAttempts returns student attempts for a quiz (Teacher view)
func GetQuizAttempts(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		quizID, err := strconv.Atoi(c.Param("id"))
		if err != nil || quizID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID kuis tidak valid"})
		}

		// Verify teacher ownership & get class_id
		var classID int
		var title string
		var passingScore float64
		err = db.QueryRow(`
			SELECT ta.class_id, q.title, q.passing_score
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND ta.staff_id = ?
		`, quizID, teacherID).Scan(&classID, &title, &passingScore)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		rows, err := db.Query(`
			SELECT s.id, s.name, COALESCE(s.nis, s.nis_siswa, ''),
			       qa.score, qa.total_correct, qa.is_passed, strftime('%Y-%m-%d %H:%M', qa.completed_at)
			FROM students s
			LEFT JOIN lms_quiz_attempts qa ON qa.student_id = s.id AND qa.quiz_id = ?
			WHERE s.class_id = ? AND s.status = 'active'
			ORDER BY qa.score DESC, s.name ASC
		`, quizID, classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca hasil kuis"})
		}
		defer rows.Close()

		attempts := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				sID int
				sName, sNIS string
				score sql.NullFloat64
				totalCorrect, isPassed sql.NullInt64
				completedAt sql.NullString
			)
			if err := rows.Scan(&sID, &sName, &sNIS, &score, &totalCorrect, &isPassed, &completedAt); err == nil {
				item := map[string]interface{}{
					"student_id":   sID,
					"student_name": sName,
					"student_nis":  sNIS,
					"attempted":    score.Valid,
				}
				if score.Valid {
					item["score"] = score.Float64
					item["total_correct"] = totalCorrect.Int64
					item["is_passed"] = isPassed.Int64 == 1
					item["completed_at"] = completedAt.String
				}
				attempts = append(attempts, item)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"quiz_id":       quizID,
			"title":         title,
			"passing_score": passingScore,
			"attempts":      attempts,
		})
	}
}

// --- STUDENT CBT QUIZ HANDLERS ---

// GetStudentQuiz retrieves a quiz for student attempt (without revealing answers)
func GetStudentQuiz(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentIDStr := studentIDFromCtx(c)
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil || studentID <= 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Tidak terautentikasi"})
		}

		quizID, err := strconv.Atoi(c.Param("id"))
		if err != nil || quizID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID kuis tidak valid"})
		}

		// Verify student is enrolled in the quiz's class
		var (
			classID, timeLimit int
			title, desc string
			passingScore float64
		)
		err = db.QueryRow(`
			SELECT ta.class_id, q.title, q.description, q.time_limit_minutes, q.passing_score
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND q.is_active = 1
		`, quizID).Scan(&classID, &title, &desc, &timeLimit, &passingScore)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Kuis tidak ditemukan atau nonaktif"})
		}

		var studentClassID int
		_ = db.QueryRow("SELECT class_id FROM students WHERE id = ?", studentID).Scan(&studentClassID)
		if studentClassID != classID {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Anda tidak memiliki akses ke kuis ini"})
		}

		// Check if already attempted
		var (
			existingScore float64
			existingPassed int
			completedAt sql.NullString
		)
		err = db.QueryRow(`
			SELECT score, is_passed, completed_at
			FROM lms_quiz_attempts
			WHERE quiz_id = ? AND student_id = ?
		`, quizID, studentID).Scan(&existingScore, &existingPassed, &completedAt)
		alreadyAttempted := (err == nil && completedAt.Valid)

		// Fetch questions
		rows, err := db.Query(`
			SELECT id, question_text, points
			FROM lms_quiz_questions
			WHERE quiz_id = ?
			ORDER BY id ASC
		`, quizID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca soal kuis"})
		}

		type rawQ struct {
			id, points int
			text string
		}
		var rawQuestions []rawQ
		for rows.Next() {
			var q rawQ
			if err := rows.Scan(&q.id, &q.text, &q.points); err == nil {
				rawQuestions = append(rawQuestions, q)
			}
		}
		rows.Close()

		questions := make([]map[string]interface{}, 0)
		for _, q := range rawQuestions {
			// Fetch options without is_correct
			optRows, _ := db.Query(`
				SELECT id, option_text
				FROM lms_quiz_options
				WHERE question_id = ?
				ORDER BY id ASC
			`, q.id)
			options := make([]map[string]interface{}, 0)
			if optRows != nil {
				for optRows.Next() {
					var (
						optID int
						optText string
					)
					if err := optRows.Scan(&optID, &optText); err == nil {
						options = append(options, map[string]interface{}{
							"id":          optID,
							"option_text": optText,
						})
					}
				}
				optRows.Close()
			}

			questions = append(questions, map[string]interface{}{
				"id":            q.id,
				"question_text": q.text,
				"points":        q.points,
				"options":       options,
			})
		}

		resp := map[string]interface{}{
			"quiz_id":            quizID,
			"title":              title,
			"description":        desc,
			"time_limit_minutes": timeLimit,
			"passing_score":      passingScore,
			"total_questions":    len(questions),
			"already_attempted":  alreadyAttempted,
			"questions":          questions,
		}

		if alreadyAttempted {
			resp["my_attempt"] = map[string]interface{}{
				"score":     existingScore,
				"is_passed": existingPassed == 1,
			}
		}

		return c.JSON(http.StatusOK, resp)
	}
}

// StudentAnswerItem defines one answer submitted by student
type StudentAnswerItem struct {
	QuestionID int `json:"question_id"`
	OptionID   int `json:"option_id"`
}

// SubmitStudentQuiz handles student quiz submission and auto-scoring
func SubmitStudentQuiz(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentIDStr := studentIDFromCtx(c)
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil || studentID <= 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Sesi tidak valid"})
		}

		var payload struct {
			QuizID  int                 `json:"quiz_id"`
			Answers []StudentAnswerItem `json:"answers"`
		}
		if err := c.Bind(&payload); err != nil || payload.QuizID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data kuis tidak valid"})
		}

		var (
			passingScore float64
			classID      int
		)
		err = db.QueryRow(`
			SELECT q.passing_score, ta.class_id
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND q.is_active = 1
		`, payload.QuizID).Scan(&passingScore, &classID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Kuis tidak ditemukan"})
		}

		// Calculate total questions in quiz
		var totalQuestions int
		_ = db.QueryRow("SELECT COUNT(*) FROM lms_quiz_questions WHERE quiz_id = ?", payload.QuizID).Scan(&totalQuestions)
		if totalQuestions == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kuis belum memiliki soal"})
		}

		// Auto-grade answers
		totalCorrect := 0
		for _, ans := range payload.Answers {
			var isCorrect int
			err := db.QueryRow(`
				SELECT is_correct
				FROM lms_quiz_options
				WHERE id = ? AND question_id = ?
			`, ans.OptionID, ans.QuestionID).Scan(&isCorrect)
			if err == nil && isCorrect == 1 {
				totalCorrect++
			}
		}

		score := (float64(totalCorrect) / float64(totalQuestions)) * 100.0
		isPassed := 0
		if score >= passingScore {
			isPassed = 1
		}

		// Save/replace attempt
		_, err = db.Exec(`
			INSERT INTO lms_quiz_attempts (quiz_id, student_id, started_at, completed_at, score, total_correct, is_passed)
			VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
			ON CONFLICT(quiz_id, student_id) DO UPDATE SET
				completed_at = CURRENT_TIMESTAMP,
				score = excluded.score,
				total_correct = excluded.total_correct,
				is_passed = excluded.is_passed
		`, payload.QuizID, studentID, score, totalCorrect, isPassed)

		if err != nil {
			// Fallback
			_, _ = db.Exec(`
				INSERT OR REPLACE INTO lms_quiz_attempts (quiz_id, student_id, started_at, completed_at, score, total_correct, is_passed)
				VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
			`, payload.QuizID, studentID, score, totalCorrect, isPassed)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":       "Kuis berhasil dikerjakan!",
			"score":         score,
			"total_correct": totalCorrect,
			"total_soal":    totalQuestions,
			"is_passed":     isPassed == 1,
			"passing_score": passingScore,
		})
	}
}

// GetQuizQuestions returns quiz questions and options with is_correct visible for teachers
func GetQuizQuestions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		quizID, err := strconv.Atoi(c.Param("id"))
		if err != nil || quizID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID kuis tidak valid"})
		}

		// Verify teacher ownership
		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND ta.staff_id = ?
		`, quizID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		rows, err := db.Query(`
			SELECT id, question_text, points
			FROM lms_quiz_questions
			WHERE quiz_id = ?
			ORDER BY id ASC
		`, quizID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca soal"})
		}

		type rawQ struct {
			id, points int
			text string
		}
		var rawQuestions []rawQ
		for rows.Next() {
			var q rawQ
			if err := rows.Scan(&q.id, &q.text, &q.points); err == nil {
				rawQuestions = append(rawQuestions, q)
			}
		}
		rows.Close()

		questions := make([]map[string]interface{}, 0)
		for _, q := range rawQuestions {
			optRows, _ := db.Query(`
				SELECT id, option_text, is_correct
				FROM lms_quiz_options
				WHERE question_id = ?
				ORDER BY id ASC
			`, q.id)
			options := make([]map[string]interface{}, 0)
			if optRows != nil {
				for optRows.Next() {
					var (
						optID, isCorrect int
						optText string
					)
					if err := optRows.Scan(&optID, &optText, &isCorrect); err == nil {
						options = append(options, map[string]interface{}{
							"id":          optID,
							"option_text": optText,
							"is_correct":  isCorrect == 1,
						})
					}
				}
				optRows.Close()
			}

			questions = append(questions, map[string]interface{}{
				"id":            q.id,
				"question_text": q.text,
				"points":        q.points,
				"options":       options,
			})
		}

		return c.JSON(http.StatusOK, questions)
	}
}

// DeleteMeetingQuiz deletes a quiz and its questions
func DeleteMeetingQuiz(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		quizID, err := strconv.Atoi(c.Param("id"))
		if err != nil || quizID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
		}

		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_quizzes q
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE q.id = ? AND ta.staff_id = ?
		`, quizID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		_, err = db.Exec("DELETE FROM lms_quizzes WHERE id = ?", quizID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus kuis"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Kuis berhasil dihapus"})
	}
}

// DeleteQuizQuestion deletes a single question from a quiz
func DeleteQuizQuestion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		questionID, err := strconv.Atoi(c.Param("id"))
		if err != nil || questionID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID soal tidak valid"})
		}

		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_quiz_questions qq
			JOIN lms_quizzes q ON q.id = qq.quiz_id
			JOIN lms_meetings m ON m.id = q.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE qq.id = ? AND ta.staff_id = ?
		`, questionID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		_, err = db.Exec("DELETE FROM lms_quiz_questions WHERE id = ?", questionID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus soal"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Soal berhasil dihapus"})
	}
}

