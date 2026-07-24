package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const defaultLogDir = "logs"

// InitLogging configures slog to write JSON logs to both stdout and
// <logDir>/app.log (mode 0640). It must be called early in main, before
// any other component starts logging. Secrets must never be logged via slog.
func InitLogging(logDir string) error {
	if logDir == "" {
		logDir = defaultLogDir
	}

	if err := os.MkdirAll(logDir, 0750); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return err
	}

	writer := io.MultiWriter(os.Stdout, file)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	return nil
}
