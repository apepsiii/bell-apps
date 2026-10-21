package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/service"
)

// --- TEACHER MATERIALS ---

// AddMeetingMaterial handles adding a material (file, link, video, or text) to a meeting
func AddMeetingMaterial(db *sql.DB, storage *service.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		meetingID, err := strconv.Atoi(c.FormValue("meeting_id"))
		if err != nil || meetingID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pertemuan tidak valid"})
		}

		// Verify teacher owns the meeting
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
		mType := strings.TrimSpace(c.FormValue("type")) // 'document', 'video', 'link', 'text'
		if title == "" || mType == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Judul dan jenis materi wajib diisi"})
		}

		var (
			contentURL  string
			fileKey     string
			fileSize    int64
			textContent string
		)

		if mType == "document" {
			fileHeader, err := c.FormFile("file")
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Berkas dokumen wajib diunggah"})
			}

			src, err := fileHeader.Open()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca berkas"})
			}
			defer src.Close()

			fileSize = fileHeader.Size
			ext := filepath.Ext(fileHeader.Filename)
			cleanName := strings.ReplaceAll(filepath.Base(fileHeader.Filename), " ", "_")
			fileKey = fmt.Sprintf("materials/%d/%d_%s", meetingID, time.Now().Unix(), cleanName)

			contentType := fileHeader.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			uploadedURL, err := storage.UploadFile(c.Request().Context(), fileKey, src, contentType)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengunggah berkas materi: " + err.Error()})
			}
			contentURL = uploadedURL
			_ = ext
		} else if mType == "video" || mType == "link" {
			contentURL = strings.TrimSpace(c.FormValue("content_url"))
			if contentURL == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tautan materi / video wajib diisi"})
			}
		} else if mType == "text" {
			textContent = strings.TrimSpace(c.FormValue("text_content"))
			if textContent == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Isi teks materi pembelajaran wajib diisi"})
			}
		}

		res, err := db.Exec(`
			INSERT INTO lms_materials (meeting_id, title, type, content_url, file_key, file_size, text_content)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, meetingID, title, mType, contentURL, fileKey, fileSize, textContent)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan materi pembelajaran"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message":     "Materi pembelajaran berhasil ditambahkan",
			"id":          id,
			"content_url": contentURL,
		})
	}
}

// DeleteMeetingMaterial deletes a learning material
func DeleteMeetingMaterial(db *sql.DB, storage *service.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID materi tidak valid"})
		}

		var fileKey string
		err = db.QueryRow(`
			SELECT mat.file_key
			FROM lms_materials mat
			JOIN lms_meetings m ON m.id = mat.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE mat.id = ? AND ta.staff_id = ?
		`, id, teacherID).Scan(&fileKey)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		_, _ = db.Exec("DELETE FROM lms_materials WHERE id = ?", id)
		if fileKey != "" {
			_ = storage.DeleteFile(c.Request().Context(), fileKey)
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Materi berhasil dihapus"})
	}
}

// --- TEACHER ASSIGNMENTS ---

// CreateMeetingAssignment creates an assignment for students
func CreateMeetingAssignment(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		meetingID, err := strconv.Atoi(c.FormValue("meeting_id"))
		if err != nil || meetingID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pertemuan tidak valid"})
		}

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
		instructions := strings.TrimSpace(c.FormValue("instructions"))
		allowedType := strings.TrimSpace(c.FormValue("allowed_type"))
		dueDate := strings.TrimSpace(c.FormValue("due_date"))
		maxPoints, _ := strconv.Atoi(c.FormValue("max_points"))

		if title == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Judul tugas wajib diisi"})
		}
		if allowedType == "" {
			allowedType = "any"
		}
		if maxPoints <= 0 {
			maxPoints = 100
		}

		res, err := db.Exec(`
			INSERT INTO lms_assignments (meeting_id, title, instructions, allowed_type, due_date, max_points, is_active)
			VALUES (?, ?, ?, ?, ?, ?, 1)
		`, meetingID, title, instructions, allowedType, dueDate, maxPoints)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat penugasan"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Penugasan berhasil dibuat",
			"id":      id,
		})
	}
}

// GetAssignmentSubmissions returns all student submissions for an assignment
func GetAssignmentSubmissions(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		assignmentID, err := strconv.Atoi(c.Param("id"))
		if err != nil || assignmentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tugas tidak valid"})
		}

		// Verify teacher ownership & get class_id
		var classID int
		var title string
		var maxPoints int
		err = db.QueryRow(`
			SELECT ta.class_id, a.title, a.max_points
			FROM lms_assignments a
			JOIN lms_meetings m ON m.id = a.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE a.id = ? AND ta.staff_id = ?
		`, assignmentID, teacherID).Scan(&classID, &title, &maxPoints)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		// Get all students in this class and their submission info
		rows, err := db.Query(`
			SELECT s.id, s.name, COALESCE(s.nis, s.nis_siswa, ''),
			       sub.id, sub.submission_type, sub.file_url, sub.link_url, sub.text_content,
			       sub.submitted_at, sub.is_late, sub.score, sub.feedback
			FROM students s
			LEFT JOIN lms_submissions sub ON sub.student_id = s.id AND sub.assignment_id = ?
			WHERE s.class_id = ? AND s.status = 'active'
			ORDER BY s.name ASC
		`, assignmentID, classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca pengumpulan tugas"})
		}
		defer rows.Close()

		submissions := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				studentID int
				name, nis string
				subID sql.NullInt64
				subType, fileURL, linkURL, textContent, submittedAt, feedback sql.NullString
				isLate sql.NullInt64
				score sql.NullFloat64
			)

			if err := rows.Scan(&studentID, &name, &nis, &subID, &subType, &fileURL, &linkURL, &textContent, &submittedAt, &isLate, &score, &feedback); err == nil {
				item := map[string]interface{}{
					"student_id":   studentID,
					"student_name": name,
					"student_nis":  nis,
					"has_submitted": subID.Valid,
				}
				if subID.Valid {
					item["submission_id"] = subID.Int64
					item["submission_type"] = subType.String
					item["file_url"] = fileURL.String
					item["link_url"] = linkURL.String
					item["text_content"] = textContent.String
					item["submitted_at"] = submittedAt.String
					item["is_late"] = isLate.Int64 == 1
					if score.Valid {
						item["score"] = score.Float64
					} else {
						item["score"] = nil
					}
					item["feedback"] = feedback.String
				}
				submissions = append(submissions, item)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"assignment_id": assignmentID,
			"title":         title,
			"max_points":    maxPoints,
			"submissions":   submissions,
		})
	}
}

// GradeSubmission records a score and feedback for a student submission
func GradeSubmission(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		submissionID, err := strconv.Atoi(c.FormValue("submission_id"))
		if err != nil || submissionID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID pengumpulan tidak valid"})
		}

		scoreStr := strings.TrimSpace(c.FormValue("score"))
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil || score < 0 || score > 100 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nilai harus berupa angka antara 0 hingga 100"})
		}

		feedback := strings.TrimSpace(c.FormValue("feedback"))

		// Verify teacher ownership
		var count int
		_ = db.QueryRow(`
			SELECT COUNT(*)
			FROM lms_submissions sub
			JOIN lms_assignments a ON a.id = sub.assignment_id
			JOIN lms_meetings m ON m.id = a.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE sub.id = ? AND ta.staff_id = ?
		`, submissionID, teacherID).Scan(&count)
		if count == 0 {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Akses ditolak"})
		}

		_, err = db.Exec(`
			UPDATE lms_submissions
			SET score = ?, feedback = ?, graded_by = ?, graded_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, score, feedback, teacherID, submissionID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan nilai"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Nilai dan feedback berhasil disimpan"})
	}
}

// --- STUDENT LMS APIS ---

// GetStudentCourses returns courses/mapel for the student's enrolled class
func GetStudentCourses(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		if studentID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Tidak terautentikasi"})
		}

		var classID int
		err := db.QueryRow("SELECT class_id FROM students WHERE id = ?", studentID).Scan(&classID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Data siswa tidak ditemukan"})
		}

		rows, err := db.Query(`
			SELECT ta.id, sub.id, sub.code, sub.name, sub.category,
			       s.id, s.name, COALESCE(s.nip, ''), ta.academic_year, ta.semester,
			       COALESCE(NULLIF(ta.banner_url, ''), COALESCE(sub.banner_url, ''))
			FROM teacher_assignments ta
			JOIN subjects sub ON sub.id = ta.subject_id
			JOIN staff s ON s.id = ta.staff_id
			WHERE ta.class_id = ? AND sub.is_active = 1
			ORDER BY sub.category ASC, sub.name ASC
		`, classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca mata pelajaran"})
		}
		defer rows.Close()

		type rawCourse struct {
			taID, subID, staffID, semester int
			subCode, subName, category, staffName, staffNIP, academicYear, bannerURL string
		}
		var rawList []rawCourse
		for rows.Next() {
			var r rawCourse
			if err := rows.Scan(&r.taID, &r.subID, &r.subCode, &r.subName, &r.category, &r.staffID, &r.staffName, &r.staffNIP, &r.academicYear, &r.semester, &r.bannerURL); err == nil {
				if r.bannerURL == "" {
					r.bannerURL = DefaultBannerForSubject(r.category, r.subCode, r.subName)
				}
				rawList = append(rawList, r)
			}
		}
		rows.Close()

		courses := make([]map[string]interface{}, 0)
		for _, r := range rawList {
			var meetCount int
			_ = db.QueryRow("SELECT COUNT(*) FROM lms_meetings WHERE assignment_id = ? AND is_published = 1", r.taID).Scan(&meetCount)

			courses = append(courses, map[string]interface{}{
				"assignment_id":  r.taID,
				"subject_id":     r.subID,
				"subject_code":   r.subCode,
				"subject_name":   r.subName,
				"category":       r.category,
				"teacher_name":   r.staffName,
				"teacher_nip":    r.staffNIP,
				"academic_year":  r.academicYear,
				"semester":       r.semester,
				"banner_url":     r.bannerURL,
				"total_meetings": meetCount,
			})
		}

		return c.JSON(http.StatusOK, courses)
	}
}

// GetStudentCourseDetail returns meetings, materials, and assignments for a course
func GetStudentCourseDetail(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := studentIDFromCtx(c)
		assignmentID, err := strconv.Atoi(c.Param("assignment_id"))
		if err != nil || assignmentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID kursus tidak valid"})
		}

		// Verify student is enrolled in this assignment's class
		var (
			classID int
			subName, subCode, subCat, teacherName, bannerURL string
		)
		err = db.QueryRow(`
			SELECT ta.class_id, sub.name, sub.code, COALESCE(sub.category, 'muatan_kejuruan'), s.name,
			       COALESCE(NULLIF(ta.banner_url, ''), COALESCE(sub.banner_url, ''))
			FROM teacher_assignments ta
			JOIN subjects sub ON sub.id = ta.subject_id
			JOIN staff s ON s.id = ta.staff_id
			WHERE ta.id = ?
		`, assignmentID).Scan(&classID, &subName, &subCode, &subCat, &teacherName, &bannerURL)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Mata pelajaran tidak ditemukan"})
		}
		if bannerURL == "" {
			bannerURL = DefaultBannerForSubject(subCat, subCode, subName)
		}

		var studentClassID int
		_ = db.QueryRow("SELECT class_id FROM students WHERE id = ?", studentID).Scan(&studentClassID)
		if studentClassID != classID {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Anda tidak terdaftar di kelas ini"})
		}

		// Get all published meetings
		rows, err := db.Query(`
			SELECT id, meeting_order, title, COALESCE(description, ''), COALESCE(strftime('%Y-%m-%d', meeting_date), meeting_date, '')
			FROM lms_meetings
			WHERE assignment_id = ? AND is_published = 1
			ORDER BY meeting_order ASC
		`, assignmentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca pertemuan"})
		}

		type rawMeeting struct {
			id, order int
			title, desc, mDate string
		}
		var rawMeetings []rawMeeting
		for rows.Next() {
			var r rawMeeting
			if err := rows.Scan(&r.id, &r.order, &r.title, &r.desc, &r.mDate); err == nil {
				rawMeetings = append(rawMeetings, r)
			}
		}
		rows.Close()

		meetings := make([]map[string]interface{}, 0)
		for _, m := range rawMeetings {
			// Attendance status for this student
			var attStatus string
			_ = db.QueryRow("SELECT status FROM lms_attendances WHERE meeting_id = ? AND student_id = ?", m.id, studentID).Scan(&attStatus)
			if attStatus == "" {
				attStatus = "-"
			}

			// Fetch materials
			matRows, _ := db.Query(`
				SELECT id, title, type, COALESCE(content_url, ''), COALESCE(file_size, 0), COALESCE(text_content, '')
				FROM lms_materials
				WHERE meeting_id = ?
				ORDER BY id ASC
			`, m.id)
			materials := make([]map[string]interface{}, 0)
			if matRows != nil {
				for matRows.Next() {
					var (
						matID int
						mTitle, mType, cURL, tContent string
						fSize int64
					)
					if err := matRows.Scan(&matID, &mTitle, &mType, &cURL, &fSize, &tContent); err == nil {
						materials = append(materials, map[string]interface{}{
							"id":           matID,
							"title":        mTitle,
							"type":         mType,
							"content_url":  cURL,
							"file_size":    fSize,
							"text_content": tContent,
						})
					}
				}
				matRows.Close()
			}

			// Fetch assignments
			assRows, _ := db.Query(`
				SELECT a.id, a.title, COALESCE(a.instructions, ''), COALESCE(a.allowed_type, 'any'), COALESCE(strftime('%Y-%m-%d %H:%M', a.due_date), ''), COALESCE(a.max_points, 100),
				       sub.id, sub.submission_type, sub.file_url, sub.link_url, sub.score, sub.feedback, sub.is_late
				FROM lms_assignments a
				LEFT JOIN lms_submissions sub ON sub.assignment_id = a.id AND sub.student_id = ?
				WHERE a.meeting_id = ? AND a.is_active = 1
				ORDER BY a.id ASC
			`, studentID, m.id)
			assignments := make([]map[string]interface{}, 0)
			if assRows != nil {
				for assRows.Next() {
					var (
						aID, maxPoints int
						aTitle, instructions, allowedType, dueDate string
						subID sql.NullInt64
						subType, fileURL, linkURL, feedback sql.NullString
						score sql.NullFloat64
						isLate sql.NullInt64
					)
					if err := assRows.Scan(&aID, &aTitle, &instructions, &allowedType, &dueDate, &maxPoints, &subID, &subType, &fileURL, &linkURL, &score, &feedback, &isLate); err == nil {
						assItem := map[string]interface{}{
							"id":           aID,
							"title":        aTitle,
							"instructions": instructions,
							"allowed_type": allowedType,
							"due_date":     dueDate,
							"max_points":   maxPoints,
							"submitted":    subID.Valid,
						}
						if subID.Valid {
							assItem["submission"] = map[string]interface{}{
								"id":              subID.Int64,
								"submission_type": subType.String,
								"file_url":        fileURL.String,
								"link_url":        linkURL.String,
								"feedback":        feedback.String,
								"is_late":         isLate.Int64 == 1,
								"score":           score.Float64,
								"is_graded":       score.Valid,
							}
						}
						assignments = append(assignments, assItem)
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
						var qCount int
						_ = db.QueryRow("SELECT COUNT(*) FROM lms_quiz_questions WHERE quiz_id = ?", qId).Scan(&qCount)

						var (
							myScore float64
							myCorrect, myPassed int
							completedAt sql.NullString
						)
						errAttempt := db.QueryRow(`
							SELECT score, total_correct, is_passed, completed_at
							FROM lms_quiz_attempts
							WHERE quiz_id = ? AND student_id = ?
						`, qId, studentID).Scan(&myScore, &myCorrect, &myPassed, &completedAt)

						qItem := map[string]interface{}{
							"id":                 qId,
							"title":              qTitle,
							"description":        qDesc,
							"time_limit_minutes": timeLimit,
							"passing_score":      passingScore,
							"total_questions":    qCount,
							"attempted":          errAttempt == nil && completedAt.Valid,
						}
						if errAttempt == nil && completedAt.Valid {
							qItem["my_attempt"] = map[string]interface{}{
								"score":         myScore,
								"total_correct": myCorrect,
								"is_passed":     myPassed == 1,
								"completed_at":  completedAt.String,
							}
						}
						quizzes = append(quizzes, qItem)
					}
				}
				quizRows.Close()
			}

			meetings = append(meetings, map[string]interface{}{
				"id":                m.id,
				"meeting_order":     m.order,
				"title":             m.title,
				"description":       m.desc,
				"meeting_date":      m.mDate,
				"attendance_status": attStatus,
				"materials":         materials,
				"assignments":       assignments,
				"quizzes":           quizzes,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"assignment_id": assignmentID,
			"subject_name":  subName,
			"subject_code":  subCode,
			"category":      subCat,
			"teacher_name":  teacherName,
			"banner_url":    bannerURL,
			"meetings":      meetings,
		})
	}
}

// SubmitStudentAssignment handles a student submitting an assignment
func SubmitStudentAssignment(db *sql.DB, storage *service.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentIDStr := studentIDFromCtx(c)
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil || studentID <= 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Sesi tidak valid"})
		}

		assignmentID, err := strconv.Atoi(c.FormValue("assignment_id"))
		if err != nil || assignmentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tugas tidak valid"})
		}

		subType := strings.TrimSpace(c.FormValue("submission_type")) // 'file', 'link', 'text'
		if subType == "" {
			subType = "file"
		}

		// Check assignment due date & student class
		var (
			dueDateStr sql.NullString
			classID    int
			studentNIS string
		)
		err = db.QueryRow(`
			SELECT a.due_date, ta.class_id
			FROM lms_assignments a
			JOIN lms_meetings m ON m.id = a.meeting_id
			JOIN teacher_assignments ta ON ta.id = m.assignment_id
			WHERE a.id = ?
		`, assignmentID).Scan(&dueDateStr, &classID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Tugas tidak ditemukan"})
		}

		_ = db.QueryRow("SELECT COALESCE(nis, nis_siswa, '') FROM students WHERE id = ?", studentID).Scan(&studentNIS)

		isLate := 0
		if dueDateStr.Valid && dueDateStr.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", dueDateStr.String); err == nil {
				if time.Now().After(t) {
					isLate = 1
				}
			}
		}

		var (
			fileURL     string
			fileKey     string
			linkURL     string
			textContent string
		)

		if subType == "file" {
			fh, err := c.FormFile("file")
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Berkas tugas wajib diunggah"})
			}
			src, err := fh.Open()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca berkas"})
			}
			defer src.Close()

			cleanName := strings.ReplaceAll(filepath.Base(fh.Filename), " ", "_")
			fileKey = fmt.Sprintf("submissions/%d/%s_%d_%s", assignmentID, studentNIS, time.Now().Unix(), cleanName)

			contentType := fh.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			uploadedURL, err := storage.UploadFile(c.Request().Context(), fileKey, src, contentType)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal upload berkas ke R2: " + err.Error()})
			}
			fileURL = uploadedURL
		} else if subType == "link" {
			linkURL = strings.TrimSpace(c.FormValue("link_url"))
			if linkURL == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Tautan tugas wajib diisi"})
			}
		} else if subType == "text" {
			textContent = strings.TrimSpace(c.FormValue("text_content"))
			if textContent == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Jawaban tugas wajib diisi"})
			}
		}

		// Upsert submission
		_, err = db.Exec(`
			INSERT INTO lms_submissions (assignment_id, student_id, submission_type, file_url, file_key, link_url, text_content, submitted_at, is_late)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
			ON CONFLICT(assignment_id, student_id) DO UPDATE SET
				submission_type = excluded.submission_type,
				file_url = excluded.file_url,
				file_key = excluded.file_key,
				link_url = excluded.link_url,
				text_content = excluded.text_content,
				submitted_at = CURRENT_TIMESTAMP,
				is_late = excluded.is_late
		`, assignmentID, studentID, subType, fileURL, fileKey, linkURL, textContent, isLate)

		if err != nil {
			// Fallback if SQLite doesn't support ON CONFLICT
			_, err = db.Exec(`
				INSERT OR REPLACE INTO lms_submissions (assignment_id, student_id, submission_type, file_url, file_key, link_url, text_content, submitted_at, is_late)
				VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
			`, assignmentID, studentID, subType, fileURL, fileKey, linkURL, textContent, isLate)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan pengumpulan tugas"})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":  "Tugas berhasil diserahkan!",
			"file_url": fileURL,
			"is_late":  isLate == 1,
		})
	}
}
