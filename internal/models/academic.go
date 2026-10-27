package models

// TeacherJournal records the official teaching log for one class meeting.
type TeacherJournal struct {
	ID                 int                `json:"id"`
	AssignmentID       int                `json:"assignment_id"`
	MeetingID          *int               `json:"meeting_id,omitempty"`
	JournalDate        string             `json:"journal_date"`
	MeetingOrder       int                `json:"meeting_order"`
	LessonTopic        string             `json:"lesson_topic"`
	LearningObjectives string             `json:"learning_objectives"`
	TeachingMethod     string             `json:"teaching_method"`
	LearningMedia      string             `json:"learning_media"`
	Activities         string             `json:"activities"`
	Assessment         string             `json:"assessment"`
	StudentIssues      string             `json:"student_issues"`
	FollowUp           string             `json:"follow_up"`
	AttendanceSummary  map[string]int     `json:"attendance_summary"`
	Status             string             `json:"status"`
	RejectionReason    string             `json:"rejection_reason,omitempty"`
	SubmittedAt        string             `json:"submitted_at,omitempty"`
	ApprovedBy         string             `json:"approved_by,omitempty"`
	ApprovedAt         string             `json:"approved_at,omitempty"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
	Assignment         *TeacherAssignment `json:"assignment,omitempty"`
}

// AcademicCalendarEvent represents a published school academic calendar item.
type AcademicCalendarEvent struct {
	ID           int    `json:"id"`
	AcademicYear string `json:"academic_year"`
	Semester     int    `json:"semester"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	IsPublished  bool   `json:"is_published"`
	CreatedBy    string `json:"created_by,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
