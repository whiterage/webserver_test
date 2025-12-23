package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// servernye nastroyki
	ServerPort string

	// postgres
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	// redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// api
	APIKey string

	// webhook
	WebhookURL           string
	WebhookRetryAttempts int
	WebhookRetryDelaySec time.Duration

	// statistika
	StatsTimeWindowMinutes int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "geo_alert"),
		PostgresSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		APIKey: getEnv("API_KEY", ""),

		WebhookURL:           getEnv("WEBHOOK_URL", "http://localhost:9090/webhook"),
		WebhookRetryAttempts: getEnvAsInt("WEBHOOK_RETRY_ATTEMPTS", 3),
		WebhookRetryDelaySec: time.Duration(getEnvAsInt("WEBHOOK_RETRY_DELAY_SECONDS", 5)) * time.Second,

		StatsTimeWindowMinutes: getEnvAsInt("STATS_TIME_WINDOW_MINUTES", 60),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY is not set")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// vozvrashyaem podkluchenie k postgres
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

// vozvrashyaem adres redis
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
