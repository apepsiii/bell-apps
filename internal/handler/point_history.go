package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type PointHistoryItem struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	StudentName  string `json:"student_name"`
	ClassName    string `json:"class_name"`
	PointsChange int    `json:"points_change"`
	Description  string `json:"description"`
	Timestamp    string `json:"timestamp"`
	RecordedBy   string `json:"recorded_by"`
}

// GetPointHistory returns point transaction history with filters
func GetPointHistory(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.QueryParam("student_id")
		typeFilter := c.QueryParam("type") // positive, negative, or empty (all)
		limitStr := c.QueryParam("limit")

		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		query := `
			SELECT sp.id, sp.student_id, s.name, COALESCE(c.name, ''), 
			       sp.points_change, sp.description, sp.timestamp, 
			       COALESCE(sp.recorded_by, '')
			FROM student_points sp
			JOIN students s ON sp.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE 1=1
		`

		args := []interface{}{}

		// Filter by student
		if studentID != "" {
			query += " AND sp.student_id = ?"
			args = append(args, studentID)
		}

		// Filter by type
		if typeFilter == "positive" {
			query += " AND sp.points_change > 0"
		} else if typeFilter == "negative" {
			query += " AND sp.points_change < 0"
		}

		query += " ORDER BY sp.timestamp DESC LIMIT ?"
		args = append(args, limit)

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var history []PointHistoryItem
		for rows.Next() {
			var item PointHistoryItem
			if err := rows.Scan(&item.ID, &item.StudentID, &item.StudentName, &item.ClassName,
				&item.PointsChange, &item.Description, &item.Timestamp, &item.RecordedBy); err != nil {
				continue
			}
			history = append(history, item)
		}

		if history == nil {
			history = []PointHistoryItem{}
		}

		return c.JSON(http.StatusOK, history)
	}
}
