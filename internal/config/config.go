package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Database
	DatabaseURL string

	// MinIO Object Storage
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// Application
	Port                 string
	SessionSecret        string
	StorageQuotaDefault  int64
	Environment          string
	LogLevel             string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "media"),
		MinIOUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		Port:           getEnv("PORT", "8080"),
		SessionSecret:  getEnv("SESSION_SECRET", ""),
		StorageQuotaDefault: getEnvInt64("STORAGE_QUOTA_DEFAULT", 524288000), // 500MB
		Environment:    getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}
	if len(config.SessionSecret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}

	return config, nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

