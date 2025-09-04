package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type HTTPConfig struct {
	Port string
}

type LoggerConfig struct {
	Level string
}

type Config struct {
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Logger   LoggerConfig
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
	}

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "subscriptions"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}
