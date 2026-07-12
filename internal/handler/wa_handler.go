package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"belsekolah/internal/service"
)

func TestWA(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		target := c.FormValue("target_phone")
		if target == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Nomor tujuan kosong"})
		}

		wa := service.NewWhatsAppService(db)
		resp, err := wa.Test(target)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error HTTP: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Request Terkirim", "api_response": resp})
	}
}
