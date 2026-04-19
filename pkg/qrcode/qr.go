package qrcode

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

type QRData struct {
	Type string
	UID  string
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

func ParseRFID(qrData string) (string, error) {
	const prefix = "RFID:"
	if len(qrData) < len(prefix) || qrData[:len(prefix)] != prefix {
		return "", fmt.Errorf("invalid QR format")
	}
	return qrData[len(prefix):], nil
}
