package router

import (
	"database/sql"
	"embed"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"belsekolah/internal/config"
	"belsekolah/internal/handler"
	appMiddleware "belsekolah/internal/middleware"
)

type templateRenderer struct {
	templates *template.Template
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// AppHandlers contains handlers that are still methods on the main App struct.
// Once those are extracted, this interface can be removed.
type AppHandlers interface {
	DashboardHandler(c echo.Context) error
	ManualAttendanceHandler(c echo.Context) error
	RecordAttendanceHandler(c echo.Context) error
	PrayerAttendanceHandler(c echo.Context) error
	TodayStatsHandler(c echo.Context) error
	RecentLogsHandler(c echo.Context) error
	VerifyFaceAttendanceHandler(c echo.Context) error
	SyncHandler(c echo.Context) error
	PublicLeaderboardHandler(c echo.Context) error
}

// Register sets up all routes on the given Echo instance.
func Register(e *echo.Echo, db *sql.DB, viewsFS embed.FS, app AppHandlers) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
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
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Renderer = &templateRenderer{
		templates: template.Must(template.ParseFS(viewsFS, "views/*.html")),
	}
	e.Static("/assets", "public/assets")

	// --- PUBLIC ROUTES ---
	e.GET("/", func(c echo.Context) error { return c.Render(http.StatusOK, "index.html", nil) })
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	}, appMiddleware.AlreadyLoggedIn())
	e.GET("/scan", handler.ScanPage(db))
	e.GET("/scan-face", handler.ScanFacePage(db))
	e.GET("/scan-sholat", handler.ScanPrayerPage(db))

	e.POST("/api/login", handler.Login())
	e.POST("/api/logout", handler.Logout())
	e.GET("/api/sync", app.SyncHandler)
	e.GET("/api/leaderboard", app.PublicLeaderboardHandler)
	e.GET("/api/point-rules", handler.GetPointRulesV2(db))
	e.GET("/api/attendance/prayer-logs", handler.PrayerLogs(db))
	e.GET("/api/attendance/record", app.RecordAttendanceHandler)
	e.GET("/api/attendance/prayer", app.PrayerAttendanceHandler)
	e.GET("/api/attendance/today-stats", app.TodayStatsHandler)
	e.GET("/api/attendance/recent-logs", app.RecentLogsHandler)
	e.POST("/api/attendance/verify-face", app.VerifyFaceAttendanceHandler)
	e.POST("/api/attendance/test-wa", handler.TestWA(db))

	// --- ADMIN ROUTES ---
	admin := e.Group("/admin", appMiddleware.AdminAuth())
	admin.GET("", app.DashboardHandler)

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

	// Recent Activity
	admin.GET("/recent-activity", handler.GetRecentActivity(db))

	admin.POST("/staff/add", handler.AddStaff(db))
	admin.POST("/staff/update/:id", handler.UpdateStaff(db))
	admin.DELETE("/staff/:id", handler.DeleteStaff(db))
	admin.POST("/staff/import", handler.ImportStaff(db))

	e.GET("/admin/announcements", handler.GetAnnouncements(db))
	e.POST("/admin/announcement/add", handler.CreateAnnouncement(db))
	e.DELETE("/admin/announcement/:id", handler.DeleteAnnouncement(db))
	e.POST("/admin/announcement/play/:id", handler.PlayAnnouncement(db))

	e.GET("/admin/holidays", handler.GetHolidays(db))
	e.POST("/admin/holiday/add", handler.AddHoliday(db))
	e.PUT("/admin/holiday/:id", handler.UpdateHoliday(db))
	e.DELETE("/admin/holiday/:id", handler.DeleteHoliday(db))
	e.POST("/admin/holidays/import-national", handler.ImportNationalHolidays(db))

	admin.GET("/point-rules", handler.GetPointRules(db))
	admin.POST("/point-rules/add", handler.AddPointRule(db))
	admin.DELETE("/point-rules/:id", handler.DeletePointRule(db))
	admin.GET("/points/student/:id", handler.GetStudentPointProfile(db))
	admin.POST("/points/transaction", handler.AddPointTransaction(db))
	admin.GET("/points/leaderboard", handler.GetLeaderboard(db))
	admin.GET("/points/history", handler.GetPointHistory(db))
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
	admin.GET("/face/status", handler.GetFaceStatus(db))
	admin.GET("/faces", handler.ListFaces(db))
	admin.DELETE("/face/:student_id", handler.DeleteFace(db))
	admin.POST("/face/verify", handler.VerifyFace(db))

	admin.GET("/settings/school", handler.GetSchoolSettings(db))
	admin.PUT("/settings/school", handler.UpdateSchoolSettings(db))

	admin.GET("/wa-logs", handler.GetWhatsAppLogs(db))

	admin.POST("/attendance/manual", app.ManualAttendanceHandler)
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

	// --- OPERATOR ROUTES ---
	serveEmbedPage := func(path string) echo.HandlerFunc {
		return func(c echo.Context) error {
			data, err := viewsFS.ReadFile(path)
			if err != nil {
				return c.String(http.StatusNotFound, "Page not found")
			}
			return c.HTMLBlob(http.StatusOK, data)
		}
	}

	e.GET("/operator/login", serveEmbedPage("views/mobile/login.html"))
	e.POST("/api/operator/login", handler.OperatorLogin(db))
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
	e.POST("/api/student/login", handler.StudentLogin(db))
	e.POST("/api/student/logout", handler.StudentLogout(db))

	studentPages := e.Group("/student")
	studentPages.Use(handler.StudentAuth(db))
	studentPages.GET("", func(c echo.Context) error { return c.Redirect(http.StatusSeeOther, "/student/app") })
	studentPages.GET("/app", serveStudentPage("app.html"))
	studentPages.GET("/dashboard", serveStudentPage("app.html"))
	studentPages.GET("/presensi", serveStudentPage("app.html"))
	studentPages.GET("/qr", serveStudentPage("app.html"))
	studentPages.GET("/profil", serveStudentPage("app.html"))

	studentAPI := e.Group("/api/student")
	studentAPI.Use(handler.StudentAuth(db))
	studentAPI.GET("/dashboard", handler.GetStudentDashboard(db))
	studentAPI.GET("/profile", handler.GetStudentPortalProfile(db))
	studentAPI.PUT("/pin", handler.ChangeStudentPIN(db))
	studentAPI.GET("/qrcard", handler.GetMyQRCard(db))
	studentAPI.GET("/calendar", handler.GetMyCalendar(db))
	studentAPI.GET("/points", handler.GetMyPoints(db))

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

	// AI Settings
	admin.GET("/ai/settings", handler.GetAISettings(db))
	admin.POST("/ai/settings", handler.UpdateAISettings(db))
	admin.POST("/ai/test", handler.TestAIConnection(db))
	admin.POST("/ai/generate-quest", handler.GenerateAIQuest(db))

	// Port config
	port := getEnv("PORT", config.GetServerPort())
	host := config.GetServerHost()
	e.Logger.Printf("Starting SMK NIBA Super Apps server on %s:%s", host, port)
	e.Logger.Fatal(e.Start(host + ":" + port))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
