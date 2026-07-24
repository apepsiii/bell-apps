package qrcode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

type QRData struct {
	Type   string `json:"type"`
	UID    string `json:"uid,omitempty"`
	ID     int    `json:"id,omitempty"`
	NIS    string `json:"nis,omitempty"`
	Name   string `json:"name,omitempty"`
	Class  string `json:"class,omitempty"`
}

func GenerateFromRFID(rfid string) (string, error) {
	qrData := fmt.Sprintf("RFID:%s", rfid)
	png, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	base64Img := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64Img, nil
}

// GenerateStudentCard encodes the student's RFID as a plain number (no
// prefix, no URL) so the QR is small and fast to scan with the kiosk camera.
// The /scan kiosk (jsQR) reads any text payload, so a bare RFID works.
// If rfid is empty, fall back to the numeric student id so the card still
// resolves (the scan handler accepts student_id too).
func GenerateStudentCard(baseURL, rfid string, id int, nis, name, class string) (string, error) {
	qrData := rfid
	if qrData == "" {
		qrData = fmt.Sprintf("%d", id)
	}

	png, err := qrcode.Encode(qrData, qrcode.High, 512)
	if err != nil {
		return "", err
	}
	base64Img := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64Img, nil
}

func ParseStudent(qrData string) (*QRData, error) {
	var data QRData
	if err := json.Unmarshal([]byte(qrData), &data); err != nil {
		return nil, err
	}
	if data.Type != "student" {
		return nil, fmt.Errorf("invalid QR type")
	}
	return &data, nil
}

func ParseRFID(qrData string) (string, error) {
	const prefix = "RFID:"
	if len(qrData) < len(prefix) || qrData[:len(prefix)] != prefix {
		return "", fmt.Errorf("invalid QR format")
	}
	return qrData[len(prefix):], nil
}
