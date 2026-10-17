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

	"belsekolah/internal/config"
	"belsekolah/internal/models"
)

// parseSqliteIntPtr safely converts interface{} values from SQLite into *int
func parseSqliteIntPtr(val interface{}) *int {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case int:
		return &v
	case int64:
		iv := int(v)
		return &iv
	case float64:
		iv := int(v)
		return &iv
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		if iv, err := strconv.Atoi(v); err == nil {
			return &iv
		}
	case []byte:
		return parseSqliteIntPtr(string(v))
	}
	return nil
}

// parseSqliteTime safely parses date/time strings returned by SQLite
func parseSqliteTime(val interface{}) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return time.Time{}
		}
		// Strip monotonic clock reading if present: " ... m=+..."
		if idx := strings.Index(v, " m="); idx != -1 {
			v = strings.TrimSpace(v[:idx])
		}
		formats := []string{
			"2006-01-02 15:04:05.999999999 -0700 MST",
			"2006-01-02 15:04:05.999999999 +0700 +07",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
			time.RFC3339,
		}
		for _, f := range formats {
			if t, err := time.Parse(f, v); err == nil {
				return t
			}
		}
		if len(v) >= 19 {
			if t, err := time.Parse("2006-01-02 15:04:05", v[:19]); err == nil {
				return t
			}
		}
		if len(v) >= 10 {
			if t, err := time.Parse("2006-01-02", v[:10]); err == nil {
				return t
			}
		}
	case []byte:
		return parseSqliteTime(string(v))
	}
	return time.Time{}
}

// SarprasPageData holds the aggregated view data for the Sarpras module
type SarprasPageData struct {
	PageTitle          string
	Username           string
	BuildDate          string
	AppVersion         string
	ActiveTab          string
	Assets             []models.Asset
	Categories         []models.AssetCategory
	Locations          []models.AssetLocation
	FundingSources     []models.AssetFundingSource
	StaffList          []models.Staff
	Students           []models.Student
	Borrowings         []models.AssetBorrowing
	MaintenanceTickets []models.AssetMaintenanceTicket
	Stats              struct {
		TotalAssets      int
		TotalValue       float64
		TotalValueFormat string
		GoodCondition    int
		NeedsRepair      int
		ActiveBorrowings int
		OpenTickets      int
	}
}

// SarprasPage serves the modern unified Sarpras & Asset Management admin console
func SarprasPage(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var data SarprasPageData
		data.PageTitle = "Sarpras & Manajemen Aset"
		data.Username = "admin"
		data.BuildDate = "2026"
		data.AppVersion = config.AppVersion

		// Detect active tab from query param or route
		tab := c.QueryParam("tab")
		if tab == "" {
			path := c.Path()
			if strings.Contains(path, "borrowing") {
				tab = "borrowings"
			} else if strings.Contains(path, "maintenance") || strings.Contains(path, "ticket") {
				tab = "maintenance"
			} else if strings.Contains(path, "categor") || strings.Contains(path, "location") || strings.Contains(path, "funding") || strings.Contains(path, "master") {
				tab = "master"
			} else {
				tab = "assets"
			}
		}
		data.ActiveTab = tab

		// 1. Fetch Assets with category, location, and funding source
		assetRows, err := db.Query(`
			SELECT 
				a.id, 
				a.inventory_code, 
				COALESCE(a.qr_code, ''), 
				a.category_id, 
				COALESCE(ac.name, '') as category_name,
				a.name, 
				COALESCE(a.specification, ''), 
				COALESCE(a.brand, ''), 
				COALESCE(a.serial_number, ''),
				COALESCE(a.acquisition_year, 0), 
				COALESCE(a.acquisition_date, ''), 
				COALESCE(a.purchase_price, 0),
				a.funding_source_id, 
				COALESCE(fs.name, '') as funding_source,
				a.location_id, 
				COALESCE(l.name, '') as location_name,
				a.pic_staff_id, 
				COALESCE(s.name, '') as pic_name,
				COALESCE(a.condition, 'good'), 
				COALESCE(a.quantity, 1), 
				COALESCE(a.unit, 'unit'), 
				COALESCE(a.notes, ''), 
				COALESCE(a.photo_url, ''),
				COALESCE(a.is_borrowable, 0), 
				COALESCE(a.created_at, '')
			FROM assets a
			LEFT JOIN asset_categories ac ON a.category_id = ac.id
			LEFT JOIN asset_funding_sources fs ON NULLIF(a.funding_source_id, '') = fs.id
			LEFT JOIN asset_locations l ON NULLIF(a.location_id, '') = l.id
			LEFT JOIN staff s ON NULLIF(a.pic_staff_id, '') = s.id
			ORDER BY a.created_at DESC
		`)
		if err == nil {
			defer assetRows.Close()
			for assetRows.Next() {
				var a models.Asset
				var isBorrowable int
				var rawAcqDate, rawCreatedAt interface{}
				var rawFundingSourceID, rawLocationID, rawPICStaffID interface{}
				if err := assetRows.Scan(
					&a.ID, &a.InventoryCode, &a.QRCode, &a.CategoryID,
					&a.CategoryName, &a.Name, &a.Specification, &a.Brand,
					&a.SerialNumber, &a.AcquisitionYear, &rawAcqDate,
					&a.PurchasePrice, &rawFundingSourceID, &a.FundingSource,
					&rawLocationID, &a.LocationName, &rawPICStaffID, &a.PICName,
					&a.Condition, &a.Quantity, &a.Unit, &a.Notes,
					&a.PhotoURL, &isBorrowable, &rawCreatedAt,
				); err == nil {
					a.IsBorrowable = (isBorrowable == 1)
					a.FundingSourceID = parseSqliteIntPtr(rawFundingSourceID)
					a.LocationID = parseSqliteIntPtr(rawLocationID)
					a.PICStaffID = parseSqliteIntPtr(rawPICStaffID)
					if t := parseSqliteTime(rawAcqDate); !t.IsZero() {
						a.AcquisitionDate = &t
					}
					a.CreatedAt = parseSqliteTime(rawCreatedAt)
					data.Assets = append(data.Assets, a)
				} else {
					slog.Warn("Failed to scan asset row", "error", err)
				}
			}
		} else {
			slog.Warn("Failed to query assets", "error", err)
		}

		// 2. Fetch Categories
		catRows, err := db.Query(`SELECT id, code, name, COALESCE(description, ''), COALESCE(created_at, '') FROM asset_categories ORDER BY code ASC`)
		if err == nil {
			defer catRows.Close()
			for catRows.Next() {
				var cat models.AssetCategory
				var rawCreated interface{}
				if err := catRows.Scan(&cat.ID, &cat.Code, &cat.Name, &cat.Description, &rawCreated); err == nil {
					cat.CreatedAt = parseSqliteTime(rawCreated)
					data.Categories = append(data.Categories, cat)
				}
			}
		}

		// 3. Fetch Locations
		locRows, err := db.Query(`
			SELECT 
				l.id, l.code, l.name, COALESCE(l.type, ''), COALESCE(l.capacity, 0),
				l.pic_staff_id, COALESCE(s.name, '') as pic_name,
				COALESCE(l.status, 'active'), COALESCE(l.created_at, '')
			FROM asset_locations l
			LEFT JOIN staff s ON NULLIF(l.pic_staff_id, '') = s.id
			ORDER BY l.name ASC
		`)
		if err == nil {
			defer locRows.Close()
			for locRows.Next() {
				var loc models.AssetLocation
				var rawCreated, rawPICStaffID interface{}
				if err := locRows.Scan(
					&loc.ID, &loc.Code, &loc.Name, &loc.Type, &loc.Capacity,
					&rawPICStaffID, &loc.PICName, &loc.Status, &rawCreated,
				); err == nil {
					loc.PICStaffID = parseSqliteIntPtr(rawPICStaffID)
					loc.CreatedAt = parseSqliteTime(rawCreated)
					data.Locations = append(data.Locations, loc)
				}
			}
		}

		// 4. Fetch Funding Sources
		fsRows, err := db.Query(`SELECT id, code, name, COALESCE(description, ''), COALESCE(created_at, '') FROM asset_funding_sources ORDER BY name ASC`)
		if err == nil {
			defer fsRows.Close()
			for fsRows.Next() {
				var fs models.AssetFundingSource
				var rawCreated interface{}
				if err := fsRows.Scan(&fs.ID, &fs.Code, &fs.Name, &fs.Description, &rawCreated); err == nil {
					fs.CreatedAt = parseSqliteTime(rawCreated)
					data.FundingSources = append(data.FundingSources, fs)
				}
			}
		}

		// 5. Fetch Staff for PIC assignment
		staffRows, err := db.Query(`SELECT id, COALESCE(nip, ''), name, COALESCE(phone, ''), COALESCE(role, '') FROM staff ORDER BY name ASC`)
		if err == nil {
			defer staffRows.Close()
			for staffRows.Next() {
				var st models.Staff
				if err := staffRows.Scan(&st.ID, &st.NIP, &st.Name, &st.Phone, &st.Role); err == nil {
					data.StaffList = append(data.StaffList, st)
				}
			}
		}

		// 6. Fetch Active Students for borrowing dropdown
		studentRows, err := db.Query(`SELECT id, nis, name FROM students WHERE status = 'active' ORDER BY name ASC`)
		if err == nil {
			defer studentRows.Close()
			for studentRows.Next() {
				var stu models.Student
				if err := studentRows.Scan(&stu.ID, &stu.NIS, &stu.Name); err == nil {
					data.Students = append(data.Students, stu)
				}
			}
		}

		// 7. Fetch Borrowings
		borrowRows, err := db.Query(`
			SELECT 
				b.id, b.asset_id, COALESCE(a.name, '') as asset_name,
				b.borrower_type, b.borrower_id,
				CASE 
					WHEN b.borrower_type = 'staff' THEN COALESCE(s.name, '')
					WHEN b.borrower_type = 'student' THEN COALESCE(st.name, '')
					ELSE ''
				END as borrower_name,
				COALESCE(b.purpose, ''), COALESCE(b.borrow_date, ''), COALESCE(b.due_date, ''), COALESCE(b.return_date, ''),
				COALESCE(b.status, 'pending'), b.approved_by, COALESCE(ap.name, '') as approved_by_name,
				COALESCE(b.approved_at, ''), COALESCE(b.rejection_reason, ''), COALESCE(b.return_condition, ''),
				COALESCE(b.return_notes, ''), COALESCE(b.created_at, '')
			FROM asset_borrowings b
			LEFT JOIN assets a ON b.asset_id = a.id
			LEFT JOIN staff s ON b.borrower_type = 'staff' AND b.borrower_id = s.id
			LEFT JOIN students st ON b.borrower_type = 'student' AND b.borrower_id = st.id
			LEFT JOIN staff ap ON b.approved_by = ap.id
			ORDER BY b.created_at DESC
		`)
		if err == nil {
			defer borrowRows.Close()
			for borrowRows.Next() {
				var b models.AssetBorrowing
				var rawBorrowDate, rawDueDate, rawReturnDate, rawApprovedAt, rawCreated interface{}
				var rawApprovedBy interface{}
				if err := borrowRows.Scan(
					&b.ID, &b.AssetID, &b.AssetName,
					&b.BorrowerType, &b.BorrowerID, &b.BorrowerName,
					&b.Purpose, &rawBorrowDate, &rawDueDate, &rawReturnDate,
					&b.Status, &rawApprovedBy, &b.ApprovedByName,
					&rawApprovedAt, &b.RejectionReason, &b.ReturnCondition,
					&b.ReturnNotes, &rawCreated,
				); err == nil {
					b.ApprovedBy = parseSqliteIntPtr(rawApprovedBy)
					b.BorrowDate = parseSqliteTime(rawBorrowDate)
					b.DueDate = parseSqliteTime(rawDueDate)
					if t := parseSqliteTime(rawReturnDate); !t.IsZero() {
						b.ReturnDate = &t
					}
					if t := parseSqliteTime(rawApprovedAt); !t.IsZero() {
						b.ApprovedAt = &t
					}
					b.CreatedAt = parseSqliteTime(rawCreated)
					data.Borrowings = append(data.Borrowings, b)
				}
			}
		}

		// 8. Fetch Maintenance Tickets
		ticketRows, err := db.Query(`
			SELECT 
				t.id, t.ticket_number, t.asset_id, COALESCE(a.name, '') as asset_name,
				t.location_id, COALESCE(l.name, '') as location_name,
				t.title, COALESCE(t.description, ''), COALESCE(t.priority, 'medium'), COALESCE(t.status, 'open'),
				t.reported_by, COALESCE(s1.name, 'Admin') as reporter_name, COALESCE(t.reported_at, ''),
				t.assigned_to, COALESCE(s2.name, '') as assigned_to_name,
				COALESCE(t.cost, 0), COALESCE(t.created_at, '')
			FROM asset_maintenance_tickets t
			LEFT JOIN assets a ON t.asset_id = a.id
			LEFT JOIN asset_locations l ON t.location_id = l.id
			LEFT JOIN staff s1 ON t.reported_by = s1.id
			LEFT JOIN staff s2 ON NULLIF(t.assigned_to, '') = s2.id
			ORDER BY t.created_at DESC
		`)
		if err == nil {
			defer ticketRows.Close()
			for ticketRows.Next() {
				var t models.AssetMaintenanceTicket
				var rawReportedAt, rawCreated, rawAssignedTo interface{}
				if err := ticketRows.Scan(
					&t.ID, &t.TicketNumber, &t.AssetID, &t.AssetName,
					&t.LocationID, &t.LocationName,
					&t.Title, &t.Description, &t.Priority, &t.Status,
					&t.ReportedBy, &t.ReporterName, &rawReportedAt,
					&rawAssignedTo, &t.AssignedToName,
					&t.Cost, &rawCreated,
				); err == nil {
					t.AssignedTo = parseSqliteIntPtr(rawAssignedTo)
					t.ReportedAt = parseSqliteTime(rawReportedAt)
					t.CreatedAt = parseSqliteTime(rawCreated)
					data.MaintenanceTickets = append(data.MaintenanceTickets, t)
				}
			}
		}

		// 9. Aggregate Stats
		db.QueryRow("SELECT COUNT(*) FROM assets").Scan(&data.Stats.TotalAssets)
		db.QueryRow("SELECT COALESCE(SUM(purchase_price * quantity), 0) FROM assets").Scan(&data.Stats.TotalValue)
		data.Stats.TotalValueFormat = fmt.Sprintf("Rp %.0f", data.Stats.TotalValue)
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE condition = 'good'").Scan(&data.Stats.GoodCondition)
		db.QueryRow("SELECT COUNT(*) FROM assets WHERE condition IN ('poor', 'broken')").Scan(&data.Stats.NeedsRepair)
		db.QueryRow("SELECT COUNT(*) FROM asset_borrowings WHERE status IN ('approved', 'borrowed')").Scan(&data.Stats.ActiveBorrowings)
		db.QueryRow("SELECT COUNT(*) FROM asset_maintenance_tickets WHERE status IN ('open', 'in_progress')").Scan(&data.Stats.OpenTickets)

		return c.Render(http.StatusOK, "sarpras.html", data)
	}
}
