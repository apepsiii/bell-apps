package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
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
)

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
			setting_key TEXT PRIMARY KEY,
			setting_value TEXT
		);`,
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

	// Migration: Add 'photo' to students
	var colCountPhoto int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('students') WHERE name='photo'").Scan(&colCountPhoto)
	if colCountPhoto == 0 {
		db.Exec("ALTER TABLE students ADD COLUMN photo TEXT DEFAULT ''")
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

func SendOneSenderMessage(to, message, token, apiUrl, recipientType, imageUrl string) (string, error) {
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
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("WA Error (Do):", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
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

	var datasets1 []ChartDataset
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

	var statusLabels []string
	var statusCounts []int
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
	_, err := a.DB.Exec("DELETE FROM students WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa dihapus"})
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
			SendOneSenderMessage(parentPhone, msg, waToken, waURL, "individual", waImage)
		}
		// 2. Send to Class Group
		if classGroup != "" {
			SendOneSenderMessage(classGroup, msg, waToken, waURL, "group", waImage)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Absensi manual berhasil disimpan"})
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

	log.Printf("Calendar data for student %s, month %s, year %s: %+v", studentID, month, year, calendarData)
	return c.JSON(http.StatusOK, calendarData)
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

	return c.JSON(http.StatusOK, calendarData)
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
	resp, err := SendOneSenderMessage(target, msg, waToken, waURL, "individual", waImage)

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
	var name, userType, photo string
	var photoNull sql.NullString

	err := a.DB.QueryRow("SELECT name, photo FROM students WHERE rfid_uid=?", rfid).Scan(&name, &photoNull)
	if err == nil {
		userType = "Siswa"
		photo = photoNull.String
	} else {
		err = a.DB.QueryRow("SELECT name FROM staff WHERE rfid_uid=?", rfid).Scan(&name)
		if err == nil {
			userType = "Staff"
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

			// Replace Template Variables
			msg = strings.ReplaceAll(msg, "{name}", name)
			msg = strings.ReplaceAll(msg, "{time}", timeStr)
			msg = strings.ReplaceAll(msg, "{status}", status)

			// 1. Send to Parent
			if parentPhone != "" {
				SendOneSenderMessage(parentPhone, msg, waToken, waURL, "individual", waImage)
			}
			// 2. Send to Class Group
			if classGroup != "" {
				SendOneSenderMessage(classGroup, msg, waToken, waURL, "group", waImage)
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
				SendOneSenderMessage(staffPhone, msg, waToken, waURL, "individual", waImage)
			}
		}()
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Absen " + status + " Berhasil",
		"name":    name,
		"type":    userType,
		"time":    timeStr,
		"photo":   photo, // Return photo filename
	})
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

// --- MAIN ---

func main() {
	db := InitDB()
	defer db.Close()
	app := &App{DB: db}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Renderer = &Template{templates: template.Must(template.ParseGlob("views/*.html"))}
	e.Static("/assets", "public/assets")

	e.GET("/", func(c echo.Context) error { return c.Render(http.StatusOK, "index.html", nil) })
	e.GET("/login", func(c echo.Context) error {
		if cookie, err := c.Cookie(CookieName); err == nil && cookie.Value == SecretKey {
			return c.Redirect(http.StatusSeeOther, "/admin")
		}
		return c.Render(http.StatusOK, "login.html", nil)
	})
	e.GET("/scan", app.ScanPageHandler) // Halaman Terminal Absensi
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

	// Staff
	admin.POST("/staff/add", app.AddStaffHandler)
	admin.POST("/staff/update/:id", app.UpdateStaffHandler)
	admin.DELETE("/staff/:id", app.DeleteStaffHandler)
	admin.POST("/staff/import", app.ImportStaffHandler)

	// Attendance
	admin.POST("/attendance/settings", app.UpdateAttendanceSettingsHandler)
	admin.POST("/attendance/manual", app.ManualAttendanceHandler)
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

	e.GET("/api/attendance/record", app.RecordAttendanceHandler) // Public API for IoT

	// Port Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
