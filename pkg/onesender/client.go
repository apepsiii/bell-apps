package onesender

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	APIURL   string
	Token    string
	Username string
	DeviceID string
	Timeout  time.Duration
}

type GowaTextPayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type GowaImagePayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
	Image   string `json:"image"`
}

type GowaResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type MessagePayload struct {
	To            string        `json:"to"`
	RecipientType string        `json:"recipient_type"`
	Type          string        `json:"type"`
	Image         *ImagePayload `json:"image,omitempty"`
	Text          *TextPayload  `json:"text,omitempty"`
}

type ImagePayload struct {
	Link    string `json:"link"`
	Caption string `json:"caption"`
}

type TextPayload struct {
	Body string `json:"body"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewClient(apiURL, token string, options ...string) *Client {
	client := &Client{
		APIURL:  apiURL,
		Token:   token,
		Timeout: 10 * time.Second,
	}
	if len(options) > 0 {
		client.Username = options[0]
	}
	if len(options) > 1 {
		client.DeviceID = options[1]
	}
	return client
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}
	if !strings.HasPrefix(phone, "62") {
		phone = "62" + phone
	}
	return phone
}

func (c *Client) SendImageMessage(to, recipientType, imageURL, caption string) (string, error) {
	if to == "" || c.Token == "" || c.APIURL == "" {
		return "", nil
	}

	to = normalizePhone(to)

	if c.Username != "" && c.DeviceID != "" {
		return c.sendGowaText(to, caption)
	}

	payload := MessagePayload{
		To:            to,
		RecipientType: recipientType,
		Type:          "image",
		Image: &ImagePayload{
			Link:    imageURL,
			Caption: caption,
		},
	}

	return c.send(payload)
}

func (c *Client) SendTextMessage(to, recipientType, body string) (string, error) {
	if to == "" || c.Token == "" || c.APIURL == "" {
		return "", nil
	}

	to = normalizePhone(to)

	if c.Username != "" && c.DeviceID != "" {
		return c.sendGowaText(to, body)
	}

	payload := MessagePayload{
		To:            to,
		RecipientType: recipientType,
		Type:          "text",
		Text: &TextPayload{
			Body: body,
		},
	}

	return c.send(payload)
}

func (c *Client) send(payload MessagePayload) (string, error) {
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.APIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println("OneSender Error (Req):", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("OneSender Error (Do):", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func (c *Client) sendGowaText(phone, message string) (string, error) {
	payload := GowaTextPayload{
		Phone:   phone,
		Message: message,
	}

	jsonPayload, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/whatsapp/send", strings.TrimRight(c.APIURL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println("Gowa Error (Req):", err)
		return "", err
	}

	authStr := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Token))
	req.Header.Set("Authorization", "Basic "+authStr)
	req.Header.Set("X-Device-Id", c.DeviceID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Gowa Error (Do):", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var gowaResp GowaResponse
	if err := json.Unmarshal(body, &gowaResp); err == nil {
		result, _ := json.Marshal(gowaResp)
		return string(result), nil
	}

	return string(body), nil
}
