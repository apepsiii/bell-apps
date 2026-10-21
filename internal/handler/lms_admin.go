package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// DefaultBannerForSubject provides an intelligent preset banner based on subject attributes
func DefaultBannerForSubject(category, code, name string) string {
	lower := strings.ToLower(code + " " + name + " " + category)
	if strings.Contains(lower, "math") || strings.Contains(lower, "matematika") || strings.Contains(lower, "aljabar") || strings.Contains(lower, "kalkulus") {
		return "/assets/banners/math.svg"
	}
	if strings.Contains(lower, "rpl") || strings.Contains(lower, "pemrograman") || strings.Contains(lower, "web") || strings.Contains(lower, "jaringan") || strings.Contains(lower, "tkj") || strings.Contains(lower, "komputer") || strings.Contains(lower, "informatika") || strings.Contains(lower, "it") || strings.Contains(lower, "sijad") || strings.Contains(lower, "software") {
		return "/assets/banners/tech.svg"
	}
	if strings.Contains(lower, "bahasa") || strings.Contains(lower, "english") || strings.Contains(lower, "indonesia") || strings.Contains(lower, "jepang") || strings.Contains(lower, "mandarin") || strings.Contains(lower, "sastra") {
		return "/assets/banners/language.svg"
	}
	if strings.Contains(lower, "ipa") || strings.Contains(lower, "fisika") || strings.Contains(lower, "kimia") || strings.Contains(lower, "biologi") || strings.Contains(lower, "sains") {
		return "/assets/banners/science.svg"
	}
	if strings.Contains(lower, "mplb") || strings.Contains(lower, "otkp") || strings.Contains(lower, "bisnis") || strings.Contains(lower, "akuntansi") || strings.Contains(lower, "manajemen") || strings.Contains(lower, "ekonomi") || strings.Contains(lower, "pemasaran") || strings.Contains(lower, "perkantoran") {
		return "/assets/banners/business.svg"
	}
	if strings.Contains(lower, "seni") || strings.Contains(lower, "dkv") || strings.Contains(lower, "desain") || strings.Contains(lower, "multimedia") || strings.Contains(lower, "budaya") || strings.Contains(lower, "musik") || strings.Contains(lower, "rupa") {
		return "/assets/banners/art.svg"
	}
	if strings.Contains(lower, "agama") || strings.Contains(lower, "pai") || strings.Contains(lower, "islam") || strings.Contains(lower, "aqidah") || strings.Contains(lower, "qur'an") || strings.Contains(lower, "syariah") || strings.Contains(lower, "budi pekerti") {
		return "/assets/banners/religion.svg"
	}
	return "/assets/banners/general.svg"
}

// --- SUBJECTS (MATA PELAJARAN) HANDLERS ---

// GetSubjects returns the list of all subjects
func GetSubjects(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT s.id, s.code, s.name, s.major_id, COALESCE(m.name, 'Semua Jurusan / Umum') as major_name,
			       s.category, COALESCE(s.banner_url, ''), s.is_active, s.created_at
			FROM subjects s
			LEFT JOIN majors m ON m.id = s.major_id
			ORDER BY s.category ASC, s.name ASC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data mata pelajaran"})
		}
		defer rows.Close()

		subjects := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, isActive int
				code, name, majorName, category, bannerURL, createdAt string
				majorID sql.NullInt64
			)
			if err := rows.Scan(&id, &code, &name, &majorID, &majorName, &category, &bannerURL, &isActive, &createdAt); err == nil {
				if bannerURL == "" {
					bannerURL = DefaultBannerForSubject(category, code, name)
				}
				item := map[string]interface{}{
					"id":         id,
					"code":       code,
					"name":       name,
					"category":   category,
					"banner_url": bannerURL,
					"is_active":  isActive == 1,
					"created_at": createdAt,
					"major_name": majorName,
				}
				if majorID.Valid {
					item["major_id"] = majorID.Int64
				} else {
					item["major_id"] = nil
				}
				subjects = append(subjects, item)
			}
		}

		return c.JSON(http.StatusOK, subjects)
	}
}

// AddSubject creates a new subject
func AddSubject(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := strings.TrimSpace(c.FormValue("code"))
		name := strings.TrimSpace(c.FormValue("name"))
		category := strings.TrimSpace(c.FormValue("category"))
		majorIDStr := strings.TrimSpace(c.FormValue("major_id"))
		bannerURL := strings.TrimSpace(c.FormValue("banner_url"))

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kode dan nama mata pelajaran wajib diisi"})
		}
		if category == "" {
			category = "muatan_kejuruan"
		}
		if bannerURL == "" {
			bannerURL = DefaultBannerForSubject(category, code, name)
		}

		var majorID interface{}
		if majorIDStr != "" && majorIDStr != "0" {
			if id, err := strconv.Atoi(majorIDStr); err == nil && id > 0 {
				majorID = id
			}
		}

		res, err := db.Exec(`
			INSERT INTO subjects (code, name, major_id, category, banner_url, is_active)
			VALUES (?, ?, ?, ?, ?, 1)
		`, code, name, majorID, category, bannerURL)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kode mata pelajaran sudah digunakan"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan mata pelajaran"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message":    "Mata pelajaran berhasil ditambahkan",
			"id":         id,
			"banner_url": bannerURL,
		})
	}
}

// UpdateSubject updates an existing subject
func UpdateSubject(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
		}

		code := strings.TrimSpace(c.FormValue("code"))
		name := strings.TrimSpace(c.FormValue("name"))
		category := strings.TrimSpace(c.FormValue("category"))
		majorIDStr := strings.TrimSpace(c.FormValue("major_id"))
		bannerURL := strings.TrimSpace(c.FormValue("banner_url"))
		isActiveStr := strings.TrimSpace(c.FormValue("is_active"))

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kode dan nama mata pelajaran wajib diisi"})
		}
		if bannerURL == "" {
			bannerURL = DefaultBannerForSubject(category, code, name)
		}

		var majorID interface{}
		if majorIDStr != "" && majorIDStr != "0" {
			if mid, err := strconv.Atoi(majorIDStr); err == nil && mid > 0 {
				majorID = mid
			}
		}

		isActive := 1
		if isActiveStr == "0" || isActiveStr == "false" {
			isActive = 0
		}

		_, err = db.Exec(`
			UPDATE subjects
			SET code = ?, name = ?, major_id = ?, category = ?, banner_url = ?, is_active = ?
			WHERE id = ?
		`, code, name, majorID, category, bannerURL, isActive, id)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kode mata pelajaran sudah digunakan"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memperbarui mata pelajaran"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":    "Mata pelajaran berhasil diperbarui",
			"banner_url": bannerURL,
		})
	}
}

// DeleteSubject deletes a subject if no dependencies
func DeleteSubject(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
		}

		// Check if subject is assigned to any teacher
		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM teacher_assignments WHERE subject_id = ?", id).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Mata pelajaran tidak dapat dihapus karena masih digunakan pada penugasan guru.",
			})
		}

		_, err = db.Exec("DELETE FROM subjects WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus mata pelajaran"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Mata pelajaran berhasil dihapus"})
	}
}

// --- TEACHER ASSIGNMENTS (PENGAMPU KELAS) HANDLERS ---

// GetTeacherAssignments lists all teacher assignments
func GetTeacherAssignments(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		classID := c.QueryParam("class_id")
		staffID := c.QueryParam("staff_id")

		query := `
			SELECT ta.id, ta.staff_id, s.name as staff_name, COALESCE(s.nip, '') as staff_nip,
			       ta.subject_id, sub.name as subject_name, sub.code as subject_code, COALESCE(sub.category, 'muatan_kejuruan'),
			       ta.class_id, c.name as class_name, ta.academic_year, ta.semester,
			       COALESCE(NULLIF(ta.banner_url, ''), COALESCE(sub.banner_url, '')) as banner_url,
			       ta.created_at
			FROM teacher_assignments ta
			JOIN staff s ON s.id = ta.staff_id
			JOIN subjects sub ON sub.id = ta.subject_id
			JOIN classes c ON c.id = ta.class_id
			WHERE 1=1
		`
		var args []interface{}

		if classID != "" {
			query += " AND ta.class_id = ?"
			args = append(args, classID)
		}
		if staffID != "" {
			query += " AND ta.staff_id = ?"
			args = append(args, staffID)
		}

		query += " ORDER BY c.name ASC, sub.name ASC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data penugasan"})
		}
		defer rows.Close()

		assignments := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, sID, subID, cID, semester int
				staffName, staffNIP, subName, subCode, subCat, className, academicYear, bannerURL, createdAt string
			)
			if err := rows.Scan(&id, &sID, &staffName, &staffNIP, &subID, &subName, &subCode, &subCat, &cID, &className, &academicYear, &semester, &bannerURL, &createdAt); err == nil {
				if bannerURL == "" {
					bannerURL = DefaultBannerForSubject(subCat, subCode, subName)
				}
				assignments = append(assignments, map[string]interface{}{
					"id":            id,
					"staff_id":      sID,
					"staff_name":    staffName,
					"staff_nip":     staffNIP,
					"subject_id":    subID,
					"subject_name":  subName,
					"subject_code":  subCode,
					"category":      subCat,
					"class_id":      cID,
					"class_name":    className,
					"academic_year": academicYear,
					"semester":      semester,
					"banner_url":    bannerURL,
					"created_at":    createdAt,
				})
			}
		}

		return c.JSON(http.StatusOK, assignments)
	}
}

// AddTeacherAssignment assigns a teacher to a subject and class
func AddTeacherAssignment(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		staffID, err1 := strconv.Atoi(c.FormValue("staff_id"))
		subjectID, err2 := strconv.Atoi(c.FormValue("subject_id"))
		classID, err3 := strconv.Atoi(c.FormValue("class_id"))
		academicYear := strings.TrimSpace(c.FormValue("academic_year"))
		bannerURL := strings.TrimSpace(c.FormValue("banner_url"))
		semester, _ := strconv.Atoi(c.FormValue("semester"))

		if err1 != nil || err2 != nil || err3 != nil || staffID <= 0 || subjectID <= 0 || classID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Guru, mata pelajaran, dan kelas wajib dipilih"})
		}

		if academicYear == "" {
			academicYear = "2026/2027"
		}
		if semester <= 0 {
			semester = 1
		}

		res, err := db.Exec(`
			INSERT INTO teacher_assignments (staff_id, subject_id, class_id, academic_year, semester, banner_url)
			VALUES (?, ?, ?, ?, ?, ?)
		`, staffID, subjectID, classID, academicYear, semester, bannerURL)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Penugasan guru ini untuk kelas dan semester tersebut sudah ada",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan penugasan"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message":    "Penugasan guru berhasil ditambahkan",
			"id":         id,
			"banner_url": bannerURL,
		})
	}
}

// DeleteTeacherAssignment removes an assignment
func DeleteTeacherAssignment(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
		}

		_, err = db.Exec("DELETE FROM teacher_assignments WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus penugasan"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Penugasan berhasil dihapus"})
	}
}

