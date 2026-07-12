package handler

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"belsekolah/pkg/utils"
)

func ImportStudents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
		}
		defer src.Close()

		reader := csv.NewReader(src)
		records, err := reader.ReadAll()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Format CSV salah"})
		}

		classMap := make(map[string]int)
		rows, _ := db.Query("SELECT id, name FROM classes")
		for rows.Next() {
			var id int
			var name string
			rows.Scan(&id, &name)
			classMap[strings.ToUpper(name)] = id
		}
		rows.Close()

		tx, _ := db.Begin()
		successCount := 0

		for i, row := range records {
			if i == 0 {
				continue
			}
			if len(row) < 5 {
				continue
			}

			nis := strings.TrimSpace(row[0])
			name := strings.TrimSpace(row[1])
			classRaw := strings.TrimSpace(row[2])
			phone := utils.FormatPhone(row[3])
			rfid := strings.TrimSpace(row[4])

			var classID int
			if id, ok := classMap[strings.ToUpper(classRaw)]; ok {
				classID = id
			} else {
				fmt.Sscanf(classRaw, "%d", &classID)
			}

			_, err := tx.Exec("INSERT OR IGNORE INTO students (rfid_uid, nis, name, parent_phone, class_id) VALUES (?, ?, ?, ?, ?)", rfid, nis, name, phone, classID)
			if err == nil {
				successCount++
			}
		}
		tx.Commit()

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import berhasil (%d data)", successCount)})
	}
}

func ImportStaff(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
		}
		defer src.Close()

		reader := csv.NewReader(src)
		records, err := reader.ReadAll()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Format CSV salah"})
		}

		tx, _ := db.Begin()
		successCount := 0

		for i, row := range records {
			if i == 0 {
				continue
			}
			if len(row) < 5 {
				continue
			}

			nip := strings.TrimSpace(row[0])
			name := strings.TrimSpace(row[1])
			role := strings.TrimSpace(row[2])
			phone := utils.FormatPhone(row[3])
			rfid := strings.TrimSpace(row[4])

			_, err := tx.Exec("INSERT OR IGNORE INTO staff (rfid_uid, nip, name, phone, role) VALUES (?, ?, ?, ?, ?)", rfid, nip, name, phone, role)
			if err == nil {
				successCount++
			}
		}
		tx.Commit()

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import berhasil (%d data)", successCount)})
	}
}

func ImportStudentsJSON(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type JSONExport []struct {
			Type string `json:"type"`
			Name string `json:"name"`
			Data []struct {
				ID        string `json:"id"`
				NISN      string `json:"nisn"`
				Nama      string `json:"nama"`
				Kelas     string `json:"kelas"`
				NamaWali  string `json:"nama_wali"`
				NomorWali string `json:"nomor_wali"`
				Foto      string `json:"foto"`
			} `json:"data,omitempty"`
		}

		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "File tidak ditemukan"})
		}
		clean := c.FormValue("clean") == "true"

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuka file"})
		}
		defer src.Close()

		byteValue := make([]byte, file.Size)
		src.Read(byteValue)

		var export JSONExport
		err = json.Unmarshal(byteValue, &export)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format JSON salah"})
		}

		tx, _ := db.Begin()
		successCount := 0

		if clean {
			tx.Exec("DELETE FROM students")
		}

		classMap := make(map[string]int)
		rows, _ := db.Query("SELECT id, name FROM classes")
		for rows.Next() {
			var id int
			var name string
			rows.Scan(&id, &name)
			classMap[strings.ToUpper(name)] = id
		}
		rows.Close()

		for _, item := range export {
			if item.Type == "table" && item.Name == "siswa" {
				for _, s := range item.Data {
					nisn := strings.TrimSpace(s.NISN)
					nama := strings.TrimSpace(s.Nama)
					kelas := strings.TrimSpace(s.Kelas)
					namaWali := strings.TrimSpace(s.NamaWali)
					nomorWali := strings.TrimSpace(s.NomorWali)
					foto := strings.TrimSpace(s.Foto)

					if nisn == "" || nama == "" {
						continue
					}

					classID, ok := classMap[strings.ToUpper(kelas)]
					if !ok {
						res, err := tx.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES (?, 0, '')", kelas)
						if err == nil {
							id, _ := res.LastInsertId()
							classID = int(id)
							classMap[strings.ToUpper(kelas)] = classID
						}
					}

					query := `INSERT INTO students (rfid_uid, nis, name, parent_name, parent_phone, class_id, photo)
						VALUES (?, ?, ?, ?, ?, ?, ?)
						ON CONFLICT(rfid_uid) DO UPDATE SET
						nis=excluded.nis, name=excluded.name, parent_name=excluded.parent_name,
						parent_phone=excluded.parent_phone, class_id=excluded.class_id, photo=excluded.photo`

					_, err := tx.Exec(query, nisn, nisn, nama, namaWali, nomorWali, classID, foto)
					if err == nil {
						successCount++
					}
				}
			}
		}

		tx.Commit()
		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("Import JSON berhasil (%d data)", successCount)})
	}
}
