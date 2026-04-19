package pdf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type ReportRecord struct {
	No              int
	ID              string
	Name            string
	ClassOrRole     string
	Status          string
	Time            string
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

func GenerateDailyPDF(data ReportData) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "LAPORAN KEHADIRAN HARIAN", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 6, "Tanggal: "+data.Period, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Tipe: "+strings.Title(data.Type), "", 1, "L", false, 0, "")
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
	pdf.CellFormat(40, 7, strconv.Itoa(data.TotalRecords), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, strconv.Itoa(data.TotalPresent), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, strconv.Itoa(data.TotalLate), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, strconv.Itoa(data.TotalSick), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, strconv.Itoa(data.TotalPermission), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, strconv.Itoa(data.TotalAbsent), "1", 1, "C", false, 0, "")
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
		pdf.CellFormat(10, 6, strconv.Itoa(record.No), "1", 0, "C", false, 0, "")
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

func GeneratePeriodPDF(data ReportData) *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, strings.ToUpper(data.Title), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 6, "Periode: "+data.Period, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Tipe: "+strings.Title(data.Type), "", 1, "L", false, 0, "")
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
		pdf.CellFormat(10, 6, strconv.Itoa(record.No), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, record.ID, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, record.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, record.ClassOrRole, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, strconv.Itoa(record.PresentCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, strconv.Itoa(record.LateCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, strconv.Itoa(record.SickCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, strconv.Itoa(record.PermissionCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 6, strconv.Itoa(record.AbsentCount), "1", 0, "C", false, 0, "")

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
