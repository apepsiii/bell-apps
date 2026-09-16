package service

import (
	"strings"
	"time"

	"belsekolah/pkg/onesender"
	"database/sql"
)

type WhatsAppService struct {
	db *sql.DB
}

func NewWhatsAppService(db *sql.DB) *WhatsAppService {
	return &WhatsAppService{db: db}
}

func (s *WhatsAppService) GetSettings() map[string]string {
	rows, _ := s.db.Query("SELECT setting_key, setting_value FROM attendance_settings")
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	return settings
}

func (s *WhatsAppService) Send(to, message, recipientType string) error {
	settings := s.GetSettings()
	apiURL := settings["onesender_api_url"]
	token := settings["onesender_api_token"]
	username := settings["onesender_username"]
	deviceID := settings["onesender_device_id"]
	imageURL := settings["wa_image_link"]

	if to == "" || token == "" || apiURL == "" {
		return nil
	}

	client := onesender.NewClient(apiURL, token, username, deviceID)
	var err error
	var resp string
	if imageURL != "" {
		resp, err = client.SendImageMessage(to, recipientType, imageURL, message)
	} else {
		resp, err = client.SendTextMessage(to, recipientType, message)
	}

	if err != nil {
		s.logMessage(to, message, "failed", err.Error())
		return err
	}

	s.logMessage(to, message, "success", resp)
	return nil
}

func (s *WhatsAppService) logMessage(target, message, status, response string) {
	s.db.Exec("INSERT INTO whatsapp_logs (target, message, status, response, timestamp) VALUES (?, ?, ?, ?, ?)",
		target, message, status, response, time.Now().Format("2006-01-02 15:04:05"))
}

func (s *WhatsAppService) SendAttendanceNotification(phone, name, status, timeStr, recipientType string) error {
	settings := s.GetSettings()
	token := settings["onesender_api_token"]
	if token == "" {
		return nil
	}

	var template string
	switch status {
	case "Terlambat":
		template = settings["wa_template_late"]
	case "Pulang":
		template = settings["wa_template_out"]
	default:
		template = settings["wa_template_in"]
	}

	if template == "" {
		template = "Halo, {name} presensi {status} pada {time}."
	}

	msg := template
	msg = strings.ReplaceAll(msg, "{name}", name)
	msg = strings.ReplaceAll(msg, "{time}", timeStr)
	msg = strings.ReplaceAll(msg, "{status}", status)
	msg = strings.ReplaceAll(msg, "{date}", time.Now().Format("02-01-2006"))

	return s.Send(phone, msg, recipientType)
}

func (s *WhatsAppService) Test(target string) (string, error) {
	settings := s.GetSettings()
	apiURL := settings["onesender_api_url"]
	token := settings["onesender_api_token"]
	username := settings["onesender_username"]
	deviceID := settings["onesender_device_id"]

	if token == "" {
		return "", nil
	}

	client := onesender.NewClient(apiURL, token, username, deviceID)
	resp, err := client.SendTextMessage(target, "individual", "Test Koneksi SMK NIBA Super Apps: Berhasil terhubung!")
	if err != nil {
		s.logMessage(target, "Test Koneksi", "failed", err.Error())
		return "", err
	}
	s.logMessage(target, "Test Koneksi", "success", resp)
	return resp, nil
}
