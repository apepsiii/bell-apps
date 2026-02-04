package main

import (
	"database/sql"
	"time"
)

// Helper function to calculate school days (exclude weekends)
func calculateSchoolDays(startDate, endDate string) int {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	days := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		// Exclude Saturday (6) and Sunday (0)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			days++
		}
	}
	return days
}

// Query function for daily report
func (a *App) queryDailyReport(date, reportType, classID string) ReportData {
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
			       al.status, al.timestamp
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN attendance_logs al ON s.rfid_uid = al.rfid_uid 
			    AND DATE(al.timestamp) = ?
			WHERE 1=1
		`
		args = append(args, date)

		if classID != "" {
			query += " AND s.class_id = ?"
			args = append(args, classID)
		}

		query += " ORDER BY c.name, s.name"
	} else {
		query = `
			SELECT st.nip, st.name, st.role,
			       al.status, al.timestamp
			FROM staff st
			LEFT JOIN attendance_logs al ON st.rfid_uid = al.rfid_uid 
			    AND DATE(al.timestamp) = ?
			ORDER BY st.name
		`
		args = append(args, date)
	}

	rows, _ := a.DB.Query(query, args...)
	defer rows.Close()

	no := 1
	for rows.Next() {
		var record ReportRecord
		var status, timestamp sql.NullString

		if reportType == "student" {
			rows.Scan(&record.ID, &record.Name, &record.ClassOrRole,
				&status, &timestamp)
		} else {
			rows.Scan(&record.ID, &record.Name, &record.ClassOrRole,
				&status, &timestamp)
		}

		record.No = no
		if status.Valid {
			record.Status = status.String
			record.Time = timestamp.String

			// Count statistics
			switch status.String {
			case "Datang":
				data.TotalPresent++
			case "Terlambat":
				data.TotalLate++
			case "Sakit":
				data.TotalSick++
			case "Izin":
				data.TotalPermission++
			case "Alpha":
				data.TotalAbsent++
			}
		} else {
			record.Status = "Tidak Hadir"
			data.TotalAbsent++
		}

		data.Records = append(data.Records, record)
		no++
	}

	data.TotalRecords = len(data.Records)
	if data.TotalRecords > 0 {
		data.AttendanceRate = float64(data.TotalPresent+data.TotalLate) /
			float64(data.TotalRecords) * 100
	}

	return data
}

// Query function for weekly/monthly report
func (a *App) queryPeriodReport(startDate, endDate, reportType, classID string) ReportData {
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
			    SUM(CASE WHEN al.status = 'Datang' THEN 1 ELSE 0 END) as present,
			    SUM(CASE WHEN al.status = 'Terlambat' THEN 1 ELSE 0 END) as late,
			    SUM(CASE WHEN al.status = 'Sakit' THEN 1 ELSE 0 END) as sick,
			    SUM(CASE WHEN al.status = 'Izin' THEN 1 ELSE 0 END) as permission,
			    SUM(CASE WHEN al.status = 'Alpha' THEN 1 ELSE 0 END) as absent
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			LEFT JOIN attendance_logs al ON s.rfid_uid = al.rfid_uid 
			    AND DATE(al.timestamp) BETWEEN ? AND ?
			WHERE 1=1
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
			    SUM(CASE WHEN al.status = 'Datang' THEN 1 ELSE 0 END) as present,
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

	rows, _ := a.DB.Query(query, args...)
	defer rows.Close()

	no := 1
	totalDays := calculateSchoolDays(startDate, endDate)

	for rows.Next() {
		var record ReportRecord

		rows.Scan(&record.ID, &record.Name, &record.ClassOrRole,
			&record.PresentCount, &record.LateCount,
			&record.SickCount, &record.PermissionCount,
			&record.AbsentCount)

		record.No = no

		// Calculate attendance rate
		totalAttendance := record.PresentCount + record.LateCount
		if totalDays > 0 {
			record.AttendanceRate = float64(totalAttendance) /
				float64(totalDays) * 100
		}

		data.Records = append(data.Records, record)

		// Aggregate statistics
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
