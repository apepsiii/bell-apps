package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/models"
)

// =====================
// ASSET BORROWINGS
// =====================

// GetAssetBorrowings returns all borrowing requests
func GetAssetBorrowings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status")

		query := `
			SELECT 
				b.id, b.asset_id, a.name as asset_name,
				b.borrower_type, b.borrower_id,
				CASE 
					WHEN b.borrower_type = 'staff' THEN s.name
					WHEN b.borrower_type = 'student' THEN st.name
					ELSE ''
				END as borrower_name,
				b.purpose, b.borrow_date, b.due_date, b.return_date,
				b.status, b.approved_by, COALESCE(ap.name, '') as approved_by_name,
				b.approved_at, b.rejection_reason, b.return_condition,
				b.return_notes, b.created_at, b.updated_at
			FROM asset_borrowings b
			LEFT JOIN assets a ON b.asset_id = a.id
			LEFT JOIN staff s ON b.borrower_type = 'staff' AND b.borrower_id = s.id
			LEFT JOIN students st ON b.borrower_type = 'student' AND b.borrower_id = st.id
			LEFT JOIN staff ap ON b.approved_by = ap.id
			WHERE 1=1
		`

		args := []interface{}{}

		if status != "" {
			query += " AND b.status = ?"
			args = append(args, status)
		}

		query += " ORDER BY b.created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var borrowings []models.AssetBorrowing
		for rows.Next() {
			var b models.AssetBorrowing
			if err := rows.Scan(
				&b.ID, &b.AssetID, &b.AssetName, &b.BorrowerType, &b.BorrowerID,
				&b.BorrowerName, &b.Purpose, &b.BorrowDate, &b.DueDate, &b.ReturnDate,
				&b.Status, &b.ApprovedBy, &b.ApprovedByName, &b.ApprovedAt,
				&b.RejectionReason, &b.ReturnCondition, &b.ReturnNotes,
				&b.CreatedAt, &b.UpdatedAt,
			); err != nil {
				continue
			}
			borrowings = append(borrowings, b)
		}

		return c.JSON(http.StatusOK, borrowings)
	}
}

// CreateAssetBorrowing creates a new borrowing request
func CreateAssetBorrowing(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		assetID := c.FormValue("asset_id")
		borrowerType := c.FormValue("borrower_type")
		borrowerID := c.FormValue("borrower_id")
		purpose := c.FormValue("purpose")
		borrowDateStr := c.FormValue("borrow_date")
		dueDateStr := c.FormValue("due_date")

		if assetID == "" || borrowerType == "" || borrowerID == "" || purpose == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Required fields missing"})
		}

		// Check if asset is borrowable
		var isBorrowable int
		var currentCondition string
		err := db.QueryRow("SELECT is_borrowable, condition FROM assets WHERE id = ?", assetID).Scan(&isBorrowable, &currentCondition)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Asset not found"})
		}
		if isBorrowable == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Asset is not borrowable"})
		}
		if currentCondition == "broken" || currentCondition == "disposed" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Asset is not in borrowable condition"})
		}

		// Check if asset is already borrowed
		var activeBorrowings int
		db.QueryRow("SELECT COUNT(*) FROM asset_borrowings WHERE asset_id = ? AND status IN ('approved', 'borrowed')", assetID).Scan(&activeBorrowings)
		if activeBorrowings > 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Asset is currently borrowed"})
		}

		borrowDate, err := time.Parse("2006-01-02", borrowDateStr)
		if err != nil {
			borrowDate = time.Now()
		}

		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			dueDate = borrowDate.AddDate(0, 0, 7) // Default 7 days
		}

		_, err = db.Exec(`
			INSERT INTO asset_borrowings (
				asset_id, borrower_type, borrower_id, purpose, 
				borrow_date, due_date, status, created_at
			) VALUES (?, ?, ?, ?, ?, ?, 'pending', ?)
		`, assetID, borrowerType, borrowerID, purpose, borrowDate, dueDate, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Borrowing request created successfully"})
	}
}

// ApproveAssetBorrowing approves a borrowing request
func ApproveAssetBorrowing(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		approvedByStr := c.FormValue("approved_by") // Staff ID from session

		approvedBy, _ := strconv.Atoi(approvedByStr)

		_, err := db.Exec(`
			UPDATE asset_borrowings
			SET status = 'approved', approved_by = ?, approved_at = ?, updated_at = ?
			WHERE id = ? AND status = 'pending'
		`, approvedBy, time.Now(), time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Borrowing approved successfully"})
	}
}

// RejectAssetBorrowing rejects a borrowing request
func RejectAssetBorrowing(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		reason := c.FormValue("reason")

		if reason == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Rejection reason is required"})
		}

		_, err := db.Exec(`
			UPDATE asset_borrowings
			SET status = 'rejected', rejection_reason = ?, updated_at = ?
			WHERE id = ? AND status = 'pending'
		`, reason, time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Borrowing rejected"})
	}
}

// ReturnAssetBorrowing marks an asset as returned
func ReturnAssetBorrowing(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		returnCondition := c.FormValue("return_condition")
		returnNotes := c.FormValue("return_notes")

		if returnCondition == "" {
			returnCondition = "good"
		}

		_, err := db.Exec(`
			UPDATE asset_borrowings
			SET status = 'returned', return_date = ?, return_condition = ?, 
				return_notes = ?, updated_at = ?
			WHERE id = ? AND status IN ('approved', 'borrowed')
		`, time.Now(), returnCondition, returnNotes, time.Now(), id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Update asset condition if needed
		if returnCondition != "good" {
			var assetID int
			db.QueryRow("SELECT asset_id FROM asset_borrowings WHERE id = ?", id).Scan(&assetID)
			db.Exec("UPDATE assets SET condition = ?, updated_at = ? WHERE id = ?", returnCondition, time.Now(), assetID)
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Asset returned successfully"})
	}
}

// =====================
// MAINTENANCE TICKETS
// =====================

// GetMaintenanceTickets returns all maintenance tickets
func GetMaintenanceTickets(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status")
		priority := c.QueryParam("priority")

		query := `
			SELECT 
				t.id, t.ticket_number, t.asset_id, COALESCE(a.name, '') as asset_name,
				t.location_id, COALESCE(l.name, '') as location_name,
				t.title, t.description, t.priority, t.status,
				t.reported_by, s1.name as reporter_name, t.reported_at,
				t.assigned_to, COALESCE(s2.name, '') as assigned_to_name,
				t.photo_urls, t.resolution_notes, t.resolved_at, t.closed_at,
				t.cost, t.created_at, t.updated_at
			FROM asset_maintenance_tickets t
			LEFT JOIN assets a ON t.asset_id = a.id
			LEFT JOIN asset_locations l ON t.location_id = l.id
			LEFT JOIN staff s1 ON t.reported_by = s1.id
			LEFT JOIN staff s2 ON t.assigned_to = s2.id
			WHERE 1=1
		`

		args := []interface{}{}

		if status != "" {
			query += " AND t.status = ?"
			args = append(args, status)
		}

		if priority != "" {
			query += " AND t.priority = ?"
			args = append(args, priority)
		}

		query += " ORDER BY t.created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var tickets []models.AssetMaintenanceTicket
		for rows.Next() {
			var t models.AssetMaintenanceTicket
			if err := rows.Scan(
				&t.ID, &t.TicketNumber, &t.AssetID, &t.AssetName,
				&t.LocationID, &t.LocationName, &t.Title, &t.Description,
				&t.Priority, &t.Status, &t.ReportedBy, &t.ReporterName,
				&t.ReportedAt, &t.AssignedTo, &t.AssignedToName,
				&t.PhotoURLs, &t.ResolutionNotes, &t.ResolvedAt, &t.ClosedAt,
				&t.Cost, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				continue
			}
			tickets = append(tickets, t)
		}

		return c.JSON(http.StatusOK, tickets)
	}
}

// GenerateTicketNumber generates a unique ticket number
func GenerateTicketNumber(db *sql.DB) string {
	now := time.Now()
	dateStr := now.Format("20060102")
	var count int
	db.QueryRow("SELECT COUNT(*) FROM asset_maintenance_tickets WHERE ticket_number LIKE ?", "TKT-"+dateStr+"%").Scan(&count)
	return fmt.Sprintf("TKT-%s-%03d", dateStr, count+1)
}

// CreateMaintenanceTicket creates a new maintenance ticket
func CreateMaintenanceTicket(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		assetID := c.FormValue("asset_id")
		locationID := c.FormValue("location_id")
		title := c.FormValue("title")
		description := c.FormValue("description")
		priority := c.FormValue("priority")
		reportedByStr := c.FormValue("reported_by")
		photoURLs := c.FormValue("photo_urls")

		if title == "" || description == "" || reportedByStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Required fields missing"})
		}

		if priority == "" {
			priority = "medium"
		}

		reportedBy, _ := strconv.Atoi(reportedByStr)
		ticketNumber := GenerateTicketNumber(db)

		_, err := db.Exec(`
			INSERT INTO asset_maintenance_tickets (
				ticket_number, asset_id, location_id, title, description,
				priority, status, reported_by, reported_at, photo_urls, created_at
			) VALUES (?, ?, ?, ?, ?, ?, 'open', ?, ?, ?, ?)
		`, ticketNumber, assetID, locationID, title, description, priority,
			reportedBy, time.Now(), photoURLs, time.Now())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Update asset condition if asset_id is provided
		if assetID != "" {
			db.Exec("UPDATE assets SET condition = 'poor', updated_at = ? WHERE id = ?", time.Now(), assetID)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":       "Maintenance ticket created successfully",
			"ticket_number": ticketNumber,
		})
	}
}

// UpdateMaintenanceTicketStatus updates ticket status
func UpdateMaintenanceTicketStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		status := c.FormValue("status")
		assignedToStr := c.FormValue("assigned_to")
		resolutionNotes := c.FormValue("resolution_notes")
		costStr := c.FormValue("cost")

		if status == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Status is required"})
		}

		query := "UPDATE asset_maintenance_tickets SET status = ?, updated_at = ?"
		args := []interface{}{status, time.Now()}

		if assignedToStr != "" {
			assignedTo, _ := strconv.Atoi(assignedToStr)
			query += ", assigned_to = ?"
			args = append(args, assignedTo)
		}

		if resolutionNotes != "" {
			query += ", resolution_notes = ?"
			args = append(args, resolutionNotes)
		}

		if costStr != "" {
			cost, _ := strconv.ParseFloat(costStr, 64)
			query += ", cost = ?"
			args = append(args, cost)
		}

		if status == "resolved" {
			query += ", resolved_at = ?"
			args = append(args, time.Now())
		}

		if status == "closed" {
			query += ", closed_at = ?"
			args = append(args, time.Now())
		}

		query += " WHERE id = ?"
		args = append(args, id)

		_, err := db.Exec(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Ticket updated successfully"})
	}
}

// DeleteMaintenanceTicket deletes a maintenance ticket
func DeleteMaintenanceTicket(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		_, err := db.Exec("DELETE FROM asset_maintenance_tickets WHERE id = ?", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Ticket deleted successfully"})
	}
}
