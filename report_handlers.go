package main

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

// Handler for daily report
func (a *App) DailyReportHandler(c echo.Context) error {
	date := c.QueryParam("date")
	reportType := c.QueryParam("type") // "student" or "staff"
	classID := c.QueryParam("class_id")
	format := c.QueryParam("format") // "pdf" or "json"

	// Validate date format
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	// Validate report type
	if reportType != "student" && reportType != "staff" {
		return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
	}

	// Query data from database
	data := a.queryDailyReport(date, reportType, classID)

	// Return JSON preview if requested
	if format == "json" {
		return c.JSON(200, data)
	}

	// Generate PDF
	pdf := generateDailyPDF(data)

	// Set headers
	filename := fmt.Sprintf("Laporan_Harian_%s_%s.pdf",
		reportType, date)
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))

	return pdf.Output(c.Response().Writer)
}

// Handler for weekly report
func (a *App) WeeklyReportHandler(c echo.Context) error {
	startDate := c.QueryParam("start")
	endDate := c.QueryParam("end")
	reportType := c.QueryParam("type")
	classID := c.QueryParam("class_id")
	format := c.QueryParam("format")

	// Validate dates
	_, err1 := time.Parse("2006-01-02", startDate)
	_, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		return c.JSON(400, map[string]string{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	// Validate report type
	if reportType != "student" && reportType != "staff" {
		return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
	}

	// Query aggregated data
	data := a.queryPeriodReport(startDate, endDate, reportType, classID)
	data.Title = "Laporan Kehadiran Mingguan"

	// Return JSON preview if requested
	if format == "json" {
		return c.JSON(200, data)
	}

	// Generate PDF
	pdf := generatePeriodPDF(data)

	filename := fmt.Sprintf("Laporan_Mingguan_%s_%s_to_%s.pdf",
		reportType, startDate, endDate)
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))

	return pdf.Output(c.Response().Writer)
}

// Handler for monthly report
func (a *App) MonthlyReportHandler(c echo.Context) error {
	month := c.QueryParam("month") // Format: YYYY-MM
	reportType := c.QueryParam("type")
	classID := c.QueryParam("class_id")
	format := c.QueryParam("format")

	// Validate month format and parse to get start and end dates
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid month format. Use YYYY-MM"})
	}

	// Validate report type
	if reportType != "student" && reportType != "staff" {
		return c.JSON(400, map[string]string{"error": "Invalid type. Use 'student' or 'staff'"})
	}

	// Calculate start and end dates for the month
	startDate := monthTime.Format("2006-01-02")
	endDate := monthTime.AddDate(0, 1, -1).Format("2006-01-02") // Last day of month

	// Query aggregated data
	data := a.queryPeriodReport(startDate, endDate, reportType, classID)
	data.Title = "Laporan Kehadiran Bulanan"
	data.Period = monthTime.Format("January 2006")

	// Return JSON preview if requested
	if format == "json" {
		return c.JSON(200, data)
	}

	// Generate PDF
	pdf := generatePeriodPDF(data)

	filename := fmt.Sprintf("Laporan_Bulanan_%s_%s.pdf",
		reportType, month)
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))

	return pdf.Output(c.Response().Writer)
}
