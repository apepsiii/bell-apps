package handler

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

const (
	maxBackupSize       = 500 * 1024 * 1024 // 500 MB
	maxUnzippedSize     = 2 * 1024 * 1024 * 1024 // 2 GB
	maxUnzippedFileSize = 500 * 1024 * 1024 // 500 MB per file
)

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

		if file.Size > maxBackupSize {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "Ukuran file backup maksimal 500 MB",
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

		// PHASE 1: Validation Pass (Dry-Run Check)
		// Validate all entries BEFORE touching, renaming, or writing any files to disk.
		var totalUnzipped int64
		hasDatabaseFile := false
		type plannedFile struct {
			zipFile       *zip.File
			destPath      string
			baseTargetDir string
		}
		var plannedFiles []plannedFile

		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}

			if f.UncompressedSize64 > maxUnzippedFileSize {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"status":  "error",
					"message": "File di dalam zip terlalu besar",
				})
			}
			totalUnzipped += int64(f.UncompressedSize64)
			if totalUnzipped > maxUnzippedSize {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"status":  "error",
					"message": "Total ukuran file backup terlalu besar (zip bomb protection)",
				})
			}

			var destPath string
			var baseTargetDir string
			if f.Name == "database.db" {
				destPath = filepath.Clean(dbPath)
				baseTargetDir = filepath.Dir(destPath)
				hasDatabaseFile = true
			} else if len(f.Name) > 8 && f.Name[:8] == "uploads/" {
				rel := f.Name[8:]
				if len(rel) > 7 && rel[:7] == "photos/" {
					baseTargetDir = filepath.Clean(config.GetPhotoPath())
					destPath = filepath.Clean(filepath.Join(baseTargetDir, rel[7:]))
				} else if len(rel) > 6 && rel[:6] == "audio/" {
					baseTargetDir = filepath.Clean(config.GetAudioPath())
					destPath = filepath.Clean(filepath.Join(baseTargetDir, rel[6:]))
				} else {
					continue
				}

				// Zip Slip Protection: Ensure destPath is strictly inside baseTargetDir
				relToTarget, err := filepath.Rel(baseTargetDir, destPath)
				if err != nil || strings.HasPrefix(relToTarget, "..") || strings.HasPrefix(relToTarget, string(filepath.Separator)) {
					return c.JSON(http.StatusBadRequest, map[string]string{
						"status":  "error",
						"message": "File backup tidak valid: terdeteksi path traversal ilegal",
					})
				}
			} else {
				continue
			}

			plannedFiles = append(plannedFiles, plannedFile{
				zipFile:       f,
				destPath:      destPath,
				baseTargetDir: baseTargetDir,
			})
		}

		if !hasDatabaseFile {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "File backup tidak valid: database.db tidak ditemukan di dalam zip",
			})
		}

		// PHASE 2: Extraction Pass
		// Backup current database before replacing
		backupPath := dbPath + ".bak"
		if _, err := os.Stat(dbPath); err == nil {
			os.Remove(backupPath) // Ensure destination is clear on Windows
			os.Rename(dbPath, backupPath)
		}

		rollback := func() {
			if _, err := os.Stat(backupPath); err == nil {
				os.Remove(dbPath)
				os.Rename(backupPath, dbPath)
			}
		}

		restoredFiles := 0
		for _, pf := range plannedFiles {
			if err := os.MkdirAll(filepath.Dir(pf.destPath), 0755); err != nil {
				rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status":  "error",
					"message": "Gagal membuat direktori tujuan restore",
				})
			}

			rc, err := pf.zipFile.Open()
			if err != nil {
				rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status":  "error",
					"message": "Gagal membaca isi zip: " + pf.zipFile.Name,
				})
			}

			out, err := os.Create(pf.destPath)
			if err != nil {
				rc.Close()
				rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status":  "error",
					"message": "Gagal menulis file tujuan restore: " + pf.destPath,
				})
			}

			_, copyErr := io.Copy(out, rc)
			out.Close()
			rc.Close()
			if copyErr != nil {
				rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status":  "error",
					"message": "Gagal mengekstrak file restore: " + pf.destPath,
				})
			}

			if pf.zipFile.Name != "database.db" {
				restoredFiles++
			}
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
