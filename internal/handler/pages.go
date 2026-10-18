package handler

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
	"belsekolah/pkg/qrcode"
)

func ScanPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "scan.html", nil)
	}
}

func ScanFacePage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "attendance_face.html", nil)
	}
}

func ScanPrayerPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "scan_prayer.html", nil)
	}
}

type ProfileData struct {
	Type       string
	ID         int
	Name       string
	IdentityNo string
	ExtraInfo  string
	Phone      string
	RFID       string
	Photo      string
}

type StudentDetailViewData struct {
	PageTitle       string
	Username        string
	Role            string
	ID              int
	Name            string
	NIS             string
	NISSiswa        string
	ClassName       string
	ParentPhone     string
	ParentName      string
	RFID            string
	Photo           string
	Status          string
	Birthday        string
	TotalPoints     int
	AchievementPts  int
	ViolationPts    int
	AttendanceRate  float64
	WorkingDays     int
	PresentDays     int
	LateDays        int
	PermissionDays  int
	SickDays        int
	AlphaDays       int
	TotalAbsentDays int
	QRImage         string
}

func StudentProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID          int
			Name        string
			NIS         string
			NISSiswa    string
			ClassName   string
			ParentPhone string
			ParentName  string
			RFID        string
			Photo       string
			Status      string
			Birthday    string
			ClassID     int
		}

		err := db.QueryRow(`
			SELECT s.id, s.name, COALESCE(s.nis, ''), COALESCE(s.nis_siswa, ''), COALESCE(c.name, '-'),
			       COALESCE(s.parent_phone, ''), COALESCE(s.parent_name, ''), COALESCE(s.rfid_uid, ''),
			       COALESCE(s.photo, ''), COALESCE(s.status, 'active'), COALESCE(s.birthday, ''), COALESCE(s.class_id, 0)
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, id).Scan(
			&s.ID, &s.Name, &s.NIS, &s.NISSiswa, &s.ClassName,
			&s.ParentPhone, &s.ParentName, &s.RFID,
			&s.Photo, &s.Status, &s.Birthday, &s.ClassID,
		)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/academics")
		}

		// Calculate points (dual track + legacy)
		var achPts, violPts, legacyPts, englishXP int
		_ = db.QueryRow(`SELECT COALESCE(SUM(points), 0) FROM student_achievement_points WHERE student_id = ?`, s.ID).Scan(&achPts)
		_ = db.QueryRow(`SELECT COALESCE(SUM(points), 0) FROM student_violation_points WHERE student_id = ?`, s.ID).Scan(&violPts)
		_ = db.QueryRow(`SELECT COALESCE(SUM(points_change), 0) FROM student_points WHERE student_id = ?`, s.ID).Scan(&legacyPts)
		_ = db.QueryRow(`SELECT COALESCE(total_xp, 0) FROM english_streaks WHERE student_id = ?`, s.ID).Scan(&englishXP)

		totalPts := legacyPts + englishXP
		if totalPts == 0 && (achPts > 0 || violPts > 0) {
			totalPts = achPts - violPts
		}

		// Calculate attendance stats this month
		var presentDays, lateDays, permDays, sickDays, alphaDays int
		_ = db.QueryRow(`
			SELECT 
				COALESCE(COUNT(CASE WHEN status IN ('Datang', 'Hadir') THEN 1 END), 0),
				COALESCE(COUNT(CASE WHEN status = 'Terlambat' THEN 1 END), 0),
				COALESCE(COUNT(CASE WHEN status LIKE 'Izin%' THEN 1 END), 0),
				COALESCE(COUNT(CASE WHEN status LIKE 'Sakit%' THEN 1 END), 0),
				COALESCE(COUNT(CASE WHEN status = 'Alpha' THEN 1 END), 0)
			FROM attendance_logs
			WHERE (student_id = ? OR rfid_uid = ?) AND strftime('%Y-%m', date) = strftime('%Y-%m', 'now')
		`, s.ID, s.RFID).Scan(&presentDays, &lateDays, &permDays, &sickDays, &alphaDays)

		totalAttended := presentDays + lateDays
		totalAbsent := permDays + sickDays + alphaDays
		workingDays := totalAttended + totalAbsent
		var attRate float64
		if workingDays > 0 {
			attRate = (float64(totalAttended) / float64(workingDays)) * 100
		}

		// Generate QR card image
		qrImage, _ := qrcode.GenerateStudentCard(config.GetDomain(), s.RFID, s.ID, s.NIS, s.Name, s.ClassName)

		data := map[string]interface{}{
			"PageTitle":       "Detail Siswa - " + s.Name,
			"Username":        "Admin",
			"Role":            "Administrator",
			"ID":              s.ID,
			"Name":            s.Name,
			"NIS":             s.NIS,
			"NISSiswa":        s.NISSiswa,
			"ClassName":       s.ClassName,
			"ParentPhone":     s.ParentPhone,
			"ParentName":      s.ParentName,
			"RFID":            s.RFID,
			"Photo":           s.Photo,
			"Status":          s.Status,
			"Birthday":        s.Birthday,
			"TotalPoints":     totalPts,
			"AchievementPts":  achPts,
			"ViolationPts":    violPts,
			"AttendanceRate":  attRate,
			"WorkingDays":     workingDays,
			"PresentDays":     presentDays,
			"LateDays":        lateDays,
			"PermissionDays":  permDays,
			"SickDays":        sickDays,
			"AlphaDays":       alphaDays,
			"TotalAbsentDays": totalAbsent,
			"QRImage":         template.URL(qrImage),
		}

		return c.Render(http.StatusOK, "student_detail.html", data)
	}
}

func StaffProfile(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID         int
			Name       string
			IdentityNo string
			Role       string
			Phone      string
			RFID       string
		}

		err := db.QueryRow(`
			SELECT id, name, COALESCE(nip, '-'), COALESCE(role, '-'), COALESCE(phone, '-'), COALESCE(rfid_uid, '-')
			FROM staff WHERE id = ?`, id).Scan(&s.ID, &s.Name, &s.IdentityNo, &s.Role, &s.Phone, &s.RFID)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/academics")
		}

		data := ProfileData{
			Type:       "Guru/Staff",
			ID:         s.ID,
			Name:       s.Name,
			IdentityNo: s.IdentityNo,
			ExtraInfo:  s.Role,
			Phone:      s.Phone,
			RFID:       s.RFID,
		}
		return c.Render(http.StatusOK, "profile.html", data)
	}
}

func FaceRegisterPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID         int
			Name       string
			ExtraInfo  string
			Photo      string
			IdentityNo string
			RFID       string
			Phone      string
		}

		err := db.QueryRow(`
			SELECT s.id, s.name, c.name, s.photo, s.nis, s.rfid_uid, s.parent_phone
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?`, id).Scan(&s.ID, &s.Name, &s.ExtraInfo, &s.Photo, &s.IdentityNo, &s.RFID, &s.Phone)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/students")
		}

		data := map[string]interface{}{
			"ID":         s.ID,
			"Name":       s.Name,
			"ExtraInfo":  s.ExtraInfo,
			"Photo":      s.Photo,
			"IdentityNo": s.IdentityNo,
			"RFID":       s.RFID,
			"Phone":      s.Phone,
			"Type":       "Siswa",
		}

		return c.Render(http.StatusOK, "face_register.html", data)
	}
}

func StaffFaceRegisterPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s struct {
			ID     int
			Name   string
			Role   string
			NIP    string
			RFID   string
			Phone  string
			Gender string
		}

		err := db.QueryRow(`
			SELECT id, name, COALESCE(role, 'Staff'), COALESCE(nip, '-'), COALESCE(rfid_uid, ''), COALESCE(phone, ''), COALESCE(gender, 'L')
			FROM staff
			WHERE id = ?`, id).Scan(&s.ID, &s.Name, &s.Role, &s.NIP, &s.RFID, &s.Phone, &s.Gender)

		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/academics")
		}

		data := map[string]interface{}{
			"ID":         s.ID,
			"Name":       s.Name,
			"ExtraInfo":  s.Role,
			"Photo":      "",
			"IdentityNo": s.NIP,
			"RFID":       s.RFID,
			"Phone":      s.Phone,
			"Type":       "Staff",
			"Gender":     s.Gender,
		}

		return c.Render(http.StatusOK, "face_register.html", data)
	}
}
