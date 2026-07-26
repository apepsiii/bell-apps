package handler

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
	"belsekolah/internal/repository"
)

// BackupDatabase creates a zip file containing the SQLite database and uploaded assets
func BackupDatabase(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if config.IsMySQL() {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "Backup otomatis hanya tersedia untuk SQLite. Untuk MySQL, gunakan mysqldump.",
			})
		}

		dbPath := config.GetDBConfig().Path
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"status":  "error",
				"message": "File database tidak ditemukan",
			})
		}

		// Checkpoint WAL to ensure all data is written
		db.Exec("PRAGMA wal_checkpoint(FULL)")

		// Create zip in memory
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		zipName := fmt.Sprintf("backup_smkniba_%s.zip", timestamp)

		c.Response().Header().Set("Content-Type", "application/zip")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipName))
		c.Response().WriteHeader(http.StatusOK)

		zw := zip.NewWriter(c.Response().Writer)
		defer zw.Close()

		// Add database file
		if err := addFileToZip(zw, dbPath, "database.db"); err != nil {
			return err
		}

		// Add uploads (photos, audio, signage) if they exist
		dirsToBackup := []struct {
			path string
			name string
		}{
			{config.GetPhotoPath(), "uploads/photos"},
			{config.GetAudioPath(), "uploads/audio"},
		}

		for _, dir := range dirsToBackup {
			if _, err := os.Stat(dir.path); os.IsNotExist(err) {
				continue
			}
			filepath.Walk(dir.path, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(dir.path, path)
				return addFileToZip(zw, path, filepath.Join(dir.name, rel))
			})
		}

		return nil
	}
}

// RestoreDatabase restores SQLite database from uploaded zip file
func RestoreDatabase(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if config.IsMySQL() {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "Restore otomatis hanya tersedia untuk SQLite.",
			})
		}

		file, err := c.FormFile("backup_file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "File backup tidak ditemukan",
			})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "error",
				"message": "Gagal membuka file backup",
			})
		}
		defer src.Close()

		// Save uploaded zip to temp
		tmpZip := filepath.Join(os.TempDir(), "restore_backup.zip")
		tmpFile, err := os.Create(tmpZip)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		io.Copy(tmpFile, src)
		tmpFile.Close()
		defer os.Remove(tmpZip)

		// Open zip
		zr, err := zip.OpenReader(tmpZip)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "File tidak valid, pastikan file adalah .zip backup dari aplikasi ini",
			})
		}
		defer zr.Close()

		dbPath := config.GetDBConfig().Path

		// Backup current database before restoring
		backupPath := dbPath + ".bak"
		if _, err := os.Stat(dbPath); err == nil {
			os.Rename(dbPath, backupPath)
		}

		restoredDB := false
		restoredFiles := 0

		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}

			var destPath string
			if f.Name == "database.db" {
				destPath = dbPath
				restoredDB = true
			} else if len(f.Name) > 8 && f.Name[:8] == "uploads/" {
				rel := f.Name[8:]
				if len(rel) > 7 && rel[:7] == "photos/" {
					destPath = filepath.Join(config.GetPhotoPath(), rel[7:])
				} else if len(rel) > 6 && rel[:6] == "audio/" {
					destPath = filepath.Join(config.GetAudioPath(), rel[6:])
				} else {
					continue
				}
				restoredFiles++
			} else {
				continue
			}

			os.MkdirAll(filepath.Dir(destPath), 0755)

			rc, err := f.Open()
			if err != nil {
				continue
			}
			out, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				continue
			}
			io.Copy(out, rc)
			out.Close()
			rc.Close()
		}

		if !restoredDB {
			// Rollback
			if _, err := os.Stat(backupPath); err == nil {
				os.Rename(backupPath, dbPath)
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "File backup tidak valid: database.db tidak ditemukan di dalam zip",
			})
		}

		// Remove old backup file
		os.Remove(backupPath)

		// Re-run migrations on restored DB
		newDB, err := sql.Open("sqlite", dbPath)
		if err == nil {
			repository.RunMigrations(newDB)
			newDB.Close()
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":         "success",
			"message":        fmt.Sprintf("Restore berhasil! Database dan %d file media telah dipulihkan. Restart aplikasi untuk menerapkan perubahan.", restoredFiles),
			"restored_files": restoredFiles,
		})
	}
}

func addFileToZip(zw *zip.Writer, filePath, zipPath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	return err
}
