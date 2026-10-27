package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func AdminListJournals(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := `SELECT j.id,j.assignment_id,j.journal_date,j.meeting_order,j.lesson_topic,j.status,
		COALESCE(j.rejection_reason,''),COALESCE(j.submitted_at,''),COALESCE(j.approved_by,''),COALESCE(j.approved_at,''),
		ta.staff_id,s.name,c.name,sub.name,j.attendance_summary
		FROM teacher_journals j
		JOIN teacher_assignments ta ON ta.id=j.assignment_id
		JOIN staff s ON s.id=ta.staff_id
		JOIN subjects sub ON sub.id=ta.subject_id
		JOIN classes c ON c.id=ta.class_id`
		args := []interface{}{}
		if st := c.QueryParam("status"); st != "" {
			q += ` WHERE j.status=?`
			args = append(args, st)
		}
		q += ` ORDER BY j.journal_date DESC, j.id DESC LIMIT 200`
		rows, err := db.Query(q, args...)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil jurnal"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, aid, staffID, order int
			var date, topic, status, rejection, submitted, approvedBy, approvedAt, staffName, className, subjectName, summary string
			if rows.Scan(&id, &aid, &date, &order, &topic, &status, &rejection, &submitted, &approvedBy, &approvedAt, &staffID, &staffName, &className, &subjectName, &summary) != nil {
				continue
			}
			out = append(out, map[string]interface{}{
				"id": id, "assignment_id": aid, "journal_date": date, "meeting_order": order,
				"lesson_topic": topic, "status": status, "rejection_reason": rejection,
				"submitted_at": submitted, "approved_by": approvedBy, "approved_at": approvedAt,
				"staff_id": staffID, "staff_name": staffName, "class_name": className,
				"subject_name": subjectName, "attendance_summary": summary,
			})
		}
		return c.JSON(http.StatusOK, out)
	}
}

func AdminApproveJournal(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		admin, _ := c.Get("user").(string)
		if admin == "" {
			admin = "admin"
		}
		var staffID int
		var topic string
		if db.QueryRow(`SELECT ta.staff_id, j.lesson_topic FROM teacher_journals j JOIN teacher_assignments ta ON ta.id=j.assignment_id WHERE j.id=?`, id).Scan(&staffID, &topic) != nil {
			return c.JSON(404, map[string]string{"message": "Jurnal tidak ditemukan"})
		}
		_, err := db.Exec(`UPDATE teacher_journals SET status='approved',approved_by=?,approved_at=CURRENT_TIMESTAMP,rejection_reason='',updated_at=CURRENT_TIMESTAMP WHERE id=?`, admin, id)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyetujui jurnal"})
		}
		createNotification(db, "teacher", staffID, "Jurnal disetujui", "Jurnal \""+topic+"\" telah disetujui admin.", "journal", id)
		return c.JSON(200, map[string]string{"message": "Jurnal disetujui"})
	}
}

func AdminRejectJournal(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		admin, _ := c.Get("user").(string)
		if admin == "" {
			admin = "admin"
		}
		reason := strings.TrimSpace(c.FormValue("reason"))
		if reason == "" {
			return c.JSON(400, map[string]string{"message": "Alasan penolakan wajib diisi"})
		}
		var staffID int
		var topic string
		if db.QueryRow(`SELECT ta.staff_id, j.lesson_topic FROM teacher_journals j JOIN teacher_assignments ta ON ta.id=j.assignment_id WHERE j.id=?`, id).Scan(&staffID, &topic) != nil {
			return c.JSON(404, map[string]string{"message": "Jurnal tidak ditemukan"})
		}
		_, err := db.Exec(`UPDATE teacher_journals SET status='rejected',rejection_reason=?,approved_by=?,approved_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=?`, reason, admin, id)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menolak jurnal"})
		}
		createNotification(db, "teacher", staffID, "Jurnal ditolak", "Jurnal \""+topic+"\" ditolak: "+reason, "journal", id)
		return c.JSON(200, map[string]string{"message": "Jurnal ditolak"})
	}
}
