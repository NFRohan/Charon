package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	AppEnvDevelopment = "development"
	AppEnvTest        = "test"
	AppEnvProduction  = "production"
)

type Config struct {
	AppEnv        string
	APIHTTPAddr   string
	PublicBaseURL string
	PostgresURL   string
	RedisURL      string
	RabbitMQURL   string
	AdminWebURL   string
	MigrationsDir string
	SeedsDir      string
}

func Load() (Config, error) {
	if err := loadDotEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:        getEnv("CHARON_APP_ENV", AppEnvDevelopment),
		APIHTTPAddr:   getEnv("CHARON_API_HTTP_ADDR", ":8080"),
		PublicBaseURL: getEnv("CHARON_PUBLIC_BASE_URL", "http://localhost:8080"),
		PostgresURL:   getEnv("CHARON_POSTGRES_URL", "postgres://charon:charon@localhost:5432/charon?sslmode=disable"),
		RedisURL:      getEnv("CHARON_REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL:   getEnv("CHARON_RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		AdminWebURL:   getEnv("CHARON_ADMIN_WEB_URL", "http://localhost:3000"),
		MigrationsDir: getEnv("CHARON_MIGRATIONS_DIR", "migrations"),
		SeedsDir:      getEnv("CHARON_SEEDS_DIR", "seeds"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	switch c.AppEnv {
	case AppEnvDevelopment, AppEnvTest, AppEnvProduction:
	default:
		errs = append(errs, fmt.Errorf("unsupported CHARON_APP_ENV %q", c.AppEnv))
	}

	required := map[string]string{
		"CHARON_API_HTTP_ADDR":   c.APIHTTPAddr,
		"CHARON_PUBLIC_BASE_URL": c.PublicBaseURL,
		"CHARON_POSTGRES_URL":    c.PostgresURL,
		"CHARON_REDIS_URL":       c.RedisURL,
		"CHARON_RABBITMQ_URL":    c.RabbitMQURL,
		"CHARON_ADMIN_WEB_URL":   c.AdminWebURL,
		"CHARON_MIGRATIONS_DIR":  c.MigrationsDir,
		"CHARON_SEEDS_DIR":       c.SeedsDir,
	}

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s must not be empty", key))
		}
	}

	return errors.Join(errs...)
}

func loadDotEnv() error {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return godotenv.Load(candidate)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", candidate, err)
		}
	}

	return nil
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}
