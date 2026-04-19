package config

import (
	"os"
)

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	DBPath   string
}

func GetDatabaseConfig() DatabaseConfig {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	cfg := DatabaseConfig{
		Driver: driver,
		DBPath: "./database.db",
	}

	if driver == "mysql" {
		cfg.Host = getEnv("DB_HOST", "localhost")
		cfg.Port = getEnv("DB_PORT", "3306")
		cfg.User = getEnv("DB_USER", "root")
		cfg.Password = getEnv("DB_PASSWORD", "")
		cfg.Database = getEnv("DB_NAME", "bell")
	}

	return cfg
}

func (c *DatabaseConfig) GetDSN() string {
	if c.Driver == "mysql" {
		return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.Database + "?parseTime=true&charset=utf8mb4"
	}
	return c.DBPath
}

func (c *DatabaseConfig) IsMySQL() bool {
	return c.Driver == "mysql"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

const (
	UploadPath  = "public/assets/audio"
	PhotoPath   = "public/assets/photos"
	SignagePath = "public/assets/signage"
	AdminUser   = "admin"
	AdminPass   = "admin123"
	CookieName  = "session_token"
	SecretKey   = "admin-secret-key-123"
	AppVersion  = "v1.2.1"
)
