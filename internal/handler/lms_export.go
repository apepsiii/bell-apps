package handler

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

func setExcelHeaders(c echo.Context, filename string) {
	h := c.Response().Header()
	h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	h.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

// ExportAssignmentGradesExcel exports per-student scores for an assignment.
func ExportAssignmentGradesExcel(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		assID, _ := strconv.Atoi(c.Param("id"))
		var title, className, subjectName string
		var maxPoints float64
		err := db.QueryRow(`SELECT a.title, a.max_points, c.name, s.name
			FROM lms_assignments a
			JOIN lms_meetings m ON m.id=a.meeting_id
			JOIN teacher_assignments ta ON ta.id=m.assignment_id
			JOIN classes c ON c.id=ta.class_id
			JOIN subjects s ON s.id=ta.subject_id
			WHERE a.id=? AND ta.staff_id=?`, assID, teacherID).Scan(&title, &maxPoints, &className, &subjectName)
		if err != nil {
			return c.JSON(403, map[string]string{"message": "Akses ditolak atau tugas tidak ditemukan"})
		}
		rows, err := db.Query(`SELECT st.nis, st.name, COALESCE(sub.score,-1), COALESCE(sub.feedback,''), COALESCE(sub.submitted_at,''), COALESCE(sub.is_late,0)
			FROM students st
			LEFT JOIN lms_submissions sub ON sub.assignment_id=? AND sub.student_id=st.id
			WHERE st.status='active' AND st.class_id=(SELECT ta.class_id FROM lms_assignments a JOIN lms_meetings m ON m.id=a.meeting_id JOIN teacher_assignments ta ON ta.id=m.assignment_id WHERE a.id=?)
			ORDER BY st.name ASC`, assID, assID)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal membaca data nilai"})
		}
		defer rows.Close()

		f := excelize.NewFile()
		defer f.Close()
		sheet := "Nilai Tugas"
		f.SetSheetName("Sheet1", sheet)
		headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"18181B"}, Pattern: 1}})
		titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
		f.SetCellValue(sheet, "A1", "Rekap Nilai Tugas")
		f.SetCellStyle(sheet, "A1", "A1", titleStyle)
		f.SetCellValue(sheet, "A2", subjectName+" — "+className)
		f.SetCellValue(sheet, "A3", fmt.Sprintf("%s (Max %.0f)", title, maxPoints))
		headers := []string{"NIS", "Nama Siswa", "Nilai", "Status", "Terlambat", "Waktu Submit", "Catatan Guru"}
		for i, htext := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 5)
			f.SetCellValue(sheet, cell, htext)
			f.SetCellStyle(sheet, cell, cell, headerStyle)
		}
		row := 6
		for rows.Next() {
			var nis, name, feedback, submittedAt string
			var score float64
			var late int
			if rows.Scan(&nis, &name, &score, &feedback, &submittedAt, &late) != nil {
				continue
			}
			status := "Belum Mengumpulkan"
			if score >= 0 {
				status = "Sudah Dinilai"
			} else if submittedAt != "" {
				status = "Menunggu Penilaian"
			}
			lateText := "Tepat Waktu"
			if late == 1 {
				lateText = "Terlambat"
			}
			vals := []interface{}{nis, name, score, status, lateText, submittedAt, feedback}
			for i, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
		f.SetColWidth(sheet, "A", "G", 18)
		f.SetColWidth(sheet, "B", "B", 28)
		f.SetColWidth(sheet, "G", "G", 40)
		setExcelHeaders(c, "Nilai_Tugas_"+time.Now().Format("20060102_150405")+".xlsx")
		return f.Write(c.Response().Writer)
	}
}

// ExportMeetingAttendanceExcel exports attendance recap for a meeting.
func ExportMeetingAttendanceExcel(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		teacherID := TeacherIDFromCtx(c)
		meetingID, _ := strconv.Atoi(c.Param("meeting_id"))
		var title, className, subjectName, meetingDate string
		err := db.QueryRow(`SELECT m.title, c.name, s.name, m.meeting_date
			FROM lms_meetings m
			JOIN teacher_assignments ta ON ta.id=m.assignment_id
			JOIN classes c ON c.id=ta.class_id
			JOIN subjects s ON s.id=ta.subject_id
			WHERE m.id=? AND ta.staff_id=?`, meetingID, teacherID).Scan(&title, &className, &subjectName, &meetingDate)
		if err != nil {
			return c.JSON(403, map[string]string{"message": "Akses ditolak atau pertemuan tidak ditemukan"})
		}
		rows, err := db.Query(`SELECT st.nis, st.name, COALESCE(a.status,'H'), COALESCE(a.notes,'')
			FROM students st
			LEFT JOIN lms_attendances a ON a.meeting_id=? AND a.student_id=st.id
			WHERE st.status='active' AND st.class_id=(SELECT ta.class_id FROM lms_meetings m JOIN teacher_assignments ta ON ta.id=m.assignment_id WHERE m.id=?)
			ORDER BY st.name ASC`, meetingID, meetingID)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal membaca data presensi"})
		}
		defer rows.Close()

		f := excelize.NewFile()
		defer f.Close()
		sheet := "Presensi KBM"
		f.SetSheetName("Sheet1", sheet)
		headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"18181B"}, Pattern: 1}})
		titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
		f.SetCellValue(sheet, "A1", "Rekap Presensi KBM")
		f.SetCellStyle(sheet, "A1", "A1", titleStyle)
		f.SetCellValue(sheet, "A2", subjectName+" — "+className)
		f.SetCellValue(sheet, "A3", title+" • "+meetingDate)
		headers := []string{"NIS", "Nama Siswa", "Status (H/S/I/A)", "Catatan"}
		for i, htext := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 5)
			f.SetCellValue(sheet, cell, htext)
			f.SetCellStyle(sheet, cell, cell, headerStyle)
		}
		row := 6
		for rows.Next() {
			var nis, name, status, notes string
			if rows.Scan(&nis, &name, &status, &notes) != nil {
				continue
			}
			vals := []interface{}{nis, name, status, notes}
			for i, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
		f.SetColWidth(sheet, "A", "D", 20)
		f.SetColWidth(sheet, "B", "B", 28)
		setExcelHeaders(c, "Presensi_KBM_"+time.Now().Format("20060102_150405")+".xlsx")
		return f.Write(c.Response().Writer)
	}
}
