package handler

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Action      string
	TableName   string
	RecordID    int
	OldData     string
	NewData     string
	PerformedBy string
	IPAddress   string
	UserAgent   string
	Reason      string
}

// LogAuditTrail logs an action to the audit trail
func LogAuditTrail(db *sql.DB, c echo.Context, action, tableName string, recordID int, oldData, newData interface{}, reason string) error {
	// Get user info from context
	performedBy := "Admin" // Default
	if user := c.Get("user"); user != nil {
		if username, ok := user.(string); ok {
			performedBy = username
		}
	}

	// Get IP and User Agent
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Convert data to JSON
	var oldDataJSON, newDataJSON string
	if oldData != nil {
		if b, err := json.Marshal(oldData); err == nil {
			oldDataJSON = string(b)
		}
	}
	if newData != nil {
		if b, err := json.Marshal(newData); err == nil {
			newDataJSON = string(b)
		}
	}

	// Insert audit log
	_, err := db.Exec(`
		INSERT INTO point_audit_log 
		(action, table_name, record_id, old_data, new_data, performed_by, ip_address, user_agent, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, action, tableName, recordID, oldDataJSON, newDataJSON, performedBy, ipAddress, userAgent, reason)

	return err
}

// GetConfigValue gets a config value from point_config table
func GetConfigValue(db *sql.DB, key string, defaultValue string) string {
	var value string
	err := db.QueryRow("SELECT config_value FROM point_config WHERE config_key = ?", key).Scan(&value)
	if err != nil {
		return defaultValue
	}
	return value
}

// CheckRateLimit checks if user has exceeded daily rate limit
func CheckRateLimit(db *sql.DB, username string, maxPerDay int) (bool, int, error) {
	var todayCount int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM student_achievement_points 
		WHERE recorded_by = ? 
		AND DATE(created_at) = DATE('now')
		AND deleted_at IS NULL
	`, username).Scan(&todayCount)
	
	if err != nil {
		return false, 0, err
	}

	exceeded := todayCount >= maxPerDay
	return exceeded, todayCount, nil
}

// CheckAdminRole checks if user has permission for certain actions
func CheckAdminRole(c echo.Context, requiredRole string) bool {
	// For now, return true. In production, implement proper role checking
	// This should check against a roles table or session data
	role := c.Get("role")
	if role == nil {
		return false
	}
	
	// Example role hierarchy: superadmin > kepala_sekolah > admin > guru
	userRole := role.(string)
	
	if requiredRole == "superadmin" {
		return userRole == "superadmin"
	}
	if requiredRole == "kepala_sekolah" {
		return userRole == "superadmin" || userRole == "kepala_sekolah"
	}
	if requiredRole == "admin" {
		return userRole == "superadmin" || userRole == "kepala_sekolah" || userRole == "admin"
	}
	
	return true // Default allow for backward compatibility
}

// SoftDeleteRecord performs soft delete on a record
func SoftDeleteRecord(db *sql.DB, tableName string, recordID int, deletedBy, reason string) error {
	query := ""
	switch tableName {
	case "student_achievement_points":
		query = `UPDATE student_achievement_points 
				 SET deleted_at = ?, deleted_by = ?, deletion_reason = ? 
				 WHERE id = ? AND deleted_at IS NULL`
	case "student_violation_points":
		query = `UPDATE student_violation_points 
				 SET deleted_at = ?, deleted_by = ?, deletion_reason = ? 
				 WHERE id = ? AND deleted_at IS NULL`
	default:
		return echo.NewHTTPError(400, "Invalid table name")
	}

	result, err := db.Exec(query, time.Now().Format("2006-01-02 15:04:05"), deletedBy, reason, recordID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return echo.NewHTTPError(404, "Record not found or already deleted")
	}

	return nil
}

// GetRecordBeforeDelete retrieves record data before deletion for audit
func GetRecordBeforeDelete(db *sql.DB, tableName string, recordID int) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	var query string
	switch tableName {
	case "student_achievement_points":
		query = `SELECT id, student_id, rule_id, points, description, recorded_by, academic_year, created_at 
				 FROM student_achievement_points WHERE id = ? AND deleted_at IS NULL`
	case "student_violation_points":
		query = `SELECT id, student_id, rule_id, occurrence, points, description, recorded_by, academic_year, created_at 
				 FROM student_violation_points WHERE id = ? AND deleted_at IS NULL`
	default:
		return nil, echo.NewHTTPError(400, "Invalid table name")
	}

	rows, err := db.Query(query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, echo.NewHTTPError(404, "Record not found")
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create a slice of interface{} to hold values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan row
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	// Map column names to values
	for i, col := range columns {
		data[col] = values[i]
	}

	return data, nil
}
