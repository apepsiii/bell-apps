package models

// Subject represents a curriculum course/mata pelajaran
type Subject struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	MajorID   *int   `json:"major_id"`
	MajorName string `json:"major_name,omitempty"`
	Category  string `json:"category"` // 'muatan_nasional', 'muatan_kejuruan', 'muatan_lokal'
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// TeacherAssignment connects Staff (Guru) with Subject and Class for an academic period
type TeacherAssignment struct {
	ID           int    `json:"id"`
	StaffID      int    `json:"staff_id"`
	StaffName    string `json:"staff_name,omitempty"`
	StaffNIP     string `json:"staff_nip,omitempty"`
	SubjectID    int    `json:"subject_id"`
	SubjectName  string `json:"subject_name,omitempty"`
	SubjectCode  string `json:"subject_code,omitempty"`
	ClassID      int    `json:"class_id"`
	ClassName    string `json:"class_name,omitempty"`
	AcademicYear string `json:"academic_year"`
	Semester     int    `json:"semester"`
	CreatedAt    string `json:"created_at"`
}

// LMSMeeting represents a class meeting / chapter session
type LMSMeeting struct {
	ID           int             `json:"id"`
	AssignmentID int             `json:"assignment_id"`
	MeetingOrder int             `json:"meeting_order"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	MeetingDate  string          `json:"meeting_date"`
	IsPublished  bool            `json:"is_published"`
	CreatedAt    string          `json:"created_at"`
	Materials    []LMSMaterial   `json:"materials,omitempty"`
	Assignments  []LMSAssignment `json:"assignments,omitempty"`
}

// LMSAttendance records student attendance during an online/offline LMS meeting
type LMSAttendance struct {
	ID              int    `json:"id"`
	MeetingID       int    `json:"meeting_id"`
	StudentID       int    `json:"student_id"`
	StudentName     string `json:"student_name,omitempty"`
	StudentNIS      string `json:"student_nis,omitempty"`
	Status          string `json:"status"` // 'H', 'S', 'I', 'A'
	Notes           string `json:"notes"`
	MarkedByStaffID int    `json:"marked_by_staff_id"`
	UpdatedAt       string `json:"updated_at"`
}

// LMSMaterial represents a course study material (document, video, link, or text)
type LMSMaterial struct {
	ID          int    `json:"id"`
	MeetingID   int    `json:"meeting_id"`
	Title       string `json:"title"`
	Type        string `json:"type"` // 'document', 'video', 'link', 'text'
	ContentURL  string `json:"content_url"`
	FileKey     string `json:"file_key"`
	FileSize    int64  `json:"file_size"`
	TextContent string `json:"text_content"`
	CreatedAt   string `json:"created_at"`
}

// LMSAssignment represents a homework/project assignment created by a teacher
type LMSAssignment struct {
	ID           int    `json:"id"`
	MeetingID    int    `json:"meeting_id"`
	Title        string `json:"title"`
	Instructions string `json:"instructions"`
	AllowedType  string `json:"allowed_type"` // 'file', 'link', 'text', 'any'
	DueDate      string `json:"due_date"`
	MaxPoints    int    `json:"max_points"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// LMSSubmission represents a student submission for an assignment
type LMSSubmission struct {
	ID             int      `json:"id"`
	AssignmentID   int      `json:"assignment_id"`
	StudentID      int      `json:"student_id"`
	StudentName    string   `json:"student_name,omitempty"`
	StudentNIS     string   `json:"student_nis,omitempty"`
	SubmissionType string   `json:"submission_type"` // 'file', 'link', 'text'
	FileURL        string   `json:"file_url"`
	FileKey        string   `json:"file_key"`
	LinkURL        string   `json:"link_url"`
	TextContent    string   `json:"text_content"`
	SubmittedAt    string   `json:"submitted_at"`
	IsLate         bool     `json:"is_late"`
	Score          *float64 `json:"score"`
	Feedback       string   `json:"feedback"`
	GradedBy       *int     `json:"graded_by"`
	GradedAt       *string  `json:"graded_at"`
}

