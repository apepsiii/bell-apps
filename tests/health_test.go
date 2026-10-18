package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"belsekolah/internal/router"
)

func TestHealthCheckEndpoint(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	e := echo.New()
	viewsFS := os.DirFS(".")
	router.RegisterRoutes(e, app.DB, viewsFS, app)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res["status"])
	assert.NotEmpty(t, res["app"])
	assert.NotEmpty(t, res["version"])
	assert.NotEmpty(t, res["git_commit"])
}

