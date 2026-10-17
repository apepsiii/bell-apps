package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func getDB(db interface{}) *sql.DB {
	if sqlDB, ok := db.(*sql.DB); ok {
		return sqlDB
	}
	return nil
}

type FaceDescriptorItem struct {
	StudentID  int         `json:"student_id"`
	Name       string      `json:"name"`
	NIS        string      `json:"nis"`
	ClassName  string      `json:"class_name"`
	Photo      string      `json:"photo"`
	Descriptor []float64   `json:"descriptor"`
}

type SaveFaceDescriptorRequest struct {
	StudentID  int         `json:"student_id"`
	Descriptor []float64   `json:"descriptor"`
}

type FaceStatusResponse struct {
	Registered bool   `json:"registered"`
	StudentID  int    `json:"student_id"`
	Name       string `json:"name,omitempty"`
	ClassName  string `json:"class_name,omitempty"`
}

func ensureFaceTable(db *sql.DB) {
	if db == nil {
		return
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS student_faces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL UNIQUE,
		descriptor TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
	)`)
}

// GetFaceDescriptors returns all enrolled students with their 128-float face descriptors
// for edge-matching on tablet/kiosk browser memory.
func GetFaceDescriptors(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}
		ensureFaceTable(database)

		rows, err := database.Query(`
			SELECT sf.student_id, sf.descriptor, s.name, COALESCE(s.nis, ''), COALESCE(c.name, ''), COALESCE(s.photo, '')
			FROM student_faces sf
			JOIN students s ON sf.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			ORDER BY s.name ASC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data wajah: " + err.Error()})
		}
		defer rows.Close()

		var list []FaceDescriptorItem
		for rows.Next() {
			var item FaceDescriptorItem
			var descStr string
			if err := rows.Scan(&item.StudentID, &descStr, &item.Name, &item.NIS, &item.ClassName, &item.Photo); err != nil {
				continue
			}

			if err := json.Unmarshal([]byte(descStr), &item.Descriptor); err == nil && len(item.Descriptor) > 0 {
				list = append(list, item)
			}
		}

		return c.JSON(http.StatusOK, list)
	}
}

// SaveFaceDescriptor saves or updates the 128-float face descriptor of a student.
func SaveFaceDescriptor(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}
		ensureFaceTable(database)

		var req SaveFaceDescriptorRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
		}

		if req.StudentID <= 0 || len(req.Descriptor) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "student_id dan descriptor wajib diisi"})
		}

		descBytes, err := json.Marshal(req.Descriptor)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal serialisasi descriptor"})
		}

		// Verify student exists
		var studentName string
		err = database.QueryRow("SELECT name FROM students WHERE id = ?", req.StudentID).Scan(&studentName)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

		// Insert or replace in SQLite
		query := `
			INSERT INTO student_faces (student_id, descriptor, updated_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(student_id) DO UPDATE SET
				descriptor = excluded.descriptor,
				updated_at = CURRENT_TIMESTAMP
		`
		if _, err := database.Exec(query, req.StudentID, string(descBytes)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan descriptor: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":     "success",
			"message":    fmt.Sprintf("Wajah %s berhasil didaftarkan", studentName),
			"student_id": req.StudentID,
		})
	}
}

// RegisterFace handles both JSON descriptors and form submissions
func RegisterFace(db interface{}) echo.HandlerFunc {
	return SaveFaceDescriptor(db)
}

// GetFaceStatus checks if a student is already enrolled in face recognition
func GetFaceStatus(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}
		ensureFaceTable(database)

		studentIDStr := c.QueryParam("student_id")
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid student ID"})
		}

		var name, className sql.NullString
		var count int
		err = database.QueryRow(`
			SELECT COUNT(sf.id), s.name, c.name
			FROM students s
			LEFT JOIN student_faces sf ON sf.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?
			GROUP BY s.id`, studentID).Scan(&count, &name, &className)

		if err != nil {
			return c.JSON(http.StatusOK, FaceStatusResponse{
				Registered: false,
				StudentID:  studentID,
			})
		}

		return c.JSON(http.StatusOK, FaceStatusResponse{
			Registered: count > 0,
			StudentID:  studentID,
			Name:       name.String,
			ClassName:  className.String,
		})
	}
}

// ListFaces returns a summary of all registered faces
func ListFaces(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}

		rows, err := database.Query(`
			SELECT sf.student_id, s.name, COALESCE(c.name, ''), sf.updated_at
			FROM student_faces sf
			JOIN students s ON sf.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			ORDER BY s.name ASC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		type FaceInfo struct {
			StudentID int    `json:"student_id"`
			Name      string `json:"name"`
			ClassName string `json:"class_name"`
			UpdatedAt string `json:"updated_at"`
		}

		var faces []FaceInfo
		for rows.Next() {
			var f FaceInfo
			if err := rows.Scan(&f.StudentID, &f.Name, &f.ClassName, &f.UpdatedAt); err == nil {
				faces = append(faces, f)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"count": len(faces),
			"faces": faces,
		})
	}
}

// DeleteFace deletes a student's face enrollment
func DeleteFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}

		studentID := c.Param("student_id")
		res, err := database.Exec("DELETE FROM student_faces WHERE student_id = ?", studentID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Data wajah tidak ditemukan"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Data wajah berhasil dihapus"})
	}
}

// VerifyFace placeholder for backward compatibility
func VerifyFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "info",
			"message": "Gunakan /api/attendance/verify-face dengan student_id hasil client-side match",
		})
	}
}
