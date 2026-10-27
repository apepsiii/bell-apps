package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func GetAcademicCalendar(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := `SELECT id,academic_year,semester,title,category,start_date,end_date,COALESCE(description,''),COALESCE(color,''),is_published,COALESCE(created_by,''),created_at,updated_at FROM academic_calendar_events WHERE is_published=1`
		args := []interface{}{}
		if y := c.QueryParam("academic_year"); y != "" {
			q += " AND academic_year=?"
			args = append(args, y)
		}
		if s := c.QueryParam("start"); s != "" {
			q += " AND end_date>=?"
			args = append(args, s)
		}
		if e := c.QueryParam("end"); e != "" {
			q += " AND start_date<=?"
			args = append(args, e)
		}
		q += " ORDER BY start_date ASC,id ASC"
		rows, err := db.Query(q, args...)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal mengambil kalender akademik"})
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, sem, pub int
			var year, title, cat, start, end, desc, color, by, created, updated string
			if rows.Scan(&id, &year, &sem, &title, &cat, &start, &end, &desc, &color, &pub, &by, &created, &updated) == nil {
				out = append(out, map[string]interface{}{"id": id, "academic_year": year, "semester": sem, "title": title, "category": cat, "start_date": start, "end_date": end, "description": desc, "color": color, "is_published": pub == 1, "created_by": by, "created_at": created, "updated_at": updated})
			}
		}
		return c.JSON(200, out)
	}
}

func CreateAcademicCalendarEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		year := strings.TrimSpace(c.FormValue("academic_year"))
		title := strings.TrimSpace(c.FormValue("title"))
		cat := strings.TrimSpace(c.FormValue("category"))
		start := strings.TrimSpace(c.FormValue("start_date"))
		end := strings.TrimSpace(c.FormValue("end_date"))
		if year == "" || title == "" || start == "" || end == "" {
			return c.JSON(400, map[string]string{"message": "Tahun ajaran, judul, dan rentang tanggal wajib diisi"})
		}
		sd, e1 := time.Parse("2006-01-02", start)
		ed, e2 := time.Parse("2006-01-02", end)
		if e1 != nil || e2 != nil || ed.Before(sd) {
			return c.JSON(400, map[string]string{"message": "Rentang tanggal tidak valid"})
		}
		sem, _ := strconv.Atoi(c.FormValue("semester"))
		if sem < 1 || sem > 2 {
			sem = 1
		}
		if cat == "" {
			cat = "other"
		}
		pub := 1
		if c.FormValue("is_published") == "0" {
			pub = 0
		}
		by, _ := c.Get("user").(string)
		res, err := db.Exec(`INSERT INTO academic_calendar_events (academic_year,semester,title,category,start_date,end_date,description,color,is_published,created_by) VALUES (?,?,?,?,?,?,?,?,?,?)`, year, sem, title, cat, start, end, c.FormValue("description"), c.FormValue("color"), pub, by)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menyimpan agenda kalender"})
		}
		id, _ := res.LastInsertId()
		return c.JSON(201, map[string]interface{}{"id": id, "message": "Agenda kalender berhasil dibuat"})
	}
}

func UpdateAcademicCalendarEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		start := c.FormValue("start_date")
		end := c.FormValue("end_date")
		sd, e1 := time.Parse("2006-01-02", start)
		ed, e2 := time.Parse("2006-01-02", end)
		if e1 != nil || e2 != nil || ed.Before(sd) {
			return c.JSON(400, map[string]string{"message": "Rentang tanggal tidak valid"})
		}
		_, err := db.Exec(`UPDATE academic_calendar_events SET academic_year=?,semester=?,title=?,category=?,start_date=?,end_date=?,description=?,color=?,is_published=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, c.FormValue("academic_year"), c.FormValue("semester"), c.FormValue("title"), c.FormValue("category"), start, end, c.FormValue("description"), c.FormValue("color"), c.FormValue("is_published") != "0", id)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal memperbarui agenda"})
		}
		return c.JSON(200, map[string]string{"message": "Agenda diperbarui"})
	}
}
func DeleteAcademicCalendarEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		res, err := db.Exec("DELETE FROM academic_calendar_events WHERE id=?", id)
		if err != nil {
			return c.JSON(500, map[string]string{"message": "Gagal menghapus agenda"})
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return c.JSON(404, map[string]string{"message": "Agenda tidak ditemukan"})
		}
		return c.JSON(200, map[string]string{"message": "Agenda dihapus"})
	}
}

var _ = http.StatusOK
