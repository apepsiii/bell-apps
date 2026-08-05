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
	e.Use(appMiddleware.SecurityHeaders())

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

	e.POST("/api/login", handler.Login(), appMiddleware.AuthRateLimiter())
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
	admin := e.Group("/admin", appMiddleware.AdminAuth(), appMiddleware.AdminRateLimiter())
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

	admin.GET("/holidays", handler.GetHolidays(db))
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
	admin.POST("/v2/student/achievement", handler.AddAchievementPoint(db))
	admin.POST("/v2/student/violation", handler.AddViolationPoint(db))
	admin.DELETE("/v2/achievement/:id", handler.DeleteAchievementPoint(db))
	admin.DELETE("/v2/violation/:id", handler.DeleteViolationPoint(db))
	admin.GET("/v2/leaderboard", handler.GetDualPointLeaderboard(db))
	admin.GET("/v2/class-summary", handler.GetClassDualPointSummary(db))
	
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
	sarpras := admin.Group("/sarpras")

	// Dashboard
	sarpras.GET("", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_dashboard.html", nil)
	})

	// Pages
	sarpras.GET("/categories", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_categories.html", nil)
	})
	sarpras.GET("/funding", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_funding.html", nil)
	})
	sarpras.GET("/locations", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_locations.html", nil)
	})
	sarpras.GET("/assets", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_assets.html", nil)
	})
	sarpras.GET("/asset/add", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_asset_add.html", nil)
	})
	sarpras.GET("/asset/:id", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_asset_detail.html", nil)
	})
	sarpras.GET("/borrowings", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_borrowings.html", nil)
	})
	sarpras.GET("/maintenance", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sarpras_maintenance.html", nil)
	})

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
	e.GET("/sarpras/scan", handler.ScanAssetQRCode(db))

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
