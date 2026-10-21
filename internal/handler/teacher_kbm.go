package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// --- LMS MEETINGS (SESI PERTEMUAN KBM) ---

// GetTeacherMeetings returns meetings for a specific teacher assignment
func GetTeacherMeetings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		assignmentID, err := strconv.Atoi(c.Param("assignment_id"))
		if err != nil || assignmentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID penugasan tidak valid"})
		}

		teacherID := TeacherIDFromCtx(c)

		// Verify teacher has access to this assignment
		var count int
		_ = db.QueryRow(
			"SELECT COUNT(*) FROM teacher_assignments WHERE id = ? AND staff_id = ?",
			assignmentID, teacherID,
		).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Anda tidak memiliki akses ke kelas ini"})
		}

		rows, err := db.Query(`
			SELECT id, assignment_id, meeting_order, title, COALESCE(description, ''),
			       COALESCE(strftime('%Y-%m-%d', meeting_date), meeting_date, ''), is_published, COALESCE(created_at, '')
			FROM lms_meetings
			WHERE assignment_id = ?
			ORDER BY meeting_order ASC, meeting_date ASC
		`, assignmentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data pertemuan"})
		}
		defer rows.Close()

		type rawMeeting struct {
			id, aID, order, isPub int
			title, desc, mDate, createdAt string
		}
		var rawMeetings []rawMeeting
		for rows.Next() {
			var r rawMeeting
			if err := rows.Scan(&r.id, &r.aID, &r.order, &r.title, &r.desc, &r.mDate, &r.isPub, &r.createdAt); err == nil {
				rawMeetings = append(rawMeetings, r)
			}
		}
		rows.Close()

		meetings := make([]map[string]interface{}, 0)
		for _, m := range rawMeetings {
			// Fetch materials for this meeting
			matRows, _ := db.Query(`SELECT id, title, type, COALESCE(content_url, ''), COALESCE(file_key, ''), COALESCE(file_size, 0), COALESCE(text_content, '') FROM lms_materials WHERE meeting_id = ? ORDER BY id ASC`, m.id)
			materials := make([]map[string]interface{}, 0)
			if matRows != nil {
				for matRows.Next() {
					var (
						mId int
						mTitle, mType, cUrl, fKey, tContent string
						fSize int64
					)
					if err := matRows.Scan(&mId, &mTitle, &mType, &cUrl, &fKey, &fSize, &tContent); err == nil {
						materials = append(materials, map[string]interface{}{
							"id":           mId,
							"title":        mTitle,
							"type":         mType,
							"content_url":  cUrl,
							"file_size":    fSize,
							"text_content": tContent,
						})
					}
				}
				matRows.Close()
			}

			// Fetch assignments for this meeting
			assRows, _ := db.Query(`SELECT id, title, COALESCE(instructions, ''), COALESCE(allowed_type, 'any'), COALESCE(strftime('%Y-%m-%d %H:%M', due_date), ''), COALESCE(max_points, 100) FROM lms_assignments WHERE meeting_id = ? ORDER BY id ASC`, m.id)
			assignments := make([]map[string]interface{}, 0)
			if assRows != nil {
				for assRows.Next() {
					var (
						aId, maxPts int
						aTitle, aInst, aType, dDate string
					)
					if err := assRows.Scan(&aId, &aTitle, &aInst, &aType, &dDate, &maxPts); err == nil {
						var subCount int
						_ = db.QueryRow("SELECT COUNT(*) FROM lms_submissions WHERE assignment_id = ?", aId).Scan(&subCount)
						assignments = append(assignments, map[string]interface{}{
							"id":           aId,
							"title":        aTitle,
							"instructions": aInst,
							"allowed_type": aType,
							"due_date":     dDate,
							"max_points":   maxPts,
							"submissions":  subCount,
						})
					}
				}
				assRows.Close()
			}

			// Fetch quizzes for this meeting
			quizRows, _ := db.Query(`SELECT id, title, COALESCE(description, ''), COALESCE(time_limit_minutes, 30), COALESCE(passing_score, 75.0) FROM lms_quizzes WHERE meeting_id = ? AND is_active = 1 ORDER BY id ASC`, m.id)
			quizzes := make([]map[string]interface{}, 0)
			if quizRows != nil {
				for quizRows.Next() {
					var (
						qId, timeLimit int
						qTitle, qDesc string
						passingScore float64
					)
					if err := quizRows.Scan(&qId, &qTitle, &qDesc, &timeLimit, &passingScore); err == nil {
						var qCount, attemptCount int
						_ = db.QueryRow("SELECT COUNT(*) FROM lms_quiz_questions WHERE quiz_id = ?", qId).Scan(&qCount)
						_ = db.QueryRow("SELECT COUNT(*) FROM lms_quiz_attempts WHERE quiz_id = ?", qId).Scan(&attemptCount)
						quizzes = append(quizzes, map[string]interface{}{
							"id":                 qId,
							"title":              qTitle,
							"description":        qDesc,
							"time_limit_minutes": timeLimit,
							"passing_score":      passingScore,
							"total_questions":    qCount,
							"attempts":           attemptCount,
						})
					}
				}
				quizRows.Close()
			}

			var attCount int
			_ = db.QueryRow("SELECT COUNT(*) FROM lms_attendances WHERE meeting_id = ?", m.id).Scan(&attCount)

			meetings = append(meetings, map[string]interface{}{
				"id":            m.id,
				"assignment_id": m.aID,
				"meeting_order": m.order,
				"title":         m.title,
				"description":   m.desc,
				"meeting_date":  m.mDate,
				"is_published":  m.isPub == 1,
				"created_at":    m.createdAt,
				"materials":     materials,
				"assignments":   assignments,
				"quizzes":       quizzes,
				"stats": map[string]int{
					"materials":   len(materials),
					"assignments": len(assignments),
					"quizzes":     len(quizzes),
					"attendances": attCount,
				},
			})
		}

		return c.JSON(http.StatusOK, meetings)
	}
}

// CreateTeacherMeeting creates a new meeting/session
func CreateTeacherMeeting(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		assignmentID, err := strconv.Atoi(c.FormValue("assignment_id"))
		if err != nil || assignmentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID penugasan tidak valid"})
		}

		var count int
		_ = db.QueryRow(
			"SELECT COUNT(*) FROM teacher_assignments WHERE id = ? AND staff_id = ?",
			assignmentID, teacherID,
		).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		title := strings.TrimSpace(c.FormValue("title"))
		description := strings.TrimSpace(c.FormValue("description"))
		meetingDate := strings.TrimSpace(c.FormValue("meeting_date"))
		order, _ := strconv.Atoi(c.FormValue("meeting_order"))

		if title == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Judul pertemuan wajib diisi"})
		}
		if meetingDate == "" {
			meetingDate = time.Now().Format("2006-01-02")
		}

		if order <= 0 {
			var maxOrder sql.NullInt64
			_ = db.QueryRow("SELECT MAX(meeting_order) FROM lms_meetings WHERE assignment_id = ?", assignmentID).Scan(&maxOrder)
			if maxOrder.Valid {
				order = int(maxOrder.Int64) + 1
			} else {
				order = 1
			}
		}

		res, err := db.Exec(`
			INSERT INTO lms_meetings (assignment_id, meeting_order, title, description, meeting_date, is_published)
			VALUES (?, ?, ?, ?, ?, 1)
		`, assignmentID, order, title, description, meetingDate)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat sesi pertemuan"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Pertemuan KBM berhasil dibuat",
			"id":      id,
		})
	}
}

// --- LMS ATTENDANCES (PRESENSI KBM PJJ OLEH GURU) ---

// GetMeetingAttendance returns attendance records for all students in the assigned class
func GetMeetingAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		meetingID, err := strconv.Atoi(c.Param("meeting_id"))
		if err != nil || meetingID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pertemuan tidak valid"})
		}

		teacherID := TeacherIDFromCtx(c)

		// Get class_id and verify ownership
		var classID, assignmentID int
		var title string
		err = db.QueryRow(`
			SELECT m.assignment_id, ta.class_id, m.title
			FROM lms_meetings m
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE m.id = ? AND ta.staff_id = ?
		`, meetingID, teacherID).Scan(&assignmentID, &classID, &title)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Pertemuan tidak ditemukan atau akses ditolak"})
		}

		// Fetch all students in the class with their attendance status
		rows, err := db.Query(`
			SELECT s.id, s.name, COALESCE(s.nis, s.nis_siswa, ''),
			       COALESCE(a.status, 'H') as status,
			       COALESCE(a.notes, '') as notes,
			       COALESCE(a.updated_at, '') as updated_at
			FROM students s
			LEFT JOIN lms_attendances a ON a.student_id = s.id AND a.meeting_id = ?
			WHERE s.class_id = ? AND s.status = 'active'
			ORDER BY s.name ASC
		`, meetingID, classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca daftar siswa"})
		}
		defer rows.Close()

		students := make([]map[string]interface{}, 0)
		stats := map[string]int{"H": 0, "S": 0, "I": 0, "A": 0}

		for rows.Next() {
			var (
				sID int
				sName, sNIS, status, notes, updatedAt string
			)
			if err := rows.Scan(&sID, &sName, &sNIS, &status, &notes, &updatedAt); err == nil {
				if _, ok := stats[status]; ok {
					stats[status]++
				}
				students = append(students, map[string]interface{}{
					"student_id": sID,
					"name":       sName,
					"nis":        sNIS,
					"status":     status,
					"notes":      notes,
					"updated_at": updatedAt,
				})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"meeting_id": meetingID,
			"title":      title,
			"students":   students,
			"stats":      stats,
		})
	}
}

// AttendanceItem represents single student record in batch update
type AttendanceItem struct {
	StudentID int    `json:"student_id"`
	Status    string `json:"status"` // 'H', 'S', 'I', 'A'
	Notes     string `json:"notes"`
}

// SaveMeetingAttendance updates attendance for multiple students at once
func SaveMeetingAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		meetingID, err := strconv.Atoi(c.Param("meeting_id"))
		if err != nil || meetingID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pertemuan tidak valid"})
		}

		teacherID := TeacherIDFromCtx(c)

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

		var payload struct {
			Attendances []AttendanceItem `json:"attendances"`
		}

		if err := c.Bind(&payload); err != nil || len(payload.Attendances) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data presensi tidak valid"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memulai transaksi"})
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(`
			INSERT INTO lms_attendances (meeting_id, student_id, status, notes, marked_by_staff_id, updated_at)
			VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(meeting_id, student_id) DO UPDATE SET
				status = excluded.status,
				notes = excluded.notes,
				marked_by_staff_id = excluded.marked_by_staff_id,
				updated_at = CURRENT_TIMESTAMP
		`)
		if err != nil {
			// Fallback if SQLite version doesn't support ON CONFLICT: use REPLACE INTO
			stmt, err = tx.Prepare(`
				INSERT OR REPLACE INTO lms_attendances (meeting_id, student_id, status, notes, marked_by_staff_id, updated_at)
				VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			`)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyiapkan query presensi"})
			}
		}
		defer stmt.Close()

		for _, item := range payload.Attendances {
			st := strings.ToUpper(strings.TrimSpace(item.Status))
			if st != "H" && st != "S" && st != "I" && st != "A" {
				st = "H"
			}
			if _, err := stmt.Exec(meetingID, item.StudentID, st, item.Notes, teacherID); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan presensi siswa"})
			}
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan perubahan presensi"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Presensi KBM berhasil disimpan"})
	}
}

