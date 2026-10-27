package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// GetUnifiedCalendar merges academic calendar events with KBM meetings for the
// authenticated actor (teacher sees own meetings, student sees class meetings).
func GetUnifiedCalendar(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		typ, aid := discussionActor(c)
		start := c.QueryParam("start")
		end := c.QueryParam("end")
		if start == "" {
			start = time.Now().Format("2006-01") + "-01"
		}
		if end == "" {
			end = time.Now().AddDate(0, 1, 0).Format("2006-01") + "-01"
		}
		events := make([]map[string]interface{}, 0)

		rows, err := db.Query(`SELECT id,title,category,start_date,end_date,COALESCE(color,''),COALESCE(description,'') FROM academic_calendar_events WHERE is_published=1 AND end_date>=? AND start_date<=? ORDER BY start_date ASC`, start, end)
		if err == nil {
			for rows.Next() {
				var id int
				var title, category, sd, ed, color, desc string
				if rows.Scan(&id, &title, &category, &sd, &ed, &color, &desc) == nil {
					events = append(events, map[string]interface{}{
						"id": id, "title": title, "category": category, "event_type": "academic",
						"start_date": sd, "end_date": ed, "color": color, "description": desc,
					})
				}
			}
			rows.Close()
		}

		mq := `SELECT m.id,m.title,m.meeting_date,sub.name,c.name FROM lms_meetings m
			JOIN teacher_assignments ta ON ta.id=m.assignment_id
			JOIN subjects sub ON sub.id=ta.subject_id
			JOIN classes c ON c.id=ta.class_id
			WHERE m.is_published=1 AND m.meeting_date BETWEEN ? AND ?`
		margs := []interface{}{start, end}
		if typ == "teacher" {
			mq += ` AND ta.staff_id=?`
			margs = append(margs, aid)
		} else {
			mq += ` AND ta.class_id=(SELECT class_id FROM students WHERE id=?)`
			margs = append(margs, aid)
		}
		mq += ` ORDER BY m.meeting_date ASC`
		mrows, merr := db.Query(mq, margs...)
		if merr == nil {
			for mrows.Next() {
				var id int
				var title, mdate, subject, class string
				if mrows.Scan(&id, &title, &mdate, &subject, &class) == nil {
					events = append(events, map[string]interface{}{
						"id": id, "title": subject + " — " + class, "category": "kbm", "event_type": "meeting",
						"start_date": mdate, "end_date": mdate, "color": "#4C1D95", "description": title,
					})
				}
			}
			mrows.Close()
		}
		return c.JSON(http.StatusOK, events)
	}
}

// GetStudentAgenda is the student-facing unified calendar (academic + class KBM).
func GetStudentAgenda(db *sql.DB) echo.HandlerFunc {
	return GetUnifiedCalendar(db)
}
