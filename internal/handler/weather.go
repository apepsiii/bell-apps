package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/config"
)

const openWeatherBaseURL = "https://api.openweathermap.org/data/2.5"

func WeatherCurrent() echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := config.GetWeatherConfig()

		if cfg.APIKey == "" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": "Weather API key not configured",
			})
		}

		city := cfg.City
		country := cfg.Country
		if city == "" {
			city = "Bogor"
		}
		if country == "" {
			country = "ID"
		}

		url := fmt.Sprintf("%s/weather?q=%s,%s&appid=%s&units=metric&lang=id",
			openWeatherBaseURL, city, country, cfg.APIKey)

		resp, err := http.Get(url)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch weather"})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse response"})
		}

		return c.JSON(resp.StatusCode, result)
	}
}

func WeatherForecast() echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := config.GetWeatherConfig()

		if cfg.APIKey == "" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": "Weather API key not configured",
			})
		}

		city := cfg.City
		country := cfg.Country
		if city == "" {
			city = "Bogor"
		}
		if country == "" {
			country = "ID"
		}

		url := fmt.Sprintf("%s/forecast?q=%s,%s&appid=%s&units=metric&lang=id&cnt=40",
			openWeatherBaseURL, city, country, cfg.APIKey)

		resp, err := http.Get(url)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch forecast"})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse response"})
		}

		return c.JSON(resp.StatusCode, result)
	}
}
