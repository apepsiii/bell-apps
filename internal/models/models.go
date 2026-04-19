package models

import (
	"database/sql"
	"sync"
	"time"
)

type App struct {
	DB                 *sql.DB
	Mu                 sync.Mutex
	ActiveAnnouncement interface{}
}

type Template struct {
	templates interface{}
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
	MajorName string
	WAGroupID string
}

type Student struct {
	ID          int
	RFID        string
	NIS         string
	Name        string
	ParentPhone string
	ParentName  string
	ClassID     int
	ClassName   string
	Photo       string
}

type Staff struct {
	ID    int
	RFID  string
	NIP   string
	Name  string
	Phone string
	Role  string
}

type AttendanceSetting struct {
	Key   string
	Value string
}

type AttendanceLog struct {
	ID        int
	RFID      string
	UserName  string
	UserType  string
	Status    string
	Method    string
	Timestamp string
	Date      string
	UserPhoto string
}

type RunningText struct {
	ID       int
	Content  string
	IsActive bool
}

type SignageMedia struct {
	ID       int
	Filename string
	FileType string
	Duration int
	IsActive bool
}

type PrayerLog struct {
	ID         int    `json:"id"`
	RFID       string `json:"rfid_uid"`
	Name       string `json:"name"`
	ClassName  string `json:"class_name"`
	PrayerType string `json:"prayer_type"`
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
	Date       string `json:"date"`
}

type StudentStatus struct {
	Student
	Status string
	Method string
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
	Announcements      []Announcement
	AttendanceLogs     []AttendanceLog
	AttendanceSettings map[string]string
	PresentStudents    []StudentStatus
	AbsentStudents     []Student
	RunningTexts       []RunningText
	SignageMedia       []SignageMedia
	ChartWeeklyClass   string
	ChartStatus        string
	ChartArrival       string
	Stats              struct {
		TotalSchedules int
		NextBell       string
		OnlineDevices  int
		TotalDevices   int
		TotalStudents  int
		TotalStaff     int
	}
	AppVersion string
}

type Announcement struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Message     string       `json:"message"`
	AudioFile   string       `json:"audio_file"`
	ScheduledAt sql.NullTime `json:"scheduled_at"`
	PlayedAt    sql.NullTime `json:"played_at"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

type Holiday struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type SchoolSetting struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Operator struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Photo     string `json:"photo"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type PointRule struct {
	ID          int    `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

type PointReward struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PointsCost  int    `json:"points_cost"`
	Stock       int    `json:"stock"`
	Description string `json:"description"`
}

type StudentPointLog struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	RuleID       *int   `json:"rule_id"`
	RewardID     *int   `json:"reward_id"`
	PointsChange int    `json:"points_change"`
	Description  string `json:"description"`
	Timestamp    string `json:"timestamp"`
	RecordedBy   string `json:"recorded_by"`
}

type StudentPointProfile struct {
	StudentID   int               `json:"student_id"`
	Name        string            `json:"name"`
	ClassName   string            `json:"class_name"`
	TotalPoints int               `json:"total_points"`
	History     []StudentPointLog `json:"history"`
}

type WhatsAppLog struct {
	ID        int
	Target    string
	Message   string
	Status    string
	Response  string
	Timestamp string
}
