package handler

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"

	"belsekolah/internal/models"
)

// ExportAssetsExcel exports all or filtered assets to a professional Excel (.xlsx) spreadsheet
func ExportAssetsExcel(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		categoryID := c.QueryParam("category_id")
		categoryName := c.QueryParam("category_name")
		locationID := c.QueryParam("location_id")
		locationName := c.QueryParam("location_name")
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
			LEFT JOIN asset_funding_sources fs ON NULLIF(a.funding_source_id, '') = fs.id
			LEFT JOIN asset_locations l ON NULLIF(a.location_id, '') = l.id
			LEFT JOIN staff s ON NULLIF(a.pic_staff_id, '') = s.id
			WHERE 1=1
		`

		var args []interface{}
		if categoryID != "" {
			query += " AND a.category_id = ?"
			args = append(args, categoryID)
		} else if categoryName != "" {
			query += " AND ac.name = ?"
			args = append(args, categoryName)
		}
		if locationID != "" {
			query += " AND a.location_id = ?"
			args = append(args, locationID)
		} else if locationName != "" {
			query += " AND l.name = ?"
			args = append(args, locationName)
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
		query += " ORDER BY a.inventory_code ASC, a.created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Gagal mengambil data aset: " + err.Error()})
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
			assets = append(assets, asset)
		}

		// Create Excel workbook
		f := excelize.NewFile()
		defer f.Close()

		sheetName := "Data Inventaris Aset"
		f.SetSheetName("Sheet1", sheetName)

		// Header styling (Dark modern slate)
		headerStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11, Family: "Segoe UI"},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#18181B"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "left", Color: "#3F3F46", Style: 1},
				{Type: "top", Color: "#3F3F46", Style: 1},
				{Type: "bottom", Color: "#3F3F46", Style: 1},
				{Type: "right", Color: "#3F3F46", Style: 1},
			},
		})
		if err != nil {
			headerStyle = 0
		}

		// Title styles
		titleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Size: 15, Color: "#09090B", Family: "Segoe UI"},
		})
		subTitleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Italic: true, Size: 10, Color: "#71717A", Family: "Segoe UI"},
		})

		// Normal bordered data style
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
			Alignment: &excelize.Alignment{Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#E4E4E7", Style: 1},
				{Type: "top", Color: "#E4E4E7", Style: 1},
				{Type: "bottom", Color: "#E4E4E7", Style: 1},
				{Type: "right", Color: "#E4E4E7", Style: 1},
			},
		})

		// Centered data style
		centerDataStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#E4E4E7", Style: 1},
				{Type: "top", Color: "#E4E4E7", Style: 1},
				{Type: "bottom", Color: "#E4E4E7", Style: 1},
				{Type: "right", Color: "#E4E4E7", Style: 1},
			},
		})

		// Currency style
		currencyStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Size: 10, Family: "Segoe UI"},
			CustomNumFmt: &[]string{"_(\"Rp\"* #,##0_);_(\"Rp\"* (#,##0);_(\"Rp\"* \"-\"_);_(@_)"}[0],
			Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#E4E4E7", Style: 1},
				{Type: "top", Color: "#E4E4E7", Style: 1},
				{Type: "bottom", Color: "#E4E4E7", Style: 1},
				{Type: "right", Color: "#E4E4E7", Style: 1},
			},
		})

		// Set Title and Subtitle
		f.SetCellValue(sheetName, "A1", "LAPORAN INVENTARIS SARANA & PRASARANA (SARPRAS)")
		f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

		downloadInfo := fmt.Sprintf("SMK NIBA Business School | Diekspor pada: %s | Total: %d Aset",
			time.Now().Format("02 Jan 2006 15:04 WIB"), len(assets))
		f.SetCellValue(sheetName, "A2", downloadInfo)
		f.SetCellStyle(sheetName, "A2", "A2", subTitleStyle)

		// Column headers
		headers := []string{
			"No", "Kode Inventaris", "Nama Barang", "Kategori KIB", "Merk / Tipe",
			"Spesifikasi", "Nomor Seri (SN)", "Tahun", "Tgl Perolehan", "Harga Satuan (Rp)",
			"Jumlah", "Satuan", "Total Nilai (Rp)", "Lokasi / Ruang", "Sumber Dana",
			"Penanggung Jawab", "Kondisi", "Bisa Dipinjam", "Keterangan",
		}

		headerRow := 4
		for colIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, headerRow)
			f.SetCellValue(sheetName, cell, h)
		}
		startCol, _ := excelize.CoordinatesToCellName(1, headerRow)
		endCol, _ := excelize.CoordinatesToCellName(len(headers), headerRow)
		f.SetCellStyle(sheetName, startCol, endCol, headerStyle)
		f.SetRowHeight(sheetName, headerRow, 28)

		// Populate Data
		currentRow := headerRow + 1
		for i, a := range assets {
			acqDateStr := "-"
			if a.AcquisitionDate != nil && !a.AcquisitionDate.IsZero() {
				acqDateStr = a.AcquisitionDate.Format("2006-01-02")
			}

			conditionLabel := "Baik"
			if a.Condition == "poor" {
				conditionLabel = "Rusak Ringan"
			} else if a.Condition == "broken" {
				conditionLabel = "Rusak Berat"
			}

			borrowLabel := "Tidak"
			if a.IsBorrowable {
				borrowLabel = "Ya"
			}

			totalValue := a.PurchasePrice * float64(a.Quantity)

			rowVals := []interface{}{
				i + 1,
				a.InventoryCode,
				a.Name,
				a.CategoryName,
				a.Brand,
				a.Specification,
				a.SerialNumber,
				a.AcquisitionYear,
				acqDateStr,
				a.PurchasePrice,
				a.Quantity,
				a.Unit,
				totalValue,
				a.LocationName,
				a.FundingSource,
				a.PICName,
				conditionLabel,
				borrowLabel,
				a.Notes,
			}

			for colIdx, val := range rowVals {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
				f.SetCellValue(sheetName, cell, val)

				// Apply styles based on column
				if colIdx == 0 || colIdx == 7 || colIdx == 8 || colIdx == 10 || colIdx == 11 || colIdx == 16 || colIdx == 17 {
					f.SetCellStyle(sheetName, cell, cell, centerDataStyle)
				} else if colIdx == 9 || colIdx == 12 {
					f.SetCellStyle(sheetName, cell, cell, currencyStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, dataStyle)
				}
			}
			f.SetRowHeight(sheetName, currentRow, 22)
			currentRow++
		}

		// Column widths
		colWidths := map[string]float64{
			"A": 6,   // No
			"B": 18,  // Kode
			"C": 28,  // Nama
			"D": 22,  // Kategori
			"E": 18,  // Merk
			"F": 32,  // Spesifikasi
			"G": 18,  // SN
			"H": 10,  // Tahun
			"I": 15,  // Tgl Perolehan
			"J": 18,  // Harga
			"K": 10,  // Jumlah
			"L": 10,  // Satuan
			"M": 20,  // Total Nilai
			"N": 20,  // Lokasi
			"O": 16,  // Sumber Dana
			"P": 20,  // PIC
			"Q": 15,  // Kondisi
			"R": 14,  // Bisa Dipinjam
			"S": 25,  // Keterangan
		}
		for col, w := range colWidths {
			f.SetColWidth(sheetName, col, col, w)
		}

		fileName := fmt.Sprintf("Inventaris_Sarpras_%s.xlsx", time.Now().Format("20060102_150405"))
		c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		return f.Write(c.Response().Writer)
	}
}

// DownloadAssetTemplateExcel provides a ready-to-fill Excel template for importing assets
func DownloadAssetTemplateExcel(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		f := excelize.NewFile()
		defer f.Close()

		sheetMain := "Template Import Aset"
		f.SetSheetName("Sheet1", sheetMain)

		// Header styling
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11, Family: "Segoe UI"},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#0F172A"}, Pattern: 1}, // Dark Navy Slate
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "left", Color: "#334155", Style: 1},
				{Type: "top", Color: "#334155", Style: 1},
				{Type: "bottom", Color: "#334155", Style: 1},
				{Type: "right", Color: "#334155", Style: 1},
			},
		})

		exampleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Size: 10, Color: "#334155", Family: "Segoe UI"},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#F8FAFC"}, Pattern: 1},
			Border: []excelize.Border{
				{Type: "left", Color: "#CBD5E1", Style: 1},
				{Type: "top", Color: "#CBD5E1", Style: 1},
				{Type: "bottom", Color: "#CBD5E1", Style: 1},
				{Type: "right", Color: "#CBD5E1", Style: 1},
			},
		})

		// Template Headers (Row 1)
		headers := []string{
			"Nama Barang*",
			"Kategori (Kode/Nama)*",
			"Merk",
			"Spesifikasi",
			"Nomor Seri (SN)",
			"Tahun Perolehan",
			"Tanggal Perolehan (YYYY-MM-DD)",
			"Harga Satuan (Rp)",
			"Jumlah",
			"Satuan",
			"Lokasi / Ruang",
			"Sumber Dana",
			"Kondisi (baik/rusak ringan/rusak berat)",
			"Bisa Dipinjam (Ya/Tidak)",
			"Catatan",
		}

		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetMain, cell, h)
		}
		endCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
		f.SetCellStyle(sheetMain, "A1", endCol, headerStyle)
		f.SetRowHeight(sheetMain, 1, 30)

		// Example rows to guide the user
		sample1 := []interface{}{
			"Laptop Lenovo Thinkpad T14",
			"KIB-B",
			"Lenovo",
			"Core i5 13th 16GB SSD 512GB",
			"PF34XYZ99",
			2026,
			"2026-02-15",
			14500000,
			1,
			"unit",
			"Home Server",
			"Dana BOS",
			"baik",
			"Ya",
			"Laptop operasional lab server",
		}
		sample2 := []interface{}{
			"Proyektor Epson EB-X500",
			"Peralatan dan Mesin",
			"Epson",
			"3600 Lumens HDMI XGA",
			"EPS778811",
			2025,
			"2025-10-10",
			6500000,
			1,
			"unit",
			"Ruang Guru",
			"Dana BOS",
			"baik",
			"Ya",
			"Proyektor presentasi ruang rapat",
		}

		for colIdx, val := range sample1 {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
			f.SetCellValue(sheetMain, cell, val)
			f.SetCellStyle(sheetMain, cell, cell, exampleStyle)
		}
		for colIdx, val := range sample2 {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 3)
			f.SetCellValue(sheetMain, cell, val)
			f.SetCellStyle(sheetMain, cell, cell, exampleStyle)
		}
		f.SetRowHeight(sheetMain, 2, 22)
		f.SetRowHeight(sheetMain, 3, 22)

		// Set column widths
		widths := []float64{30, 22, 18, 30, 18, 16, 26, 18, 12, 12, 20, 18, 26, 20, 28}
		for i, w := range widths {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheetMain, col, col, w)
		}

		// Sheet 2: Petunjuk & Daftar Referensi
		sheetRef := "Daftar Referensi"
		f.NewSheet(sheetRef)

		refHeaderStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 10, Family: "Segoe UI"},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#334155"}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})

		f.SetCellValue(sheetRef, "A1", "PANDUAN & DAFTAR REFERENSI SISTEM SARPRAS")
		f.SetCellValue(sheetRef, "A2", "Catatan: Kolom bertanda bintang (*) adalah kolom WAJIB diisi. Baris contoh pada sheet utama bisa dihapus sebelum impor.")

		// Table 1: Categories
		f.SetCellValue(sheetRef, "A4", "Kategori KIB (Kode)")
		f.SetCellValue(sheetRef, "B4", "Nama Kategori")
		f.SetCellStyle(sheetRef, "A4", "B4", refHeaderStyle)

		catRows, _ := db.Query("SELECT code, name FROM asset_categories ORDER BY code ASC")
		if catRows != nil {
			cRow := 5
			for catRows.Next() {
				var code, name string
				catRows.Scan(&code, &name)
				f.SetCellValue(sheetRef, fmt.Sprintf("A%d", cRow), code)
				f.SetCellValue(sheetRef, fmt.Sprintf("B%d", cRow), name)
				cRow++
			}
			catRows.Close()
		}

		// Table 2: Locations
		f.SetCellValue(sheetRef, "D4", "Kode Ruang")
		f.SetCellValue(sheetRef, "E4", "Nama Ruang / Lokasi")
		f.SetCellStyle(sheetRef, "D4", "E4", refHeaderStyle)

		locRows, _ := db.Query("SELECT code, name FROM asset_locations ORDER BY name ASC")
		if locRows != nil {
			lRow := 5
			for locRows.Next() {
				var code, name string
				locRows.Scan(&code, &name)
				f.SetCellValue(sheetRef, fmt.Sprintf("D%d", lRow), code)
				f.SetCellValue(sheetRef, fmt.Sprintf("E%d", lRow), name)
				lRow++
			}
			locRows.Close()
		}

		// Table 3: Funding Sources
		f.SetCellValue(sheetRef, "G4", "Kode Sumber Dana")
		f.SetCellValue(sheetRef, "H4", "Nama Sumber Dana")
		f.SetCellStyle(sheetRef, "G4", "H4", refHeaderStyle)

		fsRows, _ := db.Query("SELECT code, name FROM asset_funding_sources ORDER BY code ASC")
		if fsRows != nil {
			fsRow := 5
			for fsRows.Next() {
				var code, name string
				fsRows.Scan(&code, &name)
				f.SetCellValue(sheetRef, fmt.Sprintf("G%d", fsRow), code)
				f.SetCellValue(sheetRef, fmt.Sprintf("H%d", fsRow), name)
				fsRow++
			}
			fsRows.Close()
		}

		f.SetColWidth(sheetRef, "A", "B", 25)
		f.SetColWidth(sheetRef, "D", "E", 25)
		f.SetColWidth(sheetRef, "G", "H", 25)

		c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Response().Header().Set("Content-Disposition", "attachment; filename=\"Template_Import_Aset_Sarpras.xlsx\"")
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		return f.Write(c.Response().Writer)
	}
}

// ImportAssetsExcel processes uploaded Excel files (.xlsx) and imports asset records into SQLite
func ImportAssetsExcel(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File Excel tidak ditemukan. Pastikan Anda memilih file sebelum mengunggah."})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file unggahan."})
		}
		defer src.Close()

		f, err := excelize.OpenReader(src)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format file tidak valid atau rusak. Pastikan file berekstensi .xlsx"})
		}
		defer f.Close()

		// Determine sheet: look for "Template Import Aset" or fallback to first sheet
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Sheet di dalam file Excel kosong."})
		}

		sheetName := sheets[0]
		for _, s := range sheets {
			if strings.EqualFold(strings.TrimSpace(s), "Template Import Aset") {
				sheetName = s
				break
			}
		}

		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) < 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data di dalam sheet Excel kosong atau tidak memiliki baris data."})
		}

		// 1. Build Category lookup maps
		type catInfo struct {
			ID   int
			Code string
		}
		catMap := make(map[string]catInfo)
		var defaultCat catInfo

		cRows, _ := db.Query("SELECT id, code, name FROM asset_categories ORDER BY id ASC")
		if cRows != nil {
			for cRows.Next() {
				var ci catInfo
				var name string
				cRows.Scan(&ci.ID, &ci.Code, &name)
				catMap[strings.ToLower(strings.TrimSpace(ci.Code))] = ci
				catMap[strings.ToLower(strings.TrimSpace(name))] = ci
				if defaultCat.ID == 0 {
					defaultCat = ci
				}
			}
			cRows.Close()
		}

		// 2. Build Location lookup map
		locMap := make(map[string]int)
		lRows, _ := db.Query("SELECT id, code, name FROM asset_locations")
		if lRows != nil {
			for lRows.Next() {
				var id int
				var code, name string
				lRows.Scan(&id, &code, &name)
				locMap[strings.ToLower(strings.TrimSpace(code))] = id
				locMap[strings.ToLower(strings.TrimSpace(name))] = id
			}
			lRows.Close()
		}

		// 3. Build Funding Source lookup map
		fsMap := make(map[string]int)
		fsRows, _ := db.Query("SELECT id, code, name FROM asset_funding_sources")
		if fsRows != nil {
			for fsRows.Next() {
				var id int
				var code, name string
				fsRows.Scan(&id, &code, &name)
				fsMap[strings.ToLower(strings.TrimSpace(code))] = id
				fsMap[strings.ToLower(strings.TrimSpace(name))] = id
			}
			fsRows.Close()
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memulai transaksi database: " + err.Error()})
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(`
			INSERT INTO assets (
				inventory_code, qr_code, category_id, name, specification, brand, 
				serial_number, acquisition_year, acquisition_date, purchase_price,
				funding_source_id, location_id, pic_staff_id, condition, quantity, 
				unit, notes, photo_url, is_borrowable, created_at, created_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyiapkan query insert: " + err.Error()})
		}
		defer stmt.Close()

		insertedCount := 0
		skippedCount := 0
		categoryCounters := make(map[string]int)

		// Read cell helper
		getCell := func(row []string, idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		for idx, row := range rows {
			if idx == 0 {
				continue // Skip header row
			}

			name := getCell(row, 0)
			if name == "" {
				skippedCount++
				continue
			}

			// Skip sample rows if they match the template placeholder
			sn := getCell(row, 4)
			if (sn == "PF34XYZ99" || sn == "EPS778811") && strings.Contains(name, "Thinkpad T14") && getCell(row, 14) == "Laptop operasional lab server" {
				skippedCount++
				continue
			}

			catStr := getCell(row, 1)
			brand := getCell(row, 2)
			spec := getCell(row, 3)
			yearStr := getCell(row, 5)
			dateStr := getCell(row, 6)
			priceStr := getCell(row, 7)
			qtyStr := getCell(row, 8)
			unit := getCell(row, 9)
			locStr := getCell(row, 10)
			fundingStr := getCell(row, 11)
			conditionStr := strings.ToLower(getCell(row, 12))
			borrowStr := strings.ToLower(getCell(row, 13))
			notes := getCell(row, 14)

			// Match Category
			targetCat := defaultCat
			if ci, ok := catMap[strings.ToLower(catStr)]; ok {
				targetCat = ci
			}

			// Year
			year, _ := strconv.Atoi(yearStr)
			if year <= 1900 || year > 2100 {
				year = time.Now().Year()
			}

			// Unique inventory code generated sequentially
			catKey := fmt.Sprintf("%s-%d", targetCat.Code, year)
			currentCount, exists := categoryCounters[catKey]
			if !exists {
				var count int
				db.QueryRow("SELECT COUNT(*) FROM assets WHERE acquisition_year = ?", year).Scan(&count)
				currentCount = count
			}
			currentCount++
			categoryCounters[catKey] = currentCount
			invCode := fmt.Sprintf("%s-%d-%04d", targetCat.Code, year, currentCount)
			qrContent := fmt.Sprintf("ASSET:%s", invCode)

			// Price
			priceClean := strings.ReplaceAll(strings.ReplaceAll(priceStr, "Rp", ""), ".", "")
			priceClean = strings.ReplaceAll(priceClean, ",", ".")
			priceClean = strings.TrimSpace(priceClean)
			price, _ := strconv.ParseFloat(priceClean, 64)

			// Quantity
			qty, _ := strconv.Atoi(qtyStr)
			if qty < 1 {
				qty = 1
			}
			if unit == "" {
				unit = "unit"
			}

			// Condition normalization
			condition := "good"
			if strings.Contains(conditionStr, "rusak berat") || strings.Contains(conditionStr, "broken") {
				condition = "broken"
			} else if strings.Contains(conditionStr, "rusak ringan") || strings.Contains(conditionStr, "poor") {
				condition = "poor"
			}

			// Borrowable normalization
			isBorrowable := 0
			if borrowStr == "ya" || borrowStr == "true" || borrowStr == "1" || borrowStr == "yes" || borrowStr == "bisa" {
				isBorrowable = 1
			}

			// Date parsing
			var acqDate *time.Time
			if dateStr != "" {
				dateFormats := []string{"2006-01-02", "02/01/2006", "02-01-2006", "2006/01/02"}
				for _, df := range dateFormats {
					if t, err := time.Parse(df, dateStr); err == nil {
						acqDate = &t
						break
					}
				}
			}

			// Location matching (or auto-create if non-empty)
			var locationID *int
			if locStr != "" {
				if id, ok := locMap[strings.ToLower(locStr)]; ok {
					locationID = &id
				} else {
					// Dynamically insert new location
					res, err := tx.Exec("INSERT INTO asset_locations (code, name, type, status, created_at) VALUES (?, ?, 'Room', 'active', ?)",
						"LOC-"+strings.ToUpper(invCode), locStr, time.Now())
					if err == nil {
						newID, _ := res.LastInsertId()
						intID := int(newID)
						locationID = &intID
						locMap[strings.ToLower(locStr)] = intID
					}
				}
			}

			// Funding Source matching (or auto-create if non-empty)
			var fundingSourceID *int
			if fundingStr != "" {
				if id, ok := fsMap[strings.ToLower(fundingStr)]; ok {
					fundingSourceID = &id
				} else {
					res, err := tx.Exec("INSERT INTO asset_funding_sources (code, name, created_at) VALUES (?, ?, ?)",
						strings.ToUpper(fundingStr), fundingStr, time.Now())
					if err == nil {
						newID, _ := res.LastInsertId()
						intID := int(newID)
						fundingSourceID = &intID
						fsMap[strings.ToLower(fundingStr)] = intID
					}
				}
			}

			var picStaffID *int
			var createdBy *int

			_, err = stmt.Exec(
				invCode, qrContent, targetCat.ID, name, spec, brand,
				sn, year, acqDate, price,
				fundingSourceID, locationID, picStaffID, condition, qty,
				unit, notes, "", isBorrowable, time.Now(), createdBy,
			)
			if err != nil {
				slog.Warn("Failed to insert imported asset row", "name", name, "code", invCode, "error", err)
				skippedCount++
				continue
			}
			insertedCount++
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data ke database: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":   "success",
			"message":  fmt.Sprintf("Import selesai! Sebanyak %d aset berhasil ditambahkan ke inventaris.", insertedCount),
			"inserted": insertedCount,
			"skipped":  skippedCount,
		})
	}
}
