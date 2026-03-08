package config

import "os"

type Config struct {
	AppEnv        string
	APIHTTPAddr   string
	PublicBaseURL string
	PostgresURL   string
	RedisURL      string
	RabbitMQURL   string
	AdminWebURL   string
}

func Load() Config {
	return Config{
		AppEnv:        getEnv("CHARON_APP_ENV", "development"),
		APIHTTPAddr:   getEnv("CHARON_API_HTTP_ADDR", ":8080"),
		PublicBaseURL: getEnv("CHARON_PUBLIC_BASE_URL", "http://localhost:8080"),
		PostgresURL:   getEnv("CHARON_POSTGRES_URL", "postgres://charon:charon@localhost:5432/charon?sslmode=disable"),
		RedisURL:      getEnv("CHARON_REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL:   getEnv("CHARON_RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		AdminWebURL:   getEnv("CHARON_ADMIN_WEB_URL", "http://localhost:3000"),
	}
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}
