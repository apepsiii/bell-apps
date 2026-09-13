package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// AttendanceIndexCalculator calculates attendance index for composite scoring
type AttendanceIndexCalculator struct {
	db *sql.DB
}

// AttendanceIndexResult holds calculation results
type AttendanceIndexResult struct {
	StudentID          int     `json:"student_id"`
	StudentName        string  `json:"student_name"`
	SemesterID         int     `json:"semester_id"`
	AttendanceDays     int     `json:"attendance_days"`
	LateDays           int     `json:"late_days"`
	AbsentDays         int     `json:"absent_days"`
	TotalWorkingDays   int     `json:"total_working_days"`
	CurrentStreak      int     `json:"current_streak"`
	LongestStreak      int     `json:"longest_streak"`
	StreakBonus        float64 `json:"streak_bonus"`
	AttendanceIndex    float64 `json:"attendance_index"`
	LastCalculatedAt   string  `json:"last_calculated_at"`
}

// CalculateAttendanceIndex calculates attendance index (0-100) for a student in a semester
// Formula: 
// - Hadir tepat waktu = 100%
// - Terlambat = 50%
// - Alpha/Sakit/Dispensasi = 0%
// - Streak bonus: 10 hari berturut-turut = +10 poin indeks
func CalculateAttendanceIndex(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.QueryParam("student_id")
		semesterID := c.QueryParam("semester_id")

		if semesterID == "" {
			semesterID = "1" // default to current semester
		}

		calc := &AttendanceIndexCalculator{db: db}

		// Get semester date range
		var startDate, endDate string
		err := db.QueryRow("SELECT start_date, end_date FROM semesters WHERE id = ?", semesterID).Scan(&startDate, &endDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Semester tidak ditemukan"})
		}

		var results []AttendanceIndexResult

		if studentID != "" {
			// Calculate for single student
			result, err := calc.calculateForStudent(studentID, semesterID, startDate, endDate)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			results = append(results, result)
		} else {
			// Calculate for all active students
			results, err = calc.calculateForAllStudents(semesterID, startDate, endDate)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}

		// Save results to database
		for _, result := range results {
			err := calc.saveResult(result)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal menyimpan hasil: " + err.Error()})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("%d indeks kehadiran berhasil dihitung", len(results)),
			"results": results,
		})
	}
}

func (calc *AttendanceIndexCalculator) calculateForStudent(studentID, semesterID, startDate, endDate string) (AttendanceIndexResult, error) {
	result := AttendanceIndexResult{
		SemesterID: 0,
	}

	// Convert semesterID to int
	fmt.Sscanf(semesterID, "%d", &result.SemesterID)
	fmt.Sscanf(studentID, "%d", &result.StudentID)

	// Get student name
	calc.db.QueryRow("SELECT name FROM students WHERE id = ?", studentID).Scan(&result.StudentName)

	// Count working days in semester (exclude holidays and weekends)
	result.TotalWorkingDays = calc.countWorkingDays(startDate, endDate)

	// Get attendance statistics
	query := `
		SELECT 
			COUNT(CASE WHEN status IN ('Hadir', 'Datang') THEN 1 END) as hadir,
			COUNT(CASE WHEN status = 'Terlambat' THEN 1 END) as terlambat,
			COUNT(CASE WHEN status IN ('Alpha', 'Sakit', 'Sakit (Dengan Surat)', 'Sakit (Tanpa Surat)', 'Dispensasi') THEN 1 END) as absent
		FROM attendance_logs al
		JOIN students s ON al.rfid_uid = s.rfid_uid
		WHERE s.id = ? AND DATE(al.timestamp) BETWEEN ? AND ?
		  AND al.status != 'Pulang'
	`
	var hadir, terlambat, absent int
	err := calc.db.QueryRow(query, studentID, startDate, endDate).Scan(&hadir, &terlambat, &absent)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}

	result.AttendanceDays = hadir
	result.LateDays = terlambat
	result.AbsentDays = absent

	// Calculate attendance index
	// Formula: (hadir * 100 + terlambat * 50) / total_working_days
	if result.TotalWorkingDays > 0 {
		rawIndex := float64(hadir*100+terlambat*50) / float64(result.TotalWorkingDays)
		result.AttendanceIndex = rawIndex
	}

	// Calculate streak
	result.CurrentStreak, result.LongestStreak = calc.calculateStreak(studentID, startDate, endDate)

	// Apply streak bonus: every 10 consecutive days = +10 points (max once per calculation)
	streakBonusDays := 10 // from config
	if result.CurrentStreak >= streakBonusDays {
		result.StreakBonus = 10.0 // from config
		result.AttendanceIndex += result.StreakBonus
	}

	// Cap at 100
	if result.AttendanceIndex > 100 {
		result.AttendanceIndex = 100
	}

	result.LastCalculatedAt = time.Now().Format("2006-01-02 15:04:05")

	return result, nil
}

func (calc *AttendanceIndexCalculator) calculateForAllStudents(semesterID, startDate, endDate string) ([]AttendanceIndexResult, error) {
	var results []AttendanceIndexResult

	rows, err := calc.db.Query("SELECT id FROM students WHERE status = 'active'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var studentID int
		rows.Scan(&studentID)

		result, err := calc.calculateForStudent(fmt.Sprintf("%d", studentID), semesterID, startDate, endDate)
		if err != nil {
			continue // skip errors, continue with next student
		}
		results = append(results, result)
	}

	return results, nil
}

func (calc *AttendanceIndexCalculator) countWorkingDays(startDate, endDate string) int {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		// Check if it's a working day (skip weekends and holidays)
		isWorking, _ := IsWorkingDay(calc.db, d)
		if isWorking {
			count++
		}
	}
	return count
}

func (calc *AttendanceIndexCalculator) calculateStreak(studentID, startDate, endDate string) (int, int) {
	// Get all attendance dates for this student in the period
	query := `
		SELECT DISTINCT DATE(al.timestamp) as date
		FROM attendance_logs al
		JOIN students s ON al.rfid_uid = s.rfid_uid
		WHERE s.id = ? AND DATE(al.timestamp) BETWEEN ? AND ?
		  AND al.status IN ('Hadir', 'Datang', 'Terlambat')
		ORDER BY date ASC
	`
	rows, err := calc.db.Query(query, studentID, startDate, endDate)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var dateStr string
		rows.Scan(&dateStr)
		date, _ := time.Parse("2006-01-02", dateStr)
		dates = append(dates, date)
	}

	if len(dates) == 0 {
		return 0, 0
	}

	currentStreak := 1
	longestStreak := 1
	
	for i := 1; i < len(dates); i++ {
		dayDiff := int(dates[i].Sub(dates[i-1]).Hours() / 24)
		
		if dayDiff == 1 {
			// Consecutive day
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			// Break in streak
			currentStreak = 1
		}
	}

	return currentStreak, longestStreak
}

func (calc *AttendanceIndexCalculator) saveResult(result AttendanceIndexResult) error {
	// Update or insert into semester_scores
	_, err := calc.db.Exec(`
		INSERT INTO semester_scores 
		(student_id, semester_id, attendance_index, attendance_days, late_days, absent_days, 
		 streak_days, streak_bonus, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(student_id, semester_id) DO UPDATE SET
			attendance_index = excluded.attendance_index,
			attendance_days = excluded.attendance_days,
			late_days = excluded.late_days,
			absent_days = excluded.absent_days,
			streak_days = excluded.streak_days,
			streak_bonus = excluded.streak_bonus,
			calculated_at = excluded.calculated_at
	`, result.StudentID, result.SemesterID, result.AttendanceIndex, result.AttendanceDays,
		result.LateDays, result.AbsentDays, result.CurrentStreak, result.StreakBonus,
		result.LastCalculatedAt)

	if err != nil {
		return err
	}

	// Update streak tracking
	_, err = calc.db.Exec(`
		INSERT INTO attendance_streaks 
		(student_id, semester_id, current_streak, longest_streak, last_calculated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(student_id, semester_id) DO UPDATE SET
			current_streak = excluded.current_streak,
			longest_streak = excluded.longest_streak,
			last_calculated_at = excluded.last_calculated_at
	`, result.StudentID, result.SemesterID, result.CurrentStreak, result.LongestStreak,
		result.LastCalculatedAt)

	return err
}
