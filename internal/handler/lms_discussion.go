package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func discussionMeetingTeacher(db *sql.DB, meetingID, teacherID int) bool {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM lms_meetings m JOIN teacher_assignments a ON a.id=m.assignment_id WHERE m.id=? AND a.staff_id=?`, meetingID, teacherID).Scan(&n)
	return n > 0
}
func discussionMeetingStudent(db *sql.DB, meetingID, studentID int) bool {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM lms_meetings m JOIN teacher_assignments a ON a.id=m.assignment_id JOIN students s ON s.class_id=a.class_id WHERE m.id=? AND s.id=? AND m.is_published=1`, meetingID, studentID).Scan(&n)
	return n > 0
}
func discussionActor(c echo.Context) (string, int) {
	if id := TeacherIDFromCtx(c); id > 0 {
		return "teacher", id
	}
	sid, _ := strconv.Atoi(studentIDFromCtx(c))
	return "student", sid
}

func ListLMSDiscussions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		mid, _ := strconv.Atoi(c.Param("meeting_id"))
		typ, aid := discussionActor(c)
		if (typ == "teacher" && !discussionMeetingTeacher(db, mid, aid)) || (typ == "student" && !discussionMeetingStudent(db, mid, aid)) {
			return c.JSON(403, map[string]string{"message": "Akses ditolak"})
		}
		rows, err := db.Query(`SELECT id,author_type,author_id,body,status,created_at FROM lms_discussions WHERE meeting_id=? ORDER BY created_at DESC`, mid)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil diskusi"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, authorID int
			var at, body, status, created string
			_ = rows.Scan(&id, &at, &authorID, &body, &status, &created)
			comments := []map[string]interface{}{}
			cr, _ := db.Query(`SELECT id,parent_id,author_type,author_id,body,created_at FROM lms_discussion_comments WHERE discussion_id=? ORDER BY created_at ASC`, id)
			if cr != nil {
				for cr.Next() {
					var cid int
					var pid sql.NullInt64
					var ct, ca, cb, cc string
					_ = cr.Scan(&cid, &pid, &ct, &ca, &cb, &cc)
					comments = append(comments, map[string]interface{}{"id": cid, "parent_id": pid.Int64, "author_type": ct, "author_id": ca, "body": cb, "created_at": cc})
				}
				cr.Close()
			}
			out = append(out, map[string]interface{}{"id": id, "author_type": at, "author_id": authorID, "body": body, "status": status, "created_at": created, "comments": comments})
		}
		return c.JSON(http.StatusOK, out)
	}
}

func CreateLMSDiscussion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		mid, _ := strconv.Atoi(c.FormValue("meeting_id"))
		body := strings.TrimSpace(c.FormValue("body"))
		typ, aid := discussionActor(c)
		if body == "" || mid <= 0 {
			return c.JSON(400, map[string]string{"message": "Isi diskusi wajib diisi"})
		}
		if (typ == "teacher" && !discussionMeetingTeacher(db, mid, aid)) || (typ == "student" && !discussionMeetingStudent(db, mid, aid)) {
			return c.JSON(403, map[string]string{"message": "Akses ditolak"})
		}
		r, e := db.Exec(`INSERT INTO lms_discussions(meeting_id,author_type,author_id,body) VALUES(?,?,?,?)`, mid, typ, aid, body)
		if e != nil {
			return c.JSON(500, map[string]string{"message": "Gagal membuat diskusi"})
		}
		id, _ := r.LastInsertId()
		return c.JSON(201, map[string]interface{}{"id": id})
	}
}

func CreateLMSDiscussionComment(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		did, _ := strconv.Atoi(c.Param("id"))
		body := strings.TrimSpace(c.FormValue("body"))
		typ, aid := discussionActor(c)
		var mid int
		var status string
		if db.QueryRow(`SELECT meeting_id,status FROM lms_discussions WHERE id=?`, did).Scan(&mid, &status) != nil {
			return c.JSON(404, map[string]string{"message": "Diskusi tidak ditemukan"})
		}
		mi := mid
		if status == "closed" || body == "" || (typ == "teacher" && !discussionMeetingTeacher(db, mi, aid)) || (typ == "student" && !discussionMeetingStudent(db, mi, aid)) {
			return c.JSON(403, map[string]string{"message": "Diskusi ditutup atau akses ditolak"})
		}
		_, e := db.Exec(`INSERT INTO lms_discussion_comments(discussion_id,author_type,author_id,body,parent_id) VALUES(?,?,?,?,NULL)`, did, typ, aid, body)
		if e != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menambah komentar"})
		}
		return c.JSON(201, map[string]string{"message": "Komentar ditambahkan"})
	}
}

func DeleteLMSDiscussionComment(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		typ, aid := discussionActor(c)
		var authorType string
		var authorID int
		if db.QueryRow(`SELECT author_type,author_id FROM lms_discussion_comments WHERE id=?`, id).Scan(&authorType, &authorID) != nil {
			return c.JSON(404, map[string]string{"message": "Komentar tidak ditemukan"})
		}
		if authorType != typ || authorID != aid {
			return c.JSON(403, map[string]string{"message": "Akses ditolak"})
		}
		if _, err := db.Exec(`DELETE FROM lms_discussion_comments WHERE id=?`, id); err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menghapus komentar"})
		}
		return c.JSON(200, map[string]string{"message": "Komentar dihapus"})
	}
}

func ToggleLMSDiscussion(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		tid := TeacherIDFromCtx(c)
		status := c.FormValue("status")
		if status != "closed" {
			status = "open"
		}
		r, e := db.Exec(`UPDATE lms_discussions SET status=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND EXISTS(SELECT 1 FROM lms_meetings m JOIN teacher_assignments a ON a.id=m.assignment_id WHERE m.id=lms_discussions.meeting_id AND a.staff_id=?)`, status, id, tid)
		if e != nil {
			return c.JSON(500, e)
		}
		n, _ := r.RowsAffected()
		if n == 0 {
			return c.JSON(404, map[string]string{"message": "Diskusi tidak ditemukan"})
		}
		return c.JSON(200, map[string]string{"status": status})
	}
}

func ToggleLMSReaction(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		typ, aid := discussionActor(c)
		did, _ := strconv.Atoi(c.FormValue("discussion_id"))
		cid, _ := strconv.Atoi(c.FormValue("comment_id"))
		rt := c.FormValue("reaction_type")
		if rt == "" {
			rt = "like"
		}
		_, e := db.Exec(`INSERT OR IGNORE INTO lms_discussion_reactions(discussion_id,comment_id,author_type,author_id,reaction_type) VALUES(?,?,?,?,?)`, did, func() interface{} {
			if cid > 0 {
				return cid
			}
			return nil
		}(), typ, aid, rt)
		if e != nil {
			return c.JSON(500, e)
		}
		return c.JSON(200, map[string]string{"reaction": rt})
	}
}
