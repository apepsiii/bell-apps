package onesender

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	APIURL  string
	Token   string
	Timeout time.Duration
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

func NewClient(apiURL, token string) *Client {
	return &Client{
		APIURL:  apiURL,
		Token:   token,
		Timeout: 10 * time.Second,
	}
}

func (c *Client) SendImageMessage(to, recipientType, imageURL, caption string) (string, error) {
	if to == "" || c.Token == "" || c.APIURL == "" {
		return "", nil
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
