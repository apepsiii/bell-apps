package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type BankQuestionRequest struct {
	SubjectID    int    `json:"subject_id"`
	QuestionText string `json:"question_text"`
	Points       int    `json:"points"`
	Options      []struct {
		Text      string `json:"text"`
		IsCorrect bool   `json:"is_correct"`
	} `json:"options"`
}

func GetQuestionBank(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		q := `SELECT id,subject_id,question_text,points,created_at FROM lms_question_bank WHERE staff_id=?`
		args := []interface{}{tid}
		if s := c.QueryParam("subject_id"); s != "" {
			q += ` AND subject_id=?`
			args = append(args, s)
		}
		q += ` ORDER BY created_at DESC`
		rows, err := db.Query(q, args...)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil bank soal"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, points int
			var subjectID sql.NullInt64
			var text, created string
			if rows.Scan(&id, &subjectID, &text, &points, &created) != nil {
				continue
			}
			options := []map[string]interface{}{}
			or, _ := db.Query(`SELECT id,option_text,is_correct FROM lms_question_bank_options WHERE bank_question_id=?`, id)
			if or != nil {
				for or.Next() {
					var oid, correct int
					var otext string
					if or.Scan(&oid, &otext, &correct) == nil {
						options = append(options, map[string]interface{}{"id": oid, "text": otext, "is_correct": correct == 1})
					}
				}
				or.Close()
			}
			out = append(out, map[string]interface{}{"id": id, "subject_id": subjectID.Int64, "question_text": text, "points": points, "created_at": created, "options": options})
		}
		return c.JSON(http.StatusOK, out)
	}
}

func CreateBankQuestion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		var req BankQuestionRequest
		if err := c.Bind(&req); err != nil || req.QuestionText == "" || len(req.Options) < 2 {
			return c.JSON(400, map[string]string{"message": "Soal minimal memiliki 2 opsi jawaban"})
		}
		points := req.Points
		if points <= 0 {
			points = 10
		}
		var subjectID interface{}
		if req.SubjectID > 0 {
			subjectID = req.SubjectID
		}
		tx, err := db.Begin()
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyimpan soal"})
		}
		r, err := tx.Exec(`INSERT INTO lms_question_bank(staff_id,subject_id,question_text,points) VALUES(?,?,?,?)`, tid, subjectID, req.QuestionText, points)
		if err != nil {
			tx.Rollback()
			return c.JSON(500, map[string]string{"message": "Gagal menyimpan soal"})
		}
		qid, _ := r.LastInsertId()
		for _, o := range req.Options {
			if _, err := tx.Exec(`INSERT INTO lms_question_bank_options(bank_question_id,option_text,is_correct) VALUES(?,?,?)`, qid, o.Text, o.IsCorrect); err != nil {
				tx.Rollback()
				return c.JSON(500, map[string]string{"message": "Gagal menyimpan opsi"})
			}
		}
		if err := tx.Commit(); err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyimpan soal"})
		}
		return c.JSON(201, map[string]interface{}{"id": qid})
	}
}

func DeleteBankQuestion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		id, _ := strconv.Atoi(c.Param("id"))
		r, err := db.Exec(`DELETE FROM lms_question_bank WHERE id=? AND staff_id=?`, id, tid)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menghapus soal"})
		}
		n, _ := r.RowsAffected()
		if n == 0 {
			return c.JSON(404, map[string]string{"message": "Soal tidak ditemukan"})
		}
		return c.JSON(200, map[string]string{"message": "Soal dihapus"})
	}
}

func AddBankQuestionToQuiz(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		quizID, _ := strconv.Atoi(c.FormValue("quiz_id"))
		bankID, _ := strconv.Atoi(c.FormValue("bank_question_id"))
		var meetingID, assignmentStaff int
		if db.QueryRow(`SELECT m.id, ta.staff_id FROM lms_quizzes q JOIN lms_meetings m ON m.id=q.meeting_id JOIN teacher_assignments ta ON ta.id=m.assignment_id WHERE q.id=?`, quizID).Scan(&meetingID, &assignmentStaff) != nil || assignmentStaff != tid {
			return c.JSON(403, map[string]string{"message": "Akses ditolak"})
		}
		var text, image string
		var points int
		if db.QueryRow(`SELECT question_text,question_image,points FROM lms_question_bank WHERE id=? AND staff_id=?`, bankID, tid).Scan(&text, &image, &points) != nil {
			return c.JSON(404, map[string]string{"message": "Soal bank tidak ditemukan"})
		}
		tx, err := db.Begin()
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyalin soal"})
		}
		r, err := tx.Exec(`INSERT INTO lms_quiz_questions(quiz_id,question_text,question_image,points) VALUES(?,?,?,?)`, quizID, text, image, points)
		if err != nil {
			tx.Rollback()
			return c.JSON(500, map[string]string{"message": "Gagal menyalin soal"})
		}
		qid, _ := r.LastInsertId()
		rows, err := tx.Query(`SELECT option_text,is_correct FROM lms_question_bank_options WHERE bank_question_id=?`, bankID)
		if err != nil {
			tx.Rollback()
			return c.JSON(500, map[string]string{"message": "Gagal menyalin opsi"})
		}
		for rows.Next() {
			var otext string
			var correct bool
			if rows.Scan(&otext, &correct) == nil {
				if _, err := tx.Exec(`INSERT INTO lms_quiz_options(question_id,option_text,is_correct) VALUES(?,?,?)`, qid, otext, correct); err != nil {
					rows.Close()
					tx.Rollback()
					return c.JSON(500, map[string]string{"message": "Gagal menyalin opsi"})
				}
			}
		}
		rows.Close()
		if err := tx.Commit(); err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyalin soal"})
		}
		return c.JSON(201, map[string]interface{}{"id": qid})
	}
}
