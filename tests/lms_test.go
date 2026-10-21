package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"belsekolah/internal/handler"
	"belsekolah/internal/service"
)

func TestLMSSubjectsAndAssignments(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Seed a Major, Class, and Teacher (Staff)
	resMajor, err := app.DB.Exec("INSERT INTO majors (name) VALUES (?)", "Rekayasa Perangkat Lunak")
	require.NoError(t, err)
	majorID, _ := resMajor.LastInsertId()

	resClass, err := app.DB.Exec("INSERT INTO classes (name, major_id) VALUES (?, ?)", "XII RPL 1", majorID)
	require.NoError(t, err)
	classID, _ := resClass.LastInsertId()

	resStaff, err := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES (?, ?, ?)", "198501012010011001", "Pak Budi, S.Kom", "Guru Produktif")
	require.NoError(t, err)
	staffID, _ := resStaff.LastInsertId()

	// 2. Test Add Subject
	form := make(url.Values)
	form.Set("code", "RPL-WEB-12")
	form.Set("name", "Pemrograman Web & Mobile")
	form.Set("category", "muatan_kejuruan")
	form.Set("major_id", fmt.Sprintf("%d", majorID))

	req := httptest.NewRequest(http.MethodPost, "/admin/subject/add", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.AddSubject(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var addSubResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &addSubResp)
	assert.NoError(t, err)
	subjectID := int(addSubResp["id"].(float64))
	assert.Greater(t, subjectID, 0)

	// 3. Test Get Subjects
	req = httptest.NewRequest(http.MethodGet, "/admin/subjects", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handler.GetSubjects(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var subjects []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &subjects)
	assert.NoError(t, err)
	assert.Len(t, subjects, 1)
	assert.Equal(t, "RPL-WEB-12", subjects[0]["code"])

	// 4. Test Add Teacher Assignment
	formTA := make(url.Values)
	formTA.Set("staff_id", fmt.Sprintf("%d", staffID))
	formTA.Set("subject_id", fmt.Sprintf("%d", subjectID))
	formTA.Set("class_id", fmt.Sprintf("%d", classID))
	formTA.Set("academic_year", "2026/2027")
	formTA.Set("semester", "1")

	req = httptest.NewRequest(http.MethodPost, "/admin/teacher-assignment/add", strings.NewReader(formTA.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handler.AddTeacherAssignment(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var addTAResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &addTAResp)
	assert.NoError(t, err)
	assignmentID := int(addTAResp["id"].(float64))
	assert.Greater(t, assignmentID, 0)

	// 5. Test Get Teacher Assignments
	req = httptest.NewRequest(http.MethodGet, "/admin/teacher-assignments", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handler.GetTeacherAssignments(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var assignments []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &assignments)
	assert.NoError(t, err)
	assert.Len(t, assignments, 1)
	assert.Equal(t, "XII RPL 1", assignments[0]["class_name"])
}

func TestTeacherAuthAndKBM(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Seed Teacher, Class, Subject, Students, and Assignment
	nip := "199001012015011002"
	resStaff, err := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES (?, ?, ?)", nip, "Bu Siti Rahayu", "Guru")
	require.NoError(t, err)
	staffID, _ := resStaff.LastInsertId()

	resClass, err := app.DB.Exec("INSERT INTO classes (name) VALUES (?)", "X TKJ 1")
	require.NoError(t, err)
	classID, _ := resClass.LastInsertId()

	// Seed 2 students in that class
	resS1, err := app.DB.Exec("INSERT INTO students (nis, name, class_id, status) VALUES (?, ?, ?, 'active')", "202401", "Ahmad Dani", classID)
	require.NoError(t, err)
	s1ID, _ := resS1.LastInsertId()

	resS2, err := app.DB.Exec("INSERT INTO students (nis, name, class_id, status) VALUES (?, ?, ?, 'active')", "202402", "Bela Safitri", classID)
	require.NoError(t, err)
	s2ID, _ := resS2.LastInsertId()

	resSub, err := app.DB.Exec("INSERT INTO subjects (code, name) VALUES (?, ?)", "TKJ-NET-10", "Dasar Jaringan Komputer")
	require.NoError(t, err)
	subID, _ := resSub.LastInsertId()

	resTA, err := app.DB.Exec(
		"INSERT INTO teacher_assignments (staff_id, subject_id, class_id, academic_year, semester) VALUES (?, ?, ?, '2026/2027', 1)",
		staffID, subID, classID,
	)
	require.NoError(t, err)
	assignmentID, _ := resTA.LastInsertId()

	// 2. Test Teacher Login with Default Password (or NIP)
	formLogin := make(url.Values)
	formLogin.Set("nip", nip)
	formLogin.Set("password", handler.TeacherDefaultPassword)

	req := httptest.NewRequest(http.MethodPost, "/api/guru/login", strings.NewReader(formLogin.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.TeacherLogin(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Extract session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == handler.TeacherSessionCookie {
			sessionCookie = ck
			break
		}
	}
	require.NotNil(t, sessionCookie, "Teacher session cookie must be set")
	assert.NotEmpty(t, sessionCookie.Value)

	// 3. Test Teacher Profile endpoint with context
	req = httptest.NewRequest(http.MethodGet, "/api/guru/profile", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.GetTeacherProfile(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var profResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &profResp)
	assert.NoError(t, err)
	profile := profResp["profile"].(map[string]interface{})
	assert.Equal(t, "Bu Siti Rahayu", profile["name"])
	assignments := profResp["assignments"].([]interface{})
	assert.Len(t, assignments, 1)

	// 3b. Test Get Teacher Meetings before any meeting created (MUST return empty array [], NOT null)
	reqEmpty := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/meetings/%d", assignmentID), nil)
	recEmpty := httptest.NewRecorder()
	cEmpty := e.NewContext(reqEmpty, recEmpty)
	cEmpty.SetParamNames("assignment_id")
	cEmpty.SetParamValues(fmt.Sprintf("%d", assignmentID))
	cEmpty.Set("teacher_id", int(staffID))

	err = handler.GetTeacherMeetings(app.DB)(cEmpty)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recEmpty.Code)
	assert.Equal(t, "[]\n", recEmpty.Body.String(), "Empty meetings must serialize to [] not null")

	// 3c. Test Get Teacher Meetings for forbidden assignment
	reqForbid := httptest.NewRequest(http.MethodGet, "/api/guru/meetings/9999", nil)
	recForbid := httptest.NewRecorder()
	cForbid := e.NewContext(reqForbid, recForbid)
	cForbid.SetParamNames("assignment_id")
	cForbid.SetParamValues("9999")
	cForbid.Set("teacher_id", int(staffID))

	err = handler.GetTeacherMeetings(app.DB)(cForbid)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, recForbid.Code)

	// 4. Test Create KBM Meeting
	formMeeting := make(url.Values)
	formMeeting.Set("assignment_id", fmt.Sprintf("%d", assignmentID))
	formMeeting.Set("meeting_order", "1")
	formMeeting.Set("title", "Pengenalan Topologi Jaringan")
	formMeeting.Set("meeting_date", time.Now().Format("2006-01-02"))
	formMeeting.Set("description", "Silakan simak materi dan kerjakan tugas praktikum")

	req = httptest.NewRequest(http.MethodPost, "/api/guru/meetings", strings.NewReader(formMeeting.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.CreateTeacherMeeting(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var meetResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &meetResp)
	assert.NoError(t, err)
	meetingID := int(meetResp["id"].(float64))
	assert.Greater(t, meetingID, 0)

	// 4b. Test Get Teacher Meetings with existing meeting (materials, assignments, quizzes must be arrays, not null)
	reqMeetings := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/meetings/%d", assignmentID), nil)
	recMeetings := httptest.NewRecorder()
	cMeetings := e.NewContext(reqMeetings, recMeetings)
	cMeetings.SetParamNames("assignment_id")
	cMeetings.SetParamValues(fmt.Sprintf("%d", assignmentID))
	cMeetings.Set("teacher_id", int(staffID))

	err = handler.GetTeacherMeetings(app.DB)(cMeetings)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recMeetings.Code)

	var meetingsList []map[string]interface{}
	err = json.Unmarshal(recMeetings.Body.Bytes(), &meetingsList)
	assert.NoError(t, err)
	require.Len(t, meetingsList, 1)
	assert.Equal(t, "Pengenalan Topologi Jaringan", meetingsList[0]["title"])
	assert.NotNil(t, meetingsList[0]["materials"])
	assert.NotNil(t, meetingsList[0]["assignments"])
	assert.NotNil(t, meetingsList[0]["quizzes"])

	// 5. Test Get Meeting Attendance (all students in class should appear)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/meetings/%d/attendance", meetingID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("meeting_id")
	c.SetParamValues(fmt.Sprintf("%d", meetingID))
	c.Set("teacher_id", int(staffID))

	err = handler.GetMeetingAttendance(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var attResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &attResp)
	assert.NoError(t, err)
	students := attResp["students"].([]interface{})
	assert.Len(t, students, 2)

	// 6. Test Save Meeting Attendance (Ahmad = Hadir, Bela = Izin)
	savePayload := map[string]interface{}{
		"attendances": []map[string]interface{}{
			{
				"student_id": s1ID,
				"status":     "H",
				"notes":      "Hadir tepat waktu",
			},
			{
				"student_id": s2ID,
				"status":     "I",
				"notes":      "Surat dokter",
			},
		},
	}
	payloadBytes, _ := json.Marshal(savePayload)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/guru/meetings/%d/attendance", meetingID), bytes.NewReader(payloadBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("meeting_id")
	c.SetParamValues(fmt.Sprintf("%d", meetingID))
	c.Set("teacher_id", int(staffID))

	err = handler.SaveMeetingAttendance(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify in DB
	var hCount, iCount int
	_ = app.DB.QueryRow("SELECT COUNT(*) FROM lms_attendances WHERE meeting_id = ? AND status = 'H'", meetingID).Scan(&hCount)
	_ = app.DB.QueryRow("SELECT COUNT(*) FROM lms_attendances WHERE meeting_id = ? AND status = 'I'", meetingID).Scan(&iCount)
	assert.Equal(t, 1, hCount)
	assert.Equal(t, 1, iCount)
}

func TestLMSMaterialsAndAssignmentsLifecycle(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()
	storage := service.NewStorageService(app.DB)

	// Seed Teacher, Class, Student, Subject, Assignment, Meeting
	resStaff, err := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES (?, ?, ?)", "198808082012011003", "Pak Bambang", "Guru")
	require.NoError(t, err)
	staffID, _ := resStaff.LastInsertId()

	resClass, err := app.DB.Exec("INSERT INTO classes (name) VALUES (?)", "XI RPL 2")
	require.NoError(t, err)
	classID, _ := resClass.LastInsertId()

	resStudent, err := app.DB.Exec("INSERT INTO students (nis, name, class_id, status) VALUES (?, ?, ?, 'active')", "202499", "Citra Kirana", classID)
	require.NoError(t, err)
	studentID, _ := resStudent.LastInsertId()

	resSub, err := app.DB.Exec("INSERT INTO subjects (code, name) VALUES (?, ?)", "RPL-DB-11", "Basis Data Relasional")
	require.NoError(t, err)
	subID, _ := resSub.LastInsertId()

	resTA, err := app.DB.Exec("INSERT INTO teacher_assignments (staff_id, subject_id, class_id, academic_year, semester) VALUES (?, ?, ?, '2026/2027', 1)", staffID, subID, classID)
	require.NoError(t, err)
	taID, _ := resTA.LastInsertId()

	resM, err := app.DB.Exec("INSERT INTO lms_meetings (assignment_id, meeting_order, title, meeting_date) VALUES (?, 1, 'Pengenalan SQL DDL', '2026-10-21')", taID)
	require.NoError(t, err)
	meetingID, _ := resM.LastInsertId()

	// 1. Teacher adds Material (Link/Video & Text)
	formMat := make(url.Values)
	formMat.Set("meeting_id", fmt.Sprintf("%d", meetingID))
	formMat.Set("title", "Video Tutorial DDL")
	formMat.Set("type", "video")
	formMat.Set("content_url", "https://youtube.com/watch?v=sample")

	req := httptest.NewRequest(http.MethodPost, "/api/guru/materials", strings.NewReader(formMat.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.AddMeetingMaterial(app.DB, storage)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// 2. Teacher creates Assignment
	formAss := make(url.Values)
	formAss.Set("meeting_id", fmt.Sprintf("%d", meetingID))
	formAss.Set("title", "Tugas Praktik DDL Table Mahasiswa")
	formAss.Set("instructions", "Buat tabel menggunakan query DDL dan cantumkan link repo github")
	formAss.Set("allowed_type", "link")
	formAss.Set("max_points", "100")
	formAss.Set("due_date", "2026-12-31 23:59:00")

	req = httptest.NewRequest(http.MethodPost, "/api/guru/assignments", strings.NewReader(formAss.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.CreateMeetingAssignment(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var assResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &assResp)
	assert.NoError(t, err)
	assignmentID := int(assResp["id"].(float64))
	assert.Greater(t, assignmentID, 0)

	// 3. Student views courses
	req = httptest.NewRequest(http.MethodGet, "/api/student/courses", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.GetStudentCourses(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var courses []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &courses)
	assert.NoError(t, err)
	assert.Len(t, courses, 1)
	assert.Equal(t, "Basis Data Relasional", courses[0]["subject_name"])

	// 4. Student views course detail (meetings, materials, assignments)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/student/courses/%d", taID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("assignment_id")
	c.SetParamValues(fmt.Sprintf("%d", taID))
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.GetStudentCourseDetail(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var courseDetail map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &courseDetail)
	assert.NoError(t, err)
	meetings := courseDetail["meetings"].([]interface{})
	assert.Len(t, meetings, 1)

	// 5. Student submits Assignment
	formSub := make(url.Values)
	formSub.Set("assignment_id", fmt.Sprintf("%d", assignmentID))
	formSub.Set("submission_type", "link")
	formSub.Set("link_url", "https://github.com/citra/praktik-ddl")

	req = httptest.NewRequest(http.MethodPost, "/api/student/assignments/submit", strings.NewReader(formSub.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.SubmitStudentAssignment(app.DB, storage)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 6. Teacher views submissions
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/assignments/%d/submissions", assignmentID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", assignmentID))
	c.Set("teacher_id", int(staffID))

	err = handler.GetAssignmentSubmissions(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var subsResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &subsResp)
	assert.NoError(t, err)
	subList := subsResp["submissions"].([]interface{})
	assert.Len(t, subList, 1)
	subItem := subList[0].(map[string]interface{})
	assert.Equal(t, true, subItem["has_submitted"])
	submissionID := int(subItem["submission_id"].(float64))

	// 7. Teacher grades submission (Score: 95, Feedback: "Query rapi dan benar")
	formGrade := make(url.Values)
	formGrade.Set("submission_id", fmt.Sprintf("%d", submissionID))
	formGrade.Set("score", "95")
	formGrade.Set("feedback", "Query rapi dan benar, pertahankan!")

	req = httptest.NewRequest(http.MethodPost, "/api/guru/submissions/grade", strings.NewReader(formGrade.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.GradeSubmission(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 8. Verify score in student course detail
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/student/courses/%d", taID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("assignment_id")
	c.SetParamValues(fmt.Sprintf("%d", taID))
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.GetStudentCourseDetail(app.DB)(c)
	assert.NoError(t, err)
	err = json.Unmarshal(rec.Body.Bytes(), &courseDetail)
	assert.NoError(t, err)
	m1 := courseDetail["meetings"].([]interface{})[0].(map[string]interface{})
	a1 := m1["assignments"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, true, a1["submitted"])
	sub1 := a1["submission"].(map[string]interface{})
	assert.Equal(t, float64(95), sub1["score"])
	assert.Equal(t, "Query rapi dan benar, pertahankan!", sub1["feedback"])
}

func TestLMSCBTQuiz(t *testing.T) {
	app, teardown := setupTestApp(t)
	defer teardown()

	e := echo.New()

	// 1. Seed Teacher, Class, Student, Subject, Assignment, Meeting
	resStaff, err := app.DB.Exec("INSERT INTO staff (nip, name, role) VALUES (?, ?, ?)", "197707072005011004", "Pak Hendra", "Guru Produktif")
	require.NoError(t, err)
	staffID, _ := resStaff.LastInsertId()

	resClass, err := app.DB.Exec("INSERT INTO classes (name) VALUES (?)", "XII RPL 2")
	require.NoError(t, err)
	classID, _ := resClass.LastInsertId()

	resStudent, err := app.DB.Exec("INSERT INTO students (nis, name, class_id, status) VALUES (?, ?, ?, 'active')", "202501", "Doni Pratama", classID)
	require.NoError(t, err)
	studentID, _ := resStudent.LastInsertId()

	resSub, err := app.DB.Exec("INSERT INTO subjects (code, name) VALUES (?, ?)", "RPL-OOP-12", "Pemrograman Berorientasi Objek")
	require.NoError(t, err)
	subID, _ := resSub.LastInsertId()

	resTA, err := app.DB.Exec("INSERT INTO teacher_assignments (staff_id, subject_id, class_id, academic_year, semester) VALUES (?, ?, ?, '2026/2027', 1)", staffID, subID, classID)
	require.NoError(t, err)
	taID, _ := resTA.LastInsertId()

	resM, err := app.DB.Exec("INSERT INTO lms_meetings (assignment_id, meeting_order, title, meeting_date) VALUES (?, 1, 'Kuis Konsep Enkapsulasi & Inheritance', '2026-10-21')", taID)
	require.NoError(t, err)
	meetingID, _ := resM.LastInsertId()

	// 2. Teacher creates Quiz
	formQuiz := make(url.Values)
	formQuiz.Set("meeting_id", fmt.Sprintf("%d", meetingID))
	formQuiz.Set("title", "Kuis Harian OOP")
	formQuiz.Set("time_limit_minutes", "15")
	formQuiz.Set("passing_score", "75")

	req := httptest.NewRequest(http.MethodPost, "/api/guru/quizzes", strings.NewReader(formQuiz.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.CreateMeetingQuiz(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var qResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &qResp)
	assert.NoError(t, err)
	quizID := int(qResp["id"].(float64))
	assert.Greater(t, quizID, 0)

	// 3. Teacher adds 2 questions with multiple choice options
	// Question 1: Enkapsulasi keyword
	q1Payload := handler.AddQuizQuestionRequest{
		QuizID:       quizID,
		QuestionText: "Keyword apa yang digunakan untuk menyembunyikan data agar hanya bisa diakses dalam class yang sama?",
		Points:       50,
		CorrectIndex: 1, // 'private'
		OptionTexts:  []string{"public", "private", "protected", "default"},
	}
	q1Bytes, _ := json.Marshal(q1Payload)
	req = httptest.NewRequest(http.MethodPost, "/api/guru/quizzes/questions", bytes.NewReader(q1Bytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.AddQuizQuestion(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var q1Resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &q1Resp)
	q1ID := int(q1Resp["question_id"].(float64))

	// Question 2: Inheritance keyword in Java/PHP
	q2Payload := handler.AddQuizQuestionRequest{
		QuizID:       quizID,
		QuestionText: "Keyword untuk mewarisi sifat dari parent class adalah?",
		Points:       50,
		CorrectIndex: 0, // 'extends'
		OptionTexts:  []string{"extends", "implements", "inherits", "super"},
	}
	q2Bytes, _ := json.Marshal(q2Payload)
	req = httptest.NewRequest(http.MethodPost, "/api/guru/quizzes/questions", bytes.NewReader(q2Bytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("teacher_id", int(staffID))

	err = handler.AddQuizQuestion(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var q2Resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &q2Resp)
	q2ID := int(q2Resp["question_id"].(float64))

	// 4. Student opens Quiz (questions without is_correct)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/student/quizzes/%d", quizID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", quizID))
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.GetStudentQuiz(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var studentQuizResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &studentQuizResp)
	assert.NoError(t, err)
	questions := studentQuizResp["questions"].([]interface{})
	assert.Len(t, questions, 2)

	// Fetch option IDs for answers from DB
	var correctOptQ1, wrongOptQ2 int
	_ = app.DB.QueryRow("SELECT id FROM lms_quiz_options WHERE question_id = ? AND is_correct = 1", q1ID).Scan(&correctOptQ1)
	_ = app.DB.QueryRow("SELECT id FROM lms_quiz_options WHERE question_id = ? AND is_correct = 0 LIMIT 1", q2ID).Scan(&wrongOptQ2)

	// 5. Student submits answers (1 correct, 1 wrong -> score 50%)
	submitPayload := map[string]interface{}{
		"quiz_id": quizID,
		"answers": []map[string]interface{}{
			{"question_id": q1ID, "option_id": correctOptQ1},
			{"question_id": q2ID, "option_id": wrongOptQ2},
		},
	}
	subBytes, _ := json.Marshal(submitPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/student/quizzes/submit", bytes.NewReader(subBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("student_id", fmt.Sprintf("%d", studentID))

	err = handler.SubmitStudentQuiz(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var submitResult map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &submitResult)
	assert.NoError(t, err)
	assert.Equal(t, float64(50), submitResult["score"])
	assert.Equal(t, float64(1), submitResult["total_correct"])
	assert.Equal(t, false, submitResult["is_passed"]) // 50 < passing_score 75

	// 6. Teacher views Quiz attempts
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/quizzes/%d/attempts", quizID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", quizID))
	c.Set("teacher_id", int(staffID))

	err = handler.GetQuizAttempts(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var attemptsResp map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &attemptsResp)
	assert.NoError(t, err)
	attempts := attemptsResp["attempts"].([]interface{})
	assert.Len(t, attempts, 1)
	attemptItem := attempts[0].(map[string]interface{})
	assert.Equal(t, "Doni Pratama", attemptItem["student_name"])
	assert.Equal(t, float64(50), attemptItem["score"])
	assert.Equal(t, false, attemptItem["is_passed"])

	// 7. Teacher views Quiz Questions
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/guru/quizzes/%d/questions", quizID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", quizID))
	c.Set("teacher_id", int(staffID))

	err = handler.GetQuizQuestions(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var qList []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &qList)
	assert.NoError(t, err)
	assert.Len(t, qList, 2)

	// 8. Teacher deletes a question
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/guru/quizzes/questions/%d", q2ID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", q2ID))
	c.Set("teacher_id", int(staffID))

	err = handler.DeleteQuizQuestion(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 9. Teacher deletes the entire quiz
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/guru/quizzes/%d", quizID), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", quizID))
	c.Set("teacher_id", int(staffID))

	err = handler.DeleteMeetingQuiz(app.DB)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

