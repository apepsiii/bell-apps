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

// GenerateAIQuest generates an English quest using AI
func GenerateAIQuest(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get AI settings from database
		var baseURL, apiKey, model string
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_base_url").Scan(&baseURL)
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_api_key").Scan(&apiKey)
		db.QueryRow("SELECT setting_value FROM ai_settings WHERE setting_key = ?", "ai_model").Scan(&model)

		if baseURL == "" || apiKey == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"status":  "error",
				"message": "Konfigurasi AI belum lengkap. Silakan atur Base URL dan API Key di menu Pengaturan AI.",
			})
		}

		if model == "" {
			model = "gpt-4o-mini"
		}

		// Get request parameters
		questType := c.FormValue("quest_type")
		topic := c.FormValue("topic")
		date := c.FormValue("date")

		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		if questType == "" {
			questType = "written"
		}

		// Build AI prompt based on quest type
		prompt := buildQuestPrompt(questType, topic)

		// Call AI API
		response, err := callAIAPI(baseURL, apiKey, model, prompt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Gagal memanggil AI: %v", err),
			})
		}

		// Parse AI response and create quest data
		questData, err := parseAIResponse(response, questType, topic, date)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Gagal memproses respons AI: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Quest berhasil di-generate oleh AI",
			"data":    questData,
		})
	}
}

func buildQuestPrompt(questType, topic string) string {
	baseTopic := topic
	if baseTopic == "" {
		baseTopic = "daily English practice"
	}

	switch questType {
	case "vocabulary":
		return fmt.Sprintf(`Generate 10 English vocabulary words related to "%s" for high school students (SMK level).
Return ONLY a valid JSON object in this exact format:
{
  "title": "Vocabulary: [topic name]",
  "description": "Learn 10 new vocabulary words about [topic]",
  "topic": "%s",
  "vocabulary_words": "word1,word2,word3,word4,word5,word6,word7,word8,word9,word10",
  "xp_reward": 15
}
Do not include any explanation, just the JSON.`, baseTopic, baseTopic)

	case "quiz":
		return fmt.Sprintf(`Generate an English multiple choice quiz question about "%s" for high school students.
Return ONLY a valid JSON object in this exact format:
{
  "title": "Quiz: [topic name]",
  "description": "Test your knowledge about [topic]",
  "topic": "%s",
  "quiz_question": "What is...",
  "quiz_choices": "A. Option1|B. Option2|C. Option3|D. Option4",
  "quiz_answer": "A",
  "xp_reward": 20
}
Do not include any explanation, just the JSON.`, baseTopic, baseTopic)

	case "voice":
		return fmt.Sprintf(`Generate a voice recording task about "%s" for high school students.
Return ONLY a valid JSON object in this exact format:
{
  "title": "Voice Practice: [topic name]",
  "description": "Record yourself speaking about [topic] for 30-60 seconds",
  "topic": "%s",
  "xp_reward": 25
}
Do not include any explanation, just the JSON.`, baseTopic, baseTopic)

	default: // written
		return fmt.Sprintf(`Generate a writing prompt about "%s" for high school students (SMK level).
The prompt should ask students to write 5-10 sentences in English.
Return ONLY a valid JSON object in this exact format:
{
  "title": "Writing: [topic name]",
  "description": "Write 5-10 sentences in English about [specific writing prompt]",
  "topic": "%s",
  "xp_reward": 20
}
Do not include any explanation, just the JSON.`, baseTopic, baseTopic)
	}
}

func callAIAPI(baseURL, apiKey, model, prompt string) (string, error) {
	// Prepare request body
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiResp.Choices[0].Message.Content, nil
}

func parseAIResponse(response, questType, topic, date string) (map[string]interface{}, error) {
	// Try to parse JSON from AI response
	var questData map[string]interface{}
	
	// Clean up response - sometimes AI wraps JSON in markdown code blocks
	cleanResponse := response
	if len(response) > 0 {
		// Remove markdown code blocks if present
		if bytes.Contains([]byte(response), []byte("```json")) {
			start := bytes.Index([]byte(response), []byte("```json"))
			end := bytes.LastIndex([]byte(response), []byte("```"))
			if start >= 0 && end > start {
				cleanResponse = response[start+7 : end]
			}
		} else if bytes.Contains([]byte(response), []byte("```")) {
			start := bytes.Index([]byte(response), []byte("```"))
			end := bytes.LastIndex([]byte(response), []byte("```"))
			if start >= 0 && end > start {
				cleanResponse = response[start+3 : end]
			}
		}
	}

	err := json.Unmarshal([]byte(cleanResponse), &questData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %v", err)
	}

	// Add quest_type and date
	questData["quest_type"] = questType
	questData["date"] = date

	// Ensure required fields exist with defaults
	if _, ok := questData["title"]; !ok {
		questData["title"] = "English Quest"
	}
	if _, ok := questData["description"]; !ok {
		questData["description"] = "Complete this English quest"
	}
	if _, ok := questData["topic"]; !ok {
		questData["topic"] = topic
	}
	if _, ok := questData["xp_reward"]; !ok {
		questData["xp_reward"] = 10
	}

	// Set empty strings for unused fields
	if questType != "vocabulary" {
		questData["vocabulary_words"] = ""
	}
	if questType != "quiz" {
		questData["quiz_question"] = ""
		questData["quiz_choices"] = ""
		questData["quiz_answer"] = ""
	}

	return questData, nil
}
