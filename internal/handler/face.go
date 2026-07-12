package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type FaceServiceConfig struct {
	BaseURL string
}

var faceService = FaceServiceConfig{
	BaseURL: os.Getenv("FACE_SERVICE_URL"),
}

func SetFaceServiceURL(url string) {
	faceService.BaseURL = url
}

func GetFaceServiceURL() string {
	if faceService.BaseURL == "" {
		return "http://localhost:8001"
	}
	return faceService.BaseURL
}

type FaceRegisterRequest struct {
	StudentID   int    `json:"student_id"`
	Name        string `json:"name"`
	ClassName   string `json:"class_name"`
	ImageBase64 string `json:"image_base64"`
}

type FaceVerifyRequest struct {
	ImageBase64 string `json:"image_base64"`
}

type FaceVerifyResponse struct {
	Matched   bool    `json:"matched"`
	StudentID int     `json:"student_id"`
	Name      string  `json:"name"`
	ClassName string  `json:"class_name"`
	Distance  float64 `json:"distance"`
	Message   string  `json:"message"`
}

type FaceRegisterResponse struct {
	Success   bool   `json:"success"`
	StudentID int    `json:"student_id"`
	Message   string `json:"message"`
}

func RegisterFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentIDStr := c.QueryParam("student_id")
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid student ID"})
		}

		name := c.QueryParam("name")
		className := c.QueryParam("class_name")

		file, err := c.FormFile("image")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "No image provided"})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to open image"})
		}
		defer src.Close()

		imageData, err := io.ReadAll(src)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to read image"})
		}

		faceURL := GetFaceServiceURL() + "/register"

		// Create JSON body with all data
		reqBody := map[string]interface{}{
			"student_id": studentID,
			"name":       name,
			"class_name": className,
			"image_base64": base64.StdEncoding.EncodeToString(imageData),
		}
		reqJSON, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", faceURL, bytes.NewBuffer(reqJSON))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create request"})
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Face service returned %d: %s", resp.StatusCode, string(body)),
			})
		}

		var result FaceRegisterResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid response from face service"})
		}

		if !result.Success {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": result.Message})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Face registered for %s", name),
		})
	}
}

func VerifyFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		type VerifyRequest struct {
			Image string `json:"image_base64"`
		}

		var req VerifyRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		}

		faceURL := GetFaceServiceURL() + "/verify"

		payload, _ := json.Marshal(map[string]string{"image_base64": req.Image})

		resp, err := http.Post(faceURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Face service returned %d: %s", resp.StatusCode, string(body)),
			})
		}

		var result FaceVerifyResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid response from face service"})
		}

		if !result.Matched {
			return c.JSON(http.StatusOK, map[string]string{
				"status":  "not_matched",
				"message": result.Message,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":     "matched",
			"student_id": result.StudentID,
			"name":       result.Name,
			"class_name": result.ClassName,
			"distance":   result.Distance,
		})
	}
}

func ListFaces(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		faceURL := GetFaceServiceURL() + "/faces"

		resp, err := http.Get(faceURL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Face service returned %d", resp.StatusCode),
			})
		}

		c.Response().Header().Set("Content-Type", "application/json")
		c.Response().Write(body)
		return nil
	}
}

func DeleteFace(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Param("student_id")

		faceURL := GetFaceServiceURL() + "/face/" + studentID

		req, _ := http.NewRequest("DELETE", faceURL, nil)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Face service returned %d: %s", resp.StatusCode, string(body)),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Face deleted"})
	}
}

type FaceStatusResponse struct {
	Registered bool   `json:"registered"`
	StudentID  int    `json:"student_id"`
	Name       string `json:"name,omitempty"`
	ClassName  string `json:"class_name,omitempty"`
}

func GetFaceStatus(db interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentIDStr := c.QueryParam("student_id")
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid student ID"})
		}

		faceURL := GetFaceServiceURL() + "/faces"

		resp, err := http.Get(faceURL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Face service error: " + err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Face service returned %d", resp.StatusCode),
			})
		}

		var result struct {
			Count int `json:"count"`
			Faces []struct {
				StudentID int    `json:"student_id"`
				Name      string `json:"name"`
				ClassName string `json:"class_name"`
			} `json:"faces"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid response from face service"})
		}

		for _, face := range result.Faces {
			if face.StudentID == studentID {
				return c.JSON(http.StatusOK, FaceStatusResponse{
					Registered: true,
					StudentID:  studentID,
					Name:       face.Name,
					ClassName:  face.ClassName,
				})
			}
		}

		return c.JSON(http.StatusOK, FaceStatusResponse{
			Registered: false,
			StudentID:  studentID,
		})
	}
}
