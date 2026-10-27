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
	"os"
	"sort"
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

// validateAIBaseURL validates AI Base URL and protects against SSRF.
//
// The check is deliberately strict because the outbound request carries the
// provider API key in an Authorization header: a bypass would leak that key to
// an attacker-controlled host. It therefore blocks every non-global address
// (loopback, private, link-local, CGNAT, reserved, multicast), resolves DNS
// and re-checks every resulting address, and rejects hosts that cannot be
// resolved.
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

	// Block Cloud Metadata Endpoints (AWS, GCP, Azure, OpenStack, Alibaba)
	if hostLower == "169.254.169.254" || hostLower == "metadata.google.internal" ||
		hostLower == "instance-data" || hostLower == "100.100.100.200" {
		return errors.New("Akses ke endpoint metadata cloud dilarang (SSRF Protection)")
	}

	// Reject obviously local hostnames before any DNS lookup.
	if hostLower == "localhost" || strings.HasSuffix(hostLower, ".localhost") ||
		hostLower == "ip6-localhost" || hostLower == "ip6-loopback" {
		return errors.New("Akses ke host lokal dilarang (SSRF Protection)")
	}

	// If the host is a literal IP, validate it directly. Otherwise resolve it
	// and validate every address the name maps to, so a public-looking hostname
	// that points at an internal address is still rejected.
	ips := []net.IP{}
	if ip := net.ParseIP(hostLower); ip != nil {
		ips = append(ips, ip)
	} else {
		resolved, err := net.LookupIP(hostLower)
		if err != nil || len(resolved) == 0 {
			return errors.New("Host URL tidak dapat diselesaikan (SSRF Protection)")
		}
		ips = resolved
	}

	for _, ip := range ips {
		if allowPrivateAIURLs() {
			break
		}
		if err := checkGlobalIP(ip); err != nil {
			return err
		}
	}

	return nil
}

// allowPrivateAIURLs reports whether private/loopback AI base URLs are
// permitted. This is an explicit opt-in for local development and automated
// tests that spin up a mock provider on 127.0.0.1; it is off by default so
// production deployments keep full SSRF protection.
func allowPrivateAIURLs() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_ALLOW_PRIVATE_URLS"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// checkGlobalIP rejects any address that is not globally routable.
func checkGlobalIP(ip net.IP) error {
	if ip == nil {
		return errors.New("Alamat IP tidak valid (SSRF Protection)")
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return errors.New("Akses ke alamat internal dilarang (SSRF Protection)")
	}
	// CGNAT 100.64.0.0/10 and other reserved ranges.
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return errors.New("Akses ke rentang alamat CGNAT dilarang (SSRF Protection)")
		}
		if ip4[0] == 0 || ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0 ||
			ip4[0] == 198 && (ip4[1] == 18 || ip4[1] == 19) ||
			ip4[0] == 198 && ip4[1] == 51 && ip4[2] == 100 ||
			ip4[0] == 203 && ip4[1] == 0 && ip4[2] == 113 {
			return errors.New("Akses ke rentang alamat reserved dilarang (SSRF Protection)")
		}
	}
	return nil
}

// safeAIClient builds an HTTP client that never follows redirects. Following a
// redirect would let a public URL bounce the request (and its Authorization
// header) to an internal address after validation has already passed.
func safeAIClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
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
		var req struct {
			AIBaseURL string `json:"ai_base_url" form:"ai_base_url"`
			AIAPIKey  string `json:"ai_api_key" form:"ai_api_key"`
			AIModel   string `json:"ai_model" form:"ai_model"`
		}
		_ = c.Bind(&req)

		baseURL := strings.TrimSpace(req.AIBaseURL)
		if baseURL == "" {
			baseURL = strings.TrimSpace(c.FormValue("ai_base_url"))
		}
		apiKey := strings.TrimSpace(req.AIAPIKey)
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.FormValue("ai_api_key"))
		}
		model := strings.TrimSpace(req.AIModel)
		if model == "" {
			model = strings.TrimSpace(c.FormValue("ai_model"))
		}

		// SEC-10: Validate AI Base URL syntax & SSRF safety
		if baseURL != "" {
			if err := validateAIBaseURL(baseURL); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Terjadi kesalahan server"})
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

// resolveChatURL ensures the AI chat completions endpoint is properly formatted
func resolveChatURL(rawBaseURL string) string {
	clean := strings.TrimRight(strings.TrimSpace(rawBaseURL), "/")
	if strings.HasSuffix(clean, "/chat/completions") {
		return clean
	}
	return clean + "/chat/completions"
}

// resolveModelsURL formats the endpoint for retrieving available AI models
func resolveModelsURL(rawBaseURL string) string {
	clean := strings.TrimRight(strings.TrimSpace(rawBaseURL), "/")
	clean = strings.TrimSuffix(clean, "/chat/completions")
	if strings.HasSuffix(clean, "/models") {
		return clean
	}
	return clean + "/models"
}

// GetAIModels fetches the list of available models from the AI provider
func GetAIModels(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var reqData struct {
			AIBaseURL string `json:"ai_base_url" form:"ai_base_url"`
			AIAPIKey  string `json:"ai_api_key" form:"ai_api_key"`
		}
		_ = c.Bind(&reqData)

		var baseURL, apiKey string
		if strings.TrimSpace(reqData.AIBaseURL) != "" {
			baseURL = strings.TrimSpace(reqData.AIBaseURL)
		} else if strings.TrimSpace(c.QueryParam("base_url")) != "" {
			baseURL = strings.TrimSpace(c.QueryParam("base_url"))
		} else {
			db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_base_url").Scan(&baseURL)
		}

		if strings.TrimSpace(reqData.AIAPIKey) != "" && !isMaskedAPIKey(reqData.AIAPIKey) {
			apiKey = strings.TrimSpace(reqData.AIAPIKey)
		} else if strings.TrimSpace(c.QueryParam("api_key")) != "" && !isMaskedAPIKey(c.QueryParam("api_key")) {
			apiKey = strings.TrimSpace(c.QueryParam("api_key"))
		} else {
			db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_api_key").Scan(&apiKey)
		}

		if baseURL == "" || apiKey == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Base URL dan API Key belum diisi. Isi terlebih dahulu untuk mengambil daftar model.",
			})
		}

		// SEC-10: SSRF & URL validation
		if err := validateAIBaseURL(baseURL); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
		}

		modelsURL := resolveModelsURL(baseURL)

		req, err := http.NewRequest("GET", modelsURL, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Format URL tidak valid",
			})
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := safeAIClient()
		resp, err := client.Do(req)
		if err != nil {
			c.Logger().Errorf("Failed to connect to AI models endpoint: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal terhubung ke provider AI",
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Gagal membaca respons dari provider AI",
			})
		}

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":      "error",
				"message":     fmt.Sprintf("Provider AI mengembalikan status HTTP %d", resp.StatusCode),
				"response":    string(body),
				"status_code": resp.StatusCode,
			})
		}

		// Try parsing standard OpenAI format: {"data": [{"id": "gpt-4o"}, ...]}
		// or Ollama format: {"models": [{"name": "llama3:latest"}, ...]}
		var openAIResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Models []struct {
				Name  string `json:"name"`
				Model string `json:"model"`
			} `json:"models"`
		}

		var modelNames []string
		if err := json.Unmarshal(body, &openAIResp); err == nil {
			for _, m := range openAIResp.Data {
				if m.ID != "" {
					modelNames = append(modelNames, m.ID)
				}
			}
			if len(modelNames) == 0 {
				for _, m := range openAIResp.Models {
					if m.Name != "" {
						modelNames = append(modelNames, m.Name)
					} else if m.Model != "" {
						modelNames = append(modelNames, m.Model)
					}
				}
			}
		}

		// Deduplicate and sort
		uniqueMap := make(map[string]bool)
		var cleanModels []string
		for _, name := range modelNames {
			if !uniqueMap[name] {
				uniqueMap[name] = true
				cleanModels = append(cleanModels, name)
			}
		}
		sort.Strings(cleanModels)

		if len(cleanModels) == 0 {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "error",
				"message": "Tidak ada model yang ditemukan dalam respons API",
				"raw":     string(body),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"ok":      true,
			"models":  cleanModels,
			"count":   len(cleanModels),
			"message": fmt.Sprintf("Berhasil mengambil %d model dari API", len(cleanModels)),
		})
	}
}

// TestAIConnection tests the AI API connection
func TestAIConnection(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var reqData struct {
			AIBaseURL string `json:"ai_base_url" form:"ai_base_url"`
			AIAPIKey  string `json:"ai_api_key" form:"ai_api_key"`
			AIModel   string `json:"ai_model" form:"ai_model"`
		}
		_ = c.Bind(&reqData)

		var baseURL, apiKey, model string
		if strings.TrimSpace(reqData.AIBaseURL) != "" {
			baseURL = strings.TrimSpace(reqData.AIBaseURL)
		} else {
			db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_base_url").Scan(&baseURL)
		}

		if strings.TrimSpace(reqData.AIAPIKey) != "" && !isMaskedAPIKey(reqData.AIAPIKey) {
			apiKey = strings.TrimSpace(reqData.AIAPIKey)
		} else {
			db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_api_key").Scan(&apiKey)
		}

		if strings.TrimSpace(reqData.AIModel) != "" {
			model = strings.TrimSpace(reqData.AIModel)
		} else {
			db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_model").Scan(&model)
		}

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

		// Target chat completions endpoint safely
		chatURL := resolveChatURL(baseURL)

		// Create HTTP request
		req, err := http.NewRequest("POST", chatURL, bytes.NewBuffer(jsonData))
		if err != nil {
			c.Logger().Errorf("Failed to create HTTP request: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Format URL atau request tidak valid",
			})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		// SEC-10: Restrict client timeout to 15 seconds and never follow
		// redirects, so a validated public URL cannot bounce the request
		// (and its Authorization header) to an internal address.
		client := safeAIClient()
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
				"chat_url":     chatURL,
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
			"status":      "success",
			"ok":          true,
			"message":     "Koneksi AI berhasil! API merespons dengan baik.",
			"base_url":    baseURL,
			"chat_url":    chatURL,
			"model":       model,
			"status_code": resp.StatusCode,
		})
	}
}
