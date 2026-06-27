package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	JWTSecret     string
	DBPath        string
	WorkerCount   int
	QueueCapacity int
	LogLevel      string
	FrontendURL   string
	Port          string
	WorkflowDir   string
}

// Load reads config from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		DBPath:        getEnv("DB_PATH", "./data/platform.db"),
		WorkerCount:   getEnvInt("WORKER_COUNT", 5),
		QueueCapacity: getEnvInt("QUEUE_CAPACITY", 100),
		LogLevel:      getEnv("LOG_LEVEL", "debug"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
		Port:          getEnv("PORT", "8080"),
		WorkflowDir:   getEnv("WORKFLOW_DIR", "../workflows"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
