package handler

import (
	"database/sql"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
	"belsekolah/pkg/qrcode"
	"belsekolah/pkg/utils"
)

func AddStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rfid := c.FormValue("rfid_uid")
		nis := c.FormValue("nis")
		nisSiswa := c.FormValue("nis_siswa")
		name := c.FormValue("name")
		parentName := c.FormValue("parent_name")
		phone := utils.FormatPhone(c.FormValue("parent_phone"))
		classID := c.FormValue("class_id")
		birthday := c.FormValue("birthday")
		status := c.FormValue("status")
		password := c.FormValue("password")

		// Default status to active if not provided
		if status == "" {
			status = "active"
		}

		// Hash password if provided
		var hashedPassword string
		if password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi password"})
			}
			hashedPassword = string(hash)
		}

		photoFile := ""

		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				ext := utils.GetPhotoExtension(file.Filename)
				filename := utils.BuildPhotoFilename(nis, ext)
				dstPath := filepath.Join(config.GetPhotoPath(), filename)
				if err := utils.SaveUploadedFile(src, dstPath); err == nil {
					photoFile = filename
				}
			}
		} else {
			capturedPhoto := c.FormValue("captured_photo")
			if capturedPhoto != "" {
				filename := utils.BuildPhotoFilename(nis, ".jpg")
				dstPath := filepath.Join(config.GetPhotoPath(), filename)
				if err := utils.SaveBase64Image(capturedPhoto, dstPath); err == nil {
					photoFile = filename
				}
			}
		}

		_, err = db.Exec("INSERT INTO students (rfid_uid, nis, nis_siswa, name, parent_name, parent_phone, class_id, photo, birthday, status, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", rfid, nis, nisSiswa, name, parentName, phone, classID, photoFile, birthday, status, hashedPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal (Mungkin RFID/NIS duplikat): " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa ditambahkan"})
	}
}

func UpdateStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		rfid := c.FormValue("rfid_uid")
		nis := c.FormValue("nis")
		nisSiswa := c.FormValue("nis_siswa")
		name := c.FormValue("name")
		parentName := c.FormValue("parent_name")
		phone := utils.FormatPhone(c.FormValue("parent_phone"))
		classID := c.FormValue("class_id")
		birthday := c.FormValue("birthday")
		status := c.FormValue("status")
		password := c.FormValue("password")

		// Hash password if provided
		var hashedPassword string
		if password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi password"})
			}
			hashedPassword = string(hash)
		}

		photoFile := ""

		file, err := c.FormFile("photo")
		if err == nil {
			src, err := file.Open()
			if err == nil {
				defer src.Close()
				ext := utils.GetPhotoExtension(file.Filename)
				filename := utils.BuildPhotoFilename(nis, ext)
				dstPath := filepath.Join(config.GetPhotoPath(), filename)
				if err := utils.SaveUploadedFile(src, dstPath); err == nil {
					photoFile = filename
					// Update with photo
					if hashedPassword != "" {
						db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, photo=?, birthday=?, status=?, password=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, photoFile, birthday, status, hashedPassword, id)
					} else {
						db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, photo=?, birthday=?, status=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, photoFile, birthday, status, id)
					}
					return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
				}
			}
		} else {
			capturedPhoto := c.FormValue("captured_photo")
			if capturedPhoto != "" {
				filename := utils.BuildPhotoFilename(nis, ".jpg")
				dstPath := filepath.Join(config.GetPhotoPath(), filename)
				if err := utils.SaveBase64Image(capturedPhoto, dstPath); err == nil {
					photoFile = filename
					// Update with captured photo
					if hashedPassword != "" {
						db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, photo=?, birthday=?, status=?, password=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, photoFile, birthday, status, hashedPassword, id)
					} else {
						db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, photo=?, birthday=?, status=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, photoFile, birthday, status, id)
					}
					return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
				}
			}
		}

		// Update without photo
		if hashedPassword != "" {
			_, err = db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, birthday=?, status=?, password=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, birthday, status, hashedPassword, id)
		} else {
			_, err = db.Exec("UPDATE students SET rfid_uid=?, nis=?, nis_siswa=?, name=?, parent_name=?, parent_phone=?, class_id=?, birthday=?, status=? WHERE id=?", rfid, nis, nisSiswa, name, parentName, phone, classID, birthday, status, id)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Data siswa diperbarui"})
	}
}

func UpdateStudentStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		type StatusUpdateRequest struct {
			Status string `json:"status"`
		}

		var req StatusUpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
		}

		if req.Status != "active" && req.Status != "inactive" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Status must be 'active' or 'inactive'"})
		}

		_, err := db.Exec("UPDATE students SET status = ? WHERE id = ?", req.Status, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Status siswa diperbarui"})
	}
}

func DeleteStudent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id != "" && id != "0" {
			_, err := db.Exec("DELETE FROM students WHERE id=?", id)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Siswa dihapus"})
	}
}

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
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No IDs provided"})
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
	IDs    []int  `json:"ids"`
	Status string `json:"status"`
}

func BulkUpdateStudentStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req BulkStatusRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
		}

		if len(req.IDs) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No IDs provided"})
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

func GetStudentIDCard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		var student struct {
			ID        int
			NIS       string
			Name      string
			ClassID   int
			ClassName string
			RFID      string
		}

		err := db.QueryRow(`
			SELECT s.id, s.nis, s.name, s.class_id, c.name, s.rfid_uid
			FROM students s
			LEFT JOIN classes c ON s.class_id = c.id
			WHERE s.id = ?
		`, id).Scan(&student.ID, &student.NIS, &student.Name, &student.ClassID, &student.ClassName, &student.RFID)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Student not found"})
		}

		qrImage, err := qrcode.GenerateStudentCard(config.GetDomain(), student.RFID, student.ID, student.NIS, student.Name, student.ClassName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to generate QR"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":        student.ID,
			"nis":       student.NIS,
			"name":      student.Name,
			"class":     student.ClassName,
			"rfid":      student.RFID,
			"qr_image":  qrImage,
		})
	}
}

func GetAllStudentsForIDCard(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		classID := c.QueryParam("class_id")

		var query string
		var args []interface{}

		if classID != "" {
			query = `
				SELECT s.id, s.nis, s.name, COALESCE(c.name, '-') as class_name, s.rfid_uid
				FROM students s
				LEFT JOIN classes c ON s.class_id = c.id
				WHERE s.class_id = ? AND s.status = 'active'
				ORDER BY c.name, s.name
			`
			args = []interface{}{classID}
		} else {
			query = `
				SELECT s.id, s.nis, s.name, COALESCE(c.name, '-') as class_name, s.rfid_uid
				FROM students s
				LEFT JOIN classes c ON s.class_id = c.id
				WHERE s.status = 'active'
				ORDER BY c.name, s.name
			`
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		type StudentCard struct {
			ID      int    `json:"id"`
			NIS     string `json:"nis"`
			Name    string `json:"name"`
			Class   string `json:"class"`
			RFID    string `json:"rfid"`
			QRImage string `json:"qr_image"`
		}

		var students []StudentCard
		for rows.Next() {
			var s StudentCard
			if err := rows.Scan(&s.ID, &s.NIS, &s.Name, &s.Class, &s.RFID); err != nil {
				continue
			}
			qrImage, err := qrcode.GenerateStudentCard(config.GetDomain(), s.RFID, s.ID, s.NIS, s.Name, s.Class)
			if err == nil {
				s.QRImage = qrImage
			}
			students = append(students, s)
		}

		return c.JSON(http.StatusOK, students)
	}
}
