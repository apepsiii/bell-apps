package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// GetAISettings returns all AI configuration settings
func GetAISettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT setting_key, setting_value FROM ai_settings")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		settings := make(map[string]string)
		for rows.Next() {
			var key, value string
			rows.Scan(&key, &value)
			settings[key] = value
		}

		// Set defaults if not found
		if _, ok := settings["ai_base_url"]; !ok {
			settings["ai_base_url"] = ""
		}
		if _, ok := settings["ai_api_key"]; !ok {
			settings["ai_api_key"] = ""
		}
		if _, ok := settings["ai_model"]; !ok {
			settings["ai_model"] = "gpt-4o-mini"
		}

		return c.JSON(http.StatusOK, settings)
	}
}

// UpdateAISettings updates AI configuration settings
func UpdateAISettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		baseURL := c.FormValue("ai_base_url")
		apiKey := c.FormValue("ai_api_key")
		model := c.FormValue("ai_model")

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer tx.Rollback()

		now := time.Now().Format("2006-01-02 15:04:05")

		// Use INSERT OR REPLACE for SQLite, INSERT ... ON DUPLICATE KEY UPDATE for MySQL
		_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_base_url", baseURL, now)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_api_key", apiKey, now)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_model", model, now)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Pengaturan AI berhasil disimpan"})
	}
}

// TestAIConnection tests the AI API connection
func TestAIConnection(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get AI settings from database
		var baseURL, apiKey, model string
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_base_url").Scan(&baseURL)
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_api_key").Scan(&apiKey)
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_model").Scan(&model)

		if baseURL == "" || apiKey == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Konfigurasi AI belum lengkap. Silakan atur Base URL dan API Key terlebih dahulu.",
			})
		}

		if model == "" {
			model = "gpt-4o-mini"
		}

		// Prepare test request
		reqBody := map[string]interface{}{
			"model": model,
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": "Say hello in one word",
				},
			},
			"max_tokens": 10,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal membuat request: %v", err),
			})
		}

		// Create HTTP request
		req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal membuat HTTP request: %v", err),
			})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		// Send request
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal menghubungi API: %v", err),
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Gagal membaca response: %v", err),
			})
		}

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":       "error",
				"message":      fmt.Sprintf("API mengembalikan status %d", resp.StatusCode),
				"response":     string(body),
				"base_url":     baseURL,
				"model":        model,
				"status_code":  resp.StatusCode,
			})
		}

		// Parse response
		var aiResp map[string]interface{}
		if err := json.Unmarshal(body, &aiResp); err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":   "error",
				"message":  "Response bukan JSON yang valid",
				"response": string(body),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":       "success",
			"message":      "Koneksi AI berhasil! API merespons dengan baik.",
			"base_url":     baseURL,
			"model":        model,
			"status_code":  resp.StatusCode,
		})
	}
}
