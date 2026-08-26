package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment   string
	HTTPAddr      string
	DatabaseURL   string
	DBDriver      string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	TokenTTL      time.Duration
	CORSOrigins   []string
	RateLimit     int64
}

func Load() (Config, error) {
	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "12"))
	if err != nil || ttlHours < 1 {
		return Config{}, fmt.Errorf("JWT_TTL_HOURS must be a positive integer")
	}
	limit, err := strconv.ParseInt(getEnv("RATE_LIMIT_PER_MINUTE", "180"), 10, 64)
	if err != nil || limit < 1 {
		return Config{}, fmt.Errorf("RATE_LIMIT_PER_MINUTE must be a positive integer")
	}
	cfg := Config{
		Environment:   strings.ToLower(getEnv("APP_ENV", "development")),
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://aquaculture:aquaculture_dev_password@localhost:5432/aquaculture?sslmode=disable"),
		DBDriver:      strings.ToLower(getEnv("DB_DRIVER", "postgres")),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     getEnv("JWT_SECRET", "local-development-secret-change-me"),
		TokenTTL:      time.Duration(ttlHours) * time.Hour,
		CORSOrigins:   splitCSV(getEnv("CORS_ORIGINS", "http://localhost:18502")),
		RateLimit:     limit,
	}
	if len(cfg.JWTSecret) < 24 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 24 characters")
	}
	if cfg.DBDriver != "postgres" && cfg.DBDriver != "sqlite" {
		return Config{}, fmt.Errorf("DB_DRIVER must be postgres or sqlite")
	}
	if cfg.Environment == "production" && (len(cfg.JWTSecret) < 32 || cfg.JWTSecret == "local-development-secret-change-me" || strings.Contains(strings.ToLower(cfg.JWTSecret), "replace-with")) {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 non-default characters in production")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
