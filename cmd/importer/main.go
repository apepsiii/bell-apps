package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const DBPath = "./database.db"

type JSONExport []struct {
	Type string        `json:"type"`
	Name string        `json:"name"`
	Data []StudentJSON `json:"data,omitempty"`
}

type StudentJSON struct {
	ID        string `json:"id"`
	NISN      string `json:"nisn"`
	Nama      string `json:"nama"`
	Kelas     string `json:"kelas"`
	NamaWali  string `json:"nama_wali"`
	NomorWali string `json:"nomor_wali"`
	Foto      string `json:"foto"`
}

func main() {
	filePath := flag.String("file", "", "Path to CSV or JSON file (required)")
	clean := flag.Bool("clean", false, "Clear existing student data before import")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: go run main.go -file <path> [-clean]")
		os.Exit(1)
	}

	// Open Database
	db, err := sql.Open("sqlite", DBPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Ensure parent_name column exists (Migration)
	var colCount int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('students') WHERE name='parent_name'").Scan(&colCount)
	if err != nil {
		log.Printf("Warning: Failed to check schema: %v", err)
	} else if colCount == 0 {
		fmt.Println("Migrating database: Adding 'parent_name' column...")
		_, err = db.Exec("ALTER TABLE students ADD COLUMN parent_name TEXT DEFAULT ''")
		if err != nil {
			log.Fatal("Failed to migrate database:", err)
		}
	}

	if *clean {
		fmt.Println("Clearing existing student data...")
		_, err = db.Exec("DELETE FROM students")
		if err != nil {
			log.Fatal("Failed to clear students table:", err)
		}
		// Optional: Clear or reset auto-increment? usually not needed for sqlite unless we truncate
		_, err = db.Exec("DELETE FROM sqlite_sequence WHERE name='students'")
		if err == nil {
			fmt.Println("Reset ID sequence.")
		}
	}

	ext := strings.ToLower(filepath.Ext(*filePath))
	var records [][]string

	if ext == ".json" {
		records = parseJSON(*filePath)
	} else {
		records = parseCSV(*filePath)
	}

	// Headers for CSV: id, nisn, nama, kelas, nama_wali, nomor_wali, foto
	// Indices: 0   1     2     3      4          5           6

	fmt.Printf("Found %d records to process\n", len(records))

	successCount := 0
	skipCount := 0

	// Cache classes to reduce DB queries
	classMap := make(map[string]int)

	// Transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	for i, record := range records {
		// CSV header skip is handled in parseCSV now
		
		if len(record) < 7 {
			fmt.Printf("Row %d: Insufficient columns, skipping.\n", i+1)
			skipCount++
			continue
		}

		// Parse Fields
		// id := record[0] // Ignore, generate new
		nisn := strings.TrimSpace(record[1]) // Maps to rfid_uid AND nis
		nama := strings.TrimSpace(record[2])
		kelas := strings.TrimSpace(record[3])
		namaWali := strings.TrimSpace(record[4])
		nomorWali := strings.TrimSpace(record[5])
		foto := strings.TrimSpace(record[6])

		// Validation for Scientific Notation (Excel Issue)
		if strings.Contains(strings.ToUpper(nomorWali), "E+") {
			fmt.Printf("WARNING Row %d: Phone number '%s' appears to be in scientific notation (Excel format). Data might be truncated!\n", i+1, nomorWali)
		}

		if nisn == "" || nama == "" {
			fmt.Printf("Row %d: Missing NISN or Name, skipping.\n", i+1)
			skipCount++
			continue
		}

		// 1. Get or Create Class
		classID, ok := classMap[kelas]
		if !ok {
			// Check DB
			err := tx.QueryRow("SELECT id FROM classes WHERE name = ?", kelas).Scan(&classID)
			if err == sql.ErrNoRows {
				// Create new class
				res, err := tx.Exec("INSERT INTO classes (name, major_id, wa_group_id) VALUES (?, 0, '')", kelas)
				if err != nil {
					fmt.Printf("Row %d: Failed to create class '%s': %v\n", i+1, kelas, err)
					skipCount++
					continue
				}
				id, _ := res.LastInsertId()
				classID = int(id)
				fmt.Printf("Created new class: %s (ID: %d)\n", kelas, classID)
			} else if err != nil {
				fmt.Printf("Row %d: Database error checking class: %v\n", i+1, err)
				skipCount++
				continue
			}
			classMap[kelas] = classID
		}

		// 2. Insert or Update Student
		// We use UPSERT on rfid_uid
		query := `
			INSERT INTO students (rfid_uid, nis, name, parent_name, parent_phone, class_id, photo)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(rfid_uid) DO UPDATE SET
				nis = excluded.nis,
				name = excluded.name,
				parent_name = excluded.parent_name,
				parent_phone = excluded.parent_phone,
				class_id = excluded.class_id,
				photo = excluded.photo
		`

		_, err = tx.Exec(query, nisn, nisn, nama, namaWali, nomorWali, classID, foto)
		if err != nil {
			fmt.Printf("Row %d: Failed to upsert student '%s': %v\n", i+1, nama, err)
			skipCount++
			continue
		}

		successCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Printf("\nImport Completed.\nSuccess: %d\nSkipped: %d\n", successCount, skipCount)
}

func parseCSV(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Failed to open CSV file:", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read CSV:", err)
	}

	// Remove header if present (heuristic)
	if len(records) > 0 && strings.ToLower(records[0][0]) == "id" {
		return records[1:]
	}
	return records
}

func parseJSON(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Failed to open JSON file:", err)
	}
	defer f.Close()

	byteValue, _ := io.ReadAll(f)

	var export JSONExport
	err = json.Unmarshal(byteValue, &export)
	if err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	var parsed [][]string
	for _, item := range export {
		if item.Type == "table" && item.Name == "siswa" {
			for _, s := range item.Data {
				// Convert to same format as CSV: id, nisn, nama, kelas, nama_wali, nomor_wali, foto
				row := []string{
					s.ID,
					s.NISN,
					s.Nama,
					s.Kelas,
					s.NamaWali,
					s.NomorWali,
					s.Foto,
				}
				parsed = append(parsed, row)
			}
		}
	}
	return parsed
}
