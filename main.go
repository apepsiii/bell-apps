package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"
)

// --- CONFIGURATION ---
const (
	DBPath      = "./database.db"
	UploadPath  = "public/assets/audio"
	PhotoPath   = "public/assets/photos"
	SignagePath = "public/assets/signage"
	AdminUser   = "admin"
	AdminPass   = "admin123"
	CookieName  = "session_token"
	SecretKey   = "admin-secret-key-123"
	AppVersion  = "v1.2.1"
)

//go:embed views/*.html views/mobile/*.html
var viewsFS embed.FS

//go:embed setup.sh
var setupScript string

//go:embed setup_nginx.sh
var setupNginxScript string

// --- STRUCTS & MODELS ---

type App struct {
	DB *sql.DB
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Schedule struct {
	ID        int
	Time      string
	Label     string
	AudioFile string
}

type AudioFile struct {
	ID          int
	FileName    string
	DisplayName string
}

type Device struct {
	ID        int
	Name      string
	IPAddress string
	Status    string
	LastSync  string
}

type Major struct {
	ID   int
	Name string
}

type Class struct {
	ID        int
	Name      string
	MajorID   int
	MajorName string // For display
	WAGroupID string // OneSender Group ID
}

type Student struct {
	ID          int
	RFID        string
	NIS         string
	Name        string
	ParentPhone string
	ClassID     int
	ClassName   string // For display
	Photo       string // Filename
}

type Staff struct {
	ID    int
	RFID  string
	NIP   string
	Name  string
	Phone string
	Role  string // Guru/Staff
}

type AttendanceSetting struct {
	Key   string
	Value string
}

type AttendanceLog struct {
	ID        int
	RFID      string
	UserName  string
	UserType  string // student/staff
	Status    string // Check-In/Check-Out
	Method    string // RFID / MANUAL
	Timestamp string
	Date      string
	UserPhoto string // For UI display
}

type RunningText struct {
	ID       int
	Content  string
	IsActive bool
}

type SignageMedia struct {
	ID       int
	Filename string
	FileType string // image/video
	Duration int    // Seconds
	IsActive bool
}

type PrayerLog struct {
	ID         int    `json:"id"`
	RFID       string `json:"rfid_uid"`
	Name       string `json:"name"`
	ClassName  string `json:"class_name"`
	PrayerType string `json:"prayer_type"`
	Status     string `json:"status"` // Added status field
	Timestamp  string `json:"timestamp"`
	Date       string `json:"date"`
}

// Helper struct for Dashboard presentation
type StudentStatus struct {
	Student
	Status string // Hadir, Sakit, Izin, Terlambat
	Method string // RFID / MANUAL
	Time   string
}

type DashboardData struct {
	Username           string
	Schedules          []Schedule
	AudioFiles         []AudioFile
	Devices            []Device
	Majors             []Major
	Classes            []Class
	Students           []Student
	StaffList          []Staff
	AttendanceLogs     []AttendanceLog
	AttendanceSettings map[string]string

	// New Fields for Attendance Control
	PresentStudents []StudentStatus
	AbsentStudents  []Student

	// Signage Data
	RunningTexts []RunningText
	SignageMedia []SignageMedia

	// Charts Data (JSON Pre-rendered)
	ChartWeeklyClass string
	ChartStatus      string
	ChartArrival     string

	Stats struct {
		TotalSchedules int
		NextBell       string
		OnlineDevices  int
		TotalDevices   int
		TotalStudents  int
		TotalStaff     int
	}
	AppVersion string
}

// Report Data Structures
type ReportData struct {
	Title       string
	Period      string
	GeneratedAt string
	Type        string // "student" or "staff"

	// Statistics
	TotalRecords    int
	TotalPresent    int
	TotalLate       int
	TotalSick       int
	TotalPermission int
	TotalAbsent     int
	AttendanceRate  float64

	// Details
	Records []ReportRecord
}

type ReportRecord struct {
	No          int
	ID          string // NIS or NIP
	Name        string
	ClassOrRole string
	Status      string
	Time        string

	// For weekly/monthly
	PresentCount    int
	LateCount       int
	SickCount       int
	PermissionCount int
	AbsentCount     int
	AttendanceRate  float64
}

// --- DATABASE SETUP ---

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", DBPath)
	if err != nil {
		log.Fatal("Gagal membuka database:", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT,
			label TEXT,
			audio_file TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS audio_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT,
			display_name TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			ip_address TEXT,
			status TEXT,
			last_sync TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS majors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			major_id INTEGER,
			wa_group_id TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nis TEXT,
			name TEXT,
			parent_phone TEXT,
			parent_name TEXT,
			class_id INTEGER,
			photo TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS staff (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT UNIQUE,
			nip TEXT,
			name TEXT,
			phone TEXT,
			role TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS attendance_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE,
			setting_value TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS holidays (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS school_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			setting_key TEXT UNIQUE NOT NULL,
			setting_value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS whatsapp_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, target TEXT, message TEXT, status TEXT, response TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);`,
		`INSERT OR IGNORE INTO school_settings (setting_key, setting_value) VALUES ('work_days', '1,2,3,4,5');`,
		`CREATE TABLE IF NOT EXISTS attendance_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT,
			user_name TEXT,
			user_type TEXT,
			status TEXT,
			method TEXT DEFAULT 'RFID',
			timestamp DATETIME,
			date DATE
		);`,
		`CREATE TABLE IF NOT EXISTS running_texts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			is_active BOOLEAN DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS signage_media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT,
			file_type TEXT,
			duration INTEGER DEFAULT 10,
			is_active BOOLEAN DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS prayer_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rfid_uid TEXT,
			name TEXT,
			class_name TEXT,
			prayer_type TEXT,
			timestamp DATETIME,
			date DATE
		);`,
		// Student Point System Tables
		`CREATE TABLE IF NOT EXISTS point_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category TEXT, -- 'achievement' or 'violation'
			name TEXT,
			points INTEGER, -- Positive or negative
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS point_rewards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			points_cost INTEGER,
			stock INTEGER,
			description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS student_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER,
			rule_id INTEGER, -- Nullable if manual adjustment
			reward_id INTEGER, -- Nullable if not a redemption
			points_change INTEGER,
			description TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			recorded_by TEXT
		);`,
		// Operators Table for Mobile Prayer Management
		`CREATE TABLE IF NOT EXISTS operators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			phone TEXT,
			photo TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_start', '11:30');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('dzuhur_end', '13:00');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_start', '15:00');`,
		`INSERT OR IGNORE INTO attendance_settings (setting_key, setting_value) VALUES ('ashar_end', '16:00');`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal("Gagal migrasi tabel:", err)
		}
	}

	// Migration: Add 'method' column if not exists (Safe for existing DB)
	var colCount int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('attendance_logs') WHERE name='method'").Scan(&colCount)
	if colCount == 0 {
		db.Exec("ALTER TABLE attendance_logs ADD COLUMN method TEXT DEFAULT 'RFID'")
	}

	// Migration: Add 'wa_group_id' to classes
	var colCountClass int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('classes') WHERE name='wa_group_id'").Scan(&colCountClass)
	if colCountClass == 0 {
		db.Exec("ALTER TABLE classes ADD COLUMN wa_group_id TEXT DEFAULT ''")
	}

	// Migration: Add 'status' to prayer_logs
	var colCountPrayerStatus int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('prayer_logs') WHERE name='status'").Scan(&colCountPrayerStatus)
	if colCountPrayerStatus == 0 {
		db.Exec("ALTER TABLE prayer_logs ADD COLUMN status TEXT DEFAULT 'Hadir'")
	}

	// Migration: Add 'parent_name' to students
	var colCountParentName int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('students') WHERE name='parent_name'").Scan(&colCountParentName)
	if colCountParentName == 0 {
		db.Exec("ALTER TABLE students ADD COLUMN parent_name TEXT DEFAULT ''")
	}

	// Migration: Add 'recorded_by' to prayer_logs
	var colCountRecordedBy int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('prayer_logs') WHERE name='recorded_by'").Scan(&colCountRecordedBy)
	if colCountRecordedBy == 0 {
		db.Exec("ALTER TABLE prayer_logs ADD COLUMN recorded_by TEXT DEFAULT 'RFID'")
	}

	// Seed Attendance Settings
	var countSettings int
	db.QueryRow("SELECT COUNT(*) FROM attendance_settings").Scan(&countSettings)
	if countSettings == 0 {
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "arrival_start", "06:00")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "arrival_end", "07:15")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "departure_start", "15:30")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "departure_end", "17:00")

		// WA Settings Defaults
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "onesender_api_url", "https://onesender.my.id/api/v1/messages")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "onesender_api_token", "")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_in", "Halo, Ananda {name} telah hadir di sekolah pada pukul {time}. Status: {status}.")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_late", "Halo, Ananda {name} terlambat hadir di sekolah pada pukul {time}.")
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_out", "Halo, Ananda {name} telah pulang sekolah pada pukul {time}.")

		// Staff Templates separated
		staffIn := "✅ KONFIRMASI KEDATANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nSelamat bertugas dan semoga hari Anda menyenangkan!\n\n— Sistem Presensi Sekolah —"
		staffOut := "✅ KONFIRMASI KEPULANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nTerima kasih atas dedikasi hari ini. Selamat beristirahat.\n\n— Sistem Presensi Sekolah —"

		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_staff_in", staffIn)
		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_template_staff_out", staffOut)

		db.Exec("INSERT INTO attendance_settings VALUES (?, ?)", "wa_image_link", "https://via.placeholder.com/150")
	}

	return db
}

// --- LOGIC HELPERS ---

func DateToIndo(t time.Time) string {
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	day := days[t.Weekday()]
	month := months[t.Month()]
	return fmt.Sprintf("%s, %d %s %d", day, t.Day(), month, t.Year())
}

func (a *App) SendOneSenderMessage(to, message, token, apiUrl, recipientType, imageUrl string) (string, error) {
	if to == "" || token == "" || apiUrl == "" {
		return "", nil
	}

	var payload map[string]interface{}

	if imageUrl != "" {
		payload = map[string]interface{}{
			"to":             to,
			"recipient_type": recipientType,
			"type":           "image",
			"image": map[string]string{
				"link":    imageUrl,
				"caption": message,
			},
		}
	} else {
		// Fallback to text format (assuming it follows standard structure)
		payload = map[string]interface{}{
			"to":             to,
			"recipient_type": recipientType,
			"type":           "text",
			"text": map[string]string{
				"body": message,
			},
		}
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println("WA Error (Req):", err)
		a.DB.Exec("INSERT INTO whatsapp_logs (target, message, status, response) VALUES (?, ?, ?, ?)", to, message, "failed", err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("WA Error (Do):", err)
		a.DB.Exec("INSERT INTO whatsapp_logs (target, message, status, response) VALUES (?, ?, ?, ?)", to, message, "failed", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	responseStr := string(body)
	
	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}
	
	a.DB.Exec("INSERT INTO whatsapp_logs (target, message, status, response) VALUES (?, ?, ?, ?)", to, message, status, responseStr)

	return responseStr, nil
}

// GetWhatsAppLogsHandler - Fetch logs
func (a *App) GetWhatsAppLogsHandler(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = "100" // Default limit
	}

	rows, err := a.DB.Query("SELECT id, target, message, status, response, timestamp FROM whatsapp_logs ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	type Log struct {
		ID        int    `json:"id"`
		Target    string `json:"target"`
		Message   string `json:"message"`
		Status    string `json:"status"`
		Response  string `json:"response"`
		Timestamp string `json:"timestamp"`
	}

	var logs []Log
	for rows.Next() {
		var l Log
		var resp sql.NullString
		rows.Scan(&l.ID, &l.Target, &l.Message, &l.Status, &resp, &l.Timestamp)
		l.Response = resp.String
		logs = append(logs, l)
	}

	return c.JSON(http.StatusOK, logs)
}

func FormatPhone(phone string) string {
	// 1. Hapus karakter non-angka
	reg := regexp.MustCompile(`[^0-9]`)
	clean := reg.ReplaceAllString(phone, "")

	// 2. Normalisasi awalan
	if strings.HasPrefix(clean, "08") {
		return "62" + clean[1:]
	}
	if strings.HasPrefix(clean, "8") {
		return "62" + clean
	}
	return clean
}

func (a *App) GetNextBell() string {
	now := time.Now().Format("15:04")
	var timeStr, label string
	err := a.DB.QueryRow("SELECT time, label FROM schedules WHERE time > ? ORDER BY time ASC LIMIT 1", now).Scan(&timeStr, &label)
	if err != nil {
		err = a.DB.QueryRow("SELECT time, label FROM schedules ORDER BY time ASC LIMIT 1").Scan(&timeStr, &label)
		if err != nil {
			return "Tidak ada jadwal"
		}
		return timeStr + " (Besok) - " + label
	}
	return timeStr + " - " + label
}

// --- HANDLERS ---

func (a *App) LoginHandler(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == AdminUser && password == AdminPass {
		cookie := new(http.Cookie)
		cookie.Name = CookieName
		cookie.Value = SecretKey
		cookie.Path = "/"
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/admin")
	}
	return c.Redirect(http.StatusSeeOther, "/login?error=1")
}

func (a *App) LogoutHandler(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = CookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.MaxAge = -1
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/login")
}

func (a *App) DashboardHandler(c echo.Context) error {
	rows, _ := a.DB.Query("SELECT id, time, label, audio_file FROM schedules ORDER BY time ASC")
	var schedules []Schedule
	for rows.Next() {
		var s Schedule
		rows.Scan(&s.ID, &s.Time, &s.Label, &s.AudioFile)
		schedules = append(schedules, s)
	}
	rows.Close()

	rowsAudio, _ := a.DB.Query("SELECT id, file_name, display_name FROM audio_files ORDER BY display_name ASC")
	var audios []AudioFile
	for rowsAudio.Next() {
		var af AudioFile
		rowsAudio.Scan(&af.ID, &af.FileName, &af.DisplayName)
		audios = append(audios, af)
	}
	rowsAudio.Close()

	rowsDevice, _ := a.DB.Query("SELECT id, name, ip_address, status, last_sync FROM devices")
	var devices []Device
	onlineCount := 0
	for rowsDevice.Next() {
		var d Device
		rowsDevice.Scan(&d.ID, &d.Name, &d.IPAddress, &d.Status, &d.LastSync)
		if d.Status == "online" {
			onlineCount++
		}
		devices = append(devices, d)
	}
	rowsDevice.Close()

	// --- DATA MASTER FETCH ---

	// 1. Majors
	rowsMajor, _ := a.DB.Query("SELECT id, name FROM majors ORDER BY name ASC")
	var majors []Major
	for rowsMajor.Next() {
		var m Major
		rowsMajor.Scan(&m.ID, &m.Name)
		majors = append(majors, m)
	}
	rowsMajor.Close()

	// 2. Classes (Join Majors)
	rowsClass, _ := a.DB.Query(`
		SELECT c.id, c.name, c.major_id, m.name, c.wa_group_id 
		FROM classes c 
		LEFT JOIN majors m ON c.major_id = m.id 
		ORDER BY c.name ASC`)
	var classes []Class
	for rowsClass.Next() {
		var c Class
		var majorName sql.NullString // Handle null if major deleted
		var waGroup sql.NullString
		rowsClass.Scan(&c.ID, &c.Name, &c.MajorID, &majorName, &waGroup)
		c.MajorName = majorName.String
		c.WAGroupID = waGroup.String
		classes = append(classes, c)
	}
	rowsClass.Close()

	// 3. Students (Join Classes)
	rowsStudent, _ := a.DB.Query(`
		SELECT s.id, s.rfid_uid, s.nis, s.name, s.parent_phone, s.class_id, c.name, s.photo
		FROM students s
		LEFT JOIN classes c ON s.class_id = c.id
		ORDER BY s.name ASC`)
	var students []Student
	for rowsStudent.Next() {
		var s Student
		var className sql.NullString
		var photo sql.NullString
		rowsStudent.Scan(&s.ID, &s.RFID, &s.NIS, &s.Name, &s.ParentPhone, &s.ClassID, &className, &photo)
		s.ClassName = className.String
		s.Photo = photo.String
		students = append(students, s)
	}
	rowsStudent.Close()

	// 4. Staff
	rowsStaff, _ := a.DB.Query("SELECT id, rfid_uid, nip, name, phone, role FROM staff ORDER BY name ASC")
	var staffList []Staff
	for rowsStaff.Next() {
		var st Staff
		rowsStaff.Scan(&st.ID, &st.RFID, &st.NIP, &st.Name, &st.Phone, &st.Role)
		staffList = append(staffList, st)
	}
	rowsStaff.Close()

	// 5. Attendance (Today's Logs & Settings)
	// Logs (Join Students for Photo)
	today := time.Now().Format("2006-01-02")
	rowsLog, _ := a.DB.Query(`
		SELECT a.id, a.rfid_uid, a.user_name, a.user_type, a.status, a.method, a.timestamp, a.date, s.photo
		FROM attendance_logs a
		LEFT JOIN students s ON a.rfid_uid = s.rfid_uid
		WHERE a.date = ? 
		ORDER BY a.timestamp DESC`, today)

	var logs []AttendanceLog

	// Map to track who is present (Key: RFID) -> Value: Status Details
	presentMap := make(map[string]AttendanceLog)

	for rowsLog.Next() {
		var l AttendanceLog
		var photo sql.NullString
		rowsLog.Scan(&l.ID, &l.RFID, &l.UserName, &l.UserType, &l.Status, &l.Method, &l.Timestamp, &l.Date, &photo)
		l.UserPhoto = photo.String

		if t, err := time.Parse("2006-01-02 15:04:05", l.Timestamp); err == nil {
			l.Timestamp = t.Format("15:04")
		}
		logs = append(logs, l)

		// Only track Students for the specific lists
		if l.UserType == "Siswa" {
			presentMap[l.RFID] = l
		}
	}
	rowsLog.Close()

	// Separate Students into Present/Absent Lists
	var presentList []StudentStatus
	var absentList []Student

	for _, s := range students {
		if logData, ok := presentMap[s.RFID]; ok {
			presentList = append(presentList, StudentStatus{
				Student: s,
				Status:  logData.Status,
				Method:  logData.Method,
				Time:    logData.Timestamp,
			})
		} else {
			absentList = append(absentList, s)
		}
	}

	// Settings
	rowsSet, _ := a.DB.Query("SELECT setting_key, setting_value FROM attendance_settings")
	attSettings := make(map[string]string)
	for rowsSet.Next() {
		var k, v string
		rowsSet.Scan(&k, &v)
		attSettings[k] = v
	}
	rowsSet.Close()

	// --- SIGNAGE DATA FETCH ---

	// 1. Running Texts
	rowsRT, _ := a.DB.Query("SELECT id, content, is_active FROM running_texts ORDER BY id DESC")
	var runningTexts []RunningText
	for rowsRT.Next() {
		var rt RunningText
		rowsRT.Scan(&rt.ID, &rt.Content, &rt.IsActive)
		runningTexts = append(runningTexts, rt)
	}
	rowsRT.Close()

	// 2. Signage Media
	rowsSM, _ := a.DB.Query("SELECT id, filename, file_type, duration, is_active FROM signage_media ORDER BY id DESC")
	var signageMedia []SignageMedia
	for rowsSM.Next() {
		var sm SignageMedia
		rowsSM.Scan(&sm.ID, &sm.Filename, &sm.FileType, &sm.Duration, &sm.IsActive)
		signageMedia = append(signageMedia, sm)
	}
	rowsSM.Close()

	// --- CHARTS DATA GENERATION ---

	// 1. Weekly Class Progress (Last 7 Days)
	// Output: {labels: [Date1, ...], datasets: [{label: 'ClassA', data: [10, ...]}, ...]}
	type ChartDataset struct {
		Label       string `json:"label"`
		Data        []int  `json:"data"`
		BorderColor string `json:"borderColor"`
		Fill        bool   `json:"fill"`
	}
	type ChartStruct struct {
		Labels   []string       `json:"labels"`
		Datasets []ChartDataset `json:"datasets"`
	}

	// Generate Dates (Last 7 days)
	var chartDates []string
	dateMap := make(map[string]int) // Date -> Index
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		chartDates = append(chartDates, d)
		dateMap[d] = 6 - i
	}

	// Query DB
	rowsChart1, _ := a.DB.Query(`
		SELECT c.name, a.date, COUNT(a.id)
		FROM attendance_logs a
		JOIN students s ON a.rfid_uid = s.rfid_uid
		JOIN classes c ON s.class_id = c.id
		WHERE a.user_type = 'Siswa' 
		  AND (a.status = 'Datang' OR a.status = 'Terlambat')
		  AND a.date >= ?
		GROUP BY c.name, a.date
		ORDER BY c.name`, chartDates[0])

	classDataMap := make(map[string][]int)
	for rowsChart1.Next() {
		var cName, cDate string
		var cCount int
		rowsChart1.Scan(&cName, &cDate, &cCount)

		if _, ok := classDataMap[cName]; !ok {
			classDataMap[cName] = make([]int, 7)
		}
		if idx, ok := dateMap[cDate]; ok {
			classDataMap[cName][idx] = cCount
		}
	}
	rowsChart1.Close()

	datasets1 := []ChartDataset{}
	colors := []string{"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899", "#6366f1"}
	cIdx := 0
	for cName, cCounts := range classDataMap {
		col := colors[cIdx%len(colors)]
		datasets1 = append(datasets1, ChartDataset{
			Label:       cName,
			Data:        cCounts,
			BorderColor: col,
			Fill:        false,
		})
		cIdx++
	}

	jsonChart1, _ := json.Marshal(ChartStruct{Labels: chartDates, Datasets: datasets1})

	// 2. Status Distribution (Weekly)
	// Query
	rowsChart2, _ := a.DB.Query(`
		SELECT status, COUNT(*)
		FROM attendance_logs
		WHERE user_type = 'Siswa'
		  AND date >= ?
		GROUP BY status`, chartDates[0])

	statusLabels := []string{}
	statusCounts := []int{}
	for rowsChart2.Next() {
		var sLabel string
		var sCount int
		rowsChart2.Scan(&sLabel, &sCount)
		statusLabels = append(statusLabels, sLabel)
		statusCounts = append(statusCounts, sCount)
	}
	rowsChart2.Close()

	jsonChart2, _ := json.Marshal(map[string]interface{}{
		"labels": statusLabels,
		"datasets": []map[string]interface{}{{
			"data":            statusCounts,
			"backgroundColor": []string{"#10b981", "#ef4444", "#3b82f6", "#f59e0b", "#6b7280"},
		}},
	})

	// 3. Average Arrival Time (Weekly)
	// SQLite Time calc
	rowsChart3, _ := a.DB.Query(`
		SELECT date, AVG(strftime('%H', timestamp) * 60 + strftime('%M', timestamp))
		FROM attendance_logs
		WHERE user_type = 'Siswa'
		  AND (status = 'Datang' OR status = 'Terlambat')
		  AND date >= ?
		GROUP BY date
		ORDER BY date ASC`, chartDates[0])

	avgTimeData := make([]float64, 7)
	for rowsChart3.Next() {
		var tDate string
		var tAvg float64
		rowsChart3.Scan(&tDate, &tAvg)
		if idx, ok := dateMap[tDate]; ok {
			avgTimeData[idx] = tAvg
		}
	}
	rowsChart3.Close()

	jsonChart3, _ := json.Marshal(map[string]interface{}{
		"labels": chartDates,
		"datasets": []map[string]interface{}{{
			"label":           "Rata-rata Menit (dari 00:00)",
			"data":            avgTimeData,
			"borderColor":     "#8b5cf6",
			"backgroundColor": "rgba(139, 92, 246, 0.2)",
			"fill":            true,
		}},
	})

	data := DashboardData{
		Username:           "Administrator",
		Schedules:          schedules,
		AudioFiles:         audios,
		Devices:            devices,
		Majors:             majors,
		Classes:            classes,
		Students:           students,
		StaffList:          staffList,
		AttendanceLogs:     logs,
		AttendanceSettings: attSettings,
		PresentStudents:    presentList,
		AbsentStudents:     absentList,
		RunningTexts:       runningTexts,
		SignageMedia:       signageMedia,
		ChartWeeklyClass:   string(jsonChart1),
		ChartStatus:        string(jsonChart2),
		ChartArrival:       string(jsonChart3),
		AppVersion:         AppVersion,
	}
	data.Stats.TotalSchedules = len(schedules)
	data.Stats.NextBell = a.GetNextBell()
	data.Stats.OnlineDevices = onlineCount
	data.Stats.TotalDevices = len(devices)
	data.Stats.TotalStudents = len(students)
	data.Stats.TotalStaff = len(staffList)

	return c.Render(http.StatusOK, "admin.html", data)
}

// --- SCHEDULE CRUD ---

func (a *App) AddScheduleHandler(c echo.Context) error {
	timeVal := c.FormValue("time")
	label := c.FormValue("label")
	audio := c.FormValue("audio_file")
	_, err := a.DB.Exec("INSERT INTO schedules (time, label, audio_file) VALUES (?, ?, ?)", timeVal, label, audio)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal database: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal berhasil ditambahkan"})
}

func (a *App) UpdateScheduleHandler(c echo.Context) error {
	id := c.Param("id")
	timeVal := c.FormValue("time")
	label := c.FormValue("label")
	audio := c.FormValue("audio_file")
	_, err := a.DB.Exec("UPDATE schedules SET time=?, label=?, audio_file=? WHERE id=?", timeVal, label, audio, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal berhasil diperbarui"})
}

func (a *App) DeleteScheduleHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM schedules WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jadwal dihapus"})
}

// --- AUDIO CRUD ---

func (a *App) UploadAudioHandler(c echo.Context) error {
	displayName := c.FormValue("display_name")
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
	}
	if displayName == "" {
		displayName = file.Filename
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membaca file"})
	}
	defer src.Close()

	os.MkdirAll(UploadPath, 0755)
	dstPath := filepath.Join(UploadPath, filepath.Base(file.Filename))

	dst, err := os.Create(dstPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan file"})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyalin file"})
	}

	a.DB.Exec("INSERT INTO audio_files (file_name, display_name) VALUES (?, ?)", file.Filename, displayName)
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Audio berhasil diupload"})
}

func (a *App) RenameAudioHandler(c echo.Context) error {
	id := c.Param("id")
	newName := c.FormValue("display_name")
	_, err := a.DB.Exec("UPDATE audio_files SET display_name=? WHERE id=?", newName, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Nama audio diperbarui"})
}

func (a *App) DeleteAudioHandler(c echo.Context) error {
	id := c.Param("id")

	// 1. Ambil nama file untuk dihapus dari disk
	var fileName string
	err := a.DB.QueryRow("SELECT file_name FROM audio_files WHERE id=?", id).Scan(&fileName)
	if err == nil {
		os.Remove(filepath.Join(UploadPath, fileName))
	}

	// 2. Hapus dari DB
	_, err = a.DB.Exec("DELETE FROM audio_files WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Audio dihapus"})
}

// --- DEVICE CRUD ---

func (a *App) AddDeviceHandler(c echo.Context) error {
	name := c.FormValue("name")
	ip := c.FormValue("ip_address")
	_, err := a.DB.Exec("INSERT INTO devices (name, ip_address, status, last_sync) VALUES (?, ?, 'offline', '-')", name, ip)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Perangkat ditambahkan"})
}

func (a *App) UpdateDeviceHandler(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	ip := c.FormValue("ip_address")
	_, err := a.DB.Exec("UPDATE devices SET name=?, ip_address=? WHERE id=?", name, ip, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Perangkat diperbarui"})
}

func (a *App) DeleteDeviceHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Perangkat dihapus"})
}

// --- ACADEMIC DATA CRUD ---

// 1. MAJORS (JURUSAN)
func (a *App) AddMajorHandler(c echo.Context) error {
	name := c.FormValue("name")
	_, err := a.DB.Exec("INSERT INTO majors (name) VALUES (?)", name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan ditambahkan"})
}
func (a *App) UpdateMajorHandler(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	_, err := a.DB.Exec("UPDATE majors SET name=? WHERE id=?", name, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan diperbarui"})
}
func (a *App) DeleteMajorHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM majors WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Jurusan dihapus"})
}

// 2. CLASSES (KELAS)
func (a *App) AddClassHandler(c echo.Context) error {
	name := c.FormValue("name")
	majorID := c.FormValue("major_id")
	waGroup := c.FormValue("wa_group_id")
	_, err := a.DB.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES (?, ?, ?)", name, majorID, waGroup)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas ditambahkan"})
}
func (a *App) UpdateClassHandler(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	majorID := c.FormValue("major_id")
	waGroup := c.FormValue("wa_group_id")
	_, err := a.DB.Exec("UPDATE classes SET name=?, major_id=?, wa_group_id=? WHERE id=?", name, majorID, waGroup, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas diperbarui"})
}
func (a *App) DeleteClassHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM classes WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Kelas dihapus"})
}

// 3. STUDENTS (SISWA)
func (a *App) AddStudentHandler(c echo.Context) error {
	rfid := c.FormValue("rfid_uid")
	nis := c.FormValue("nis")
	name := c.FormValue("name")
	phone := FormatPhone(c.FormValue("parent_phone")) // Format HP
	classID := c.FormValue("class_id")

	// Photo Upload
	photoFile := ""
	file, err := c.FormFile("photo")
	if err == nil {
		src, err := file.Open()
		if err == nil {
			defer src.Close()
			os.MkdirAll(PhotoPath, 0755)
			ext := filepath.Ext(file.Filename)
			newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext) // NIS_Timestamp.jpg
			dstPath := filepath.Join(PhotoPath, newFilename)
			dst, err := os.Create(dstPath)
			if err == nil {
				defer dst.Close()
				io.Copy(dst, src)
				photoFile = newFilename
			}
		}
	}

	_, err = a.DB.Exec("INSERT INTO students (rfid_uid, nis, name, parent_phone, class_id, photo) VALUES (?, ?, ?, ?, ?, ?)", rfid, nis, name, phone, classID, photoFile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal (Mungkin RFID/NIS duplikat): " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa ditambahkan"})
}
func (a *App) UpdateStudentHandler(c echo.Context) error {
	id := c.Param("id")
	rfid := c.FormValue("rfid_uid")
	nis := c.FormValue("nis")
	name := c.FormValue("name")
	phone := FormatPhone(c.FormValue("parent_phone"))
	classID := c.FormValue("class_id")

	// Handle Photo
	file, err := c.FormFile("photo")
	if err == nil {
		// New photo uploaded
		src, err := file.Open()
		if err == nil {
			defer src.Close()
			os.MkdirAll(PhotoPath, 0755)
			ext := filepath.Ext(file.Filename)
			newFilename := fmt.Sprintf("%s_%s%s", nis, time.Now().Format("20060102150405"), ext)
			dstPath := filepath.Join(PhotoPath, newFilename)
			dst, err := os.Create(dstPath)
			if err == nil {
				defer dst.Close()
				io.Copy(dst, src)

				// Update with photo
				_, err = a.DB.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=?, photo=? WHERE id=?", rfid, nis, name, phone, classID, newFilename, id)
			}
		}
	} else {
		// No new photo, keep old
		_, err = a.DB.Exec("UPDATE students SET rfid_uid=?, nis=?, name=?, parent_phone=?, class_id=? WHERE id=?", rfid, nis, name, phone, classID, id)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
}
func (a *App) DeleteStudentHandler(c echo.Context) error {
	id := c.Param("id")
	// Check if ID is not 0 (meaning parsed successfully) to avoid accidental deletions of ID 0 if any
	if id != "" && id != "0" {
		_, err := a.DB.Exec("DELETE FROM students WHERE id=?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa dihapus"})
}

type PromoteRequest struct {
	StudentIDs    []int `json:"student_ids"`
	TargetClassID int   `json:"target_class_id"`
}

func (a *App) GetStudentsJSONHandler(c echo.Context) error {
	classID := c.QueryParam("class_id")
	if classID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Class ID required"})
	}

	rows, err := a.DB.Query("SELECT id, nis, name FROM students WHERE class_id = ? ORDER BY name ASC", classID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.NIS, &s.Name); err != nil {
			continue
		}
		students = append(students, s)
	}

	return c.JSON(http.StatusOK, students)
}

func (a *App) PromoteStudentsHandler(c echo.Context) error {
	var req PromoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	if len(req.StudentIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "No students selected"})
	}

	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database transaction failed"})
	}

	query := "UPDATE students SET class_id = ? WHERE id IN ("
	args := make([]interface{}, len(req.StudentIDs)+1)
	args[0] = req.TargetClassID
	
	for i, id := range req.StudentIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i+1] = id
	}
	query += ")"

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to promote students: " + err.Error()})
	}

	tx.Commit()
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Berhasil memindahkan %d siswa", len(req.StudentIDs)),
	})
}

type BulkDeleteRequest struct {
	IDs []int `json:"ids"`
}

func (a *App) BulkDeleteStudentsHandler(c echo.Context) error {
	var req BulkDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
	}

	if len(req.IDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "No IDs provided"})
	}

	// Construct query: DELETE FROM students WHERE id IN (?, ?, ...)
	query := "DELETE FROM students WHERE id IN ("
	args := make([]interface{}, len(req.IDs))
	for i, id := range req.IDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database transaction failed"})
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete students: " + err.Error()})
	}

	tx.Commit()
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Berhasil menghapus %d siswa", len(req.IDs)),
	})
}

// --- IMPORT HANDLERS ---

func (a *App) ImportStudentHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Format CSV salah"})
	}

	// Pre-load Classes for Lookup (Name -> ID)
	classMap := make(map[string]int)
	rows, _ := a.DB.Query("SELECT id, name FROM classes")
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		classMap[strings.ToUpper(name)] = id
	}
	rows.Close()

	tx, _ := a.DB.Begin()
	successCount := 0

	// Skip header (row 0)
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		} // Minimal 5 kolom: NIS, Nama, Kelas, HP, RFID

		// Format CSV: NIS, Nama, Kelas (Nama/ID), HP, RFID
		nis := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		classRaw := strings.TrimSpace(row[2])
		phone := FormatPhone(row[3])
		rfid := strings.TrimSpace(row[4])

		// Determine Class ID
		var classID int
		if id, ok := classMap[strings.ToUpper(classRaw)]; ok {
			classID = id
		} else {
			// If not found by name, try parsing as ID just in case
			fmt.Sscanf(classRaw, "%d", &classID)
		}

		// Insert (IGNORE duplicates to prevent failure of entire batch)
		_, err := tx.Exec("INSERT OR IGNORE INTO students (rfid_uid, nis, name, parent_phone, class_id) VALUES (?, ?, ?, ?, ?)", rfid, nis, name, phone, classID)
		if err == nil {
			successCount++
		}
	}
	tx.Commit()

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import berhasil (%d data)", successCount)})
}

func (a *App) ImportStaffHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Format CSV salah"})
	}

	tx, _ := a.DB.Begin()
	successCount := 0

	// Skip header
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 5 {
			continue
		} // NIP, Nama, Role, HP, RFID

		nip := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		role := strings.TrimSpace(row[2])
		phone := FormatPhone(row[3])
		rfid := strings.TrimSpace(row[4])

		_, err := tx.Exec("INSERT OR IGNORE INTO staff (rfid_uid, nip, name, phone, role) VALUES (?, ?, ?, ?, ?)", rfid, nip, name, phone, role)
		if err == nil {
			successCount++
		}
	}
	tx.Commit()

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import berhasil (%d data)", successCount)})
}

func (a *App) ManualAttendanceHandler(c echo.Context) error {
	studentID := c.FormValue("student_id")
	status := c.FormValue("status") // Hadir, Sakit, Izin, Alpha

	// 1. Get Student Data
	var rfid, name string
	err := a.DB.QueryRow("SELECT rfid_uid, name FROM students WHERE id=?", studentID).Scan(&rfid, &name)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}

	// 2. Prepare Data
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05")
	dateStr := now.Format("2006-01-02")

	// 3. Insert Log
	_, err = a.DB.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, 'Siswa', ?, 'MANUAL', ?, ?)",
		rfid, name, status, timestamp, dateStr)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	// --- NOTIFIKASI WA (Async) ---
	// Ambil Settings
	var waURL, waToken, waTemplateIn, waTemplateLate, waTemplateOut, waTemplateStaff, waImage string

	// Better to load all at once or query smartly, but for now individual queries to ensure data freshness
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='onesender_api_url'").Scan(&waURL)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='onesender_api_token'").Scan(&waToken)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_template_in'").Scan(&waTemplateIn)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_template_late'").Scan(&waTemplateLate)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_template_out'").Scan(&waTemplateOut)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_template_staff'").Scan(&waTemplateStaff)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_image_link'").Scan(&waImage)

	// Determine Template based on Status/Type
	// Note: Manual handler hardcodes 'Siswa', logic here needs to be generic if staff manual added, but currently only students in manual list
	var msgTemplate string
	if status == "Datang" {
		msgTemplate = waTemplateIn
	}
	if status == "Terlambat" {
		msgTemplate = waTemplateLate
	}
	if status == "Pulang" {
		msgTemplate = waTemplateOut
	}
	if msgTemplate == "" {
		msgTemplate = waTemplateIn
	} // Fallback

	// Ambil Data HP Ortu & Grup Kelas
	var parentPhone, classGroup string
	a.DB.QueryRow(`
		SELECT s.parent_phone, c.wa_group_id 
		FROM students s 
		JOIN classes c ON s.class_id = c.id 
		WHERE s.rfid_uid=?`, rfid).Scan(&parentPhone, &classGroup)

	go func() {
		// Replace Template Variables
		msg := msgTemplate
		msg = strings.ReplaceAll(msg, "{name}", name)
		msg = strings.ReplaceAll(msg, "{time}", time.Now().Format("15:04"))
		msg = strings.ReplaceAll(msg, "{status}", status)

		// 1. Send to Parent
		if parentPhone != "" {
			a.SendOneSenderMessage(parentPhone, msg, waToken, waURL, "individual", waImage)
		}
		// 2. Send to Class Group
		if classGroup != "" {
			a.SendOneSenderMessage(classGroup, msg, waToken, waURL, "group", waImage)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Absensi manual berhasil disimpan"})
}

// --- JSON Import Structs ---
type JSONExport []struct {
	Type string        `json:"type"`
	Name string        `json:"name"`
	Data []StudentJSON `json:"data,omitempty"`
}

type StudentJSON struct {
	ID        string `json:"id"`
	NISN      string `json:"nisn"`
	Nama      string `json:"nama"`
	Kelas     string `json:"kelas"`
	NamaWali  string `json:"nama_wali"`
	NomorWali string `json:"nomor_wali"`
	Foto      string `json:"foto"`
}

func (a *App) ImportStudentJSONHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
	}
	clean := c.FormValue("clean") == "true"

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
	}
	defer src.Close()

	byteValue, _ := io.ReadAll(src)

	var export JSONExport
	err = json.Unmarshal(byteValue, &export)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format JSON salah"})
	}

	tx, _ := a.DB.Begin()
	successCount := 0

	if clean {
		tx.Exec("DELETE FROM students")
	}

	// Cache classes
	classMap := make(map[string]int)
	rows, _ := a.DB.Query("SELECT id, name FROM classes")
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		classMap[strings.ToUpper(name)] = id
	}
	rows.Close()

	for _, item := range export {
		if item.Type == "table" && item.Name == "siswa" {
			for _, s := range item.Data {
				nisn := strings.TrimSpace(s.NISN)
				nama := strings.TrimSpace(s.Nama)
				kelas := strings.TrimSpace(s.Kelas)
				namaWali := strings.TrimSpace(s.NamaWali)
				nomorWali := strings.TrimSpace(s.NomorWali)
				foto := strings.TrimSpace(s.Foto)

				if nisn == "" || nama == "" {
					continue
				}

				// 1. Get or Create Class
				classID, ok := classMap[strings.ToUpper(kelas)]
				if !ok {
					// Check DB inside Tx just in case created in previous loop (though map should handle unique names if updated)
					// Simpler: Just try to insert class if not mapped.
					res, err := tx.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES (?, 0, '')", kelas)
					if err == nil {
						id, _ := res.LastInsertId()
						classID = int(id)
						classMap[strings.ToUpper(kelas)] = classID
					}
				}

				// 2. Insert or Upsert Student
				query := `INSERT INTO students (rfid_uid, nis, name, parent_name, parent_phone, class_id, photo) 
                          VALUES (?, ?, ?, ?, ?, ?, ?)
                          ON CONFLICT(rfid_uid) DO UPDATE SET 
                          nis=excluded.nis, name=excluded.name, parent_name=excluded.parent_name, 
                          parent_phone=excluded.parent_phone, class_id=excluded.class_id, photo=excluded.photo`
				
				_, err := tx.Exec(query, nisn, nisn, nama, namaWali, nomorWali, classID, foto)
				if err == nil {
					successCount++
				}
			}
		}
	}

	tx.Commit()
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import JSON berhasil (%d data)", successCount)})
}

func (a *App) GetStudentCalendarHandler(c echo.Context) error {
	studentID := c.QueryParam("id")
	month := c.QueryParam("month") // 01-12
	year := c.QueryParam("year")   // 2026

	if studentID == "" || month == "" || year == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
	}

	// 1. Get Student RFID
	var rfid string
	err := a.DB.QueryRow("SELECT rfid_uid FROM students WHERE id=?", studentID).Scan(&rfid)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
	}

	// 2. Query Logs for Month
	// SQLite strftime: %Y, %m
	query := `
		SELECT status, timestamp, date 
		FROM attendance_logs 
		WHERE rfid_uid = ? 
		  AND strftime('%m', date) = ? 
		  AND strftime('%Y', date) = ?
		ORDER BY timestamp ASC`

	rows, err := a.DB.Query(query, rfid, month, year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	// 3. Process Data
	// Output: map[DayString][]Log
	type CalendarLog struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}

	calendarData := make(map[int][]CalendarLog)

	for rows.Next() {
		var status, ts, dateStr string
		rows.Scan(&status, &ts, &dateStr)

		// Parse Date to get Day
		// dateStr can be in format "2006-01-02" or "2006-01-02T15:04:05Z"
		var day int
		
		// Try parsing as full timestamp first
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			// Try parsing as date only
			t, err = time.Parse("2006-01-02", dateStr)
		}
		
		if err == nil {
			day = t.Day()
		} else {
			// Last resort: extract from string
			parts := strings.Split(dateStr, "-")
			if len(parts) >= 3 {
				// Extract just the day part (remove any time component)
				dayStr := strings.Split(parts[2], "T")[0]
				day, err = strconv.Atoi(dayStr)
				if err != nil {
					log.Printf("ERROR: Failed to parse day from '%s': %v", dateStr, err)
					continue
				}
			} else {
				log.Printf("ERROR: Invalid date format '%s'", dateStr)
				continue
			}
		}

		// Parse Timestamp to get Time using string slice to avoid timezone confusion
		timeOnly := ""
		if len(ts) >= 16 {
			timeOnly = ts[11:16]
		} else {
			// Fallback if format is unexpected
			tTime, _ := time.Parse("2006-01-02 15:04:05", ts)
			timeOnly = tTime.Format("15:04")
		}

		calendarData[day] = append(calendarData[day], CalendarLog{
			Status: status,
			Time:   timeOnly,
		})
	}

	// Calculate Stats
	totalPresent := 0
	totalLate := 0
	totalSick := 0
	totalPermission := 0
	totalAlpha := 0

	monthInt, _ := strconv.Atoi(month)
	monthType := time.Month(monthInt)
	yearInt, _ := strconv.Atoi(year)
	
	workingDays := a.GetWorkingDaysInMonth(yearInt, monthType)

	for _, logs := range calendarData {
		for _, log := range logs {
			switch log.Status {
			case "Datang", "Hadir":
				totalPresent++
			case "Terlambat":
				totalLate++
			case "Sakit":
				totalSick++
			case "Izin":
				totalPermission++
			case "Alpha":
				totalAlpha++
			}
		}
	}

	attendanceRate := 0.0
	if workingDays > 0 {
		attendanceRate = float64(totalPresent+totalLate) / float64(workingDays) * 100
	}

	response := map[string]interface{}{
		"calendar": calendarData,
		"stats": map[string]interface{}{
			"working_days":     workingDays,
			"present":          totalPresent,
			"late":             totalLate,
			"sick":             totalSick,
			"permission":       totalPermission,
			"alpha":            totalAlpha,
			"attendance_rate":  attendanceRate,
		},
		"holidays": a.GetHolidaysForMonth(year, month),
	}

	return c.JSON(http.StatusOK, response)
}



func (a *App) GetStaffCalendarHandler(c echo.Context) error {
	staffID := c.QueryParam("id")
	month := c.QueryParam("month")
	year := c.QueryParam("year")

	if staffID == "" || month == "" || year == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Parameter tidak lengkap"})
	}

	var rfid string
	err := a.DB.QueryRow("SELECT rfid_uid FROM staff WHERE id=?", staffID).Scan(&rfid)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Staff tidak ditemukan"})
	}

	query := `
		SELECT status, timestamp, date 
		FROM attendance_logs 
		WHERE rfid_uid = ? 
		  AND strftime('%m', date) = ? 
		  AND strftime('%Y', date) = ?
		ORDER BY timestamp ASC`

	rows, err := a.DB.Query(query, rfid, month, year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	type CalendarLog struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	calendarData := make(map[int][]CalendarLog)

	for rows.Next() {
		var status, ts, dateStr string
		rows.Scan(&status, &ts, &dateStr)
		t, _ := time.Parse("2006-01-02", dateStr)

		timeOnly := ""
		if len(ts) >= 16 {
			timeOnly = ts[11:16]
		} else {
			tTime, _ := time.Parse("2006-01-02 15:04:05", ts)
			timeOnly = tTime.Format("15:04")
		}

		calendarData[t.Day()] = append(calendarData[t.Day()], CalendarLog{Status: status, Time: timeOnly})
	}

	// Calculate Stats
	totalPresent := 0
	totalLate := 0
	totalSick := 0
	totalPermission := 0
	totalAlpha := 0

	monthInt, _ := strconv.Atoi(month)
	monthType := time.Month(monthInt)
	yearInt, _ := strconv.Atoi(year)
	
	workingDays := a.GetWorkingDaysInMonth(yearInt, monthType)

	for _, logs := range calendarData {
		for _, log := range logs {
			switch log.Status {
			case "Datang":
				totalPresent++
			case "Terlambat":
				totalLate++
			case "Sakit":
				totalSick++
			case "Izin":
				totalPermission++
			case "Alpha":
				totalAlpha++
			}
		}
	}

	attendanceRate := 0.0
	if workingDays > 0 {
		attendanceRate = float64(totalPresent+totalLate) / float64(workingDays) * 100
	}

	response := map[string]interface{}{
		"calendar": calendarData,
		"stats": map[string]interface{}{
			"working_days":     workingDays,
			"present":          totalPresent,
			"late":             totalLate,
			"sick":             totalSick,
			"permission":       totalPermission,
			"alpha":            totalAlpha,
			"attendance_rate":  attendanceRate,
		},
		"holidays": a.GetHolidaysForMonth(year, month),
	}

	return c.JSON(http.StatusOK, response)
}

// --- PROFILE HANDLERS ---

type ProfileData struct {
	Type       string // "Siswa" or "Guru/Staff"
	ID         int
	Name       string
	IdentityNo string // NIS or NIP
	ExtraInfo  string // Class or Role
	Phone      string
	RFID       string
	Photo      string
}

func (a *App) StudentProfileHandler(c echo.Context) error {
	id := c.Param("id")
	var s Student
	var className sql.NullString
	var photo sql.NullString

	err := a.DB.QueryRow(`
		SELECT s.id, s.rfid_uid, s.nis, s.name, s.parent_phone, c.name, s.photo
		FROM students s
		LEFT JOIN classes c ON s.class_id = c.id
		WHERE s.id = ?`, id).Scan(&s.ID, &s.RFID, &s.NIS, &s.Name, &s.ParentPhone, &className, &photo)

	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin")
	}

	data := ProfileData{
		Type:       "Siswa",
		ID:         s.ID,
		Name:       s.Name,
		IdentityNo: s.NIS,
		ExtraInfo:  className.String,
		Phone:      s.ParentPhone,
		RFID:       s.RFID,
		Photo:      photo.String,
	}
	return c.Render(http.StatusOK, "profile.html", data)
}

func (a *App) StaffProfileHandler(c echo.Context) error {
	id := c.Param("id")
	var s Staff

	err := a.DB.QueryRow("SELECT id, rfid_uid, nip, name, phone, role FROM staff WHERE id = ?", id).Scan(&s.ID, &s.RFID, &s.NIP, &s.Name, &s.Phone, &s.Role)

	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin")
	}

	data := ProfileData{
		Type:       "Guru/Staff",
		ID:         s.ID,
		Name:       s.Name,
		IdentityNo: s.NIP,
		ExtraInfo:  s.Role,
		Phone:      s.Phone,
		RFID:       s.RFID,
	}
	return c.Render(http.StatusOK, "profile.html", data)
}

// 4. STAFF (GURU & TENDIK)
func (a *App) AddStaffHandler(c echo.Context) error {
	rfid := c.FormValue("rfid_uid")
	nip := c.FormValue("nip")
	name := c.FormValue("name")
	phone := FormatPhone(c.FormValue("phone"))
	role := c.FormValue("role")

	_, err := a.DB.Exec("INSERT INTO staff (rfid_uid, nip, name, phone, role) VALUES (?, ?, ?, ?, ?)", rfid, nip, name, phone, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staff ditambahkan"})
}
func (a *App) UpdateStaffHandler(c echo.Context) error {
	id := c.Param("id")
	rfid := c.FormValue("rfid_uid")
	nip := c.FormValue("nip")
	name := c.FormValue("name")
	phone := FormatPhone(c.FormValue("phone"))
	role := c.FormValue("role")

	_, err := a.DB.Exec("UPDATE staff SET rfid_uid=?, nip=?, name=?, phone=?, role=? WHERE id=?", rfid, nip, name, phone, role, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data staff diperbarui"})
}
func (a *App) DeleteStaffHandler(c echo.Context) error {
	id := c.Param("id")
	_, err := a.DB.Exec("DELETE FROM staff WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Staff dihapus"})
}

// --- ATTENDANCE HANDLERS ---

func (a *App) UpdateAttendanceSettingsHandler(c echo.Context) error {
	arrivalStart := c.FormValue("arrival_start")
	arrivalEnd := c.FormValue("arrival_end")
	departureStart := c.FormValue("departure_start")
	departureEnd := c.FormValue("departure_end")

	// WA Settings
	waURL := c.FormValue("onesender_api_url")
	waToken := c.FormValue("onesender_api_token")
	waTemplateIn := c.FormValue("wa_template_in")
	waTemplateLate := c.FormValue("wa_template_late")
	waTemplateOut := c.FormValue("wa_template_out")
	waTemplateStaffIn := c.FormValue("wa_template_staff_in")
	waTemplateStaffOut := c.FormValue("wa_template_staff_out")
	waImage := c.FormValue("wa_image_link")

	// Prayer Times
	dzStart := c.FormValue("dzuhur_start")
	dzEnd := c.FormValue("dzuhur_end")
	asStart := c.FormValue("ashar_start")
	asEnd := c.FormValue("ashar_end")

	tx, _ := a.DB.Begin()
	// Gunakan INSERT OR REPLACE agar jika setting belum ada (karena migrasi), akan otomatis dibuat
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_start", arrivalStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "arrival_end", arrivalEnd)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_start", departureStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "departure_end", departureEnd)

	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_url", waURL)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "onesender_api_token", waToken)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_in", waTemplateIn)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_late", waTemplateLate)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_out", waTemplateOut)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_in", waTemplateStaffIn)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_template_staff_out", waTemplateStaffOut)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "wa_image_link", waImage)

	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_start", dzStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "dzuhur_end", dzEnd)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_start", asStart)
	tx.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)", "ashar_end", asEnd)

	tx.Commit()

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Pengaturan disimpan"})
}

func (a *App) ExportAttendanceHandler(c echo.Context) error {
	exportType := c.QueryParam("type") // daily, monthly, custom
	dateVal := c.QueryParam("date")
	monthVal := c.QueryParam("month")
	yearVal := c.QueryParam("year")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	var query string
	var args []interface{}
	filename := "attendance_export.csv"

	baseQuery := `
		SELECT a.date, a.timestamp, s.nis, s.name, c.name, a.status, a.method 
		FROM attendance_logs a
		LEFT JOIN students s ON a.rfid_uid = s.rfid_uid
		LEFT JOIN classes c ON s.class_id = c.id
		WHERE a.user_type = 'Siswa'`

	switch exportType {
	case "daily":
		query = baseQuery + " AND a.date = ? ORDER BY a.timestamp ASC"
		args = append(args, dateVal)
		filename = fmt.Sprintf("presensi_harian_%s.csv", dateVal)
	case "monthly":
		query = baseQuery + " AND strftime('%m', a.date) = ? AND strftime('%Y', a.date) = ? ORDER BY a.date ASC, a.timestamp ASC"
		args = append(args, monthVal, yearVal)
		filename = fmt.Sprintf("presensi_bulanan_%s_%s.csv", monthVal, yearVal)
	case "custom":
		query = baseQuery + " AND a.date BETWEEN ? AND ? ORDER BY a.date ASC, a.timestamp ASC"
		args = append(args, startDate, endDate)
		filename = fmt.Sprintf("presensi_custom_%s_sd_%s.csv", startDate, endDate)
	default:
		return c.String(http.StatusBadRequest, "Tipe export tidak valid")
	}

	rows, err := a.DB.Query(query, args...)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	// Set Headers
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Response().Writer)
	defer writer.Flush()

	// CSV Header
	writer.Write([]string{"Tanggal", "Waktu", "NIS", "Nama", "Kelas", "Status", "Metode"})

	for rows.Next() {
		var date, timestamp, nis, name, className, status, method sql.NullString
		rows.Scan(&date, &timestamp, &nis, &name, &className, &status, &method)

		// Clean timestamp (HH:mm)
		ts := timestamp.String
		if len(ts) >= 16 {
			ts = ts[11:16]
		}

		writer.Write([]string{
			date.String,
			ts,
			nis.String,
			name.String,
			className.String,
			status.String,
			method.String,
		})
	}

	return nil
}

func (a *App) TestWAHandler(c echo.Context) error {
	target := c.FormValue("target_phone")
	if target == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nomor tujuan kosong"})
	}

	// Get Settings from DB
	var waURL, waToken, waImage string
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='onesender_api_url'").Scan(&waURL)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='onesender_api_token'").Scan(&waToken)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='wa_image_link'").Scan(&waImage)

	if waToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Token API belum disetting"})
	}

	msg := "Test Koneksi SmartBell: Berhasil terhubung!"
	resp, err := a.SendOneSenderMessage(target, msg, waToken, waURL, "individual", waImage)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error HTTP: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Request Terkirim", "api_response": resp})
}

// API for RFID Reader (IoT Device)
// Endpoint: /api/attendance/record?rfid=...
func (a *App) RecordAttendanceHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID kosong"})
	}

	// 1. Identify User
	var name, userType, photo, identityNo, className string
	var photoNull, classNameNull sql.NullString

	// Try Student first
	err := a.DB.QueryRow(`
		SELECT s.name, s.photo, s.nis, c.name 
		FROM students s 
		LEFT JOIN classes c ON s.class_id = c.id 
		WHERE s.rfid_uid=?`, rfid).Scan(&name, &photoNull, &identityNo, &classNameNull)
	
	if err == nil {
		userType = "Siswa"
		photo = photoNull.String
		className = classNameNull.String
	} else {
		// Try Staff
		err = a.DB.QueryRow("SELECT name, nip FROM staff WHERE rfid_uid=?", rfid).Scan(&name, &identityNo)
		if err == nil {
			userType = "Staff"
			className = "" // Staff doesn't have class
		} else {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Kartu tidak dikenali"})
		}
	}

	// 2. Determine Logic (Check-In / Check-Out) based on Time
	now := time.Now()
	timeStr := now.Format("15:04")
	dateStr := now.Format("2006-01-02")
	timestamp := now.Format("2006-01-02 15:04:05")

	// Get Settings
	var arrStart, arrEnd, depStart, depEnd string
	var waURL, waToken, waTemplateIn, waTemplateLate, waTemplateOut, waTemplateStaffIn, waTemplateStaffOut, waImage string // WA Settings

	rows, _ := a.DB.Query("SELECT setting_key, setting_value FROM attendance_settings")
	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	rows.Close()

	arrStart, arrEnd = settings["arrival_start"], settings["arrival_end"]
	depStart, depEnd = settings["departure_start"], settings["departure_end"]
	waURL, waToken = settings["onesender_api_url"], settings["onesender_api_token"]
	waTemplateIn, waTemplateLate = settings["wa_template_in"], settings["wa_template_late"]
	waTemplateOut, waImage = settings["wa_template_out"], settings["wa_image_link"]
	waTemplateStaffIn, waTemplateStaffOut = settings["wa_template_staff_in"], settings["wa_template_staff_out"]

	status := "Unknown"

	// Logic Sederhana:
	// Jika jam sekarang antara Arrival Start - End => Datang (Tepat Waktu / Terlambat logic bisa dikembangkan)
	// Jika jam sekarang antara Departure Start - End => Pulang
	// Jika di luar itu, tolak atau anggap terlambat/izin

	if timeStr >= arrStart && timeStr <= arrEnd {
		status = "Datang"
		// Cek duplikasi hari ini
		var count int
		a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE rfid_uid=? AND date=? AND status='Datang'", rfid, dateStr).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen datang hari ini", "name": name})
		}
	} else if timeStr >= depStart && timeStr <= depEnd {
		status = "Pulang"
		var count int
		a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE rfid_uid=? AND date=? AND status='Pulang'", rfid, dateStr).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen pulang hari ini", "name": name})
		}
	} else {
		// Di luar jam (bisa dianggap Terlambat datang jika > arrEnd && < depStart)
		if timeStr > arrEnd && timeStr < depStart {
			status = "Terlambat"
			// Cek duplikat
			var count int
			a.DB.QueryRow("SELECT COUNT(*) FROM attendance_logs WHERE rfid_uid=? AND date=? AND (status='Datang' OR status='Terlambat')", rfid, dateStr).Scan(&count)
			if count > 0 {
				return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen hari ini", "name": name})
			}
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Di luar jam absen", "name": name})
		}
	}

	// 3. Record
	_, err = a.DB.Exec("INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (?, ?, ?, ?, 'RFID', ?, ?)",
		rfid, name, userType, status, timestamp, dateStr)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	// Send Broadcast (WA) Only for Students
	if userType == "Siswa" {
		go func() {
			// Determine Template based on Status
			msg := waTemplateIn
			if status == "Terlambat" {
				msg = waTemplateLate
			}
			if status == "Pulang" {
				msg = waTemplateOut
			}

			if msg == "" {
				msg = "Halo, {name} presensi {status} pada {time}."
			}

			// Ambil Data HP Ortu & Grup Kelas
			var parentPhone, classGroup string
			a.DB.QueryRow(`
				SELECT s.parent_phone, c.wa_group_id 
				FROM students s 
				JOIN classes c ON s.class_id = c.id 
				WHERE s.rfid_uid=?`, rfid).Scan(&parentPhone, &classGroup)

			// Format Date properly (e.g. "Senin, 02 Januari 2006")
			formattedDate := now.Format("02-01-2006") // Default fallback
			
			// Replace Template Variables
			msg = strings.ReplaceAll(msg, "{name}", name)
			msg = strings.ReplaceAll(msg, "{time}", timeStr)
			msg = strings.ReplaceAll(msg, "{status}", status)
			msg = strings.ReplaceAll(msg, "{date}", formattedDate)

			// 1. Send to Parent
			if parentPhone != "" {
				a.SendOneSenderMessage(parentPhone, msg, waToken, waURL, "individual", waImage)
			}
			// 2. Send to Class Group
			if classGroup != "" {
				a.SendOneSenderMessage(classGroup, msg, waToken, waURL, "group", waImage)
			}
		}()
	} else if userType == "Staff" {
		// Staff Broadcast
		go func() {
			var msg string
			// Select template based on status
			if status == "Pulang" {
				msg = waTemplateStaffOut
			} else {
				msg = waTemplateStaffIn
			}

			// Fallback if empty
			if msg == "" && status == "Pulang" {
				msg = "✅ KONFIRMASI KEPULANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nTerima kasih atas dedikasi hari ini. Selamat beristirahat.\n\n— Sistem Presensi Sekolah —"
			} else if msg == "" {
				msg = "✅ KONFIRMASI KEDATANGAN GURU/STAF\n\nYth. Bapak/Ibu {teacher_name},\n\nPresensi {type} Anda pada hari {date} telah berhasil dicatat sistem.\n\n🕒 Pukul: {time} WIB\n\nSelamat bertugas dan semoga hari Anda menyenangkan!\n\n— Sistem Presensi Sekolah —"
			}

			// Replacements
			msg = strings.ReplaceAll(msg, "{teacher_name}", name)
			msg = strings.ReplaceAll(msg, "{name}", name) // Fallback support
			msg = strings.ReplaceAll(msg, "{type}", status)
			msg = strings.ReplaceAll(msg, "{status}", status) // Fallback support
			msg = strings.ReplaceAll(msg, "{time}", timeStr)
			msg = strings.ReplaceAll(msg, "{date}", DateToIndo(now))

			// Get Staff Phone
			var staffPhone string
			a.DB.QueryRow("SELECT phone FROM staff WHERE rfid_uid=?", rfid).Scan(&staffPhone)

			if staffPhone != "" {
				a.SendOneSenderMessage(staffPhone, msg, waToken, waURL, "individual", waImage)
			}
		}()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "success",
		"message":     "Absen " + status + " Berhasil",
		"name":        name,
		"type":        userType,
		"time":        timeStr,
		"photo":       photo,
		"identity_no": identityNo, // NIS or NIP
		"class_name":  className,  // Class name for students
	})
}

// API for Today's Statistics (for Kiosk Display)
// Endpoint: /api/attendance/today-stats
func (a *App) TodayStatsHandler(c echo.Context) error {
	dateStr := time.Now().Format("2006-01-02")

	// Count students
	var totalStudents, presentStudents, lateStudents int
	a.DB.QueryRow("SELECT COUNT(*) FROM students").Scan(&totalStudents)
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM attendance_logs WHERE date=? AND user_type='Siswa' AND status='Datang'", dateStr).Scan(&presentStudents)
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM attendance_logs WHERE date=? AND user_type='Siswa' AND status='Terlambat'", dateStr).Scan(&lateStudents)

	// Count staff
	var totalStaff, presentStaff int
	a.DB.QueryRow("SELECT COUNT(*) FROM staff").Scan(&totalStaff)
	a.DB.QueryRow("SELECT COUNT(DISTINCT rfid_uid) FROM attendance_logs WHERE date=? AND user_type='Staff' AND (status='Datang' OR status='Terlambat')", dateStr).Scan(&presentStaff)

	// Calculate percentage
	var attendancePercentage float64
	if totalStudents > 0 {
		attendancePercentage = float64(presentStudents+lateStudents) / float64(totalStudents) * 100
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_students":        totalStudents,
		"present_students":      presentStudents,
		"late_students":         lateStudents,
		"total_staff":           totalStaff,
		"present_staff":         presentStaff,
		"attendance_percentage": attendancePercentage,
		"date":                  dateStr,
	})
}

// API for Recent Attendance Logs (for Kiosk Display)
// Endpoint: /api/attendance/recent-logs
func (a *App) RecentLogsHandler(c echo.Context) error {
	limit := 5 // Last 5 logs

	rows, err := a.DB.Query(`
		SELECT al.user_name, al.user_type, al.status, al.timestamp, s.photo
		FROM attendance_logs al
		LEFT JOIN students s ON al.rfid_uid = s.rfid_uid
		ORDER BY al.timestamp DESC
		LIMIT ?`, limit)
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var name, userType, status, timestamp string
		var photoNull sql.NullString
		rows.Scan(&name, &userType, &status, &timestamp, &photoNull)
		
		// Parse timestamp to get time only
		timeOnly := "00:00"
		if len(timestamp) >= 16 {
			timeOnly = timestamp[11:16]
		}
		
		logs = append(logs, map[string]interface{}{
			"name":      name,
			"type":      userType,
			"status":    status,
			"time":      timeOnly,
			"timestamp": timestamp,
			"photo":     photoNull.String,
		})
	}

	return c.JSON(http.StatusOK, logs)
}

func (a *App) SyncHandler(c echo.Context) error {
	rows, _ := a.DB.Query("SELECT time, label, audio_file FROM schedules ORDER BY time ASC")
	defer rows.Close()

	var schedules []map[string]string
	for rows.Next() {
		var t, l, af string
		rows.Scan(&t, &l, &af)
		schedules = append(schedules, map[string]string{
			"time": t, "label": l, "audio": af, "audio_url": "/assets/audio/" + af,
		})
	}
	return c.JSON(http.StatusOK, schedules)
}

func (a *App) ScanPageHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "scan.html", nil)
}

func (a *App) ScanPrayerPageHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "scan_prayer.html", nil)
}

// --- MAIN ---

func main() {
	if len(os.Args) > 1 && os.Args[1] == "wizard" {
		runWizard()
		return
	}

	db := InitDB()
	defer db.Close()
	
	// Create default operator account
	CreateDefaultOperator(db)
	
	app := &App{DB: db}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Recover())

	// Custom Error Handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}

		if code == http.StatusNotFound {
			// Check if it's an API request
			if strings.HasPrefix(c.Request().URL.Path, "/api") {
				c.JSON(http.StatusNotFound, map[string]string{"message": "Not Found"})
				return
			}
			// Render 404 page
			if err := c.Render(http.StatusNotFound, "404.html", nil); err != nil {
				c.Logger().Error(err)
			}
			return
		}

		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Renderer = &Template{templates: template.Must(template.ParseFS(viewsFS, "views/*.html"))}
	e.Static("/assets", "public/assets")

	e.GET("/", func(c echo.Context) error { return c.Render(http.StatusOK, "index.html", nil) })
	e.GET("/login", func(c echo.Context) error {
		if cookie, err := c.Cookie(CookieName); err == nil && cookie.Value == SecretKey {
			return c.Redirect(http.StatusSeeOther, "/admin")
		}
		return c.Render(http.StatusOK, "login.html", nil)
	})
	e.GET("/scan", app.ScanPageHandler)             // Public Scan Page (Presence)
	e.GET("/scan-sholat", app.ScanPrayerPageHandler) // Public Scan Page (Prayer)

	// API Endpointsi
	e.POST("/api/login", app.LoginHandler)
	e.POST("/api/logout", app.LogoutHandler)
	e.GET("/api/sync", app.SyncHandler)

	admin := e.Group("/admin", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(CookieName)
			if err != nil || cookie.Value != SecretKey {
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			return next(c)
		}
	})
	admin.GET("", app.DashboardHandler)

	// Schedule Routes
	admin.POST("/schedule/add", app.AddScheduleHandler)
	admin.POST("/schedule/update/:id", app.UpdateScheduleHandler)
	admin.DELETE("/schedule/:id", app.DeleteScheduleHandler)

	// Audio Routes
	admin.POST("/audio/upload", app.UploadAudioHandler)
	admin.POST("/audio/rename/:id", app.RenameAudioHandler)
	admin.DELETE("/audio/:id", app.DeleteAudioHandler)

	// Device Routes
	admin.POST("/device/add", app.AddDeviceHandler)
	admin.POST("/device/update/:id", app.UpdateDeviceHandler)
	admin.DELETE("/device/:id", app.DeleteDeviceHandler)

	// --- ACADEMIC ROUTES ---

	// Majors
	admin.POST("/major/add", app.AddMajorHandler)
	admin.POST("/major/update/:id", app.UpdateMajorHandler)
	admin.DELETE("/major/:id", app.DeleteMajorHandler)

	// Classes
	admin.POST("/class/add", app.AddClassHandler)
	admin.POST("/class/update/:id", app.UpdateClassHandler)
	admin.DELETE("/class/:id", app.DeleteClassHandler)

	// Students
	admin.POST("/student/add", app.AddStudentHandler)
	admin.POST("/student/update/:id", app.UpdateStudentHandler)
	admin.DELETE("/student/:id", app.DeleteStudentHandler)
	admin.POST("/student/import", app.ImportStudentHandler)
	admin.POST("/student/import-json", app.ImportStudentJSONHandler) // New JSON Import
	admin.POST("/students/delete-multiple", app.BulkDeleteStudentsHandler) // New Bulk Delete Route
	admin.GET("/students/json", app.GetStudentsJSONHandler)               // Fetch students for promotion
	admin.POST("/students/promote", app.PromoteStudentsHandler)           // Promote students

	// Staff
	admin.POST("/staff/add", app.AddStaffHandler)
	admin.POST("/staff/update/:id", app.UpdateStaffHandler)
	admin.DELETE("/staff/:id", app.DeleteStaffHandler)
	admin.POST("/staff/import", app.ImportStaffHandler)

	// Attendance
	admin.POST("/attendance/settings", app.UpdateAttendanceSettingsHandler)
	admin.POST("/attendance/manual", app.ManualAttendanceHandler)
	admin.GET("/attendance/daily", app.GetDailyAttendanceHandler) // New Bulk Attendance Data
	admin.POST("/attendance/bulk", app.BulkAttendanceHandler)     // New Bulk Attendance Save
	admin.POST("/attendance/test-wa", app.TestWAHandler)
	admin.GET("/student/calendar", app.GetStudentCalendarHandler)
	admin.GET("/staff/calendar", app.GetStaffCalendarHandler) // New Staff API

	// Profile Pages
	admin.GET("/student/:id", app.StudentProfileHandler)
	admin.GET("/staff/:id", app.StaffProfileHandler)

	// Report Routes
	admin.GET("/report/daily", app.DailyReportHandler)
	admin.GET("/report/weekly", app.WeeklyReportHandler)
	admin.GET("/report/monthly", app.MonthlyReportHandler)

	e.GET("/api/attendance/record", app.RecordAttendanceHandler)    // Public API for IoT
	e.GET("/api/attendance/today-stats", app.TodayStatsHandler)     // Public API for Kiosk Stats
	e.GET("/api/attendance/recent-logs", app.RecentLogsHandler)     // Public API for Recent Logs
	e.GET("/api/attendance/prayer", app.PrayerAttendanceHandler)    // Public API for Prayer Attendance (TAP)
	e.GET("/api/attendance/prayer-logs", app.PrayerLogsListHandler) // Public API for Prayer Logs List

	// Prayer Admin
	admin.GET("/prayer/attendance", app.GetPrayerAttendanceHandler)
	admin.POST("/prayer/attendance", app.BulkPrayerAttendanceHandler)
	admin.GET("/prayer/report", app.PrayerReportHandler)
	e.GET("/api/student/lookup", app.GetStudentByRFIDHandler)       // Public API for Student Lookup (Points)
	
	// WhatsApp Logs
	admin.GET("/wa-logs", app.GetWhatsAppLogsHandler)

	// Holiday API
	admin.GET("/holidays", app.GetHolidaysHandler)
	admin.POST("/holidays", app.AddHolidayHandler)
	admin.PUT("/holidays/:id", app.UpdateHolidayHandler)
	admin.DELETE("/holidays/:id", app.DeleteHolidayHandler)
	admin.POST("/holidays/import-national", app.ImportNationalHolidaysHandler)

	// Student Point System API
	admin.GET("/point-rules", app.GetPointRulesHandler)
	admin.POST("/point-rules/add", app.AddPointRuleHandler)
	admin.DELETE("/point-rules/:id", app.DeletePointRuleHandler)
	admin.GET("/points/student/:id", app.GetStudentPointProfileHandler)
	admin.POST("/points/transaction", app.AddPointTransactionHandler)
	admin.GET("/points/leaderboard", app.GetLeaderboardHandler)

	// Reward System API
	admin.GET("/point-rewards", app.GetPointRewardsHandler)
	admin.POST("/point-rewards/add", app.AddPointRewardHandler)
	admin.DELETE("/point-rewards/:id", app.DeletePointRewardHandler)
	admin.POST("/points/redeem", app.RedeemRewardHandler)

	// ===== OPERATOR ROUTES (Mobile Prayer Management) =====
	
	// Public Login Page
	e.GET("/operator/login", func(c echo.Context) error {
		data, err := viewsFS.ReadFile("views/mobile/login.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	
	// Authentication API
	e.POST("/api/operator/login", app.OperatorLoginHandler)
	e.POST("/api/operator/logout", app.OperatorLogoutHandler)
	
	// Protected Operator Pages
	operatorPages := e.Group("/operator")
	operatorPages.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return app.OperatorAuthMiddleware(next)
	})
	
	operatorPages.GET("/dashboard", func(c echo.Context) error {
		data, err := viewsFS.ReadFile("views/mobile/dashboard.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	operatorPages.GET("/scan", func(c echo.Context) error {
		data, err := viewsFS.ReadFile("views/mobile/scan.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	operatorPages.GET("/manual", func(c echo.Context) error {
		data, err := viewsFS.ReadFile("views/mobile/manual.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	operatorPages.GET("/profile", func(c echo.Context) error {
		data, err := viewsFS.ReadFile("views/mobile/profile.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
		return c.HTMLBlob(http.StatusOK, data)
	})
	
	// Protected Operator API
	operatorAPI := e.Group("/api/operator")
	operatorAPI.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return app.OperatorAuthMiddleware(next)
	})
	
	operatorAPI.GET("/prayer-stats", app.OperatorPrayerStatsHandler)
	operatorAPI.POST("/scan-qr", app.ScanQRCodeHandler)
	operatorAPI.GET("/classes", app.GetClassesHandler)
	operatorAPI.GET("/recent-logs", app.GetRecentPrayerLogsHandler)
	operatorAPI.GET("/students", app.GetStudentsForPrayerHandler)
	operatorAPI.POST("/prayer-attendance", app.SavePrayerAttendanceHandler)
	operatorAPI.GET("/profile", app.GetOperatorProfileHandler)
	operatorAPI.PUT("/profile", app.UpdateOperatorProfileHandler)
	operatorAPI.PUT("/password", app.ChangeOperatorPasswordHandler)
	
	// QR Code Generation (can be used by admin too)
	admin.GET("/qr-generate", app.GenerateQRCodeHandler)

	// School Settings API
	admin.GET("/settings/school", app.GetSchoolSettingsHandler)
	admin.PUT("/settings/school", app.UpdateSchoolSettingsHandler)

	// Port Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func runWizard() {
	fmt.Println("=========================================")
	fmt.Println("   SmartBell All-in-One Wizard 🚀        ")
	fmt.Println("=========================================")

	if os.Getuid() != 0 {
		fmt.Println("❌ Harap jalankan dengan sudo (sudo ./bell_linux wizard)")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Detect Executable Name
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("❌ Gagal mendeteksi nama file aplikasi.")
		return
	}
	exeName := filepath.Base(exePath)

	for {
		fmt.Println("\nPilih menu:")
		fmt.Println("1) Install Baru (Fresh Install)")
		fmt.Println("2) Update Aplikasi (Update Service ke File Ini)")
		fmt.Println("3) Setup Domain & SSL")
		fmt.Println("4) Keluar")
		fmt.Print("Masukkan pilihan: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1", "2":
			// Option 2 now also runs setup to update the service to point to THIS file
			if choice == "2" {
				fmt.Println("--- Update Service Systemd ---")
			} else {
				fmt.Println("--- Menjalankan Installer ---")
			}
			runScript(setupScript, "setup.sh", exeName)
		case "3":
			fmt.Println("--- Setup Domain & SSL ---")
			runScript(setupNginxScript, "setup_nginx.sh")
		case "4":
			fmt.Println("Bye! 👋")
			return
		default:
			fmt.Println("Pilihan tidak valid.")
		}
	}
}

func runScript(content, name string, args ...string) {
	// Write to temp file
	tmpFile := "/tmp/" + name
	err := os.WriteFile(tmpFile, []byte(content), 0755)
	if err != nil {
		fmt.Printf("❌ Gagal membuat script sementara: %v\n", err)
		return
	}
	// Run it
	cmd := exec.Command("bash", append([]string{tmpFile}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Script error: %v\n", err)
	}
	// Cleanup
	os.Remove(tmpFile)
}

func (a *App) PrayerAttendanceHandler(c echo.Context) error {
	rfid := c.QueryParam("rfid")
	if rfid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "RFID kosong"})
	}

	// 1. Identify User (Students Only)
	var name, className_res string
	var classNameNull sql.NullString
	err := a.DB.QueryRow(`
		SELECT s.name, c.name 
		FROM students s 
		LEFT JOIN classes c ON s.class_id = c.id 
		WHERE s.rfid_uid=?`, rfid).Scan(&name, &classNameNull)
	
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Kartu tidak dikenali"})
	}
	className_res = classNameNull.String

	// 2. Determine Prayer Type based on Time
	now := time.Now()
	timeStr := now.Format("15:04")
	dateStr := now.Format("2006-01-02")
	timestamp := now.Format("2006-01-02 15:04:05")

	var dzStart, dzEnd, asStart, asEnd string
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='dzuhur_start'").Scan(&dzStart)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='dzuhur_end'").Scan(&dzEnd)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='ashar_start'").Scan(&asStart)
	a.DB.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key='ashar_end'").Scan(&asEnd)

	// Defaults if empty
	if dzStart == "" { dzStart = "11:30" }
	if dzEnd == "" { dzEnd = "13:00" }
	if asStart == "" { asStart = "15:00" }
	if asEnd == "" { asEnd = "16:00" }

	prayerType := ""

	if timeStr >= dzStart && timeStr <= dzEnd {
		prayerType = "Dzuhur"
	} else if timeStr >= asStart && timeStr <= asEnd {
		prayerType = "Ashar"
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Bukan waktu sholat"})
	}

	// 3. Check Duplicate
	var count int
	a.DB.QueryRow("SELECT COUNT(*) FROM prayer_logs WHERE rfid_uid=? AND date=? AND prayer_type=?", rfid, dateStr, prayerType).Scan(&count)
	if count > 0 {
		return c.JSON(http.StatusOK, map[string]string{"status": "duplicate", "message": "Sudah absen sholat " + prayerType, "name": name})
	}

	// 4. Record
	_, err = a.DB.Exec("INSERT INTO prayer_logs (rfid_uid, name, class_name, prayer_type, timestamp, date) VALUES (?, ?, ?, ?, ?, ?)",
		rfid, name, className_res, prayerType, timestamp, dateStr)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":      "success",
		"message":     "Absen Sholat " + prayerType + " Berhasil",
		"name":        name,
		"prayer_type": prayerType,
		"time":        timeStr,
	})
}

func (a *App) PrayerLogsListHandler(c echo.Context) error {
	// Get logs for today by default, or limit
	rows, err := a.DB.Query("SELECT id, rfid_uid, name, class_name, prayer_type, timestamp, date FROM prayer_logs ORDER BY id DESC LIMIT 100")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	defer rows.Close()

	var logs []PrayerLog
	for rows.Next() {
		var l PrayerLog
		var className sql.NullString
		rows.Scan(&l.ID, &l.RFID, &l.Name, &className, &l.PrayerType, &l.Timestamp, &l.Date)
		l.ClassName = className.String
		logs = append(logs, l)
	}
	
	if logs == nil {
		logs = []PrayerLog{}
	}

	return c.JSON(http.StatusOK, logs)
}

// --- BULK ATTENDANCE HANDLERS ---

type DailyAttendanceItem struct {
	StudentID int    `json:"student_id"`
	NIS       string `json:"nis"`
	Name      string `json:"name"`
	RFID      string `json:"rfid"`
	Status    string `json:"status"` // Hadir, Sakit, Izin, Alpha, Unknown
	Time      string `json:"time"`
	LogID     int    `json:"log_id"`
}

func (a *App) GetDailyAttendanceHandler(c echo.Context) error {
	classID := c.QueryParam("class_id")
	date := c.QueryParam("date") // Format: YYYY-MM-DD

	if classID == "" || date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Class ID and Date are required"})
	}

	// 1. Get all students in class
	query := `
		SELECT s.id, s.nis, s.name, s.rfid_uid, 
		       COALESCE(l.status, 'Unknown') as status, 
		       COALESCE(strftime('%H:%M', l.timestamp), '') as time,
		       COALESCE(l.id, 0) as log_id
		FROM students s
		LEFT JOIN attendance_logs l ON s.rfid_uid = l.rfid_uid AND l.date = ?
		WHERE s.class_id = ?
		ORDER BY s.name ASC`

	rows, err := a.DB.Query(query, date, classID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error: " + err.Error()})
	}
	defer rows.Close()

	var result []DailyAttendanceItem
	for rows.Next() {
		var item DailyAttendanceItem
		var rfid sql.NullString // Handle null RFID
		if err := rows.Scan(&item.StudentID, &item.NIS, &item.Name, &rfid, &item.Status, &item.Time, &item.LogID); err != nil {
			continue
		}
		item.RFID = rfid.String
		result = append(result, item)
	}

	return c.JSON(http.StatusOK, result)
}

type BulkAttendanceItem struct {
	StudentID int    `json:"student_id"`
	RFID      string `json:"rfid"` // Optional, used for lookup if needed
	Status    string `json:"status"` // Hadir, Sakit, Izin, Alpha
}

type BulkAttendanceRequest struct {
	ClassID int                  `json:"class_id"`
	Date    string               `json:"date"` // YYYY-MM-DD
	Items   []BulkAttendanceItem `json:"students"`
}

func (a *App) BulkAttendanceHandler(c echo.Context) error {
	var req BulkAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	if req.Date == "" || len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Date and Students are required"})
	}

	tx, err := a.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Transaction failed"})
	}

	// Fetch class name for logging (optional, skipped for now to save query)

	for _, item := range req.Items {
		// 1. Get Student Data (Need Name for log)
		var rfid, name string
		// Try to get RFID from DB if not provided or just verify
		// We use StudentID as the reliable key
		err := tx.QueryRow("SELECT rfid_uid, name FROM students WHERE id=?", item.StudentID).Scan(&rfid, &name)
		if err != nil {
			continue // Skip if student not found
		}

		// 2. Check if log exists for this Date & RFID
		var logID int
		err = tx.QueryRow("SELECT id FROM attendance_logs WHERE rfid_uid=? AND date=?", rfid, req.Date).Scan(&logID)

		timestamp := fmt.Sprintf("%s 07:00:00", req.Date) // Default time for manual entry
		if item.Status == "Unknown" {
			// If status set to Unknown, maybe delete the log?
			// For now, let's just ignore or delete. "Alpha" is usually explicit.
			// Let's implement Delete if Status is empty/Unknown to allow "Reset"
			if logID != 0 {
				tx.Exec("DELETE FROM attendance_logs WHERE id=?", logID)
			}
			continue
		}

		if err == sql.ErrNoRows {
			// Insert
			_, err = tx.Exec(`
				INSERT INTO attendance_logs (rfid_uid, user_name, user_type, status, method, timestamp, date)
				VALUES (?, ?, 'Siswa', ?, 'Manual', ?, ?)
			`, rfid, name, item.Status, timestamp, req.Date)
		} else {
			// Update
			_, err = tx.Exec(`
				UPDATE attendance_logs 
				SET status=?, method='Manual' 
				WHERE id=?
			`, item.Status, logID)
		}
	}

	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Commit failed: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Presensi berhasil disimpan"})
}

