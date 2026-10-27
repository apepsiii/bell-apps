package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func teacherAssignmentOwned(db *sql.DB, assignmentID, teacherID int) bool {
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM teacher_assignments WHERE id = ? AND staff_id = ?", assignmentID, teacherID).Scan(&count)
	return count > 0
}

func GetTeacherJournals(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		rows, err := db.Query(`SELECT j.id,j.assignment_id,j.meeting_id,j.journal_date,j.meeting_order,j.lesson_topic,
		COALESCE(j.learning_objectives,''),COALESCE(j.teaching_method,''),COALESCE(j.learning_media,''),COALESCE(j.activities,''),
		COALESCE(j.assessment,''),COALESCE(j.student_issues,''),COALESCE(j.follow_up,''),COALESCE(j.attendance_summary,'{}'),j.status,
		COALESCE(j.rejection_reason,''),COALESCE(j.submitted_at,''),COALESCE(j.approved_by,''),COALESCE(j.approved_at,''),j.created_at,j.updated_at,
		ta.subject_id,s.name,c.name,ta.academic_year,ta.semester
		FROM teacher_journals j JOIN teacher_assignments ta ON ta.id=j.assignment_id JOIN subjects s ON s.id=ta.subject_id JOIN classes c ON c.id=ta.class_id
		WHERE ta.staff_id=? AND (?='' OR j.journal_date>=?) AND (?='' OR j.journal_date<=?) ORDER BY j.journal_date DESC,j.meeting_order DESC`, teacherID, c.QueryParam("start"), c.QueryParam("start"), c.QueryParam("end"), c.QueryParam("end"))
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil jurnal"})
		}
		defer rows.Close()
		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, aid, order, sid, semester int
			var meetingID sql.NullInt64
			var date, topic, objectives, method, media, activities, assessment, issues, followup, summary, status, rejection, submitted, approvedBy, approved, created, updated, subject, className, year string
			if err := rows.Scan(&id, &aid, &meetingID, &date, &order, &topic, &objectives, &method, &media, &activities, &assessment, &issues, &followup, &summary, &status, &rejection, &submitted, &approvedBy, &approved, &created, &updated, &sid, &subject, &className, &year, &semester); err != nil {
				continue
			}
			var att map[string]int
			_ = json.Unmarshal([]byte(summary), &att)
			if att == nil {
				att = map[string]int{}
			}
			result = append(result, map[string]interface{}{"id": id, "assignment_id": aid, "meeting_id": meetingID.Int64, "journal_date": date, "meeting_order": order, "lesson_topic": topic, "learning_objectives": objectives, "teaching_method": method, "learning_media": media, "activities": activities, "assessment": assessment, "student_issues": issues, "follow_up": followup, "attendance_summary": att, "status": status, "rejection_reason": rejection, "submitted_at": submitted, "approved_by": approvedBy, "approved_at": approved, "created_at": created, "updated_at": updated, "assignment": map[string]interface{}{"subject_id": sid, "subject_name": subject, "class_name": className, "academic_year": year, "semester": semester}})
		}
		return c.JSON(http.StatusOK, result)
	}
}

func CreateTeacherJournal(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		aid, _ := strconv.Atoi(c.FormValue("assignment_id"))
		if aid <= 0 || !teacherAssignmentOwned(db, aid, teacherID) {
			return c.JSON(403, map[string]string{"message": "Akses ditolak"})
		}
		date := strings.TrimSpace(c.FormValue("journal_date"))
		topic := strings.TrimSpace(c.FormValue("lesson_topic"))
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		if _, err := time.Parse("2006-01-02", date); err != nil || topic == "" {
			return c.JSON(400, map[string]string{"message": "Tanggal dan topik wajib valid"})
		}
		order, _ := strconv.Atoi(c.FormValue("meeting_order"))
		meetingID, _ := strconv.Atoi(c.FormValue("meeting_id"))
		if meetingID > 0 {
			var meetingAssignment int
			if err := db.QueryRow(`SELECT assignment_id FROM lms_meetings WHERE id=?`, meetingID).Scan(&meetingAssignment); err != nil || meetingAssignment != aid {
				return c.JSON(400, map[string]string{"message": "Pertemuan tidak sesuai dengan penugasan"})
			}
		} else {
			meetingID = 0
		}
		summary := c.FormValue("attendance_summary")
		if summary == "" && meetingID > 0 {
			rows, err := db.Query(`SELECT status, COUNT(*) FROM lms_attendances WHERE meeting_id=? GROUP BY status`, meetingID)
			if err == nil {
				counts := map[string]int{}
				defer rows.Close()
				for rows.Next() {
					var status string
					var count int
					if rows.Scan(&status, &count) == nil {
						counts[status] = count
					}
				}
				encoded, _ := json.Marshal(counts)
				summary = string(encoded)
			}
		}
		if summary == "" {
			summary = "{}"
		}
		res, err := db.Exec(`INSERT INTO teacher_journals (assignment_id,meeting_id,journal_date,meeting_order,lesson_topic,learning_objectives,teaching_method,learning_media,activities,assessment,student_issues,follow_up,attendance_summary) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, aid, nullableInt(meetingID), date, order, topic, c.FormValue("learning_objectives"), c.FormValue("teaching_method"), c.FormValue("learning_media"), c.FormValue("activities"), c.FormValue("assessment"), c.FormValue("student_issues"), c.FormValue("follow_up"), summary)
		if err != nil {
			return c.JSON(409, map[string]string{"message": "Jurnal pada tanggal dan pertemuan tersebut sudah ada"})
		}
		id, _ := res.LastInsertId()
		return c.JSON(201, map[string]interface{}{"id": id, "message": "Jurnal berhasil disimpan"})
	}
}

func UpdateTeacherJournal(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		teacherID := TeacherIDFromCtx(c)
		var aid int
		var status string
		if db.QueryRow(`SELECT j.assignment_id,j.status FROM teacher_journals j JOIN teacher_assignments ta ON ta.id=j.assignment_id WHERE j.id=? AND ta.staff_id=?`, id, teacherID).Scan(&aid, &status) != nil {
			return c.JSON(404, map[string]string{"message": "Jurnal tidak ditemukan"})
		}
		if status == "approved" || status == "submitted" {
			return c.JSON(409, map[string]string{"message": "Jurnal sudah dikirim dan tidak dapat diedit"})
		}
		_, err := db.Exec(`UPDATE teacher_journals SET lesson_topic=?,learning_objectives=?,teaching_method=?,learning_media=?,activities=?,assessment=?,student_issues=?,follow_up=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, c.FormValue("lesson_topic"), c.FormValue("learning_objectives"), c.FormValue("teaching_method"), c.FormValue("learning_media"), c.FormValue("activities"), c.FormValue("assessment"), c.FormValue("student_issues"), c.FormValue("follow_up"), id)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal memperbarui jurnal"})
		}
		return c.JSON(200, map[string]string{"message": "Jurnal diperbarui"})
	}
}

func SubmitTeacherJournal(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		teacherID := TeacherIDFromCtx(c)
		res, err := db.Exec(`UPDATE teacher_journals SET status='submitted',submitted_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=? AND status IN ('draft','rejected') AND assignment_id IN (SELECT id FROM teacher_assignments WHERE staff_id=?)`, id, teacherID)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengirim jurnal"})
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return c.JSON(409, map[string]string{"message": "Jurnal tidak ditemukan atau tidak dapat dikirim"})
		}
		return c.JSON(200, map[string]string{"message": "Jurnal dikirim untuk persetujuan"})
	}
}

func nullableInt(v int) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

var _ = fmt.Sprintf
