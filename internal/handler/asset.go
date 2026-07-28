package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/models"
)

// =====================
// ASSET CATEGORIES
// =====================

// GetAssetCategories returns all asset categories
func GetAssetCategories(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT id, code, name, description, created_at
			FROM asset_categories
			ORDER BY code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var categories []models.AssetCategory
		for rows.Next() {
			var cat models.AssetCategory
			if err := rows.Scan(&cat.ID, &cat.Code, &cat.Name, &cat.Description, &cat.CreatedAt); err != nil {
				continue
			}
			categories = append(categories, cat)
		}

		return c.JSON(http.StatusOK, categories)
	}
}

// AddAssetCategory creates a new asset category
func AddAssetCategory(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.FormValue("code")
		name := c.FormValue("name")
		description := c.FormValue("description")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		_, err := db.Exec(`
			INSERT INTO asset_categories (code, name, description, created_at)
			VALUES (?, ?, ?, ?)
		`, code, name, description, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Category added successfully"})
	}
}

// UpdateAssetCategory updates an existing asset category
func UpdateAssetCategory(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		code := c.FormValue("code")
		name := c.FormValue("name")
		description := c.FormValue("description")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		_, err := db.Exec(`
			UPDATE asset_categories
			SET code = ?, name = ?, description = ?
			WHERE id = ?
		`, code, name, description, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Category updated successfully"})
	}
}

// DeleteAssetCategory deletes an asset category
func DeleteAssetCategory(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if category is being used
		var count int
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE category_id = ?", id).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete category with existing assets"})
		}

		_, err := db.Exec("DELETE FROM asset_categories WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Category deleted successfully"})
	}
}

// =====================
// ASSET FUNDING SOURCES
// =====================

// GetAssetFundingSources returns all funding sources
func GetAssetFundingSources(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT id, code, name, description, created_at
			FROM asset_funding_sources
			ORDER BY code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var sources []models.AssetFundingSource
		for rows.Next() {
			var src models.AssetFundingSource
			if err := rows.Scan(&src.ID, &src.Code, &src.Name, &src.Description, &src.CreatedAt); err != nil {
				continue
			}
			sources = append(sources, src)
		}

		return c.JSON(http.StatusOK, sources)
	}
}

// AddAssetFundingSource creates a new funding source
func AddAssetFundingSource(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.FormValue("code")
		name := c.FormValue("name")
		description := c.FormValue("description")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		_, err := db.Exec(`
			INSERT INTO asset_funding_sources (code, name, description, created_at)
			VALUES (?, ?, ?, ?)
		`, code, name, description, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Funding source added successfully"})
	}
}

// UpdateAssetFundingSource updates an existing funding source
func UpdateAssetFundingSource(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		code := c.FormValue("code")
		name := c.FormValue("name")
		description := c.FormValue("description")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		_, err := db.Exec(`
			UPDATE asset_funding_sources
			SET code = ?, name = ?, description = ?
			WHERE id = ?
		`, code, name, description, id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Funding source updated successfully"})
	}
}

// DeleteAssetFundingSource deletes a funding source
func DeleteAssetFundingSource(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if source is being used
		var count int
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE funding_source_id = ?", id).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete funding source with existing assets"})
		}

		_, err := db.Exec("DELETE FROM asset_funding_sources WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Funding source deleted successfully"})
	}
}

// =====================
// ASSET LOCATIONS
// =====================

// GetAssetLocations returns all asset locations
func GetAssetLocations(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query(`
			SELECT 
				l.id, l.code, l.name, l.type, l.capacity, 
				l.pic_staff_id, COALESCE(s.name, '') as pic_name,
				l.status, l.created_at, l.updated_at
			FROM asset_locations l
			LEFT JOIN staff s ON l.pic_staff_id = s.id
			ORDER BY l.code
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var locations []models.AssetLocation
		for rows.Next() {
			var loc models.AssetLocation
			if err := rows.Scan(&loc.ID, &loc.Code, &loc.Name, &loc.Type, &loc.Capacity, 
				&loc.PICStaffID, &loc.PICName, &loc.Status, &loc.CreatedAt, &loc.UpdatedAt); err != nil {
				continue
			}
			locations = append(locations, loc)
		}

		return c.JSON(http.StatusOK, locations)
	}
}

// AddAssetLocation creates a new location
func AddAssetLocation(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.FormValue("code")
		name := c.FormValue("name")
		locType := c.FormValue("type")
		capacityStr := c.FormValue("capacity")
		picStaffIDStr := c.FormValue("pic_staff_id")
		status := c.FormValue("status")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		var capacity int
		if capacityStr != "" {
			capacity, _ = strconv.Atoi(capacityStr)
		}

		var picStaffID *int
		if picStaffIDStr != "" {
			id, _ := strconv.Atoi(picStaffIDStr)
			picStaffID = &id
		}

		if status == "" {
			status = "active"
		}

		_, err := db.Exec(`
			INSERT INTO asset_locations (code, name, type, capacity, pic_staff_id, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, code, name, locType, capacity, picStaffID, status, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Location added successfully"})
	}
}

// UpdateAssetLocation updates an existing location
func UpdateAssetLocation(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		code := c.FormValue("code")
		name := c.FormValue("name")
		locType := c.FormValue("type")
		capacityStr := c.FormValue("capacity")
		picStaffIDStr := c.FormValue("pic_staff_id")
		status := c.FormValue("status")

		if code == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code and name are required"})
		}

		var capacity int
		if capacityStr != "" {
			capacity, _ = strconv.Atoi(capacityStr)
		}

		var picStaffID *int
		if picStaffIDStr != "" {
			staffID, _ := strconv.Atoi(picStaffIDStr)
			picStaffID = &staffID
		}

		_, err := db.Exec(`
			UPDATE asset_locations
			SET code = ?, name = ?, type = ?, capacity = ?, pic_staff_id = ?, status = ?, updated_at = ?
			WHERE id = ?
		`, code, name, locType, capacity, picStaffID, status, time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Location updated successfully"})
	}
}

// DeleteAssetLocation deletes a location
func DeleteAssetLocation(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if location is being used
		var count int
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE location_id = ?", id).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete location with existing assets"})
		}

		_, err := db.Exec("DELETE FROM asset_locations WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Location deleted successfully"})
	}
}
