package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// maskAPIKey masks sensitive API keys, revealing only prefix and suffix
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// isMaskedAPIKey checks if the string appears to be a masked API key
func isMaskedAPIKey(key string) bool {
	return strings.Contains(key, "****") || key == "********"
}

// validateAIBaseURL validates AI Base URL and protects against SSRF
func validateAIBaseURL(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return errors.New("URL AI tidak boleh kosong")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("Format URL tidak valid: %v", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("Hanya protokol http dan https yang diizinkan untuk AI Base URL")
	}

	host := u.Hostname()
	hostLower := strings.ToLower(host)
	if hostLower == "" {
		return errors.New("Host URL tidak boleh kosong")
	}

	// Block Cloud Metadata Endpoints (AWS, GCP, Azure, OpenStack)
	if hostLower == "169.254.169.254" || hostLower == "metadata.google.internal" || hostLower == "instance-data" {
		return errors.New("Akses ke endpoint metadata cloud dilarang (SSRF Protection)")
	}

	ip := net.ParseIP(hostLower)
	if ip != nil {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return errors.New("Akses ke alamat link-local dilarang (SSRF Protection)")
		}
	}

	return nil
}

// GetAISettings returns all AI configuration settings with masked API key
func GetAISettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT setting_key, setting_value FROM ai_settings")
		if err != nil {
			c.Logger().Errorf("Failed to query ai_settings: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memuat pengaturan AI"})
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
		} else {
			// SEC-12: Mask API Key when returning to frontend
			settings["ai_api_key"] = maskAPIKey(settings["ai_api_key"])
		}
		if _, ok := settings["ai_model"]; !ok {
			settings["ai_model"] = "gpt-4o-mini"
		}

		return c.JSON(http.StatusOK, settings)
	}
}

// UpdateAISettings updates AI configuration settings securely
func UpdateAISettings(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		baseURL := strings.TrimSpace(c.FormValue("ai_base_url"))
		apiKey := strings.TrimSpace(c.FormValue("ai_api_key"))
		model := strings.TrimSpace(c.FormValue("ai_model"))

		// SEC-10: Validate AI Base URL syntax & SSRF safety
		if baseURL != "" {
			if err := validateAIBaseURL(baseURL); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
			}
		}

		tx, err := db.Begin()
		if err != nil {
			c.Logger().Errorf("Failed to begin transaction: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal memulai transaksi database"})
		}
		defer tx.Rollback()

		now := time.Now().Format("2006-01-02 15:04:05")

		_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_base_url", baseURL, now)
		if err != nil {
			c.Logger().Errorf("Failed to update ai_base_url: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan URL AI"})
		}

		// SEC-12: Only update API key if provided and not masked
		if apiKey != "" && !isMaskedAPIKey(apiKey) {
			_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_api_key", apiKey, now)
			if err != nil {
				c.Logger().Errorf("Failed to update ai_api_key: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan API Key AI"})
			}
		}

		_, err = tx.Exec("INSERT OR REPLACE INTO ai_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)", "ai_model", model, now)
		if err != nil {
			c.Logger().Errorf("Failed to update ai_model: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan model AI"})
		}

		if err := tx.Commit(); err != nil {
			c.Logger().Errorf("Failed to commit transaction: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data ke database"})
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

		// SEC-10: SSRF & URL validation
		if err := validateAIBaseURL(baseURL); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
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
			c.Logger().Errorf("Failed to marshal JSON payload: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal menyiapkan payload request AI",
			})
		}

		// Create HTTP request
		req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			c.Logger().Errorf("Failed to create HTTP request: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Format URL atau request tidak valid",
			})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		// SEC-10: Restrict client timeout to 10 seconds
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			c.Logger().Errorf("Failed to connect to AI API: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal menghubungi endpoint AI: koneksi gagal atau timeout",
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.Logger().Errorf("Failed to read AI response body: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal membaca respons dari AI provider",
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
