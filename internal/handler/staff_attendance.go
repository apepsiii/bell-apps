package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// ManualStaffAttendance handles manual attendance recording for teachers and staff.
// Endpoint: POST /admin/attendance/staff-manual
func ManualStaffAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		staffIDStr := c.FormValue("staff_id")
		status := strings.TrimSpace(c.FormValue("status")) // Hadir, Terlambat, Sakit, Izin, Alpha
		timeStr := strings.TrimSpace(c.FormValue("time"))
		note := strings.TrimSpace(c.FormValue("note"))

		staffID, err := strconv.Atoi(staffIDStr)
		if err != nil || staffID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID Guru/Staff wajib diisi"})
		}

		if status == "" {
			status = "Hadir"
		}

		// 1. Get Staff Info
		var name, role string
		var rfidNull sql.NullString
		err = db.QueryRow("SELECT name, COALESCE(rfid_uid, ''), role FROM staff WHERE id = ?", staffID).Scan(&name, &rfidNull, &role)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Data guru/staff tidak ditemukan"})
		}
		rfid := rfidNull.String

		// 2. Prepare timestamp & date
		now := time.Now()
		dateStr := now.Format("2006-01-02")
		var timestamp string

		if (status == "Hadir" || status == "Terlambat") && timeStr != "" {
			if _, err := time.Parse("15:04", timeStr); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format jam tidak valid (gunakan format JJ:MM)"})
			}
			timestamp = fmt.Sprintf("%s %s:00", dateStr, timeStr)
		} else {
			timestamp = now.Format("2006-01-02 15:04:05")
		}

		// 3. Insert Log into attendance_logs
		_, err = db.Exec(`
			INSERT INTO attendance_logs (student_id, rfid_uid, user_name, user_type, status, method, timestamp, date, note)
			VALUES (0, ?, ?, 'Staff', ?, 'MANUAL', ?, ?, ?)`,
			rfid, name, status, timestamp, dateStr, note)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan presensi staff: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Presensi %s (%s) berhasil dicatat sebagai %s", name, role, status),
		})
	}
}

// ExportStaffAttendance exports staff attendance logs as a CSV file.
// Endpoint: GET /admin/attendance/staff-export
func ExportStaffAttendance(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		month := c.QueryParam("month")
		year := c.QueryParam("year")
		date := c.QueryParam("date")

		now := time.Now()
		if year == "" {
			year = now.Format("2006")
		}
		if month == "" && date == "" {
			month = now.Format("01")
		}

		var query string
		var args []interface{}

		if date != "" {
			query = `
				SELECT s.nip, s.name, s.role, al.date, al.status, al.method, al.timestamp, COALESCE(al.note, '')
				FROM staff s
				LEFT JOIN attendance_logs al ON ((s.rfid_uid != '' AND s.rfid_uid = al.rfid_uid) OR s.name = al.user_name)
					AND al.user_type = 'Staff' AND al.date = ?
				ORDER BY s.name ASC, al.timestamp ASC`
			args = append(args, date)
		} else {
			// Ensure two-digit month
			if len(month) == 1 {
				month = "0" + month
			}
			query = `
				SELECT s.nip, s.name, s.role, al.date, al.status, al.method, al.timestamp, COALESCE(al.note, '')
				FROM attendance_logs al
				JOIN staff s ON ((al.rfid_uid != '' AND al.rfid_uid = s.rfid_uid) OR al.user_name = s.name)
				WHERE al.user_type = 'Staff'
				  AND strftime('%m', al.date) = ?
				  AND strftime('%Y', al.date) = ?
				ORDER BY al.date ASC, s.name ASC, al.timestamp ASC`
			args = append(args, month, year)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal query data presensi: "+err.Error())
		}
		defer rows.Close()

		filename := fmt.Sprintf("presensi_guru_staff_%s_%s.csv", year, month)
		if date != "" {
			filename = fmt.Sprintf("presensi_guru_staff_%s.csv", date)
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/csv; charset=utf-8")
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
		c.Response().WriteHeader(http.StatusOK)

		// Write UTF-8 BOM for Microsoft Excel compatibility
		c.Response().Write([]byte("\xEF\xBB\xBF"))

		writer := csv.NewWriter(c.Response().Writer)
		defer writer.Flush()

		// Write CSV Header
		writer.Write([]string{"No", "NIP / ID", "Nama Guru / Staff", "Jabatan", "Tanggal", "Status", "Metode", "Waktu", "Catatan"})

		no := 1
		for rows.Next() {
			var nip, name, role, logDate, status, method, ts, note string
			var nipNull, dateNull, statusNull, methodNull, tsNull sql.NullString
			if err := rows.Scan(&nipNull, &name, &role, &dateNull, &statusNull, &methodNull, &tsNull, &note); err != nil {
				continue
			}
			nip = nipNull.String
			logDate = dateNull.String
			status = statusNull.String
			method = methodNull.String
			ts = tsNull.String

			timeOnly := "-"
			if len(ts) >= 16 {
				timeOnly = ts[11:16]
			}

			if status == "" {
				status = "Belum Hadir"
			}
			if logDate == "" && date != "" {
				logDate = date
			}

			writer.Write([]string{
				strconv.Itoa(no),
				nip,
				name,
				role,
				logDate,
				status,
				method,
				timeOnly,
				note,
			})
			no++
		}

		return nil
	}
}

