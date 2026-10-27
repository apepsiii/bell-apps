package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"belsekolah/internal/config"
)

// AdminChangePassword lets the authenticated admin rotate their password.
// The old password is verified against the configured credential (bcrypt hash
// or legacy plaintext), and the new password is stored as a bcrypt hash in
// config.yaml via config.UpdateAdminPassword.
func AdminChangePassword(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			OldPassword     string `json:"old_password" form:"old_password"`
			NewPassword     string `json:"new_password" form:"new_password"`
			ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
		}
		if strings.TrimSpace(req.NewPassword) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password baru wajib diisi"})
		}
		if len(req.NewPassword) < 6 {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password baru minimal 6 karakter"})
		}
		if req.NewPassword != req.ConfirmPassword {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Konfirmasi password tidak cocok"})
		}

		stored := config.GetAdminPass()
		valid := false
		if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") {
			valid = bcrypt.CompareHashAndPassword([]byte(stored), []byte(req.OldPassword)) == nil
		} else {
			valid = stored == req.OldPassword
		}
		if !valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Password lama salah"})
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengenkripsi password"})
		}
		if err := config.UpdateAdminPassword(string(hash)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan password baru"})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Password berhasil diubah"})
	}
}
