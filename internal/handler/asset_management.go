package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"

	"belsekolah/internal/models"
)

// =====================
// ASSETS MANAGEMENT
// =====================

// GetAssets returns all assets with filters
func GetAssets(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		categoryID := c.QueryParam("category_id")
		locationID := c.QueryParam("location_id")
		condition := c.QueryParam("condition")
		search := c.QueryParam("search")

		query := `
			SELECT 
				a.id, a.inventory_code, COALESCE(a.qr_code, ''), a.category_id, 
				COALESCE(ac.name, '') as category_name,
				a.name, COALESCE(a.specification, ''), COALESCE(a.brand, ''), COALESCE(a.serial_number, ''),
				COALESCE(a.acquisition_year, 0), COALESCE(a.acquisition_date, ''), COALESCE(a.purchase_price, 0),
				a.funding_source_id, COALESCE(fs.name, '') as funding_source,
				a.location_id, COALESCE(l.name, '') as location_name,
				a.pic_staff_id, COALESCE(s.name, '') as pic_name,
				COALESCE(a.condition, 'good'), COALESCE(a.quantity, 1), COALESCE(a.unit, 'unit'), COALESCE(a.notes, ''), COALESCE(a.photo_url, ''),
				COALESCE(a.is_borrowable, 0), COALESCE(a.created_at, ''), COALESCE(a.updated_at, ''), COALESCE(a.created_by, '')
			FROM assets a
			LEFT JOIN asset_categories ac ON a.category_id = ac.id
			LEFT JOIN asset_funding_sources fs ON a.funding_source_id = fs.id
			LEFT JOIN asset_locations l ON a.location_id = l.id
			LEFT JOIN staff s ON a.pic_staff_id = s.id
			WHERE 1=1
		`

		args := []interface{}{}

		if categoryID != "" {
			query += " AND a.category_id = ?"
			args = append(args, categoryID)
		}

		if locationID != "" {
			query += " AND a.location_id = ?"
			args = append(args, locationID)
		}

		if condition != "" {
			query += " AND a.condition = ?"
			args = append(args, condition)
		}

		if search != "" {
			query += " AND (a.name LIKE ? OR a.inventory_code LIKE ? OR a.brand LIKE ?)"
			searchTerm := "%" + search + "%"
			args = append(args, searchTerm, searchTerm, searchTerm)
		}

		query += " ORDER BY a.created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}
		defer rows.Close()

		var assets []models.Asset
		for rows.Next() {
			var asset models.Asset
			var isBorrowable int
			var rawAcqDate, rawCreatedAt, rawUpdatedAt interface{}
			var rawFundingSourceID, rawLocationID, rawPICStaffID, rawCreatedBy interface{}
			if err := rows.Scan(
				&asset.ID, &asset.InventoryCode, &asset.QRCode, &asset.CategoryID,
				&asset.CategoryName, &asset.Name, &asset.Specification, &asset.Brand,
				&asset.SerialNumber, &asset.AcquisitionYear, &rawAcqDate,
				&asset.PurchasePrice, &rawFundingSourceID, &asset.FundingSource,
				&rawLocationID, &asset.LocationName, &rawPICStaffID, &asset.PICName,
				&asset.Condition, &asset.Quantity, &asset.Unit, &asset.Notes,
				&asset.PhotoURL, &isBorrowable, &rawCreatedAt, &rawUpdatedAt,
				&rawCreatedBy,
			); err != nil {
				continue
			}
			asset.IsBorrowable = (isBorrowable == 1)
			asset.FundingSourceID = parseSqliteIntPtr(rawFundingSourceID)
			asset.LocationID = parseSqliteIntPtr(rawLocationID)
			asset.PICStaffID = parseSqliteIntPtr(rawPICStaffID)
			asset.CreatedBy = parseSqliteIntPtr(rawCreatedBy)
			if t := parseSqliteTime(rawAcqDate); !t.IsZero() {
				asset.AcquisitionDate = &t
			}
			asset.CreatedAt = parseSqliteTime(rawCreatedAt)
			if t := parseSqliteTime(rawUpdatedAt); !t.IsZero() {
				asset.UpdatedAt = &t
			}
			assets = append(assets, asset)
		}

		return c.JSON(http.StatusOK, assets)
	}
}

// GetAssetByID returns a single asset by ID
func GetAssetByID(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var asset models.Asset
		var isBorrowable int
		var rawAcqDate, rawCreatedAt, rawUpdatedAt interface{}
		var rawFundingSourceID, rawLocationID, rawPICStaffID, rawCreatedBy interface{}
		err := db.QueryRow(`
			SELECT 
				a.id, a.inventory_code, COALESCE(a.qr_code, ''), a.category_id, 
				COALESCE(ac.name, '') as category_name,
				a.name, COALESCE(a.specification, ''), COALESCE(a.brand, ''), COALESCE(a.serial_number, ''),
				COALESCE(a.acquisition_year, 0), COALESCE(a.acquisition_date, ''), COALESCE(a.purchase_price, 0),
				a.funding_source_id, COALESCE(fs.name, '') as funding_source,
				a.location_id, COALESCE(l.name, '') as location_name,
				a.pic_staff_id, COALESCE(s.name, '') as pic_name,
				COALESCE(a.condition, 'good'), COALESCE(a.quantity, 1), COALESCE(a.unit, 'unit'), COALESCE(a.notes, ''), COALESCE(a.photo_url, ''),
				COALESCE(a.is_borrowable, 0), COALESCE(a.created_at, ''), COALESCE(a.updated_at, ''), COALESCE(a.created_by, '')
			FROM assets a
			LEFT JOIN asset_categories ac ON a.category_id = ac.id
			LEFT JOIN asset_funding_sources fs ON NULLIF(a.funding_source_id, '') = fs.id
			LEFT JOIN asset_locations l ON NULLIF(a.location_id, '') = l.id
			LEFT JOIN staff s ON NULLIF(a.pic_staff_id, '') = s.id
			WHERE a.id = ?
		`, id).Scan(
			&asset.ID, &asset.InventoryCode, &asset.QRCode, &asset.CategoryID,
			&asset.CategoryName, &asset.Name, &asset.Specification, &asset.Brand,
			&asset.SerialNumber, &asset.AcquisitionYear, &rawAcqDate,
			&asset.PurchasePrice, &rawFundingSourceID, &asset.FundingSource,
			&rawLocationID, &asset.LocationName, &rawPICStaffID, &asset.PICName,
			&asset.Condition, &asset.Quantity, &asset.Unit, &asset.Notes,
			&asset.PhotoURL, &isBorrowable, &rawCreatedAt, &rawUpdatedAt,
			&rawCreatedBy,
		)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Asset not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		asset.IsBorrowable = (isBorrowable == 1)
		asset.FundingSourceID = parseSqliteIntPtr(rawFundingSourceID)
		asset.LocationID = parseSqliteIntPtr(rawLocationID)
		asset.PICStaffID = parseSqliteIntPtr(rawPICStaffID)
		asset.CreatedBy = parseSqliteIntPtr(rawCreatedBy)
		if t := parseSqliteTime(rawAcqDate); !t.IsZero() {
			asset.AcquisitionDate = &t
		}
		asset.CreatedAt = parseSqliteTime(rawCreatedAt)
		if t := parseSqliteTime(rawUpdatedAt); !t.IsZero() {
			asset.UpdatedAt = &t
		}

		return c.JSON(http.StatusOK, asset)
	}
}

// GenerateInventoryCode generates a unique inventory code
func GenerateInventoryCode(db *sql.DB, categoryCode string, year int) string {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM assets WHERE acquisition_year = ?", year).Scan(&count)
	return fmt.Sprintf("%s-%d-%04d", categoryCode, year, count+1)
}

// AddAsset creates a new asset
func AddAsset(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		categoryID := c.FormValue("category_id")
		name := c.FormValue("name")
		specification := c.FormValue("specification")
		brand := c.FormValue("brand")
		serialNumber := c.FormValue("serial_number")
		acquisitionYearStr := c.FormValue("acquisition_year")
		acquisitionDateStr := c.FormValue("acquisition_date")
		purchasePriceStr := c.FormValue("purchase_price")
		condition := c.FormValue("condition")
		quantityStr := c.FormValue("quantity")
		unit := c.FormValue("unit")
		notes := c.FormValue("notes")
		photoURL := c.FormValue("photo_url")
		isBorrowableStr := c.FormValue("is_borrowable")

		var fundingSourceID *int
		if fsStr := c.FormValue("funding_source_id"); fsStr != "" {
			if id, err := strconv.Atoi(fsStr); err == nil && id > 0 {
				fundingSourceID = &id
			}
		}

		var locationID *int
		if locStr := c.FormValue("location_id"); locStr != "" {
			if id, err := strconv.Atoi(locStr); err == nil && id > 0 {
				locationID = &id
			}
		}

		var picStaffID *int
		if picStr := c.FormValue("pic_staff_id"); picStr != "" {
			if id, err := strconv.Atoi(picStr); err == nil && id > 0 {
				picStaffID = &id
			}
		}

		if categoryID == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category and name are required"})
		}

		// Get category code for inventory code generation
		var categoryCode string
		db.QueryRow("SELECT code FROM asset_categories WHERE id = ?", categoryID).Scan(&categoryCode)

		acquisitionYear, _ := strconv.Atoi(acquisitionYearStr)
		if acquisitionYear == 0 {
			acquisitionYear = time.Now().Year()
		}

		// Generate inventory code
		inventoryCode := GenerateInventoryCode(db, categoryCode, acquisitionYear)

		// Generate QR code content
		qrContent := fmt.Sprintf("ASSET:%s", inventoryCode)

		purchasePrice, _ := strconv.ParseFloat(purchasePriceStr, 64)
		quantity, _ := strconv.Atoi(quantityStr)
		if quantity == 0 {
			quantity = 1
		}

		isBorrowable := 0
		if isBorrowableStr == "1" || isBorrowableStr == "true" {
			isBorrowable = 1
		}

		if condition == "" {
			condition = "good"
		}

		var acquisitionDate *time.Time
		if acquisitionDateStr != "" {
			t, err := time.Parse("2006-01-02", acquisitionDateStr)
			if err == nil {
				acquisitionDate = &t
			}
		}

		var createdBy *int

		result, err := db.Exec(`
			INSERT INTO assets (
				inventory_code, qr_code, category_id, name, specification, brand, 
				serial_number, acquisition_year, acquisition_date, purchase_price,
				funding_source_id, location_id, pic_staff_id, condition, quantity, 
				unit, notes, photo_url, is_borrowable, created_at, created_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, inventoryCode, qrContent, categoryID, name, specification, brand,
			serialNumber, acquisitionYear, acquisitionDate, purchasePrice,
			fundingSourceID, locationID, picStaffID, condition, quantity,
			unit, notes, photoURL, isBorrowable, time.Now(), createdBy)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		assetID, _ := result.LastInsertId()

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":        "Asset added successfully",
			"id":             assetID,
			"inventory_code": inventoryCode,
		})
	}
}

// UpdateAsset updates an existing asset
func UpdateAsset(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		categoryID := c.FormValue("category_id")
		name := c.FormValue("name")
		specification := c.FormValue("specification")
		brand := c.FormValue("brand")
		serialNumber := c.FormValue("serial_number")
		acquisitionYearStr := c.FormValue("acquisition_year")
		acquisitionDateStr := c.FormValue("acquisition_date")
		purchasePriceStr := c.FormValue("purchase_price")
		condition := c.FormValue("condition")
		quantityStr := c.FormValue("quantity")
		unit := c.FormValue("unit")
		notes := c.FormValue("notes")
		photoURL := c.FormValue("photo_url")
		isBorrowableStr := c.FormValue("is_borrowable")

		var fundingSourceID *int
		if fsStr := c.FormValue("funding_source_id"); fsStr != "" {
			if id, err := strconv.Atoi(fsStr); err == nil && id > 0 {
				fundingSourceID = &id
			}
		}

		var locationID *int
		if locStr := c.FormValue("location_id"); locStr != "" {
			if id, err := strconv.Atoi(locStr); err == nil && id > 0 {
				locationID = &id
			}
		}

		var picStaffID *int
		if picStr := c.FormValue("pic_staff_id"); picStr != "" {
			if id, err := strconv.Atoi(picStr); err == nil && id > 0 {
				picStaffID = &id
			}
		}

		if categoryID == "" || name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category and name are required"})
		}

		acquisitionYear, _ := strconv.Atoi(acquisitionYearStr)
		purchasePrice, _ := strconv.ParseFloat(purchasePriceStr, 64)
		quantity, _ := strconv.Atoi(quantityStr)

		isBorrowable := 0
		if isBorrowableStr == "1" || isBorrowableStr == "true" {
			isBorrowable = 1
		}

		var acquisitionDate *time.Time
		if acquisitionDateStr != "" {
			t, err := time.Parse("2006-01-02", acquisitionDateStr)
			if err == nil {
				acquisitionDate = &t
			}
		}

		_, err := db.Exec(`
			UPDATE assets
			SET category_id = ?, name = ?, specification = ?, brand = ?, 
				serial_number = ?, acquisition_year = ?, acquisition_date = ?, 
				purchase_price = ?, funding_source_id = ?, location_id = ?, 
				pic_staff_id = ?, condition = ?, quantity = ?, unit = ?, 
				notes = ?, photo_url = ?, is_borrowable = ?, updated_at = ?
			WHERE id = ?
		`, categoryID, name, specification, brand, serialNumber, acquisitionYear,
			acquisitionDate, purchasePrice, fundingSourceID, locationID,
			picStaffID, condition, quantity, unit, notes, photoURL,
			isBorrowable, time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Asset updated successfully"})
	}
}

// DeleteAsset deletes an asset
func DeleteAsset(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		// Check if asset has borrowing records
		var count int
		db.QueryRow("SELECT COUNT(*) FROM asset_borrowings WHERE asset_id = ? AND status IN ('pending', 'approved', 'borrowed')", id).Scan(&count)
		if count > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete asset with active borrowings"})
		}

		_, err := db.Exec("DELETE FROM assets WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Asset deleted successfully"})
	}
}

// GenerateAssetQRCode generates QR code for an asset
func GenerateAssetQRCode(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var inventoryCode string
		err := db.QueryRow("SELECT inventory_code FROM assets WHERE id = ?", id).Scan(&inventoryCode)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Asset not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		// Generate QR code
		qrContent := fmt.Sprintf("ASSET:%s", inventoryCode)
		png, err := qrcode.Encode(qrContent, qrcode.Medium, 256)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		return c.Blob(http.StatusOK, "image/png", png)
	}
}

// ScanAssetQRCode handles QR code scanning
func ScanAssetQRCode(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := c.QueryParam("code")

		if code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Code parameter is required"})
		}

		// Parse QR code (format: ASSET:INV-CODE or just INV-CODE)
		inventoryCode := code
		if len(code) > 6 && code[:6] == "ASSET:" {
			inventoryCode = code[6:]
		}

		var asset models.Asset
		var isBorrowable int
		err := db.QueryRow(`
			SELECT 
				a.id, a.inventory_code, a.name, 
				ac.name as category_name,
				COALESCE(l.name, '') as location_name,
				a.condition, a.is_borrowable
			FROM assets a
			LEFT JOIN asset_categories ac ON a.category_id = ac.id
			LEFT JOIN asset_locations l ON a.location_id = l.id
			WHERE a.inventory_code = ? OR a.qr_code = ?
		`, inventoryCode, code).Scan(
			&asset.ID, &asset.InventoryCode, &asset.Name,
			&asset.CategoryName, &asset.LocationName,
			&asset.Condition, &isBorrowable,
		)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Asset not found"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan server"})
		}

		asset.IsBorrowable = isBorrowable == 1

		return c.JSON(http.StatusOK, asset)
	}
}

// GetAssetStats returns statistics for dashboard
func GetAssetStats(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var stats struct {
			TotalAssets      int     `json:"total_assets"`
			TotalValue       float64 `json:"total_value"`
			GoodCondition    int     `json:"good_condition"`
			NeedsRepair      int     `json:"needs_repair"`
			ActiveBorrowings int     `json:"active_borrowings"`
			OpenTickets      int     `json:"open_tickets"`
		}

		db.QueryRow("SELECT COUNT(*) FROM assets").Scan(&stats.TotalAssets)
		db.QueryRow("SELECT COALESCE(SUM(purchase_price * quantity), 0) FROM assets").Scan(&stats.TotalValue)
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE condition = 'good'").Scan(&stats.GoodCondition)
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE condition IN ('poor', 'broken')").Scan(&stats.NeedsRepair)
		db.QueryRow("SELECT COUNT(*) FROM asset_borrowings WHERE status IN ('approved', 'borrowed')").Scan(&stats.ActiveBorrowings)
		db.QueryRow("SELECT COUNT(*) FROM asset_maintenance_tickets WHERE status IN ('open', 'in_progress')").Scan(&stats.OpenTickets)

		return c.JSON(http.StatusOK, stats)
	}
}
