package router

import (
	"database/sql"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	appMiddleware "belsekolah/internal/middleware"
	"belsekolah/internal/service"
)

type templateRenderer struct {
	templates *template.Template
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	data = withVersionData(data)
	return t.templates.ExecuteTemplate(w, name, data)
}

func withVersionData(data interface{}) interface{} {
	versionInfo := map[string]string{
		"version":    os.Getenv("APP_VERSION"),
		"build_date": os.Getenv("BUILD_DATE"),
		"git_commit": os.Getenv("GIT_COMMIT"),
		"app_name":   os.Getenv("APP_NAME"),
	}
	if versionInfo["version"] == "" {
		versionInfo["version"] = "2.1.0"
	}
	if versionInfo["app_name"] == "" {
		versionInfo["app_name"] = "NIBA SuperApps"
	}
	if versionInfo["build_date"] == "" {
		versionInfo["build_date"] = "unknown"
	}
	if versionInfo["git_commit"] == "" {
		versionInfo["git_commit"] = "unknown"
	}

	appVersion := "v" + versionInfo["version"]

	if data == nil {
		return map[string]interface{}{
			"AppVersion":  appVersion,
			"AppName":     versionInfo["app_name"],
			"BuildDate":   versionInfo["build_date"],
			"GitCommit":   versionInfo["git_commit"],
			"FullVersion": versionInfo["app_name"] + " " + appVersion + " (build " + versionInfo["build_date"] + ")",
			"VersionInfo": versionInfo,
		}
	}
	if m, ok := data.(map[string]interface{}); ok {
		m["AppVersion"] = appVersion
		m["AppName"] = versionInfo["app_name"]
		m["BuildDate"] = versionInfo["build_date"]
		m["GitCommit"] = versionInfo["git_commit"]
		m["FullVersion"] = versionInfo["app_name"] + " " + appVersion + " (build " + versionInfo["build_date"] + ")"
		m["VersionInfo"] = versionInfo
		return m
	}
	return data
}

// AppHandlers contains handlers that are still methods on the main App struct.
// Once those are extracted, this interface can be removed.
type AppHandlers interface {
	DashboardHandler(c echo.Context) error
	SchedulesPageHandler(c echo.Context) error
	AudioPageHandler(c echo.Context) error
	AttendancePageHandler(c echo.Context) error
	AcademicsPageHandler(c echo.Context) error
	PointsPageHandler(c echo.Context) error
	SettingsPageHandler(c echo.Context) error
	ReportsPageHandler(c echo.Context) error
	ManualAttendanceHandler(c echo.Context) error
	RecordAttendanceHandler(c echo.Context) error
	PrayerAttendanceHandler(c echo.Context) error
	TodayStatsHandler(c echo.Context) error
	RecentLogsHandler(c echo.Context) error
	VerifyFaceAttendanceHandler(c echo.Context) error
	SyncHandler(c echo.Context) error
	PublicLeaderboardHandler(c echo.Context) error
	VersionHandler(c echo.Context) error
}

// Register sets up all routes on the given Echo instance and starts the HTTP server.
func Register(e *echo.Echo, db *sql.DB, viewsFS fs.FS, app AppHandlers) {
	RegisterRoutes(e, db, viewsFS, app)

	port := getEnv("PORT", config.GetServerPort())
	host := config.GetServerHost()
	e.Logger.Printf("Starting SMK NIBA Super Apps server on %s:%s", host, port)
	e.Logger.Fatal(e.Start(host + ":" + port))
}

// RegisterRoutes sets up all routes and middlewares on the given Echo instance without starting the HTTP server.
func RegisterRoutes(e *echo.Echo, db *sql.DB, viewsFS fs.FS, app AppHandlers) {
	storageService := service.NewStorageService(db)

	// SEC-09: Secure IP Extraction against IP spoofing via X-Forwarded-For / X-Real-IP headers
	e.IPExtractor = echo.ExtractIPDirect()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(appMiddleware.SecurityHeaders())
	e.Use(appMiddleware.CSRF())

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		var message interface{} = "Terjadi kesalahan server"
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message
		}

		if code == http.StatusNotFound {
			if strings.HasPrefix(c.Request().URL.Path, "/api") {
				c.JSON(http.StatusNotFound, map[string]string{"message": "Not Found"})
				return
			}
			if err := c.Render(http.StatusNotFound, "404.html", nil); err != nil {
				c.Logger().Error(err)
			}
			return
		}

		if code == http.StatusForbidden {
			if strings.HasPrefix(c.Request().URL.Path, "/api") {
				msgStr, ok := message.(string)
				if !ok || msgStr == "" {
					msgStr = "Akses ditolak"
				}
				c.JSON(http.StatusForbidden, map[string]string{"message": msgStr})
				return
			}
			if err := c.Render(http.StatusForbidden, "403.html", nil); err != nil {
				c.Logger().Error(err)
			}
			return
		}

		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Renderer = &templateRenderer{
		templates: template.Must(template.ParseFS(viewsFS, "views/*.html", "views/admin/*.html")),
	}
	e.Static("/assets", "public/assets")
	e.Static("/assets/uploads", "public/uploads")

	// --- PUBLIC ROUTES ---
	e.GET("/", func(c echo.Context) error { return c.Render(http.StatusOK, "index.html", nil) })
	e.GET("/404", func(c echo.Context) error { return c.Render(http.StatusNotFound, "404.html", nil) })
	e.GET("/403", func(c echo.Context) error { return c.Render(http.StatusForbidden, "403.html", nil) })
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	}, appMiddleware.AlreadyLoggedIn())
	e.GET("/scan", handler.ScanPage(db))
	e.GET("/scan-face", handler.ScanFacePage(db))
	e.GET("/scan-sholat", handler.ScanPrayerPage(db))

	e.POST("/api/login", handler.Login(db), appMiddleware.AuthRateLimiter())
	e.GET("/logout", handler.Logout())
	e.POST("/logout", handler.Logout())
	e.GET("/api/logout", handler.Logout())
	e.POST("/api/logout", handler.Logout())
	e.GET("/api/version", app.VersionHandler)
	e.GET("/api/sync", app.SyncHandler)
	e.GET("/api/leaderboard", app.PublicLeaderboardHandler)
	e.GET("/health", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}
		appName := os.Getenv("APP_NAME")
		if appName == "" {
			appName = "NIBA SuperApps"
		}
		appVer := os.Getenv("APP_VERSION")
		if appVer == "" {
			appVer = "2.1.0"
		}
		if !strings.HasPrefix(appVer, "v") {
			appVer = "v" + appVer
		}
		gitCommit := os.Getenv("GIT_COMMIT")
		if gitCommit == "" {
			gitCommit = "unknown"
		}
		return c.JSON(http.StatusOK, map[string]string{
			"status":     "ok",
			"app":        appName,
			"version":    appVer,
			"git_commit": gitCommit,
		})
	})
	e.GET("/api/point-rules", handler.GetPointRulesV2(db))
	e.GET("/api/attendance/prayer-logs", handler.PrayerLogs(db))
	e.GET("/api/attendance/record", app.RecordAttendanceHandler, handler.DeviceAuth(db))
	e.POST("/api/attendance/record", app.RecordAttendanceHandler, handler.DeviceAuth(db))
	e.GET("/api/attendance/prayer", app.PrayerAttendanceHandler, handler.DeviceAuth(db))
	e.POST("/api/attendance/prayer", app.PrayerAttendanceHandler, handler.DeviceAuth(db))
	e.GET("/api/attendance/today-stats", app.TodayStatsHandler, handler.DeviceAuth(db))
	e.GET("/api/attendance/recent-logs", app.RecentLogsHandler, handler.DeviceAuth(db))
	e.POST("/api/attendance/verify-face", app.VerifyFaceAttendanceHandler, handler.DeviceAuth(db))
	e.GET("/api/face/descriptors", handler.GetFaceDescriptors(db), handler.DeviceAuth(db))
	e.POST("/api/attendance/test-wa", handler.TestWA(db), appMiddleware.AdminAuth())

	// --- ADMIN ROUTES ---
	admin := e.Group("/admin", appMiddleware.AdminAuth())
	admin.GET("", app.DashboardHandler)
	admin.GET("/dashboard", app.DashboardHandler)
	admin.GET("/logout", handler.Logout())
	admin.POST("/logout", handler.Logout())
	admin.GET("/schedules", app.SchedulesPageHandler)
	admin.GET("/audio", app.AudioPageHandler)
	admin.GET("/attendance", app.AttendancePageHandler)
	admin.GET("/academics", app.AcademicsPageHandler)
	admin.GET("/points", app.PointsPageHandler)
	admin.GET("/settings", app.SettingsPageHandler)
	admin.GET("/reports", app.ReportsPageHandler)
	admin.GET("/journals", func(c echo.Context) error { return c.Render(http.StatusOK, "journals.html", nil) })

	// Admin account
	admin.PUT("/password", handler.AdminChangePassword(db))

	// Journal approval (jurnal mengajar guru)
	admin.GET("/journals/json", handler.AdminListJournals(db))
	admin.POST("/journals/:id/approve", handler.AdminApproveJournal(db))
	admin.POST("/journals/:id/reject", handler.AdminRejectJournal(db))

	// Global search & LMS analytics
	admin.GET("/search", handler.AdminGlobalSearch(db))
	admin.GET("/lms-analytics", handler.AdminLMSAnalytics(db))

	admin.GET("/announcements", handler.GetAnnouncements(db))
	admin.POST("/announcement/add", handler.CreateAnnouncement(db))
	admin.DELETE("/announcement/:id", handler.DeleteAnnouncement(db))
	admin.POST("/announcement/play/:id", handler.PlayAnnouncement(db))

	admin.POST("/schedule/add", handler.AddSchedule(db))
	admin.POST("/schedule/update/:id", handler.UpdateSchedule(db))
	admin.DELETE("/schedule/:id", handler.DeleteSchedule(db))

	admin.POST("/audio/upload", handler.UploadAudio(db))
	admin.POST("/audio/rename/:id", handler.RenameAudio(db))
	admin.DELETE("/audio/:id", handler.DeleteAudio(db))

	admin.POST("/device/add", handler.AddDevice(db))
	admin.POST("/device/update/:id", handler.UpdateDevice(db))
	admin.DELETE("/device/:id", handler.DeleteDevice(db))

	admin.POST("/major/add", handler.AddMajor(db))
	admin.POST("/major/update/:id", handler.UpdateMajor(db))
	admin.DELETE("/major/:id", handler.DeleteMajor(db))

	admin.POST("/class/add", handler.AddClass(db))
	admin.POST("/class/update/:id", handler.UpdateClass(db))
	admin.DELETE("/class/:id", handler.DeleteClass(db))
	admin.GET("/classes/json", handler.GetClassesJSON(db))

	admin.POST("/student/add", handler.AddStudent(db))
	admin.POST("/student/update/:id", handler.UpdateStudent(db))
	admin.POST("/student/status/:id", handler.UpdateStudentStatus(db))
	admin.DELETE("/student/:id", handler.DeleteStudent(db))
	admin.POST("/student/import", handler.ImportStudents(db))
	admin.POST("/student/import-json", handler.ImportStudentsJSON(db))
	admin.POST("/students/delete-multiple", handler.BulkDeleteStudents(db))
	admin.GET("/students/json", handler.GetStudentsJSON(db))
	admin.GET("/idcard", func(c echo.Context) error { return c.Render(http.StatusOK, "idcard.html", nil) })
	admin.POST("/students/promote", handler.PromoteStudents(db))
	admin.POST("/students/bulk-status", handler.BulkUpdateStudentStatus(db))
	admin.GET("/student/:id", handler.StudentProfile(db))
	admin.GET("/student/idcard/:id", handler.GetStudentIDCard(db))
	admin.GET("/students/idcard", handler.GetAllStudentsForIDCard(db))
	admin.GET("/staff/:id", handler.StaffProfile(db))
	admin.GET("/face/register/:id", handler.FaceRegisterPage(db))
	admin.GET("/face/register-staff/:id", handler.StaffFaceRegisterPage(db))

	// Recent Activity
	admin.GET("/recent-activity", handler.GetRecentActivity(db))

	admin.POST("/staff/add", handler.AddStaff(db))
	admin.POST("/staff/update/:id", handler.UpdateStaff(db))
	admin.DELETE("/staff/:id", handler.DeleteStaff(db))
	admin.POST("/staff/import", handler.ImportStaff(db))

	// Curriculum & LMS Subject Management
	admin.GET("/subjects", handler.GetSubjects(db))
	admin.POST("/subject/add", handler.AddSubject(db))
	admin.POST("/subject/update/:id", handler.UpdateSubject(db))
	admin.DELETE("/subject/:id", handler.DeleteSubject(db))
	admin.GET("/teacher-assignments", handler.GetTeacherAssignments(db))
	admin.POST("/teacher-assignment/add", handler.AddTeacherAssignment(db))
	admin.DELETE("/teacher-assignment/:id", handler.DeleteTeacherAssignment(db))

	admin.GET("/holidays", handler.GetHolidays(db))
	admin.GET("/academic-calendar", handler.GetAcademicCalendar(db))
	admin.POST("/academic-calendar", handler.CreateAcademicCalendarEvent(db))
	admin.PUT("/academic-calendar/:id", handler.UpdateAcademicCalendarEvent(db))
	admin.DELETE("/academic-calendar/:id", handler.DeleteAcademicCalendarEvent(db))
	admin.POST("/holiday/add", handler.AddHoliday(db))
	admin.PUT("/holiday/:id", handler.UpdateHoliday(db))
	admin.DELETE("/holiday/:id", handler.DeleteHoliday(db))
	admin.POST("/holidays/import-national", handler.ImportNationalHolidays(db))

	admin.GET("/point-rules", handler.GetPointRules(db))
	admin.POST("/point-rules/add", handler.AddPointRule(db))
	admin.DELETE("/point-rules/:id", handler.DeletePointRule(db))
	admin.GET("/points/student/:id", handler.GetStudentPointProfile(db))
	admin.POST("/points/transaction", handler.AddPointTransaction(db))
	admin.GET("/points/leaderboard", handler.GetLeaderboard(db))
	admin.GET("/points/history", handler.GetPointHistory(db))
	admin.DELETE("/points/log/:id", handler.DeletePointLog(db))

	// Point Management: Reset & Adjust
	admin.POST("/points/reset/:id", handler.ResetStudentPoints(db))
	admin.POST("/points/reset-bulk", handler.BulkResetPoints(db))
	admin.POST("/points/adjust", handler.AdjustStudentPoints(db))
	admin.POST("/points/adjust-bulk", handler.BulkAdjustPoints(db))

	admin.GET("/points/student-profile", func(c echo.Context) error {
		return c.Render(http.StatusOK, "student_point_profile.html", nil)
	})
	admin.GET("/point-rules-v2", handler.GetPointRulesV2(db))
	admin.POST("/point-claims", handler.SubmitPointClaim(db))
	admin.GET("/point-claims", handler.GetPointClaims(db))
	admin.POST("/point-claims/:id/approve", handler.ApprovePointClaim(db))
	admin.POST("/point-claims/:id/reject", handler.RejectPointClaim(db))
	admin.GET("/point-claims-page", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_point_claims.html", nil)
	})

	admin.GET("/point-rewards", handler.GetPointRewards(db))
	admin.POST("/point-rewards/add", handler.AddPointReward(db))
	admin.DELETE("/point-rewards/:id", handler.DeletePointReward(db))
	admin.POST("/points/redeem", handler.RedeemReward(db))
	admin.GET("/points/search-student", handler.SearchStudent(db))

	admin.GET("/qr-generate", handler.GenerateQR(db))

	admin.POST("/face/register", handler.RegisterFace(db))
	admin.POST("/face/save-descriptor", handler.SaveFaceDescriptor(db))
	admin.GET("/face/status", handler.GetFaceStatus(db))
	admin.GET("/faces", handler.ListFaces(db))
	admin.DELETE("/face/:student_id", handler.DeleteFace(db))
	admin.DELETE("/face/staff/:staff_id", handler.DeleteStaffFace(db))
	admin.POST("/face/verify", handler.VerifyFace(db))

	admin.GET("/settings/school", handler.GetSchoolSettings(db))
	admin.PUT("/settings/school", handler.UpdateSchoolSettings(db))
	admin.POST("/settings/school", handler.UpdateSchoolSettings(db))

	admin.GET("/wa-logs", handler.GetWhatsAppLogs(db))
	admin.POST("/attendance/test-wa", handler.TestWA(db))

	admin.POST("/attendance/manual", app.ManualAttendanceHandler)
	admin.POST("/attendance/staff-manual", handler.ManualStaffAttendance(db))
	admin.GET("/attendance/staff-export", handler.ExportStaffAttendance(db))
	admin.GET("/attendance/daily", handler.GetDailyAttendance(db))
	admin.GET("/attendance/sheet", handler.GetClassAttendanceSheet(db))
	admin.POST("/attendance/bulk", handler.BulkAttendance(db))
	admin.POST("/attendance/settings", handler.UpdateAttendanceSettings(db))

	admin.GET("/prayer/attendance", handler.GetPrayerAttendance(db))
	admin.POST("/prayer/attendance", handler.BulkPrayerAttendance(db))
	admin.GET("/prayer/report", handler.PrayerReport(db))

	admin.GET("/student/calendar", handler.GetStudentCalendar(db))
	admin.GET("/staff/calendar", handler.GetStaffCalendar(db))

	admin.GET("/report/daily", handler.DailyReport(db))
	admin.GET("/report/weekly", handler.WeeklyReport(db))
	admin.GET("/report/monthly", handler.MonthlyReport(db))

	// --- COMPOSITE SCORING (Phase 1) ---
	admin.GET("/scoring/attendance-index", handler.CalculateAttendanceIndex(db))

	// --- COMPOSITE SCORING (Phase 2) ---
	admin.POST("/scoring/achievement/submit", handler.SubmitAchievement(db))
	admin.POST("/scoring/achievement/approve", handler.ApproveAchievement(db))
	admin.GET("/scoring/achievement/pending", handler.GetPendingAchievements(db))
	admin.GET("/scoring/composite", handler.GetCompositeScores(db))
	admin.GET("/composite-scores/leaderboard", handler.GetCompositeScores(db))
	admin.POST("/composite-scores/recalculate-all", handler.RecalculateAllCompositeScores(db))

	// Direct API endpoints for composite scores (matching AJAX calls in points.html)
	e.GET("/api/composite-scores/leaderboard", handler.GetCompositeScores(db), appMiddleware.AdminAuth())
	e.POST("/api/composite-scores/recalculate-all", handler.RecalculateAllCompositeScores(db), appMiddleware.AdminAuth())

	// --- COMPOSITE SCORING (Phase 3) ---
	admin.POST("/scoring/violation/record", handler.RecordViolation(db))
	admin.GET("/scoring/violation/list", handler.GetStudentViolations(db))
	admin.POST("/scoring/redemption/submit", handler.SubmitRedemption(db))
	admin.POST("/scoring/redemption/verify", handler.VerifyRedemption(db))
	admin.GET("/scoring/redemption/pending", handler.GetPendingRedemptions(db))

	// --- OPERATOR ROUTES ---
	serveEmbedPage := func(path string) echo.HandlerFunc {
		return func(c echo.Context) error {
			data, err := fs.ReadFile(viewsFS, path)
			if err != nil {
				return c.String(http.StatusNotFound, "Page not found")
			}
			return c.HTMLBlob(http.StatusOK, data)
		}
	}

	e.GET("/operator/login", serveEmbedPage("views/mobile/login.html"))
	e.POST("/api/operator/login", handler.OperatorLogin(db), appMiddleware.AuthRateLimiter())
	e.POST("/api/operator/logout", handler.OperatorLogout(db))

	operatorPages := e.Group("/operator")
	operatorPages.Use(handler.OperatorAuth(db))
	operatorPages.GET("/dashboard", serveEmbedPage("views/mobile/dashboard.html"))
	operatorPages.GET("/scan", serveEmbedPage("views/mobile/scan.html"))
	operatorPages.GET("/manual", serveEmbedPage("views/mobile/manual.html"))
	operatorPages.GET("/profile", serveEmbedPage("views/mobile/profile.html"))

	operatorAPI := e.Group("/api/operator")
	operatorAPI.Use(handler.OperatorAuth(db))
	operatorAPI.GET("/prayer-stats", handler.GetOperatorPrayerStats(db))
	operatorAPI.POST("/scan-qr", handler.ScanQR(db))
	operatorAPI.GET("/classes", handler.GetClasses(db))
	operatorAPI.GET("/recent-logs", handler.GetRecentPrayerLogs(db))
	operatorAPI.GET("/students", handler.GetPrayerAttendance(db))
	operatorAPI.POST("/prayer-attendance", handler.BulkPrayerAttendance(db))
	operatorAPI.GET("/profile", handler.GetOperatorProfile(db))
	operatorAPI.PUT("/profile", handler.UpdateOperatorProfile(db))
	operatorAPI.PUT("/password", handler.ChangeOperatorPassword(db))

	// --- GURU / TEACHER PORTAL ROUTES ---
	serveTeacherPage := func(name string) echo.HandlerFunc {
		return serveEmbedPage("views/teacher/" + name)
	}

	e.GET("/guru/login", func(c echo.Context) error {
		if cookie, err := c.Cookie(handler.TeacherSessionCookie); err == nil && cookie.Value != "" {
			var staffID int
			if db.QueryRow("SELECT staff_id FROM teacher_sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP", cookie.Value).Scan(&staffID) == nil {
				return c.Redirect(http.StatusSeeOther, "/guru/dashboard")
			}
		}
		return serveTeacherPage("login.html")(c)
	})
	e.POST("/api/guru/login", handler.TeacherLogin(db), appMiddleware.AuthRateLimiter())
	e.GET("/guru/logout", handler.TeacherLogout(db))
	e.POST("/guru/logout", handler.TeacherLogout(db))
	e.POST("/api/guru/logout", handler.TeacherLogout(db))

	guruPages := e.Group("/guru", handler.TeacherAuth(db))
	guruPages.GET("", func(c echo.Context) error { return c.Redirect(http.StatusSeeOther, "/guru/dashboard") })
	guruPages.GET("/dashboard", serveTeacherPage("dashboard.html"))
	guruPages.GET("/class/:id", serveTeacherPage("class_detail.html"))

	guruAPI := e.Group("/api/guru", handler.TeacherAuth(db))
	guruAPI.GET("/profile", handler.GetTeacherProfile(db))
	guruAPI.PUT("/password", handler.TeacherChangePassword(db))
	guruAPI.GET("/meetings/:assignment_id", handler.GetTeacherMeetings(db))
	guruAPI.POST("/meetings", handler.CreateTeacherMeeting(db))
	guruAPI.GET("/journals", handler.GetTeacherJournals(db))
	guruAPI.POST("/journals", handler.CreateTeacherJournal(db))
	guruAPI.PUT("/journals/:id", handler.UpdateTeacherJournal(db))
	guruAPI.POST("/journals/:id/submit", handler.SubmitTeacherJournal(db))
	guruAPI.GET("/calendar", handler.GetAcademicCalendar(db))
	guruAPI.GET("/meetings/:meeting_id/attendance", handler.GetMeetingAttendance(db))
	guruAPI.GET("/meetings/:meeting_id/discussions", handler.ListLMSDiscussions(db))
	guruAPI.POST("/discussions", handler.CreateLMSDiscussion(db))
	guruAPI.POST("/discussions/:id/comments", handler.CreateLMSDiscussionComment(db))
	guruAPI.POST("/discussions/:id/toggle", handler.ToggleLMSDiscussion(db))
	guruAPI.POST("/discussions/reactions", handler.ToggleLMSReaction(db))
	guruAPI.DELETE("/discussions/comments/:id", handler.DeleteLMSDiscussionComment(db))
	guruAPI.GET("/notifications", handler.GetTeacherNotifications(db))
	guruAPI.POST("/notifications/read-all", handler.MarkAllTeacherNotificationsRead(db))
	guruAPI.POST("/notifications/:id/read", handler.MarkTeacherNotificationRead(db))
	guruAPI.GET("/agenda", handler.GetUnifiedCalendar(db))
	guruAPI.GET("/question-bank", handler.GetQuestionBank(db))
	guruAPI.POST("/question-bank", handler.CreateBankQuestion(db))
	guruAPI.DELETE("/question-bank/:id", handler.DeleteBankQuestion(db))
	guruAPI.POST("/quizzes/from-bank", handler.AddBankQuestionToQuiz(db))
	guruAPI.GET("/assignments/:id/export-excel", handler.ExportAssignmentGradesExcel(db))
	guruAPI.GET("/meetings/:meeting_id/attendance/export-excel", handler.ExportMeetingAttendanceExcel(db))
	guruAPI.POST("/meetings/:meeting_id/attendance", handler.SaveMeetingAttendance(db))
	guruAPI.POST("/materials", handler.AddMeetingMaterial(db, storageService))
	guruAPI.DELETE("/materials/:id", handler.DeleteMeetingMaterial(db, storageService))
	guruAPI.POST("/assignments", handler.CreateMeetingAssignment(db))
	guruAPI.GET("/assignments/:id/submissions", handler.GetAssignmentSubmissions(db))
	guruAPI.POST("/submissions/grade", handler.GradeSubmission(db))
	guruAPI.POST("/quizzes", handler.CreateMeetingQuiz(db))
	guruAPI.GET("/quizzes/:id/questions", handler.GetQuizQuestions(db))
	guruAPI.POST("/quizzes/questions", handler.AddQuizQuestion(db))
	guruAPI.DELETE("/quizzes/questions/:id", handler.DeleteQuizQuestion(db))
	guruAPI.DELETE("/quizzes/:id", handler.DeleteMeetingQuiz(db))
	guruAPI.GET("/quizzes/:id/attempts", handler.GetQuizAttempts(db))

	// --- STUDENT PORTAL ROUTES ---
	serveStudentPage := func(name string) echo.HandlerFunc {
		return serveEmbedPage("views/student/" + name)
	}

	e.GET("/student/login", func(c echo.Context) error {
		if cookie, err := c.Cookie(handler.StudentSessionCookie); err == nil && cookie.Value != "" {
			var sid string
			if db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = ?",
				"student_session_"+cookie.Value).Scan(&sid) == nil {
				return c.Redirect(http.StatusSeeOther, "/student/app")
			}
		}
		return serveStudentPage("login.html")(c)
	})
	e.POST("/api/student/login", handler.StudentLogin(db), appMiddleware.AuthRateLimiter())
	e.POST("/api/student/logout", handler.StudentLogout(db))

	// --- REVERSE OTP & MAGIC LINK AUTHENTICATION ROUTES ---
	e.POST("/api/auth/magic-request", handler.RequestMagicLogin(db), appMiddleware.AuthRateLimiter())
	e.GET("/api/auth/magic-status", handler.CheckMagicStatus(db), appMiddleware.AuthRateLimiter())
	e.GET("/auth/magic", handler.ConsumeMagicLink(db), appMiddleware.AuthRateLimiter())
	e.POST("/api/webhook/whatsapp", handler.HandleWhatsAppWebhook(db), appMiddleware.AuthRateLimiter())
	e.POST("/api/webhook/onesender", handler.HandleWhatsAppWebhook(db), appMiddleware.AuthRateLimiter())

	// Parent Portal APIs
	e.GET("/api/parent/info", handler.GetParentInfo(db))
	e.POST("/api/parent/switch-child", handler.SwitchActiveChild(db))

	// Weather API (public, proxied from backend to hide API key)
	e.GET("/api/weather/current", handler.WeatherCurrent())
	e.GET("/api/weather/forecast", handler.WeatherForecast())

	studentPages := e.Group("/student")
	studentPages.Use(handler.StudentAuth(db))
	studentPages.GET("", func(c echo.Context) error { return c.Redirect(http.StatusSeeOther, "/student/app") })
	studentPages.GET("/app", serveStudentPage("app.html"))
	studentPages.GET("/dashboard", serveStudentPage("app.html"))
	studentPages.GET("/presensi", serveStudentPage("app.html"))
	studentPages.GET("/qr", serveStudentPage("app.html"))
	studentPages.GET("/profil", serveStudentPage("app.html"))
	studentPages.GET("/belajar", serveStudentPage("app.html"))
	studentPages.GET("/lms", serveStudentPage("app.html"))

	studentAPI := e.Group("/api/student")
	studentAPI.Use(handler.StudentAuth(db))
	studentAPI.GET("/dashboard", handler.GetStudentDashboard(db))
	studentAPI.GET("/profile", handler.GetStudentPortalProfile(db))
	studentAPI.PUT("/profile", handler.UpdateStudentPortalProfile(db))
	studentAPI.POST("/photo", handler.UploadStudentPhoto(db))
	studentAPI.PUT("/pin", handler.ChangeStudentPIN(db))
	studentAPI.GET("/qrcard", handler.GetMyQRCard(db))
	studentAPI.GET("/calendar", handler.GetMyCalendar(db))
	studentAPI.GET("/points", handler.GetMyPoints(db))
	studentAPI.GET("/claim-rules", handler.GetStudentClaimRules(db))
	studentAPI.GET("/claims", handler.GetMyClaimsHistory(db))
	studentAPI.POST("/claims", handler.SubmitStudentClaim(db))
	studentAPI.POST("/announcement/:id/read", handler.MarkAnnouncementRead(db))
	studentAPI.POST("/announcements/read-all", handler.MarkAllAnnouncementsRead(db))
	studentAPI.GET("/doa", handler.GetDailyPrayers())

	// LMS Student Courses & Assignments
	studentAPI.GET("/courses", handler.GetStudentCourses(db))
	studentAPI.GET("/courses/:assignment_id", handler.GetStudentCourseDetail(db))
	studentAPI.GET("/meetings/:meeting_id/discussions", handler.ListLMSDiscussions(db))
	studentAPI.POST("/discussions", handler.CreateLMSDiscussion(db))
	studentAPI.POST("/discussions/:id/comments", handler.CreateLMSDiscussionComment(db))
	studentAPI.POST("/discussions/reactions", handler.ToggleLMSReaction(db))
	studentAPI.DELETE("/discussions/comments/:id", handler.DeleteLMSDiscussionComment(db))
	studentAPI.GET("/notifications", handler.GetStudentNotifications(db))
	studentAPI.POST("/notifications/read-all", handler.MarkAllStudentNotificationsRead(db))
	studentAPI.POST("/notifications/:id/read", handler.MarkStudentNotificationRead(db))
	studentAPI.GET("/agenda", handler.GetStudentAgenda(db))
	studentAPI.POST("/assignments/submit", handler.SubmitStudentAssignment(db, storageService))
	studentAPI.GET("/quizzes/:id", handler.GetStudentQuiz(db))
	studentAPI.POST("/quizzes/submit", handler.SubmitStudentQuiz(db))

	// --- ENGLISH DAILY QUEST ROUTES ---
	studentPages.GET("/english", serveStudentPage("app.html"))

	studentAPI.GET("/english/quest", handler.GetMyEnglishQuest(db))
	studentAPI.POST("/english/submit", handler.SubmitEnglishQuest(db))
	studentAPI.GET("/english/profile", handler.GetMyEnglishProfile(db))
	studentAPI.GET("/english/leaderboard", handler.GetEnglishLeaderboard(db))

	admin.GET("/english/quest", handler.GetTodayEnglishQuest(db))
	admin.POST("/english/quest", handler.CreateEnglishQuest(db))
	admin.GET("/english/submissions", handler.GetEnglishSubmissions(db))
	admin.POST("/english/submissions/:id/review", handler.ReviewEnglishSubmission(db))
	admin.GET("/english/leaderboard", handler.GetEnglishLeaderboard(db))
	admin.GET("/english/progress", handler.GetEnglishProgressReport(db))

	// AI Settings (Both /admin/ai and /api/ai aliases)
	admin.GET("/ai/settings", handler.GetAISettings(db))
	admin.POST("/ai/settings", handler.UpdateAISettings(db))
	admin.GET("/ai/test", handler.TestAIConnection(db))
	admin.POST("/ai/test", handler.TestAIConnection(db))
	admin.GET("/ai/models", handler.GetAIModels(db))
	admin.POST("/ai/models", handler.GetAIModels(db))
	admin.POST("/ai/generate-quest", handler.GenerateAIQuest(db))

	aiAPI := e.Group("/api/ai", appMiddleware.AdminAuth())
	aiAPI.GET("/settings", handler.GetAISettings(db))
	aiAPI.POST("/settings", handler.UpdateAISettings(db))
	aiAPI.GET("/test", handler.TestAIConnection(db))
	aiAPI.POST("/test", handler.TestAIConnection(db))
	aiAPI.GET("/models", handler.GetAIModels(db))
	aiAPI.POST("/models", handler.GetAIModels(db))
	aiAPI.POST("/generate-quest", handler.GenerateAIQuest(db))

	// Backup & Restore
	admin.GET("/backup", handler.BackupDatabase(db))
	admin.POST("/restore", handler.RestoreDatabase(db))

	// === DUAL-TRACK POINT SYSTEM V2 ===
	admin.GET("/v2/achievement-rules", handler.GetAchievementRules(db))
	admin.GET("/v2/violation-rules", handler.GetViolationRules(db))
	admin.POST("/v2/achievement-rule/add", handler.AddAchievementRule(db))
	admin.POST("/v2/achievement-rule/update/:id", handler.UpdateAchievementRule(db))
	admin.DELETE("/v2/achievement-rule/:id", handler.DeleteAchievementRule(db))
	admin.POST("/v2/achievement-rule/toggle/:id", handler.ToggleAchievementRuleStatus(db))
	admin.POST("/v2/violation-rule/add", handler.AddViolationRule(db))
	admin.POST("/v2/violation-rule/update/:id", handler.UpdateViolationRule(db))
	admin.DELETE("/v2/violation-rule/:id", handler.DeleteViolationRule(db))
	admin.POST("/v2/violation-rule/toggle/:id", handler.ToggleViolationRuleStatus(db))
	admin.GET("/v2/achievement-rules/export", handler.ExportAchievementRules(db))
	admin.POST("/v2/achievement-rules/import", handler.ImportAchievementRules(db))
	admin.GET("/v2/violation-rules/export", handler.ExportViolationRules(db))
	admin.POST("/v2/violation-rules/import", handler.ImportViolationRules(db))
	admin.GET("/v2/student/:id/points", handler.GetStudentDualPointProfile(db))
	admin.GET("/v2/student-detail", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			id = c.Param("id")
		}
		c.SetParamNames("id")
		c.SetParamValues(id)
		return handler.GetStudentDualPointProfile(db)(c)
	})
	admin.POST("/v2/student/achievement", handler.AddAchievementPoint(db))
	admin.POST("/v2/student/violation", handler.AddViolationPoint(db))
	admin.DELETE("/v2/achievement/:id", handler.DeleteAchievementPoint(db))
	admin.DELETE("/v2/violation/:id", handler.DeleteViolationPoint(db))
	admin.GET("/v2/leaderboard", handler.GetDualPointLeaderboard(db))
	admin.GET("/v2/class-summary", handler.GetClassDualPointSummary(db))
	admin.GET("/v2/recent-activities", handler.GetDualPointRecentActivities(db))

	// Dual-Track Point System Pages
	admin.GET("/dual-point/dashboard", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_dashboard.html", nil)
	})
	admin.GET("/dual-point/input", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_input.html", nil)
	})
	admin.GET("/dual-point/leaderboard", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_leaderboard.html", nil)
	})
	admin.GET("/dual-point/reports", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_reports.html", nil)
	})
	admin.GET("/dual-point/student/:id", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_student_detail.html", nil)
	})
	admin.GET("/dual-point/rules", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin_dual_point_rules.html", nil)
	})

	// === SARPRAS (ASSET MANAGEMENT) ROUTES ===
	admin.GET("/sarpras", handler.SarprasPage(db))
	sarpras := admin.Group("/sarpras")

	// Dashboard & Integrated shadcn Pages
	sarpras.GET("", handler.SarprasPage(db))
	sarpras.GET("/dashboard", handler.SarprasPage(db))
	sarpras.GET("/categories", handler.SarprasPage(db))
	sarpras.GET("/funding", handler.SarprasPage(db))
	sarpras.GET("/locations", handler.SarprasPage(db))
	sarpras.GET("/assets", handler.SarprasPage(db))
	sarpras.GET("/asset/add", handler.SarprasPage(db))
	sarpras.GET("/asset/:id", handler.SarprasPage(db))
	sarpras.GET("/borrowings", handler.SarprasPage(db))
	sarpras.GET("/maintenance", handler.SarprasPage(db))

	// API Endpoints
	sarprasAPI := sarpras.Group("/api")

	// Asset Categories API
	sarprasAPI.GET("/categories", handler.GetAssetCategories(db))
	sarprasAPI.POST("/category/add", handler.AddAssetCategory(db))
	sarprasAPI.POST("/category/update/:id", handler.UpdateAssetCategory(db))
	sarprasAPI.DELETE("/category/:id", handler.DeleteAssetCategory(db))

	// Asset Funding Sources
	sarprasAPI.GET("/funding-sources", handler.GetAssetFundingSources(db))
	sarprasAPI.POST("/funding-source/add", handler.AddAssetFundingSource(db))
	sarprasAPI.POST("/funding-source/update/:id", handler.UpdateAssetFundingSource(db))
	sarprasAPI.DELETE("/funding-source/:id", handler.DeleteAssetFundingSource(db))

	// Asset Locations
	sarprasAPI.GET("/locations", handler.GetAssetLocations(db))
	sarprasAPI.POST("/location/add", handler.AddAssetLocation(db))
	sarprasAPI.POST("/location/update/:id", handler.UpdateAssetLocation(db))
	sarprasAPI.DELETE("/location/:id", handler.DeleteAssetLocation(db))

	// Assets Management
	sarprasAPI.GET("/assets", handler.GetAssets(db))
	sarprasAPI.GET("/asset/:id", handler.GetAssetByID(db))
	sarprasAPI.POST("/asset/add", handler.AddAsset(db))
	sarprasAPI.POST("/asset/update/:id", handler.UpdateAsset(db))
	sarprasAPI.DELETE("/asset/:id", handler.DeleteAsset(db))
	sarprasAPI.GET("/asset/:id/qr", handler.GenerateAssetQRCode(db))
	sarprasAPI.GET("/stats", handler.GetAssetStats(db))
	sarprasAPI.GET("/asset/export-excel", handler.ExportAssetsExcel(db))
	sarprasAPI.GET("/asset/template-excel", handler.DownloadAssetTemplateExcel(db))
	sarprasAPI.POST("/asset/import-excel", handler.ImportAssetsExcel(db))

	// Asset Borrowings
	sarprasAPI.GET("/borrowings", handler.GetAssetBorrowings(db))
	sarprasAPI.POST("/borrowing/create", handler.CreateAssetBorrowing(db))
	sarprasAPI.POST("/borrowing/:id/approve", handler.ApproveAssetBorrowing(db))
	sarprasAPI.POST("/borrowing/:id/reject", handler.RejectAssetBorrowing(db))
	sarprasAPI.POST("/borrowing/:id/return", handler.ReturnAssetBorrowing(db))

	// Maintenance Tickets
	sarprasAPI.GET("/tickets", handler.GetMaintenanceTickets(db))
	sarprasAPI.POST("/ticket/create", handler.CreateMaintenanceTicket(db))
	sarprasAPI.POST("/ticket/:id/update", handler.UpdateMaintenanceTicketStatus(db))
	sarprasAPI.DELETE("/ticket/:id", handler.DeleteMaintenanceTicket(db))

	// Public scan endpoint (for QR code scanning)
	e.GET("/sarpras/scan", handler.ScanAssetQRCode(db), handler.DeviceAuth(db))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
