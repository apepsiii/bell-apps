package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WhatsAppLogEntry struct {
	ID        int    `json:"id"`
	Target    string `json:"target"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	Response  string `json:"response"`
	Timestamp string `json:"timestamp"`
}

func GetWhatsAppLogs(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		limit := c.QueryParam("limit")
		if limit == "" {
			limit = "100"
		}

		rows, err := db.Query("SELECT id, target, message, status, response, timestamp FROM whatsapp_logs ORDER BY id DESC LIMIT ?", limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}
		defer rows.Close()

		var logs []WhatsAppLogEntry
		for rows.Next() {
			var l WhatsAppLogEntry
			var resp sql.NullString
			rows.Scan(&l.ID, &l.Target, &l.Message, &l.Status, &resp, &l.Timestamp)
			l.Response = resp.String
			logs = append(logs, l)
		}

		return c.JSON(http.StatusOK, logs)
	}
}
