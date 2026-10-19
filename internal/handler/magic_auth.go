package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/service"
	"belsekolah/pkg/utils"
)

const (
	// MagicLinkTTL is the lifetime for magic login tokens.
	MagicLinkTTL = 10 * time.Minute
	// ParentSessionCookie is the cookie name for parent sessions.
	ParentSessionCookie = "parent_session"
)

// generateMagicCode generates an unambiguous 6-character uppercase alphanumeric code.
func generateMagicCode() string {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based code
		return fmt.Sprintf("%06X", time.Now().UnixNano()%0xFFFFFF)
	}
	res := make([]byte, 6)
	for i := 0; i < 6; i++ {
		res[i] = charset[int(b[i])%len(charset)]
	}
	return string(res)
}

// generateSecureToken creates a 32-byte hex string for cryptographic URL tokens.
func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return GenerateSessionToken()
	}
	return hex.EncodeToString(b)
}

// RequestMagicLogin initiates a reverse OTP login request.
// It generates a request code and token, saves it, and returns the WhatsApp click-to-chat URL.
func RequestMagicLogin(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Opportunistically clean up expired tokens
		_, _ = db.Exec("DELETE FROM magic_login_tokens WHERE expires_at < ?", time.Now())

		code := generateMagicCode()
		token := generateSecureToken()
		expiresAt := time.Now().Add(MagicLinkTTL)

		// Insert new pending token
		_, err := db.Exec(`
			INSERT INTO magic_login_tokens (token, request_code, status, expires_at)
			VALUES (?, ?, 'pending', ?)
		`, token, code, expiresAt)
		if err != nil {
			log.Printf("[MagicAuth] Failed to create token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Gagal membuat sesi login WhatsApp",
			})
		}

		// Retrieve official school WA bot number
		var schoolWAPhone string
		_ = db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'school_wa_phone'").Scan(&schoolWAPhone)
		schoolWAPhone = utils.FormatPhone(schoolWAPhone)

		var waURL string
		messageText := fmt.Sprintf("LOGIN %s", code)
		if schoolWAPhone != "" {
			waURL = fmt.Sprintf("https://wa.me/%s?text=%s", schoolWAPhone, strings.ReplaceAll(messageText, " ", "%20"))
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":          "success",
			"request_code":    code,
			"school_wa_phone": schoolWAPhone,
			"wa_url":          waURL,
			"message_text":    messageText,
			"expires_in":      int(MagicLinkTTL.Seconds()),
		})
	}
}

// CheckMagicStatus allows frontend polling to detect when WhatsApp message is received.
func CheckMagicStatus(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		code := strings.TrimSpace(strings.ToUpper(c.QueryParam("code")))
		if code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Kode wajib diisi"})
		}

		var (
			token     string
			status    string
			userType  string
			expiresAt time.Time
		)
		err := db.QueryRow(`
			SELECT token, status, user_type, expires_at 
			FROM magic_login_tokens 
			WHERE request_code = ?
		`, code).Scan(&token, &status, &userType, &expiresAt)

		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"status": "not_found", "message": "Kode tidak ditemukan"})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Kesalahan server"})
		}

		if time.Now().After(expiresAt) {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":  "expired",
				"message": "Kode telah kadaluarsa",
			})
		}

		if status == "verified" {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":       "verified",
				"user_type":    userType,
				"redirect_url": "/auth/magic?token=" + token,
			})
		}

		if status == "used" {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":       "used",
				"user_type":    userType,
				"redirect_url": "/student/app",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "pending",
		})
	}
}

// extractWebhookPayload parses various incoming WhatsApp webhook formats (OneSender, GOWA, JSON, Form).
func extractWebhookPayload(c echo.Context) (sender string, body string) {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var payload map[string]interface{}
		rawBody, err := io.ReadAll(c.Request().Body)
		if err == nil && len(rawBody) > 0 {
			_ = json.Unmarshal(rawBody, &payload)
		}

		if payload != nil {
			// 1. OneSender nested format: data.from, data.body / data.message
			if data, ok := payload["data"].(map[string]interface{}); ok {
				if f, ok := data["from"].(string); ok && f != "" {
					sender = f
				} else if p, ok := data["phone"].(string); ok && p != "" {
					sender = p
				} else if s, ok := data["sender"].(string); ok && s != "" {
					sender = s
				}

				if b, ok := data["body"].(string); ok && b != "" {
					body = b
				} else if m, ok := data["message"].(string); ok && m != "" {
					body = m
				} else if t, ok := data["text"].(string); ok && t != "" {
					body = t
				}
			}

			// 2. GOWA payload format: payload.from, payload.body
			if sender == "" {
				if gPayload, ok := payload["payload"].(map[string]interface{}); ok {
					if f, ok := gPayload["from"].(string); ok {
						sender = f
					}
					if b, ok := gPayload["body"].(string); ok {
						body = b
					}
				}
			}

			// 3. Flat JSON: from, body / message / text
			if sender == "" {
				if f, ok := payload["from"].(string); ok && f != "" {
					sender = f
				} else if p, ok := payload["phone"].(string); ok && p != "" {
					sender = p
				} else if s, ok := payload["sender"].(string); ok && s != "" {
					sender = s
				}
			}
			if body == "" {
				if b, ok := payload["body"].(string); ok && b != "" {
					body = b
				} else if m, ok := payload["message"].(string); ok && m != "" {
					body = m
				} else if t, ok := payload["text"].(string); ok && t != "" {
					body = t
				}
			}
		}
	}

	// 4. Fallback to Form values if JSON was empty or form-encoded
	if sender == "" {
		sender = c.FormValue("from")
		if sender == "" {
			sender = c.FormValue("phone")
		}
		if sender == "" {
			sender = c.FormValue("sender")
		}
	}
	if body == "" {
		body = c.FormValue("message")
		if body == "" {
			body = c.FormValue("body")
		}
		if body == "" {
			body = c.FormValue("text")
		}
	}

	return sender, body
}

// cleanSenderPhone strips WhatsApp JID suffixes and normalizes phone to 628xxx format.
func cleanSenderPhone(raw string) string {
	if strings.Contains(raw, "@") {
		parts := strings.Split(raw, "@")
		raw = parts[0]
	}
	return utils.FormatPhone(raw)
}

// matchLoginCode extracts a 4-10 character code from incoming message.
var loginCodeRegex = regexp.MustCompile(`(?i)(?:LOGIN|MASUK)\s*[:#\- ]*\s*([A-Za-z0-9]{4,10})`)

// HandleWhatsAppWebhook handles incoming messages from OneSender / GOWA.
// When a valid LOGIN code is received from an authorized phone number, it validates the request,
// updates the token status to verified, and replies via WhatsApp with the single-use Magic Link.
func HandleWhatsAppWebhook(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawSender, message := extractWebhookPayload(c)
		senderPhone := cleanSenderPhone(rawSender)

		if senderPhone == "" || strings.TrimSpace(message) == "" {
			return c.JSON(http.StatusOK, map[string]string{
				"status": "ignored",
				"reason": "empty_phone_or_message",
			})
		}

		// Extract code from message text
		var code string
		if matches := loginCodeRegex.FindStringSubmatch(message); len(matches) > 1 {
			code = strings.ToUpper(strings.TrimSpace(matches[1]))
		} else {
			// If message is just a standalone 6-char code
			trimmed := strings.ToUpper(strings.TrimSpace(message))
			if len(trimmed) == 6 {
				var count int
				_ = db.QueryRow("SELECT COUNT(*) FROM magic_login_tokens WHERE request_code = ? AND status = 'pending'", trimmed).Scan(&count)
				if count > 0 {
					code = trimmed
				}
			}
		}

		if code == "" {
			// Not a login message, return 200 OK so webhook provider doesn't retry
			return c.JSON(http.StatusOK, map[string]string{
				"status": "ignored",
				"reason": "no_login_code_matched",
			})
		}

		// Look up pending token
		var (
			tokenID   int
			token     string
			status    string
			expiresAt time.Time
		)
		err := db.QueryRow(`
			SELECT id, token, status, expires_at 
			FROM magic_login_tokens 
			WHERE request_code = ?
		`, code).Scan(&tokenID, &token, &status, &expiresAt)

		waService := service.NewWhatsAppService(db)

		if err == sql.ErrNoRows || time.Now().After(expiresAt) || status != "pending" {
			reply := "Maaf, kode masuk WhatsApp Anda tidak valid atau telah kadaluarsa.\n\nSilakan buka kembali halaman login untuk mendapatkan kode baru:\n👉 https://" + c.Request().Host + "/student/login"
			_ = waService.Send(senderPhone, reply, "individual")
			return c.JSON(http.StatusOK, map[string]string{
				"status": "failed",
				"reason": "invalid_or_expired_code",
			})
		}

		// Phone matching variations: 628..., 08..., +628...
		altPhone1 := "0" + strings.TrimPrefix(senderPhone, "62")
		altPhone2 := "+" + senderPhone

		var (
			userType    string
			userID      int
			displayName string
		)

		// 1. Check parent in students table
		var (
			studentID   int
			studentName string
			parentName  string
		)
		err = db.QueryRow(`
			SELECT id, name, COALESCE(parent_name, '') 
			FROM students 
			WHERE (parent_phone = ? OR parent_phone = ? OR parent_phone = ?)
			  AND COALESCE(status, 'active') != 'inactive'
			ORDER BY id ASC LIMIT 1
		`, senderPhone, altPhone1, altPhone2).Scan(&studentID, &studentName, &parentName)

		if err == nil {
			userType = "parent"
			userID = studentID
			if parentName != "" {
				displayName = parentName
			} else {
				displayName = "Orang Tua dari " + studentName
			}
		}

		// 2. If not parent, check student's own phone
		if userType == "" {
			err = db.QueryRow(`
				SELECT id, name 
				FROM students 
				WHERE (phone = ? OR phone = ? OR phone = ?)
				  AND COALESCE(status, 'active') != 'inactive'
				ORDER BY id ASC LIMIT 1
			`, senderPhone, altPhone1, altPhone2).Scan(&studentID, &studentName)
			if err == nil {
				userType = "student"
				userID = studentID
				displayName = studentName
			}
		}

		// 3. If not student, check staff
		if userType == "" {
			var staffID int
			var staffName string
			err = db.QueryRow(`
				SELECT id, name 
				FROM staff 
				WHERE phone = ? OR phone = ? OR phone = ?
				ORDER BY id ASC LIMIT 1
			`, senderPhone, altPhone1, altPhone2).Scan(&staffID, &staffName)
			if err == nil {
				userType = "staff"
				userID = staffID
				displayName = staffName
			}
		}

		// 4. If not staff, check operator / admin
		if userType == "" {
			var opID int
			var opName string
			err = db.QueryRow(`
				SELECT id, name 
				FROM operators 
				WHERE phone = ? OR phone = ? OR phone = ?
				ORDER BY id ASC LIMIT 1
			`, senderPhone, altPhone1, altPhone2).Scan(&opID, &opName)
			if err == nil {
				userType = "operator"
				userID = opID
				displayName = opName
			}
		}

		// Unregistered phone number
		if userType == "" {
			reply := fmt.Sprintf("Assalamu'alaikum / Halo! 👋\n\nNomor WhatsApp Anda (%s) belum terdaftar di sistem SMK NIBA Super Apps.\n\nJika Anda adalah orang tua siswa, guru, atau staf sekolah, silakan hubungi pihak sekolah/wali kelas untuk memperbarui nomor kontak Anda.\n\n— SMK NIBA Super Apps —", senderPhone)
			_ = waService.Send(senderPhone, reply, "individual")
			return c.JSON(http.StatusOK, map[string]string{
				"status": "unregistered_phone",
			})
		}

		// Determine base host and protocol
		scheme := "https"
		if c.Request().TLS == nil && c.Request().Header.Get("X-Forwarded-Proto") != "https" {
			// Check if localhost
			if strings.HasPrefix(c.Request().Host, "localhost") || strings.HasPrefix(c.Request().Host, "127.0.0.1") {
				scheme = "http"
			}
		}
		host := c.Request().Host

		magicURL := fmt.Sprintf("%s://%s/auth/magic?token=%s", scheme, host, token)

		// Mark token verified in database
		_, err = db.Exec(`
			UPDATE magic_login_tokens 
			SET phone = ?, user_type = ?, user_id = ?, status = 'verified' 
			WHERE id = ?
		`, senderPhone, userType, userID, tokenID)
		if err != nil {
			log.Printf("[MagicAuth] Error updating token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database error"})
		}

		// Compose friendly response message
		var replyMessage string
		switch userType {
		case "parent":
			replyMessage = fmt.Sprintf("Assalamu'alaikum / Halo Bapak/Ibu %s! 👋\n\nBerikut tautan masuk ke Portal Orang Tua SMK NIBA:\n\n👉 %s\n\nℹ️ *Catatan Keamanan:*\n• Tautan ini hanya berlaku selama 10 menit.\n• Jika Anda membuka halaman login di browser, layar tersebut akan otomatis masuk tanpa perlu diklik.\n\n— SMK NIBA Super Apps —", displayName, magicURL)
		case "student":
			replyMessage = fmt.Sprintf("Halo %s! 👋\n\nBerikut tautan masuk ke Portal Siswa SMK NIBA:\n\n👉 %s\n\nℹ️ *Catatan:* Tautan berlaku 10 menit dan sekali pakai. Jika tab login Anda masih terbuka, layar akan otomatis masuk.\n\n— SMK NIBA Super Apps —", displayName, magicURL)
		default:
			replyMessage = fmt.Sprintf("Halo Bapak/Ibu %s! 👋\n\nBerikut tautan masuk ke sistem SMK NIBA:\n\n👉 %s\n\nℹ️ Tautan berlaku 10 menit dan sekali pakai demi keamanan akun Anda.\n\n— SMK NIBA Super Apps —", displayName, magicURL)
		}

		_ = waService.Send(senderPhone, replyMessage, "individual")

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "success",
			"user_type": userType,
			"user_id":   userID,
			"phone":     senderPhone,
		})
	}
}

// ConsumeMagicLink processes single-use magic login link (GET /auth/magic?token=xyz).
// It issues the appropriate session cookie and redirects user directly to their portal.
func ConsumeMagicLink(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := strings.TrimSpace(c.QueryParam("token"))
		if token == "" {
			return c.Redirect(http.StatusSeeOther, "/student/login?magic_error=missing_token")
		}

		var (
			id        int
			phone     string
			userType  string
			userID    int
			status    string
			expiresAt time.Time
		)
		err := db.QueryRow(`
			SELECT id, phone, user_type, user_id, status, expires_at 
			FROM magic_login_tokens 
			WHERE token = ?
		`, token).Scan(&id, &phone, &userType, &userID, &status, &expiresAt)

		if err == sql.ErrNoRows {
			return c.Redirect(http.StatusSeeOther, "/student/login?magic_error=invalid_token")
		} else if err != nil {
			return c.Redirect(http.StatusSeeOther, "/student/login?magic_error=server_error")
		}

		if time.Now().After(expiresAt) {
			return c.Redirect(http.StatusSeeOther, "/student/login?magic_error=expired")
		}

		if status == "pending" {
			return c.Redirect(http.StatusSeeOther, "/student/login?magic_error=not_verified")
		}

		// Mark token as used to enforce single-use
		_, _ = db.Exec("UPDATE magic_login_tokens SET status = 'used' WHERE id = ?", id)

		sessionExpiry := time.Now().Add(7 * 24 * time.Hour)
		isSecure := c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"

		switch userType {
		case "parent":
			// 1. Create parent session
			parentSessionToken := GenerateSessionToken()
			var parentName string
			_ = db.QueryRow("SELECT COALESCE(parent_name, '') FROM students WHERE id = ?", userID).Scan(&parentName)

			_, _ = db.Exec(`
				REPLACE INTO parent_sessions (token, parent_phone, parent_name, student_id, expires_at)
				VALUES (?, ?, ?, ?, ?)
			`, parentSessionToken, phone, parentName, userID, sessionExpiry)

			c.SetCookie(&http.Cookie{
				Name:     ParentSessionCookie,
				Value:    parentSessionToken,
				Path:     "/",
				Expires:  sessionExpiry,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			// 2. Also create student session so all student APIs & dashboard work seamlessly!
			studentSessionToken := GenerateSessionToken()
			_, _ = db.Exec(`
				REPLACE INTO student_sessions (token, student_id, expires_at)
				VALUES (?, ?, ?)
			`, studentSessionToken, userID, sessionExpiry)

			// Transitional compatibility write
			_, _ = db.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
				studentSessionPrefix+studentSessionToken, userID)

			c.SetCookie(&http.Cookie{
				Name:     StudentSessionCookie,
				Value:    studentSessionToken,
				Path:     "/",
				Expires:  sessionExpiry,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			return c.Redirect(http.StatusSeeOther, "/student/app")

		case "student":
			sessionToken := GenerateSessionToken()
			_, _ = db.Exec(`
				REPLACE INTO student_sessions (token, student_id, expires_at)
				VALUES (?, ?, ?)
			`, sessionToken, userID, sessionExpiry)

			_, _ = db.Exec("REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
				studentSessionPrefix+sessionToken, userID)

			c.SetCookie(&http.Cookie{
				Name:     StudentSessionCookie,
				Value:    sessionToken,
				Path:     "/",
				Expires:  sessionExpiry,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			return c.Redirect(http.StatusSeeOther, "/student/app")

		case "operator":
			opSessionToken := GenerateSessionToken()
			opExpiry := time.Now().Add(24 * time.Hour)
			_, _ = db.Exec(`
				REPLACE INTO operator_sessions (token, operator_id, expires_at)
				VALUES (?, ?, ?)
			`, opSessionToken, userID, opExpiry)

			_, _ = db.Exec("INSERT OR REPLACE INTO attendance_settings (setting_key, setting_value) VALUES (?, ?)",
				"session_"+opSessionToken, userID)

			c.SetCookie(&http.Cookie{
				Name:     "operator_session",
				Value:    opSessionToken,
				Path:     "/",
				Expires:  opExpiry,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			return c.Redirect(http.StatusSeeOther, "/admin/dashboard")

		case "staff":
			// If staff, issue student or operator session or redirect to attendance
			opSessionToken := GenerateSessionToken()
			opExpiry := time.Now().Add(24 * time.Hour)
			_, _ = db.Exec(`
				REPLACE INTO operator_sessions (token, operator_id, expires_at)
				VALUES (?, ?, ?)
			`, opSessionToken, userID, opExpiry)

			c.SetCookie(&http.Cookie{
				Name:     "operator_session",
				Value:    opSessionToken,
				Path:     "/",
				Expires:  opExpiry,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			return c.Redirect(http.StatusSeeOther, "/admin/attendance")

		default:
			return c.Redirect(http.StatusSeeOther, "/student/app")
		}
	}
}

