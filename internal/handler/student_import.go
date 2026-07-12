package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PromoteRequest struct {
	StudentIDs    []int `json:"student_ids"`
	TargetClassID int   `json:"target_class_id"`
}

type BulkDeleteRequest struct {
	IDs []int `json:"ids"`
}

type StudentBasic struct {
	ID   int    `json:"id"`
	NIS  string `json:"nis"`
	Name string `json:"name"`
}

func GetStudentsJSON(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		classID := c.QueryParam("class_id")
		if classID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Class ID required"})
		}

		rows, err := db.Query("SELECT id, nis, name FROM students WHERE class_id = ? AND status = 'active' ORDER BY name ASC", classID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var students []StudentBasic
		for rows.Next() {
			var s StudentBasic
			if err := rows.Scan(&s.ID, &s.NIS, &s.Name); err != nil {
				continue
			}
			students = append(students, s)
		}

		return c.JSON(http.StatusOK, students)
	}
}

func PromoteStudents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req PromoteRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		if len(req.StudentIDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No students selected"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database transaction failed"})
		}

		query := "UPDATE students SET class_id = ? WHERE id IN ("
		args := make([]interface{}, len(req.StudentIDs)+1)
		args[0] = req.TargetClassID

		for i, id := range req.StudentIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args[i+1] = id
		}
		query += ")"

		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to promote students: " + err.Error()})
		}

		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Berhasil memindahkan siswa",
		})
	}
}

func BulkDeleteStudents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkDeleteRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
		}

		if len(req.IDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No students selected"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Transaction failed"})
		}

		query := "DELETE FROM students WHERE id IN ("
		args := make([]interface{}, len(req.IDs))
		for i, id := range req.IDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args[i] = id
		}
		query += ")"

		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to delete students: " + err.Error()})
		}

		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Berhasil menghapus siswa",
		})
	}
}

type BulkStatusRequest struct {
	IDs     []int  `json:"ids"`
	Status  string `json:"status"`
}

func BulkUpdateStudentStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkStatusRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
		}

		if len(req.IDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No students selected"})
		}

		if req.Status != "active" && req.Status != "inactive" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Status must be 'active' or 'inactive'"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Transaction failed"})
		}

		query := "UPDATE students SET status = ? WHERE id IN ("
		args := make([]interface{}, len(req.IDs)+1)
		args[0] = req.Status

		for i, id := range req.IDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args[i+1] = id
		}
		query += ")"

		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update status: " + err.Error()})
		}

		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Berhasil mengupdate status siswa",
		})
	}
}
