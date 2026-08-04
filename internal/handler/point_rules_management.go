package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// =====================
// ACHIEVEMENT RULES MANAGEMENT
// =====================

// AddAchievementRule creates a new achievement rule
func AddAchievementRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.FormValue("code")
		category := c.FormValue("category")
		categoryCode := c.FormValue("category_code")
		name := c.FormValue("name")
		description := c.FormValue("description")
		pointsStr := c.FormValue("points")
		minPointsStr := c.FormValue("min_points")
		maxPointsStr := c.FormValue("max_points")

		// Validation
		if code == "" || category == "" || categoryCode == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode, kategori, kode kategori, dan nama wajib diisi",
			})
		}

		points, _ := strconv.Atoi(pointsStr)
		minPoints, _ := strconv.Atoi(minPointsStr)
		maxPoints, _ := strconv.Atoi(maxPointsStr)

		if points < 0 || minPoints < 0 || maxPoints < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin tidak boleh negatif",
			})
		}

		if maxPoints > 0 && minPoints > maxPoints {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin minimal tidak boleh lebih besar dari poin maksimal",
			})
		}

		// Check for duplicate code
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM achievement_rules WHERE code = ?", code).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode rule sudah digunakan",
			})
		}

		// Insert
		_, err = db.Exec(`
			INSERT INTO achievement_rules (code, category, category_code, name, description, points, min_points, max_points, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, code, category, categoryCode, name, description, points, minPoints, maxPoints)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal menambahkan rule: " + err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule penghargaan berhasil ditambahkan",
		})
	}
}

// UpdateAchievementRule updates an existing achievement rule
func UpdateAchievementRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		code := c.FormValue("code")
		category := c.FormValue("category")
		categoryCode := c.FormValue("category_code")
		name := c.FormValue("name")
		description := c.FormValue("description")
		pointsStr := c.FormValue("points")
		minPointsStr := c.FormValue("min_points")
		maxPointsStr := c.FormValue("max_points")

		// Validation
		if code == "" || category == "" || categoryCode == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode, kategori, kode kategori, dan nama wajib diisi",
			})
		}

		points, _ := strconv.Atoi(pointsStr)
		minPoints, _ := strconv.Atoi(minPointsStr)
		maxPoints, _ := strconv.Atoi(maxPointsStr)

		if points < 0 || minPoints < 0 || maxPoints < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin tidak boleh negatif",
			})
		}

		if maxPoints > 0 && minPoints > maxPoints {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin minimal tidak boleh lebih besar dari poin maksimal",
			})
		}

		// Check for duplicate code (exclude current record)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM achievement_rules WHERE code = ? AND id != ?", code, id).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode rule sudah digunakan",
			})
		}

		// Update
		result, err := db.Exec(`
			UPDATE achievement_rules 
			SET code = ?, category = ?, category_code = ?, name = ?, description = ?, 
			    points = ?, min_points = ?, max_points = ?
			WHERE id = ?
		`, code, category, categoryCode, name, description, points, minPoints, maxPoints, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal mengupdate rule: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule penghargaan berhasil diupdate",
		})
	}
}

// DeleteAchievementRule soft-deletes an achievement rule
func DeleteAchievementRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if rule is being used
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM student_achievement_points WHERE rule_id = ? AND deleted_at IS NULL", id).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Rule ini sedang digunakan oleh %d transaksi. Nonaktifkan saja jika tidak ingin digunakan lagi.", count),
			})
		}

		// Soft delete (set is_active = 0)
		result, err := db.Exec("UPDATE achievement_rules SET is_active = 0 WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal menghapus rule: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule penghargaan berhasil dinonaktifkan",
		})
	}
}

// ToggleAchievementRuleStatus toggles the is_active status
func ToggleAchievementRuleStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		result, err := db.Exec("UPDATE achievement_rules SET is_active = 1 - is_active WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal mengubah status: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Status rule berhasil diubah",
		})
	}
}

// =====================
// VIOLATION RULES MANAGEMENT
// =====================

// AddViolationRule creates a new violation rule
func AddViolationRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.FormValue("code")
		category := c.FormValue("category")
		categoryCode := c.FormValue("category_code")
		name := c.FormValue("name")
		description := c.FormValue("description")
		points1Str := c.FormValue("points_1")
		points2Str := c.FormValue("points_2")
		points3Str := c.FormValue("points_3")

		// Validation
		if code == "" || category == "" || categoryCode == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode, kategori, kode kategori, dan nama wajib diisi",
			})
		}

		points1, _ := strconv.Atoi(points1Str)
		points2, _ := strconv.Atoi(points2Str)
		points3, _ := strconv.Atoi(points3Str)

		if points1 < 0 || points2 < 0 || points3 < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin tidak boleh negatif",
			})
		}

		// Check for duplicate code
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM violation_rules WHERE code = ?", code).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode rule sudah digunakan",
			})
		}

		// Insert
		_, err = db.Exec(`
			INSERT INTO violation_rules (code, category, category_code, name, description, points_1, points_2, points_3, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, code, category, categoryCode, name, description, points1, points2, points3)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal menambahkan rule: " + err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule pelanggaran berhasil ditambahkan",
		})
	}
}

// UpdateViolationRule updates an existing violation rule
func UpdateViolationRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		code := c.FormValue("code")
		category := c.FormValue("category")
		categoryCode := c.FormValue("category_code")
		name := c.FormValue("name")
		description := c.FormValue("description")
		points1Str := c.FormValue("points_1")
		points2Str := c.FormValue("points_2")
		points3Str := c.FormValue("points_3")

		// Validation
		if code == "" || category == "" || categoryCode == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode, kategori, kode kategori, dan nama wajib diisi",
			})
		}

		points1, _ := strconv.Atoi(points1Str)
		points2, _ := strconv.Atoi(points2Str)
		points3, _ := strconv.Atoi(points3Str)

		if points1 < 0 || points2 < 0 || points3 < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Poin tidak boleh negatif",
			})
		}

		// Check for duplicate code (exclude current record)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM violation_rules WHERE code = ? AND id != ?", code, id).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Kode rule sudah digunakan",
			})
		}

		// Update
		result, err := db.Exec(`
			UPDATE violation_rules 
			SET code = ?, category = ?, category_code = ?, name = ?, description = ?, 
			    points_1 = ?, points_2 = ?, points_3 = ?
			WHERE id = ?
		`, code, category, categoryCode, name, description, points1, points2, points3, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal mengupdate rule: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule pelanggaran berhasil diupdate",
		})
	}
}

// DeleteViolationRule soft-deletes a violation rule
func DeleteViolationRule(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if rule is being used
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM student_violation_points WHERE rule_id = ? AND deleted_at IS NULL", id).Scan(&count)
		if err == nil && count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Rule ini sedang digunakan oleh %d transaksi. Nonaktifkan saja jika tidak ingin digunakan lagi.", count),
			})
		}

		// Soft delete (set is_active = 0)
		result, err := db.Exec("UPDATE violation_rules SET is_active = 0 WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal menghapus rule: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Rule pelanggaran berhasil dinonaktifkan",
		})
	}
}

// ToggleViolationRuleStatus toggles the is_active status
func ToggleViolationRuleStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		result, err := db.Exec("UPDATE violation_rules SET is_active = 1 - is_active WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal mengubah status: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Rule tidak ditemukan",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Status rule berhasil diubah",
		})
	}
}

// =====================
// IMPORT / EXPORT
// =====================

// ExportAchievementRules exports achievement rules to CSV
func ExportAchievementRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT code, category, category_code, name, description, points, min_points, max_points, is_active
			FROM achievement_rules
			ORDER BY category_code, code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		// Set headers for CSV download
		c.Response().Header().Set(echo.HeaderContentType, "text/csv")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=achievement_rules.csv")
		c.Response().WriteHeader(http.StatusOK)

		writer := csv.NewWriter(c.Response().Writer)
		defer writer.Flush()

		// Write header
		writer.Write([]string{"Kode", "Kategori", "Kode Kategori", "Nama", "Deskripsi", "Poin", "Poin Min", "Poin Max", "Aktif"})

		// Write data
		for rows.Next() {
			var code, category, categoryCode, name, description string
			var points, minPoints, maxPoints, isActive int
			rows.Scan(&code, &category, &categoryCode, &name, &description, &points, &minPoints, &maxPoints, &isActive)
			
			activeStr := "Tidak"
			if isActive == 1 {
				activeStr = "Ya"
			}

			writer.Write([]string{
				code, category, categoryCode, name, description,
				strconv.Itoa(points), strconv.Itoa(minPoints), strconv.Itoa(maxPoints), activeStr,
			})
		}

		return nil
	}
}

// ImportAchievementRules imports achievement rules from CSV
func ImportAchievementRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "File tidak ditemukan",
			})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal membuka file",
			})
		}
		defer src.Close()

		reader := csv.NewReader(src)
		records, err := reader.ReadAll()
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Format CSV tidak valid",
			})
		}

		if len(records) < 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "File CSV kosong",
			})
		}

		successCount := 0
		errorCount := 0
		var errors []string

		// Skip header row
		for i, record := range records[1:] {
			if len(record) < 8 {
				errorCount++
				errors = append(errors, fmt.Sprintf("Baris %d: kolom tidak lengkap", i+2))
				continue
			}

			code := strings.TrimSpace(record[0])
			category := strings.TrimSpace(record[1])
			categoryCode := strings.TrimSpace(record[2])
			name := strings.TrimSpace(record[3])
			description := strings.TrimSpace(record[4])
			points, _ := strconv.Atoi(strings.TrimSpace(record[5]))
			minPoints, _ := strconv.Atoi(strings.TrimSpace(record[6]))
			maxPoints, _ := strconv.Atoi(strings.TrimSpace(record[7]))

			if code == "" || name == "" {
				errorCount++
				errors = append(errors, fmt.Sprintf("Baris %d: kode atau nama kosong", i+2))
				continue
			}

			// Check if code exists
			var existingID int
			err := db.QueryRow("SELECT id FROM achievement_rules WHERE code = ?", code).Scan(&existingID)
			
			if err == sql.ErrNoRows {
				// Insert new
				_, err = db.Exec(`
					INSERT INTO achievement_rules (code, category, category_code, name, description, points, min_points, max_points, is_active)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
				`, code, category, categoryCode, name, description, points, minPoints, maxPoints)
				
				if err != nil {
					errorCount++
					errors = append(errors, fmt.Sprintf("Baris %d: %s", i+2, err.Error()))
				} else {
					successCount++
				}
			} else {
				// Update existing
				_, err = db.Exec(`
					UPDATE achievement_rules 
					SET category = ?, category_code = ?, name = ?, description = ?, 
					    points = ?, min_points = ?, max_points = ?
					WHERE code = ?
				`, category, categoryCode, name, description, points, minPoints, maxPoints, code)
				
				if err != nil {
					errorCount++
					errors = append(errors, fmt.Sprintf("Baris %d: %s", i+2, err.Error()))
				} else {
					successCount++
				}
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": fmt.Sprintf("Import selesai: %d berhasil, %d gagal", successCount, errorCount),
			"success": successCount,
			"errors":  errorCount,
			"details": errors,
		})
	}
}

// ExportViolationRules exports violation rules to CSV
func ExportViolationRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT code, category, category_code, name, description, points_1, points_2, points_3, is_active
			FROM violation_rules
			ORDER BY category_code, code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		// Set headers for CSV download
		c.Response().Header().Set(echo.HeaderContentType, "text/csv")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=violation_rules.csv")
		c.Response().WriteHeader(http.StatusOK)

		writer := csv.NewWriter(c.Response().Writer)
		defer writer.Flush()

		// Write header
		writer.Write([]string{"Kode", "Kategori", "Kode Kategori", "Nama", "Deskripsi", "Poin Ke-1", "Poin Ke-2", "Poin Ke-3", "Aktif"})

		// Write data
		for rows.Next() {
			var code, category, categoryCode, name, description string
			var points1, points2, points3, isActive int
			rows.Scan(&code, &category, &categoryCode, &name, &description, &points1, &points2, &points3, &isActive)
			
			activeStr := "Tidak"
			if isActive == 1 {
				activeStr = "Ya"
			}

			writer.Write([]string{
				code, category, categoryCode, name, description,
				strconv.Itoa(points1), strconv.Itoa(points2), strconv.Itoa(points3), activeStr,
			})
		}

		return nil
	}
}

// ImportViolationRules imports violation rules from CSV
func ImportViolationRules(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "File tidak ditemukan",
			})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Gagal membuka file",
			})
		}
		defer src.Close()

		reader := csv.NewReader(src)
		records, err := reader.ReadAll()
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Format CSV tidak valid",
			})
		}

		if len(records) < 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "File CSV kosong",
			})
		}

		successCount := 0
		errorCount := 0
		var errors []string

		// Skip header row
		for i, record := range records[1:] {
			if len(record) < 8 {
				errorCount++
				errors = append(errors, fmt.Sprintf("Baris %d: kolom tidak lengkap", i+2))
				continue
			}

			code := strings.TrimSpace(record[0])
			category := strings.TrimSpace(record[1])
			categoryCode := strings.TrimSpace(record[2])
			name := strings.TrimSpace(record[3])
			description := strings.TrimSpace(record[4])
			points1, _ := strconv.Atoi(strings.TrimSpace(record[5]))
			points2, _ := strconv.Atoi(strings.TrimSpace(record[6]))
			points3, _ := strconv.Atoi(strings.TrimSpace(record[7]))

			if code == "" || name == "" {
				errorCount++
				errors = append(errors, fmt.Sprintf("Baris %d: kode atau nama kosong", i+2))
				continue
			}

			// Check if code exists
			var existingID int
			err := db.QueryRow("SELECT id FROM violation_rules WHERE code = ?", code).Scan(&existingID)
			
			if err == sql.ErrNoRows {
				// Insert new
				_, err = db.Exec(`
					INSERT INTO violation_rules (code, category, category_code, name, description, points_1, points_2, points_3, is_active)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
				`, code, category, categoryCode, name, description, points1, points2, points3)
				
				if err != nil {
					errorCount++
					errors = append(errors, fmt.Sprintf("Baris %d: %s", i+2, err.Error()))
				} else {
					successCount++
				}
			} else {
				// Update existing
				_, err = db.Exec(`
					UPDATE violation_rules 
					SET category = ?, category_code = ?, name = ?, description = ?, 
					    points_1 = ?, points_2 = ?, points_3 = ?
					WHERE code = ?
				`, category, categoryCode, name, description, points1, points2, points3, code)
				
				if err != nil {
					errorCount++
					errors = append(errors, fmt.Sprintf("Baris %d: %s", i+2, err.Error()))
				} else {
					successCount++
				}
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": fmt.Sprintf("Import selesai: %d berhasil, %d gagal", successCount, errorCount),
			"success": successCount,
			"errors":  errorCount,
			"details": errors,
		})
	}
}
