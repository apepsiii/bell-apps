package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Paths    PathsConfig    `yaml:"paths"`
	Auth     AuthConfig     `yaml:"auth"`
	Log      LogConfig      `yaml:"log"`
	Weather  WeatherConfig  `yaml:"weather"`
}

type WeatherConfig struct {
	APIKey  string `yaml:"openweather_api_key"`
	City    string `yaml:"city"`
	Country string `yaml:"country"`
}

type ServerConfig struct {
	Port   string `yaml:"port"`
	Host   string `yaml:"host"`
	Domain string `yaml:"domain"`
}

type LogConfig struct {
	Dir string `yaml:"dir"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
}

type PathsConfig struct {
	Upload   string `yaml:"upload"`
	Photo    string `yaml:"photo"`
	Signage  string `yaml:"signage"`
	Audio    string `yaml:"audio"`
}

type AuthConfig struct {
	AdminUser   string `yaml:"admin_user"`
	AdminPass   string `yaml:"admin_password"`
	CookieName  string `yaml:"cookie_name"`
	SecretKey   string `yaml:"secret_key"`
}

var (
	AppVersion  string
	config      *AppConfig
	configPath  = "config.yaml"
)

// Load parses the given YAML bytes as the default configuration. If a
// config.yaml file exists on disk next to the binary, it overrides the
// embedded default (so operators can still customize without rebuilding).
// Environment variables override file values for specific keys.
func Load(yamlData []byte) error {
	config = &AppConfig{}

	// Start from embedded default.
	if err := yaml.Unmarshal(yamlData, config); err != nil {
		return err
	}

	// Override with on-disk config.yaml if present (operator customization).
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		// No on-disk config: write the embedded default so operators can see
		// and edit it on first run.
		_ = os.WriteFile(configPath, yamlData, 0644)
	} else {
		return err
	}

	AppVersion = getEnv("APP_VERSION", "v1.3.0")

	return nil
}

// InitConfig is the legacy entry point retained for compatibility. It reads
// config.yaml from disk, creating a default if missing. New code should
// prefer Load([]byte) with an embedded default.
func InitConfig() error {
	config = &AppConfig{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	AppVersion = getEnv("APP_VERSION", "v1.3.0")

	return nil
}

func createDefaultConfig() error {
	defaultConfig := AppConfig{
		Server: ServerConfig{
			Port:   "8080",
			Host:   "0.0.0.0",
			Domain: "",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			Path:   "./database.db",
		},
		Paths: PathsConfig{
			Upload:   "public/assets/audio",
			Photo:    "public/assets/photos",
			Signage:  "public/assets/signage",
			Audio:    "public/assets/audio",
		},
		Auth: AuthConfig{
			AdminUser:   "admin",
			AdminPass:   "admin123",
			CookieName:  "session_token",
			SecretKey:   "admin-secret-key-change-me",
		},
		Log: LogConfig{
			Dir: "logs",
		},
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func GetConfig() *AppConfig {
	if config == nil {
		InitConfig()
	}
	return config
}

func GetServerPort() string {
	if config != nil && config.Server.Port != "" {
		return config.Server.Port
	}
	return getEnv("SERVER_PORT", "8080")
}

func GetServerHost() string {
	if config != nil && config.Server.Host != "" {
		return config.Server.Host
	}
	return getEnv("SERVER_HOST", "0.0.0.0")
}

func GetDBConfig() DatabaseConfig {
	if config != nil {
		return config.Database
	}
	return DatabaseConfig{
		Driver: getEnv("DB_DRIVER", "sqlite"),
		Host:   getEnv("DB_HOST", "localhost"),
		Port:   getEnv("DB_PORT", "3306"),
		User:   getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:   getEnv("DB_NAME", "bell"),
		Path:   getEnv("DB_PATH", "./database.db"),
	}
}

func (c *DatabaseConfig) GetDSN() string {
	if c.Driver == "mysql" {
		return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.Name + "?parseTime=true&charset=utf8mb4"
	}
	return c.Path
}

func IsMySQL() bool {
	db := GetDBConfig()
	return db.Driver == "mysql"
}

func GetPaths() PathsConfig {
	if config != nil {
		return config.Paths
	}
	return PathsConfig{
		Upload:   getEnv("UPLOAD_PATH", "public/assets/audio"),
		Photo:    getEnv("PHOTO_PATH", "public/assets/photos"),
		Signage:  getEnv("SIGNAGE_PATH", "public/assets/signage"),
		Audio:    getEnv("AUDIO_PATH", "public/assets/audio"),
	}
}

func GetAuth() AuthConfig {
	if config != nil {
		return config.Auth
	}
	return AuthConfig{
		AdminUser:  getEnv("ADMIN_USER", "admin"),
		AdminPass:  getEnv("ADMIN_PASSWORD", "admin123"),
		CookieName: getEnv("COOKIE_NAME", "session_token"),
		SecretKey:  getEnv("SECRET_KEY", "admin-secret-key-change-me"),
	}
}

func GetUploadPath() string {
	return GetPaths().Upload
}

func GetPhotoPath() string {
	return GetPaths().Photo
}

func GetSignagePath() string {
	return GetPaths().Signage
}

func GetAudioPath() string {
	return GetPaths().Audio
}

func GetLogDir() string {
	if config != nil && config.Log.Dir != "" {
		return config.Log.Dir
	}
	return getEnv("LOG_DIR", "logs")
}

// GetDomain returns the configured public domain (without scheme/port).
// Empty when not set — callers should fall back to building a URL from the
// request host so QR codes still work on LAN during setup.
func GetDomain() string {
	if config != nil {
		return config.Server.Domain
	}
	return ""
}

func GetAdminUser() string {
	return GetAuth().AdminUser
}

func GetAdminPass() string {
	return GetAuth().AdminPass
}

func GetCookieName() string {
	return GetAuth().CookieName
}

func GetSecretKey() string {
	return GetAuth().SecretKey
}

func GetConfigPath() string {
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), "config.yaml")
}

func GetWeatherConfig() WeatherConfig {
	if config != nil {
		return config.Weather
	}
	return WeatherConfig{
		APIKey:  getEnv("OPENWEATHER_API_KEY", ""),
		City:    getEnv("WEATHER_CITY", "Bogor"),
		Country: getEnv("WEATHER_COUNTRY", "ID"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
