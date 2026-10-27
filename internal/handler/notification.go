package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func createNotification(db *sql.DB, recipientType string, recipientID int, title, message, notifType string, refID int) {
	if recipientID <= 0 || title == "" {
		return
	}
	_, _ = db.Exec(`INSERT INTO notifications (recipient_type, recipient_id, title, message, notif_type, ref_id) VALUES (?,?,?,?,?,?)`, recipientType, recipientID, title, message, notifType, refID)
}

func notifyClassStudents(db *sql.DB, classID int, title, message, notifType string, refID int) {
	rows, err := db.Query(`SELECT id FROM students WHERE class_id=? AND status='active'`, classID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sid int
		if rows.Scan(&sid) == nil {
			createNotification(db, "student", sid, title, message, notifType, refID)
		}
	}
}

func notifyAssignmentTeachers(db *sql.DB, assignmentID int, title, message, notifType string, refID int) {
	rows, err := db.Query(`SELECT staff_id FROM teacher_assignments WHERE id=?`, assignmentID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tid int
		if rows.Scan(&tid) == nil {
			createNotification(db, "teacher", tid, title, message, notifType, refID)
		}
	}
}

func GetStudentNotifications(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		sid, _ := strconv.Atoi(studentIDFromCtx(c))
		rows, err := db.Query(`SELECT id,title,message,notif_type,ref_id,is_read,created_at FROM notifications WHERE recipient_type='student' AND recipient_id=? ORDER BY created_at DESC LIMIT 50`, sid)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil notifikasi"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, refID, read int
			var title, message, ntype, created string
			if rows.Scan(&id, &title, &message, &ntype, &refID, &read, &created) == nil {
				out = append(out, map[string]interface{}{"id": id, "title": title, "message": message, "notif_type": ntype, "ref_id": refID, "is_read": read == 1, "created_at": created})
			}
		}
		return c.JSON(http.StatusOK, out)
	}
}

func MarkStudentNotificationRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		sid, _ := strconv.Atoi(studentIDFromCtx(c))
		id, _ := strconv.Atoi(c.Param("id"))
		db.Exec(`UPDATE notifications SET is_read=1 WHERE id=? AND recipient_type='student' AND recipient_id=?`, id, sid)
		return c.JSON(200, map[string]string{"message": "ok"})
	}
}

func MarkAllStudentNotificationsRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		sid, _ := strconv.Atoi(studentIDFromCtx(c))
		db.Exec(`UPDATE notifications SET is_read=1 WHERE recipient_type='student' AND recipient_id=?`, sid)
		return c.JSON(200, map[string]string{"message": "ok"})
	}
}

func GetTeacherNotifications(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		rows, err := db.Query(`SELECT id,title,message,notif_type,ref_id,is_read,created_at FROM notifications WHERE recipient_type='teacher' AND recipient_id=? ORDER BY created_at DESC LIMIT 50`, tid)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil notifikasi"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, refID, read int
			var title, message, ntype, created string
			if rows.Scan(&id, &title, &message, &ntype, &refID, &read, &created) == nil {
				out = append(out, map[string]interface{}{"id": id, "title": title, "message": message, "notif_type": ntype, "ref_id": refID, "is_read": read == 1, "created_at": created})
			}
		}
		return c.JSON(http.StatusOK, out)
	}
}

func MarkAllTeacherNotificationsRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		db.Exec(`UPDATE notifications SET is_read=1 WHERE recipient_type='teacher' AND recipient_id=?`, tid)
		return c.JSON(200, map[string]string{"message": "ok"})
	}
}

func MarkTeacherNotificationRead(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		tid := TeacherIDFromCtx(c)
		id, _ := strconv.Atoi(c.Param("id"))
		db.Exec(`UPDATE notifications SET is_read=1 WHERE id=? AND recipient_type='teacher' AND recipient_id=?`, id, tid)
		return c.JSON(200, map[string]string{"message": "ok"})
	}
}
