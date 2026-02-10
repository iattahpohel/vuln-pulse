package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	Host     string
	Port     int
	Password string
	DB       int
}

type RabbitMQConfig struct {
	URL       string
	QueueName string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://vulnpulse:devpassword@localhost:5432/vulnpulse?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			URL:       getEnv("RABBITMQ_URL", "amqp://vulnpulse:devpassword@localhost:5672/"),
			QueueName: getEnv("RABBITMQ_QUEUE", "vulnpulse-jobs"),
		},
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT_SEC", 15)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT_SEC", 15)) * time.Second,
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", ""),
			TokenDuration: time.Duration(getEnvAsInt("JWT_DURATION_HOURS", 24)) * time.Hour,
		},
	}

	// Validate required fields
	if cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Parse Redis URL if host/port not set separately
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = cfg.Redis.URL
		cfg.Redis.Port = 6379
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
