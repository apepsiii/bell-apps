package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func getDB(db interface{}) *sql.DB {
	if sqlDB, ok := db.(*sql.DB); ok {
		return sqlDB
	}
	return nil
}

type FaceDescriptorItem struct {
	ID         int       `json:"id"`
	StudentID  int       `json:"student_id,omitempty"`
	StaffID    int       `json:"staff_id,omitempty"`
	UserType   string    `json:"user_type"` // "Siswa" or "Staff"
	Name       string    `json:"name"`
	NIS        string    `json:"nis,omitempty"`
	NIP        string    `json:"nip,omitempty"`
	ClassName  string    `json:"class_name,omitempty"`
	Role       string    `json:"role,omitempty"`
	Gender     string    `json:"gender,omitempty"`
	Honorific  string    `json:"honorific,omitempty"`
	Photo      string    `json:"photo"`
	Descriptor []float64 `json:"descriptor"`
}

type SaveFaceDescriptorRequest struct {
	StudentID  int       `json:"student_id"`
	StaffID    int       `json:"staff_id"`
	UserType   string    `json:"user_type"` // "Siswa" or "Staff"
	Descriptor []float64 `json:"descriptor"`
}

type FaceStatusResponse struct {
	Registered bool   `json:"registered"`
	UserType   string `json:"user_type"`
	StudentID  int    `json:"student_id,omitempty"`
	StaffID    int    `json:"staff_id,omitempty"`
	Name       string `json:"name,omitempty"`
	ClassName  string `json:"class_name,omitempty"`
	Role       string `json:"role,omitempty"`
}

// DetermineStaffHonorific determines "Bapak" or "Ibu" based on gender or name clues
func DetermineStaffHonorific(gender, name string) string {
	g := strings.ToUpper(strings.TrimSpace(gender))
	if g == "P" || g == "PEREMPUAN" || g == "FEMALE" || g == "IBU" {
		return "Ibu"
	}
	if g == "L" || g == "LAKI-LAKI" || g == "MALE" || g == "BAPAK" {
		return "Bapak"
	}

	lower := strings.ToLower(name)
	femaleKeywords := []string{"ibu", "neneng", "ida", "siti", "dewi", "putri", "ayu", "ani", "nur", "sri", "endang", "retno", "wulan", "tri", "lestari", "rahayu", "diah", "ratna", "riri", "rina", "rini", "dian", "yuliana", "fitri"}
	for _, kw := range femaleKeywords {
		if strings.Contains(lower, kw) {
			return "Ibu"
		}
	}
	return "Bapak"
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
	db.Exec(`CREATE TABLE IF NOT EXISTS staff_faces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		staff_id INTEGER NOT NULL UNIQUE,
		descriptor TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (staff_id) REFERENCES staff(id) ON DELETE CASCADE
	)`)
}

// GetFaceDescriptors returns all enrolled students AND staff with their 128-float face descriptors
// for edge-matching on tablet/kiosk browser memory.
func GetFaceDescriptors(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}
		ensureFaceTable(database)

		var list []FaceDescriptorItem

		// 1. Enrolled Students
		studentRows, err := database.Query(`
			SELECT sf.student_id, sf.descriptor, s.name, COALESCE(s.nis, ''), COALESCE(c.name, ''), COALESCE(s.photo, '')
			FROM student_faces sf
			JOIN students s ON sf.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.status = 'active'
			ORDER BY s.name ASC
		`)
		if err == nil {
			defer studentRows.Close()
			for studentRows.Next() {
				var item FaceDescriptorItem
				var descStr string
				if err := studentRows.Scan(&item.StudentID, &descStr, &item.Name, &item.NIS, &item.ClassName, &item.Photo); err == nil {
					if err := json.Unmarshal([]byte(descStr), &item.Descriptor); err == nil && len(item.Descriptor) > 0 {
						item.ID = item.StudentID
						item.UserType = "Siswa"
						list = append(list, item)
					}
				}
			}
		}

		// 2. Enrolled Staff / Teachers
		staffRows, err := database.Query(`
			SELECT sf.staff_id, sf.descriptor, s.name, COALESCE(s.nip, '-'), COALESCE(s.role, 'Staff'), COALESCE(s.gender, 'L')
			FROM staff_faces sf
			JOIN staff s ON sf.staff_id = s.id
			ORDER BY s.name ASC
		`)
		if err == nil {
			defer staffRows.Close()
			for staffRows.Next() {
				var staffID int
				var descStr, name, nip, role, gender string
				if err := staffRows.Scan(&staffID, &descStr, &name, &nip, &role, &gender); err == nil {
					var desc []float64
					if err := json.Unmarshal([]byte(descStr), &desc); err == nil && len(desc) > 0 {
						honorific := DetermineStaffHonorific(gender, name)
						list = append(list, FaceDescriptorItem{
							ID:         staffID,
							StaffID:    staffID,
							UserType:   "Staff",
							Name:       name,
							NIP:        nip,
							Role:       role,
							ClassName:  role,
							Gender:     gender,
							Honorific:  honorific,
							Photo:      "",
							Descriptor: desc,
						})
					}
				}
			}
		}

		return c.JSON(http.StatusOK, list)
	}
}

// SaveFaceDescriptor saves or updates the 128-float face descriptor of a student or staff.
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

		if len(req.Descriptor) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "descriptor wajah wajib diisi"})
		}

		descBytes, err := json.Marshal(req.Descriptor)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal serialisasi descriptor"})
		}

		// Handle Staff Face Enrollment
		if req.UserType == "Staff" || req.StaffID > 0 {
			staffID := req.StaffID
			if staffID <= 0 {
				staffID = req.StudentID
			}
			if staffID <= 0 {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "staff_id wajib diisi"})
			}

			var staffName, staffRole, staffGender string
			err = database.QueryRow("SELECT name, COALESCE(role, 'Staff'), COALESCE(gender, 'L') FROM staff WHERE id = ?", staffID).Scan(&staffName, &staffRole, &staffGender)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"message": "Guru / Staff tidak ditemukan"})
			}

			query := `
				INSERT INTO staff_faces (staff_id, descriptor, updated_at)
				VALUES (?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT(staff_id) DO UPDATE SET
					descriptor = excluded.descriptor,
					updated_at = CURRENT_TIMESTAMP
			`
			if _, err := database.Exec(query, staffID, string(descBytes)); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan wajah staff: " + err.Error()})
			}

			honorific := DetermineStaffHonorific(staffGender, staffName)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":    "success",
				"message":   fmt.Sprintf("Wajah %s %s berhasil didaftarkan", honorific, staffName),
				"user_type": "Staff",
				"staff_id":  staffID,
			})
		}

		// Handle Student Face Enrollment
		if req.StudentID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "student_id wajib diisi"})
		}

		var studentName string
		err = database.QueryRow("SELECT name FROM students WHERE id = ?", req.StudentID).Scan(&studentName)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Siswa tidak ditemukan"})
		}

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
			"user_type":  "Siswa",
			"student_id": req.StudentID,
		})
	}
}

// RegisterFace handles both JSON descriptors and form submissions
func RegisterFace(db interface{}) echo.HandlerFunc {
	return SaveFaceDescriptor(db)
}

// GetFaceStatus checks if a student or staff is already enrolled in face recognition
func GetFaceStatus(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}
		ensureFaceTable(database)

		staffIDStr := c.QueryParam("staff_id")
		if staffIDStr != "" {
			staffID, _ := strconv.Atoi(staffIDStr)
			if staffID > 0 {
				var name, role sql.NullString
				var count int
				err := database.QueryRow(`
					SELECT COUNT(sf.id), s.name, s.role
					FROM staff s
					LEFT JOIN staff_faces sf ON sf.staff_id = s.id
					WHERE s.id = ?
					GROUP BY s.id`, staffID).Scan(&count, &name, &role)

				if err != nil {
					return c.JSON(http.StatusOK, FaceStatusResponse{
						Registered: false,
						UserType:   "Staff",
						StaffID:    staffID,
					})
				}

				return c.JSON(http.StatusOK, FaceStatusResponse{
					Registered: count > 0,
					UserType:   "Staff",
					StaffID:    staffID,
					Name:       name.String,
					Role:       role.String,
				})
			}
		}

		studentIDStr := c.QueryParam("student_id")
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
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
				UserType:   "Siswa",
				StudentID:  studentID,
			})
		}

		return c.JSON(http.StatusOK, FaceStatusResponse{
			Registered: count > 0,
			UserType:   "Siswa",
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
		ensureFaceTable(database)

		type FaceInfo struct {
			ID        int    `json:"id"`
			UserType  string `json:"user_type"`
			Name      string `json:"name"`
			ExtraInfo string `json:"extra_info"`
			UpdatedAt string `json:"updated_at"`
		}

		var faces []FaceInfo

		// Students
		rows, err := database.Query(`
			SELECT sf.student_id, s.name, COALESCE(c.name, ''), sf.updated_at
			FROM student_faces sf
			JOIN students s ON sf.student_id = s.id
			LEFT JOIN classes c ON s.class_id = c.id
			ORDER BY s.name ASC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var f FaceInfo
				if err := rows.Scan(&f.ID, &f.Name, &f.ExtraInfo, &f.UpdatedAt); err == nil {
					f.UserType = "Siswa"
					faces = append(faces, f)
				}
			}
		}

		// Staff
		staffRows, err := database.Query(`
			SELECT sf.staff_id, s.name, COALESCE(s.role, 'Staff'), sf.updated_at
			FROM staff_faces sf
			JOIN staff s ON sf.staff_id = s.id
			ORDER BY s.name ASC
		`)
		if err == nil {
			defer staffRows.Close()
			for staffRows.Next() {
				var f FaceInfo
				if err := staffRows.Scan(&f.ID, &f.Name, &f.ExtraInfo, &f.UpdatedAt); err == nil {
					f.UserType = "Staff"
					faces = append(faces, f)
				}
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

// DeleteStaffFace deletes a staff's face enrollment
func DeleteStaffFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		database := getDB(db)
		if database == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}

		staffID := c.Param("staff_id")
		res, err := database.Exec("DELETE FROM staff_faces WHERE staff_id = ?", staffID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Data wajah staff tidak ditemukan"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Data wajah staff berhasil dihapus"})
	}
}

// VerifyFace placeholder for backward compatibility
func VerifyFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "info",
			"message": "Gunakan /api/attendance/verify-face dengan student_id atau staff_id hasil client-side match",
		})
	}
}
