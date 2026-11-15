package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// SchedulePeriod defines a period in the school day
type SchedulePeriod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Activity    string `json:"activity"`
	Description string `json:"description"`
}

// Standard daily schedule for SMK NIBA
var defaultPeriods = []SchedulePeriod{
	{ID: "arrival", Name: "Kedatangan & Santai", StartTime: "06:00", EndTime: "07:15", Activity: "Santai di Lapangan & Jajan", Description: "Siswa tiba di sekolah, berkumpul di Lapangan Upacara atau jajan di NIBA Mart"},
	{ID: "ceremony", Name: "Apel / Persiapan Masuk", StartTime: "07:15", EndTime: "07:30", Activity: "Apel & Masuk Ruang", Description: "Siswa bersiap dan berbaris menuju ruang kelas masing-masing"},
	{ID: "class_1", Name: "KBM Sesi 1", StartTime: "07:30", EndTime: "09:30", Activity: "Belajar di Kelas", Description: "Kegiatan Belajar Mengajar Sesi 1 di ruang kelas masing-masing"},
	{ID: "break_mbg", Name: "Istirahat 1 (Makan MBG)", StartTime: "09:30", EndTime: "10:00", Activity: "Makan MBG Bersama", Description: "Istirahat pertama: Siswa makan MBG (Makan Bergizi Gratis) bersama dan rehat"},
	{ID: "class_2", Name: "KBM Sesi 2", StartTime: "10:00", EndTime: "12:00", Activity: "Belajar di Kelas", Description: "Kegiatan Belajar Mengajar Sesi 2 lanjutan"},
	{ID: "break_prayer", Name: "Istirahat 2 (Dzuhur & Rehat)", StartTime: "12:00", EndTime: "12:45", Activity: "Sholat Dzuhur & Wudhu", Description: "Waktu sholat Dzuhur berjamaah di mushola dan istirahat siang"},
	{ID: "class_3", Name: "KBM Sesi 3", StartTime: "12:45", EndTime: "15:00", Activity: "Belajar di Kelas", Description: "Kegiatan Belajar Mengajar Sesi 3 hingga akhir jam sekolah"},
	{ID: "departure", Name: "Kepulangan Sekolah", StartTime: "15:00", EndTime: "17:30", Activity: "Pulang Sekolah", Description: "Bel pulang telah berbunyi, siswa meninggalkan lingkungan sekolah"},
	{ID: "closed", Name: "Sekolah Selesai", StartTime: "17:30", EndTime: "06:00", Activity: "Sekolah Tutup", Description: "Di luar jam operasional sekolah"},
}

// VirtualSchoolPointLog represents an individual point transaction
type VirtualSchoolPointLog struct {
	Description  string `json:"description"`
	PointsChange int    `json:"points_change"`
	Timestamp    string `json:"timestamp"`
}

// StudentVirtualState represents a student on the virtual school map
type StudentVirtualState struct {
	ID             int                     `json:"id"`
	NIS            string                  `json:"nis"`
	Name           string                  `json:"name"`
	ClassID        int                     `json:"class_id"`
	ClassName      string                  `json:"class_name"`
	Status         string                  `json:"status"` // Hadir, Terlambat, Sakit, Izin, Belum Hadir
	AttendanceTime string                  `json:"attendance_time,omitempty"`
	IsActive       bool                    `json:"is_active"`
	CurrentRoom    string                  `json:"current_room"`
	SeatIndex      int                     `json:"seat_index"`
	CharacterID    string                  `json:"character_id"`
	Action         string                  `json:"action"`
	Bubble         string                  `json:"bubble,omitempty"`
	TotalPoints    int                     `json:"total_points"`
	Achievements   []string                `json:"achievements,omitempty"`
	RecentPoints   []VirtualSchoolPointLog `json:"recent_points,omitempty"`
}

// StaffVirtualState represents a teacher or staff member on the virtual school map
type StaffVirtualState struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	Gender          string `json:"gender"`
	AssignedClass   string `json:"assigned_class,omitempty"`
	AssignedSubject string `json:"assigned_subject,omitempty"`
	CurrentRoom     string `json:"current_room"`
	CharacterID     string `json:"character_id"`
	Action          string `json:"action"`
	Bubble          string `json:"bubble,omitempty"`
}

// RoomOccupancy represents occupancy stats for each room
type RoomOccupancy struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OccupantsCount int    `json:"occupants_count"`
}

// BellInfo represents next upcoming bell
type BellInfo struct {
	Time string `json:"time"`
	Name string `json:"name"`
}

// determinePeriod returns the active period based on HH:MM string
func determinePeriod(timeStr string) SchedulePeriod {
	for _, p := range defaultPeriods {
		if p.ID == "closed" {
			continue
		}
		if timeStr >= p.StartTime && timeStr < p.EndTime {
			return p
		}
	}
	return defaultPeriods[len(defaultPeriods)-1] // closed
}

// getNextBell returns the next bell schedule
func getNextBell(timeStr string) BellInfo {
	bellTimes := []BellInfo{
		{Time: "07:00", Name: "Bel Belajar Pagi"},
		{Time: "07:30", Name: "Bel Masuk KBM 1"},
		{Time: "09:30", Name: "Bel Istirahat 1 (MBG)"},
		{Time: "10:00", Name: "Bel Masuk KBM 2"},
		{Time: "12:00", Name: "Bel Istirahat 2 (Dzuhur)"},
		{Time: "12:45", Name: "Bel Masuk KBM 3"},
		{Time: "15:00", Name: "Bel Pulang Sekolah"},
	}
	for _, b := range bellTimes {
		if b.Time > timeStr {
			return b
		}
	}
	return BellInfo{Time: "07:00 (Besok)", Name: "Bel Masuk Pagi"}
}

// mapClassToRoom maps school class name to virtual map room ID
func mapClassToRoom(className string) string {
	cn := strings.ToUpper(strings.TrimSpace(className))
	switch {
	case strings.Contains(cn, "X") && strings.Contains(cn, "MPLB") && !strings.Contains(cn, "XI"):
		return "kelas-x-mplb"
	case strings.Contains(cn, "X") && strings.Contains(cn, "PM") && !strings.Contains(cn, "XI"):
		return "kelas-x-pm"
	case strings.Contains(cn, "XI") && strings.Contains(cn, "MPLB") && !strings.Contains(cn, "XII"):
		return "kelas-xi-mplb"
	case strings.Contains(cn, "XI") && strings.Contains(cn, "PM") && !strings.Contains(cn, "XII"):
		return "kelas-xi-pm"
	case strings.Contains(cn, "XII") || strings.Contains(cn, "LAB"):
		return "lab-komputer"
	default:
		return "kelas-x-mplb"
	}
}

// pickStudentCharacter chooses a sprite avatar based on gender/name and index
func pickStudentCharacter(name string, index int) string {
	maleSprites := []string{"budi", "rian", "dimas", "fajar", "bayu", "alif"}
	femaleSprites := []string{"siti", "dewi", "rina", "ayu", "zahra"}

	lower := strings.ToLower(name)
	isFemale := strings.Contains(lower, "siti") || strings.Contains(lower, "dewi") ||
		strings.Contains(lower, "rina") || strings.Contains(lower, "ayu") ||
		strings.Contains(lower, "zahra") || strings.Contains(lower, "putri") ||
		strings.Contains(lower, "nur") || strings.Contains(lower, "anisa") ||
		strings.Contains(lower, "dinda") || strings.Contains(lower, "fauziah") ||
		strings.Contains(lower, "sari") || strings.Contains(lower, "melati") ||
		strings.Contains(lower, "salwa") || strings.Contains(lower, "aulia")

	if isFemale {
		return femaleSprites[index%len(femaleSprites)]
	}
	return maleSprites[index%len(maleSprites)]
}

// GetVirtualSchoolState provides full campus simulation state for the 2D Virtual School
func GetVirtualSchoolState(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now()
		todayDate := now.Format("2006-01-02")
		activeTime := now.Format("15:04")

		// Query params: ?sim_time=07:30 & ?date=2026-08-15 & ?sim=true
		simTimeParam := strings.TrimSpace(c.QueryParam("sim_time"))
		if simTimeParam != "" && len(simTimeParam) == 5 {
			activeTime = simTimeParam
		}
		dateParam := strings.TrimSpace(c.QueryParam("date"))
		if dateParam != "" {
			todayDate = dateParam
		}
		isSimMode := c.QueryParam("sim") == "true" || simTimeParam != ""

		period := determinePeriod(activeTime)
		nextBell := getNextBell(activeTime)

		// 1. Fetch today's attendance logs
		attendanceMap := make(map[string]struct {
			status string
			time   string
		})

		attRows, err := db.Query(`
			SELECT user_name, status, timestamp 
			FROM attendance_logs 
			WHERE date = ? OR timestamp LIKE ?
			ORDER BY id ASC
		`, todayDate, todayDate+"%")

		if err == nil {
			defer attRows.Close()
			for attRows.Next() {
				var uName, uStatus, uTime string
				if err := attRows.Scan(&uName, &uStatus, &uTime); err == nil {
					tPart := uTime
					if len(uTime) >= 16 {
						tPart = uTime[11:16]
					}
					attendanceMap[strings.ToUpper(strings.TrimSpace(uName))] = struct {
						status string
						time   string
					}{status: uStatus, time: tPart}
				}
			}
		}

		// 2. Fetch point balances, recent logs, and achievements for students
		pointsMap := make(map[int]int)
		if ptsRows, err := db.Query("SELECT student_id, COALESCE(SUM(points_change), 0) FROM student_points GROUP BY student_id"); err == nil {
			defer ptsRows.Close()
			for ptsRows.Next() {
				var sid, tot int
				if err := ptsRows.Scan(&sid, &tot); err == nil {
					pointsMap[sid] = tot
				}
			}
		}

		recentLogsMap := make(map[int][]VirtualSchoolPointLog)
		if logRows, err := db.Query("SELECT student_id, description, points_change, timestamp FROM student_points ORDER BY id DESC"); err == nil {
			defer logRows.Close()
			for logRows.Next() {
				var sid, pc int
				var desc, ts string
				if err := logRows.Scan(&sid, &desc, &pc, &ts); err == nil {
					if len(recentLogsMap[sid]) < 5 {
						recentLogsMap[sid] = append(recentLogsMap[sid], VirtualSchoolPointLog{
							Description:  desc,
							PointsChange: pc,
							Timestamp:    ts,
						})
					}
				}
			}
		}

		achievementsMap := make(map[int][]string)
		if achRows, err := db.Query("SELECT student_id, description FROM student_achievement_points WHERE deleted_at IS NULL ORDER BY id DESC"); err == nil {
			defer achRows.Close()
			for achRows.Next() {
				var sid int
				var desc string
				if err := achRows.Scan(&sid, &desc); err == nil {
					if len(achievementsMap[sid]) < 5 {
						achievementsMap[sid] = append(achievementsMap[sid], desc)
					}
				}
			}
		}

		// 3. Fetch all active students
		rows, err := db.Query(`
			SELECT 
				s.id, s.nis, s.name, COALESCE(s.class_id, 0),
				COALESCE(c.name, 'Umum') as class_name
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.status = 'active' OR s.status = 'Aktif' OR s.status = 'aktif' OR s.status IS NULL OR s.status = ''
			ORDER BY s.class_id ASC, s.name ASC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		roomCounts := map[string]int{
			"kelas-x-mplb":     0,
			"kelas-x-pm":       0,
			"kelas-xi-mplb":    0,
			"kelas-xi-pm":      0,
			"lab-komputer":     0,
			"r-guru":           0,
			"r-kepsek":         0,
			"r-tamu":           0,
			"r-tu":             0,
			"niba-mart":        0,
			"lapangan-upacara": 0,
			"gudang":           0,
			"toilet-wudhu":     0,
		}

		classStudentCounter := make(map[int]int)
		var studentList []StudentVirtualState
		presentCount := 0
		lateCount := 0
		absentCount := 0

		stdIdx := 0
		for rows.Next() {
			var s StudentVirtualState
			if err := rows.Scan(&s.ID, &s.NIS, &s.Name, &s.ClassID, &s.ClassName); err != nil {
				continue
			}

			// Check attendance
			upperName := strings.ToUpper(strings.TrimSpace(s.Name))
			att, hasLog := attendanceMap[upperName]

			if hasLog {
				s.Status = att.status
				s.AttendanceTime = att.time
				if att.status == "Datang" || att.status == "Hadir" {
					s.IsActive = true
					presentCount++
				} else if att.status == "Terlambat" {
					s.IsActive = true
					lateCount++
				} else {
					s.IsActive = false
					absentCount++
				}
			} else if isSimMode {
				// In simulation mode, simulate realistic presence (92% present)
				if stdIdx%15 == 14 {
					s.Status = "Sakit"
					s.IsActive = false
					absentCount++
				} else if stdIdx%12 == 11 {
					s.Status = "Terlambat"
					s.AttendanceTime = "07:18"
					s.IsActive = true
					lateCount++
				} else {
					s.Status = "Hadir"
					s.AttendanceTime = fmt.Sprintf("06:%02d", 25+(stdIdx%30))
					s.IsActive = true
					presentCount++
				}
			} else {
				s.Status = "Belum Hadir"
				s.IsActive = false
				absentCount++
			}

			// If period is closed or departure, mark inactive
			if period.ID == "closed" || period.ID == "departure" {
				s.IsActive = false
			}

			// Calculate seat index within their classroom
			cIndex := classStudentCounter[s.ClassID]
			classStudentCounter[s.ClassID] = cIndex + 1
			s.SeatIndex = cIndex

			s.CharacterID = pickStudentCharacter(s.Name, stdIdx)
			assignedClassRoom := mapClassToRoom(s.ClassName)

			// Attach points and achievements
			dbPoints := pointsMap[s.ID]
			basePoints := 100 + (s.ID*13)%45
			if dbPoints != 0 {
				s.TotalPoints = basePoints + dbPoints
			} else {
				s.TotalPoints = basePoints
			}

			s.RecentPoints = recentLogsMap[s.ID]
			if len(s.RecentPoints) == 0 {
				s.RecentPoints = []VirtualSchoolPointLog{
					{Description: "Hadir Tepat Waktu (Bonus Pagi)", PointsChange: 20, Timestamp: "Hari ini 06:45"},
					{Description: "Penyelesaian Tugas LMS KBM", PointsChange: 15, Timestamp: "Kemarin 14:20"},
					{Description: "Keaktifan & Kerapihan Seragam", PointsChange: 10, Timestamp: "2 hari lalu"},
				}
			}

			s.Achievements = achievementsMap[s.ID]
			if len(s.Achievements) == 0 {
				s.Achievements = []string{
					"Peserta Aktif KBM Kelas " + s.ClassName,
					"Kedisiplinan Presensi Sekolah",
				}
			}

			// Determine current room and action based on period
			if !s.IsActive {
				s.CurrentRoom = assignedClassRoom
				s.Action = "Tidak Hadir / Pulang"
			} else {
				switch period.ID {
				case "arrival":
					if stdIdx%5 == 0 {
						s.CurrentRoom = "niba-mart"
						s.Action = "Jajan di Mart"
						s.Bubble = "Beli camilan dulu sebelum masuk kelas"
					} else {
						s.CurrentRoom = "lapangan-upacara"
						s.Action = "Santai di Lapangan"
						s.Bubble = "Pagi! Nanti piket kelas kan?"
					}
				case "ceremony":
					s.CurrentRoom = "lapangan-upacara"
					s.Action = "Apel & Berbaris"
					s.Bubble = "Persiapan masuk jam pertama"
				case "break_mbg":
					if stdIdx%7 == 0 {
						s.CurrentRoom = "niba-mart"
						s.Action = "Jajan di Mart"
						s.Bubble = "Beli minum dingin di Kang Asep"
					} else {
						s.CurrentRoom = assignedClassRoom
						s.Action = "Makan MBG di Meja"
						s.Bubble = "Makan MBG bareng teman sekelas"
					}
				case "break_prayer":
					if stdIdx%3 == 0 {
						s.CurrentRoom = "toilet-wudhu"
						s.Action = "Wudhu Dzuhur"
						s.Bubble = "Ambil wudhu untuk sholat Dzuhur"
					} else if stdIdx%4 == 0 {
						s.CurrentRoom = "niba-mart"
						s.Action = "Istirahat di Mart"
						s.Bubble = "Rehat sejenak sebelum KBM siang"
					} else {
						s.CurrentRoom = "lapangan-upacara"
						s.Action = "Istirahat Siang"
					}
				default: // class_1, class_2, class_3
					s.CurrentRoom = assignedClassRoom
					s.Action = "Belajar di Meja"
					if stdIdx%6 == 0 {
						s.Bubble = "Sedang fokus belajar di kelas"
					}
				}
				roomCounts[s.CurrentRoom]++
			}

			studentList = append(studentList, s)
			stdIdx++
		}

		// 3. Staff / Teachers
		var staffList []StaffVirtualState
		staffRows, err := db.Query(`
			SELECT 
				s.id, s.name, s.role, s.gender,
				COALESCE(c.name, '') as assigned_class,
				COALESCE(sub.name, '') as assigned_subject
			FROM staff s
			LEFT JOIN teacher_assignments ta ON s.id = ta.staff_id
			LEFT JOIN classes c ON ta.class_id = c.id
			LEFT JOIN subjects sub ON ta.subject_id = sub.id
			ORDER BY s.id ASC
		`)

		if err == nil {
			defer staffRows.Close()
			for staffRows.Next() {
				var st StaffVirtualState
				if err := staffRows.Scan(&st.ID, &st.Name, &st.Role, &st.Gender, &st.AssignedClass, &st.AssignedSubject); err == nil {
					st.CharacterID = "pak-joko"
					if st.Gender == "P" || strings.Contains(strings.ToLower(st.Name), "sri") || strings.Contains(strings.ToLower(st.Name), "ibu") {
						st.CharacterID = "bu-sri"
					}

					// Position staff based on period
					switch period.ID {
					case "class_1", "class_2", "class_3":
						if st.AssignedClass != "" {
							st.CurrentRoom = mapClassToRoom(st.AssignedClass)
							st.Action = fmt.Sprintf("Mengajar %s", st.AssignedSubject)
							st.Bubble = fmt.Sprintf("Menerangkan materi %s", st.AssignedSubject)
						} else {
							st.CurrentRoom = "r-guru"
							st.Action = "Persiapan Bahan Ajar"
							st.Bubble = "Menyiapkan lembar kerja siswa"
						}
					case "break_mbg", "break_prayer":
						st.CurrentRoom = "r-guru"
						st.Action = "Rehat & Istirahat"
						st.Bubble = "Rehat kopi dan koordinasi guru"
					default: // arrival, ceremony, departure, closed
						st.CurrentRoom = "r-guru"
						st.Action = "Di Ruang Guru"
					}
					roomCounts[st.CurrentRoom]++
					staffList = append(staffList, st)
				}
			}
		}

		// Ensure Pak Joko, Bu Sri, and Kang Asep are always present in virtual school
		hasKepsek := false
		hasMart := false
		for _, st := range staffList {
			if strings.Contains(strings.ToLower(st.Role), "kepala sekolah") || strings.Contains(strings.ToLower(st.Name), "sri") {
				hasKepsek = true
			}
			if strings.Contains(strings.ToLower(st.Name), "asep") {
				hasMart = true
			}
		}

		if !hasKepsek {
			kepsek := StaffVirtualState{
				ID:          991,
				Name:        "Dra. Hj. Sri Wahyuni",
				Role:        "Kepala Sekolah",
				Gender:      "P",
				CurrentRoom: "r-kepsek",
				CharacterID: "bu-sri",
				Action:      "Memantau KBM & Administrasi",
				Bubble:      "Memastikan pembelajaran berjalan tertib",
			}
			roomCounts[kepsek.CurrentRoom]++
			staffList = append(staffList, kepsek)
		}

		if !hasMart {
			mart := StaffVirtualState{
				ID:          992,
				Name:        "Kang Asep",
				Role:        "Pengelola NIBA Mart",
				Gender:      "L",
				CurrentRoom: "niba-mart",
				CharacterID: "kang-asep",
				Action:      "Melayani Kasir NIBA Mart",
				Bubble:      "Silakan, snack dan ATK lengkap!",
			}
			roomCounts[mart.CurrentRoom]++
			staffList = append(staffList, mart)
		}

		// 4. Build room occupancy list
		roomNames := map[string]string{
			"kelas-x-mplb":     "Kelas X MPLB",
			"kelas-x-pm":       "Kelas X PM",
			"kelas-xi-mplb":    "Kelas XI MPLB",
			"kelas-xi-pm":      "Kelas XI PM",
			"lab-komputer":     "Lab Komputer",
			"r-guru":           "Ruang Guru",
			"r-kepsek":         "Ruang Kepala Sekolah",
			"r-tamu":           "Ruang Tamu",
			"r-tu":             "Ruang Tata Usaha",
			"niba-mart":        "NIBA Mart",
			"lapangan-upacara": "Lapangan Upacara",
			"gudang":           "Gudang",
			"toilet-wudhu":     "Toilet & Tempat Wudhu",
		}

		var roomList []RoomOccupancy
		for rID, rName := range roomNames {
			roomList = append(roomList, RoomOccupancy{
				ID:             rID,
				Name:           rName,
				OccupantsCount: roomCounts[rID],
			})
		}

		// 5. Build recent RFID / Face attendance taps
		type RecentTapItem struct {
			StudentID   int    `json:"student_id"`
			StudentName string `json:"student_name"`
			ClassName   string `json:"class_name"`
			Status      string `json:"status"`
			Time        string `json:"time"`
			Method      string `json:"method"`
		}

		var recentTaps []RecentTapItem
		tapRows, tapErr := db.Query(`
			SELECT COALESCE(s.id, 0), COALESCE(s.name, a.user_name), COALESCE(c.name, 'Umum'), a.status, a.timestamp, COALESCE(a.method, 'RFID Gate')
			FROM attendance_logs a
			LEFT JOIN students s ON (a.student_id > 0 AND a.student_id = s.id) OR (a.rfid_uid != '' AND a.rfid_uid = s.rfid_uid)
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE a.date = ? OR a.timestamp LIKE ?
			ORDER BY a.id DESC
			LIMIT 10
		`, todayDate, todayDate+"%")

		if tapErr == nil {
			defer tapRows.Close()
			for tapRows.Next() {
				var sid int
				var sname, cname, stat, ts, meth string
				if err := tapRows.Scan(&sid, &sname, &cname, &stat, &ts, &meth); err == nil {
					tPart := ts
					if len(ts) >= 16 {
						tPart = ts[11:16]
					}
					recentTaps = append(recentTaps, RecentTapItem{
						StudentID:   sid,
						StudentName: sname,
						ClassName:   cname,
						Status:      stat,
						Time:        tPart,
						Method:      meth,
					})
				}
			}
		}

		// Fallback for simulation mode or empty morning attendance
		if len(recentTaps) == 0 {
			for _, st := range studentList {
				if len(recentTaps) < 8 {
					tVal := st.AttendanceTime
					if tVal == "" {
						tVal = "06:45"
					}
					method := "RFID Gate"
					if st.ID%3 == 0 {
						method = "Face AI"
					}
					stat := st.Status
					if stat == "" || stat == "Belum Hadir" {
						stat = "Hadir"
					}
					recentTaps = append(recentTaps, RecentTapItem{
						StudentID:   st.ID,
						StudentName: st.Name,
						ClassName:   st.ClassName,
						Status:      stat,
						Time:        tVal,
						Method:      method,
					})
				}
			}
		}

		dayNames := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
		dayName := dayNames[int(now.Weekday())]

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"meta": map[string]interface{}{
				"server_time":        now.Format("15:04:05"),
				"active_time":        activeTime,
				"date":               todayDate,
				"day_name":           dayName,
				"is_simulated":       isSimMode,
				"period_id":          period.ID,
				"period_name":        period.Name,
				"period_activity":    period.Activity,
				"period_description": period.Description,
				"next_bell":          nextBell,
			},
			"stats": map[string]interface{}{
				"total_students": len(studentList),
				"present_count":  presentCount,
				"late_count":     lateCount,
				"absent_count":   absentCount,
				"total_staff":    len(staffList),
			},
			"schedule_periods": defaultPeriods,
			"students":         studentList,
			"staff":            staffList,
			"rooms":            roomList,
			"recent_taps":      recentTaps,
		})
	}
}

// GetVirtualSchoolBells returns all bell schedule items
func GetVirtualSchoolBells(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type BellItem struct {
			ID        int    `json:"id"`
			Time      string `json:"time"`
			Label     string `json:"label"`
			AudioFile string `json:"audio_file,omitempty"`
		}

		rows, err := db.Query("SELECT id, time, label, COALESCE(audio_file, '') FROM schedules ORDER BY time ASC")
		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "success",
				"bells":   defaultPeriods,
				"message": "Using standard default periods",
			})
		}
		defer rows.Close()

		var bells []BellItem
		for rows.Next() {
			var b BellItem
			if err := rows.Scan(&b.ID, &b.Time, &b.Label, &b.AudioFile); err == nil {
				bells = append(bells, b)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":   "success",
			"bells":    bells,
			"defaults": defaultPeriods,
		})
	}
}

// VirtualSchoolAddPointRequest represents payload to award or deduct points
type VirtualSchoolAddPointRequest struct {
	StudentID    int    `json:"student_id"`
	PointsChange int    `json:"points_change"`
	Description  string `json:"description"`
	RecordedBy   string `json:"recorded_by"`
}

// AddVirtualSchoolPoint handles direct point mutation from virtual school
func AddVirtualSchoolPoint(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req VirtualSchoolAddPointRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Invalid request payload",
			})
		}

		if req.StudentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "student_id is required",
			})
		}

		if req.Description == "" {
			if req.PointsChange >= 0 {
				req.Description = "Poin Apresiasi KBM"
			} else {
				req.Description = "Catatan Pelanggaran Tata Tertib"
			}
		}

		if req.RecordedBy == "" {
			req.RecordedBy = "Guru / Wali Kelas"
		}

		nowStr := time.Now().Format("2006-01-02T15:04:05Z")

		// Insert into student_points
		res, err := db.Exec(`
			INSERT INTO student_points (student_id, points_change, description, timestamp, recorded_by)
			VALUES (?, ?, ?, ?, ?)
		`, req.StudentID, req.PointsChange, req.Description, nowStr, req.RecordedBy)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal menyimpan poin: %v", err),
			})
		}

		lastID, _ := res.LastInsertId()

		// Get updated total points for student
		var totalPoints int
		_ = db.QueryRow("SELECT COALESCE(SUM(points_change), 0) FROM student_points WHERE student_id = ?", req.StudentID).Scan(&totalPoints)
		basePoints := 100 + (req.StudentID*13)%45
		finalTotal := basePoints + totalPoints

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":        "success",
			"message":       fmt.Sprintf("Berhasil mencatat poin (%+d): %s", req.PointsChange, req.Description),
			"point_id":      lastID,
			"student_id":    req.StudentID,
			"points_change": req.PointsChange,
			"total_points":  finalTotal,
			"timestamp":     nowStr,
		})
	}
}

// VirtualSchoolAttendanceRequest represents manual attendance mutation
type VirtualSchoolAttendanceRequest struct {
	StudentID int    `json:"student_id"`
	Status    string `json:"status"` // "Hadir", "Terlambat", "Pulang", "Sakit", "Izin"
	Note      string `json:"note"`
}

// UpdateVirtualSchoolAttendance records or updates student attendance from virtual school
func UpdateVirtualSchoolAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req VirtualSchoolAttendanceRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Invalid request payload",
			})
		}

		if req.StudentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "student_id is required",
			})
		}

		if req.Status == "" {
			req.Status = "Hadir"
		}

		// Find student name and class
		var studentName, nis string
		var classID int
		err := db.QueryRow("SELECT name, COALESCE(nis, ''), COALESCE(class_id, 0) FROM students WHERE id = ?", req.StudentID).Scan(&studentName, &nis, &classID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"status":  "error",
				"message": "Siswa tidak ditemukan",
			})
		}

		now := time.Now()
		todayDate := now.Format("2006-01-02")
		timeStr := now.Format("15:04:05")
		fullTs := todayDate + "T" + timeStr + "Z"

		// Insert or replace attendance log
		_, err = db.Exec(`
			INSERT INTO attendance_logs (user_name, user_type, status, timestamp, date, method, student_id)
			VALUES (?, 'Siswa', ?, ?, ?, 'Virtual School Manual', ?)
		`, studentName, req.Status, fullTs, todayDate, req.StudentID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal mencatat presensi: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":          "success",
			"message":         fmt.Sprintf("Presensi %s berhasil dicatat sebagai '%s'", studentName, req.Status),
			"student_id":      req.StudentID,
			"student_name":    studentName,
			"status_presensi": req.Status,
			"time":            timeStr[:5],
			"timestamp":       fullTs,
		})
	}
}
