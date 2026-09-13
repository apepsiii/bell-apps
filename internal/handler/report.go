package handler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v4"
)

type ReportRecord struct {
	No              int
	ID              string
	Name            string
	ClassOrRole     string
	Status          string
	Time            string
	TimeOut         string
	Note            string
	PresentCount    int
	LateCount       int
	SickCount       int
	PermissionCount int
	AbsentCount     int
	AttendanceRate  float64
}

type ReportData struct {
	Title           string
	Period          string
	GeneratedAt     string
	Type            string
	TotalRecords    int
	TotalPresent    int
	TotalLate       int
	TotalSick       int
	TotalPermission int
	TotalAbsent     int
	AttendanceRate  float64
	Records         []ReportRecord
}

func DailyReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		date := c.QueryParam("date")
		reportType := c.QueryParam("type")
		classID := c.QueryParam("class_id")
		format := c.QueryParam("format")

		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid date format. Use YYYY-MM-DD"})
		}

		if reportType != "student" && reportType != "staff" {
			return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
		}

		data := queryDailyReport(db, date, reportType, classID)

		if format == "json" {
			return c.JSON(200, data)
		}

		pdf := generateDailyReportPDF(data)

		filename := fmt.Sprintf("Laporan_Harian_%s_%s.pdf", reportType, date)
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		return pdf.Output(c.Response().Writer)
	}
}

func WeeklyReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		startDate := c.QueryParam("start")
		endDate := c.QueryParam("end")
		reportType := c.QueryParam("type")
		classID := c.QueryParam("class_id")
		format := c.QueryParam("format")

		_, err1 := time.Parse("2006-01-02", startDate)
		_, err2 := time.Parse("2006-01-02", endDate)
		if err1 != nil || err2 != nil {
			return c.JSON(400, map[string]string{"error": "Invalid date format. Use YYYY-MM-DD"})
		}

		if reportType != "student" && reportType != "staff" {
			return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
		}

		data := queryPeriodReport(db, startDate, endDate, reportType, classID)
		data.Title = "Laporan Kehadiran Mingguan"

		if format == "json" {
			return c.JSON(200, data)
		}

		pdf := generatePeriodReportPDF(data)

		filename := fmt.Sprintf("Laporan_Mingguan_%s_%s_to_%s.pdf", reportType, startDate, endDate)
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		return pdf.Output(c.Response().Writer)
	}
}

func MonthlyReport(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		month := c.QueryParam("month")
		reportType := c.QueryParam("type")
		classID := c.QueryParam("class_id")
		format := c.QueryParam("format")

		monthTime, err := time.Parse("2006-01", month)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid month format. Use YYYY-MM"})
		}

		if reportType != "student" && reportType != "staff" {
			return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
		}

		startDate := monthTime.Format("2006-01-02")
		endDate := monthTime.AddDate(0, 1, -1).Format("2006-01-02")

		data := queryPeriodReport(db, startDate, endDate, reportType, classID)
		data.Title = "Laporan Kehadiran Bulanan"
		data.Period = monthTime.Format("January 2006")

		if format == "json" {
			return c.JSON(200, data)
		}

		pdf := generatePeriodReportPDF(data)

		filename := fmt.Sprintf("Laporan_Bulanan_%s_%s.pdf", reportType, month)
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		return pdf.Output(c.Response().Writer)
	}
}

func calculateSchoolDays(db *sql.DB, startDate, endDate string) int {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	days := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		isWork, _ := IsWorkingDay(db, d)
		if isWork {
			days++
		}
	}
	return days
}

func queryDailyReport(db *sql.DB, date, reportType, classID string) ReportData {
	data := ReportData{
		Title:       "Laporan Kehadiran Harian",
		Period:      date,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Type:        reportType,
	}

	var query string
	var args []interface{}

	if reportType == "student" {
		query = `
			SELECT s.nis, s.name, c.name as class_name, 
			       COALESCE(al_in.status, 'Tidak Hadir') as status,
			       COALESCE(al_in.timestamp, '') as time_in,
			       COALESCE(al_out.timestamp, '') as time_out,
			       COALESCE(al_in.note, '') as note
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN attendance_logs al_in ON s.rfid_uid = al_in.rfid_uid 
			    AND DATE(al_in.timestamp) = ?
			    AND al_in.id = (
			        SELECT MAX(id) FROM attendance_logs 
			        WHERE rfid_uid = s.rfid_uid AND DATE(timestamp) = ? AND status != 'Pulang'
			    )
			LEFT JOIN attendance_logs al_out ON s.rfid_uid = al_out.rfid_uid 
			    AND DATE(al_out.timestamp) = ?
			    AND al_out.id = (
			        SELECT MAX(id) FROM attendance_logs 
			        WHERE rfid_uid = s.rfid_uid AND DATE(timestamp) = ? AND status = 'Pulang'
			    )
			WHERE s.status = 'active'
		`
		args = append(args, date, date, date, date)

		if classID != "" {
			query += " AND s.class_id = ?"
			args = append(args, classID)
		}

		query += " ORDER BY c.name, s.name"
	} else {
		query = `
			SELECT st.nip, st.name, st.role,
			       COALESCE(al_in.status, 'Tidak Hadir') as status,
			       COALESCE(al_in.timestamp, '') as time_in,
			       COALESCE(al_out.timestamp, '') as time_out,
			       COALESCE(al_in.note, '') as note
			FROM staff st
			LEFT JOIN attendance_logs al_in ON st.rfid_uid = al_in.rfid_uid 
			    AND DATE(al_in.timestamp) = ?
			    AND al_in.id = (
			        SELECT MAX(id) FROM attendance_logs 
			        WHERE rfid_uid = st.rfid_uid AND DATE(timestamp) = ? AND status != 'Pulang'
			    )
			LEFT JOIN attendance_logs al_out ON st.rfid_uid = al_out.rfid_uid 
			    AND DATE(al_out.timestamp) = ?
			    AND al_out.id = (
			        SELECT MAX(id) FROM attendance_logs 
			        WHERE rfid_uid = st.rfid_uid AND DATE(timestamp) = ? AND status = 'Pulang'
			    )
			ORDER BY st.name
		`
		args = append(args, date, date, date, date)
	}

	rows, _ := db.Query(query, args...)
	defer rows.Close()

	no := 1
	for rows.Next() {
		var record ReportRecord
		var status, timeIn, timeOut, note string

		if reportType == "student" {
			rows.Scan(&record.ID, &record.Name, &record.ClassOrRole, &status, &timeIn, &timeOut, &note)
		} else {
			rows.Scan(&record.ID, &record.Name, &record.ClassOrRole, &status, &timeIn, &timeOut, &note)
		}

		record.No = no
		record.Status = status
		record.Time = timeIn
		record.TimeOut = timeOut
		record.Note = note

		switch status {
		case "Datang", "Hadir":
			data.TotalPresent++
		case "Terlambat":
			data.TotalLate++
		case "Sakit", "Sakit (Dengan Surat)", "Sakit (Tanpa Surat)":
			data.TotalSick++
		case "Dispensasi", "Izin":
			data.TotalPermission++
		case "Alpha":
			data.TotalAbsent++
		case "Tidak Hadir":
			data.TotalAbsent++
		}

		data.Records = append(data.Records, record)
		no++
	}

	data.TotalRecords = len(data.Records)
	if data.TotalRecords > 0 {
		data.AttendanceRate = float64(data.TotalPresent+data.TotalLate) / float64(data.TotalRecords) * 100
	}

	return data
}

func queryPeriodReport(db *sql.DB, startDate, endDate, reportType, classID string) ReportData {
	data := ReportData{
		Title:       "Laporan Kehadiran Periode",
		Period:      startDate + " s/d " + endDate,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Type:        reportType,
	}

	var query string
	var args []interface{}

	if reportType == "student" {
		query = `
			SELECT s.nis, s.name, c.name as class_name,
			    SUM(CASE WHEN al.status IN ('Datang', 'Hadir') THEN 1 ELSE 0 END) as present,
			    SUM(CASE WHEN al.status = 'Terlambat' THEN 1 ELSE 0 END) as late,
			    SUM(CASE WHEN al.status = 'Sakit' THEN 1 ELSE 0 END) as sick,
			    SUM(CASE WHEN al.status = 'Izin' THEN 1 ELSE 0 END) as permission,
			    SUM(CASE WHEN al.status = 'Alpha' THEN 1 ELSE 0 END) as absent
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN attendance_logs al ON s.rfid_uid = al.rfid_uid 
			    AND DATE(al.timestamp) BETWEEN ? AND ?
			WHERE s.status = 'active'
		`
		args = append(args, startDate, endDate)

		if classID != "" {
			query += " AND s.class_id = ?"
			args = append(args, classID)
		}

		query += " GROUP BY s.id ORDER BY c.name, s.name"
	} else {
		query = `
			SELECT st.nip, st.name, st.role,
			    SUM(CASE WHEN al.status IN ('Datang', 'Hadir') THEN 1 ELSE 0 END) as present,
			    SUM(CASE WHEN al.status = 'Terlambat' THEN 1 ELSE 0 END) as late,
			    SUM(CASE WHEN al.status = 'Sakit' THEN 1 ELSE 0 END) as sick,
			    SUM(CASE WHEN al.status = 'Izin' THEN 1 ELSE 0 END) as permission,
			    SUM(CASE WHEN al.status = 'Alpha' THEN 1 ELSE 0 END) as absent
			FROM staff st
			LEFT JOIN attendance_logs al ON st.rfid_uid = al.rfid_uid 
			    AND DATE(al.timestamp) BETWEEN ? AND ?
			GROUP BY st.id ORDER BY st.name
		`
		args = append(args, startDate, endDate)
	}

	rows, _ := db.Query(query, args...)
	defer rows.Close()

	no := 1
	totalDays := calculateSchoolDays(db, startDate, endDate)

	for rows.Next() {
		var record ReportRecord

		rows.Scan(&record.ID, &record.Name, &record.ClassOrRole,
			&record.PresentCount, &record.LateCount,
			&record.SickCount, &record.PermissionCount,
			&record.AbsentCount)

		record.No = no

		totalAttendance := record.PresentCount + record.LateCount
		if totalDays > 0 {
			record.AttendanceRate = float64(totalAttendance) / float64(totalDays) * 100
		}

		data.Records = append(data.Records, record)

		data.TotalPresent += record.PresentCount
		data.TotalLate += record.LateCount
		data.TotalSick += record.SickCount
		data.TotalPermission += record.PermissionCount
		data.TotalAbsent += record.AbsentCount

		no++
	}

	data.TotalRecords = len(data.Records)

	return data
}

func generateDailyReportPDF(data ReportData) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "LAPORAN KEHADIRAN HARIAN", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 6, "Tanggal: "+data.Period, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Tipe: "+reportTitleCase(data.Type), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(40, 7, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Hadir", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Terlambat", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Sakit", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Izin", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Alpha", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(40, 7, fmt.Sprintf("%d", data.TotalRecords), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", data.TotalPresent), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", data.TotalLate), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", data.TotalSick), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", data.TotalPermission), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", data.TotalAbsent), "1", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 220, 255)
	pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "NIS/NIP", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Nama", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Kelas/Role", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Waktu", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	for _, record := range data.Records {
		pdf.CellFormat(10, 6, fmt.Sprintf("%d", record.No), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, record.ID, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, record.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, record.ClassOrRole, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, record.Status, "1", 0, "C", false, 0, "")

		timeStr := ""
		if record.Time != "" && len(record.Time) >= 16 {
			timeStr = record.Time[11:16]
		}
		pdf.CellFormat(25, 6, timeStr, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 6, "Dicetak pada: "+data.GeneratedAt, "", 1, "L", false, 0, "")

	return pdf
}

func generatePeriodReportPDF(data ReportData) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, reportToUpper(data.Title), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 6, "Periode: "+data.Period, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Tipe: "+reportTitleCase(data.Type), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 220, 255)
	pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "NIS/NIP", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Nama", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Kelas/Role", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Hadir", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Terlambat", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Sakit", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Izin", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 7, "Alpha", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "% Kehadiran", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	for _, record := range data.Records {
		pdf.CellFormat(10, 6, fmt.Sprintf("%d", record.No), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, record.ID, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, record.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, record.ClassOrRole, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", record.PresentCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", record.LateCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", record.SickCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", record.PermissionCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", record.AbsentCount), "1", 0, "C", false, 0, "")

		if record.AttendanceRate < 80 {
			pdf.SetTextColor(255, 0, 0)
		}
		pdf.CellFormat(30, 6, fmt.Sprintf("%.1f%%", record.AttendanceRate), "1", 1, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 6, "Dicetak pada: "+data.GeneratedAt, "", 1, "L", false, 0, "")

	return pdf
}

func reportToUpper(s string) string {
	result := ""
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			result += string(c - 32)
		} else {
			result += string(c)
		}
	}
	return result
}

func reportTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
