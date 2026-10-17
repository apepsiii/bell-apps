package handler

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	doaCacheMutex sync.RWMutex
	doaCacheData  []byte
	doaCacheTime  time.Time
	doaTTL        = 24 * time.Hour
)

// GetDailyPrayers fetches daily prayers from equran.id API with in-memory caching.
func GetDailyPrayers() echo.HandlerFunc {
	return func(c echo.Context) error {
		doaCacheMutex.RLock()
		if len(doaCacheData) > 0 && time.Since(doaCacheTime) < doaTTL {
			data := doaCacheData
			doaCacheMutex.RUnlock()
			return c.Blob(http.StatusOK, "application/json; charset=utf-8", data)
		}
		doaCacheMutex.RUnlock()

		// Fetch from upstream
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://equran.id/api/doa")
		if err != nil {
			// If error but we have stale cache, serve it
			doaCacheMutex.RLock()
			if len(doaCacheData) > 0 {
				data := doaCacheData
				doaCacheMutex.RUnlock()
				return c.Blob(http.StatusOK, "application/json; charset=utf-8", data)
			}
			doaCacheMutex.RUnlock()
			return c.JSON(http.StatusBadGateway, map[string]string{
				"message": "Gagal mengambil data doa dari server pusat: " + err.Error(),
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Gagal membaca data respons",
			})
		}

		if resp.StatusCode == http.StatusOK && len(body) > 0 {
			doaCacheMutex.Lock()
			doaCacheData = body
			doaCacheTime = time.Now()
			doaCacheMutex.Unlock()
		}

		return c.Blob(resp.StatusCode, "application/json; charset=utf-8", body)
	}
}

