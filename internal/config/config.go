package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv              string
	BindAddress         string
	PublicOrigin        string
	DataDir             string
	DBPath              string
	UploadsDir          string
	StaticDir           string
	OpsToken            string
	SafeMode            bool
	OpenBrowser         bool
	BackupRetentionDays int
	MaintenanceInterval time.Duration
	ReadHeaderTimeout   time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	MaxHeaderBytes      int
}

func Load() Config {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}

	appEnv := resolveAppEnv()
	dataDir := envOr("UNIVERSALD_DATA_DIR", baseDir)
	bind := resolveBindAddress()
	dbPath := envOr("UNIVERSALD_DB", filepath.Join(dataDir, "universald.db"))
	uploadsDir := envOr("UNIVERSALD_UPLOADS", filepath.Join(dataDir, "panel_uploads"))
	staticDir := envOr("UNIVERSALD_WEB", filepath.Join(baseDir, "web"))

	return Config{
		AppEnv:              appEnv,
		BindAddress:         bind,
		PublicOrigin:        normalizePublicOrigin(os.Getenv("UNIVERSALD_PUBLIC_ORIGIN")),
		DataDir:             dataDir,
		DBPath:              dbPath,
		UploadsDir:          uploadsDir,
		StaticDir:           staticDir,
		OpsToken:            strings.TrimSpace(os.Getenv("UNIVERSALD_OPS_TOKEN")),
		SafeMode:            envBool("UNIVERSALD_SAFE_MODE", true),
		OpenBrowser:         envBool("UNIVERSALD_OPEN", defaultOpenBrowser()),
		BackupRetentionDays: envInt("UNIVERSALD_BACKUP_RETENTION_DAYS", 14),
		MaintenanceInterval: envDurationSeconds("UNIVERSALD_MAINTENANCE_INTERVAL_SEC", 60),
		ReadHeaderTimeout:   envDurationSeconds("UNIVERSALD_READ_HEADER_TIMEOUT_SEC", 5),
		ReadTimeout:         envDurationSeconds("UNIVERSALD_READ_TIMEOUT_SEC", 10),
		WriteTimeout:        envDurationSeconds("UNIVERSALD_WRITE_TIMEOUT_SEC", 15),
		IdleTimeout:         envDurationSeconds("UNIVERSALD_IDLE_TIMEOUT_SEC", 60),
		ShutdownTimeout:     envDurationSeconds("UNIVERSALD_SHUTDOWN_TIMEOUT_SEC", 5),
		MaxHeaderBytes:      envInt("UNIVERSALD_MAX_HEADER_BYTES", 1<<20),
	}
}

func envOr(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envDurationSeconds(key string, fallbackSeconds int) time.Duration {
	return time.Duration(envInt(key, fallbackSeconds)) * time.Second
}

func resolveBindAddress() string {
	if bind := strings.TrimSpace(os.Getenv("UNIVERSALD_BIND")); bind != "" {
		return bind
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return "0.0.0.0:" + port
	}
	return "127.0.0.1:7788"
}

func resolveAppEnv() string {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("UNIVERSALD_APP_ENV")))
	switch value {
	case "production", "prod":
		return "production"
	case "staging", "stage":
		return "staging"
	case "development", "dev", "local":
		return "development"
	}
	if envBool("RENDER", false) || strings.TrimSpace(os.Getenv("PORT")) != "" {
		return "production"
	}
	return "development"
}

func normalizePublicOrigin(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimRight(value, "/")
	return value
}

func defaultOpenBrowser() bool {
	if strings.TrimSpace(os.Getenv("PORT")) != "" {
		return false
	}
	if envBool("CI", false) {
		return false
	}
	if envBool("RENDER", false) {
		return false
	}
	return true
}
