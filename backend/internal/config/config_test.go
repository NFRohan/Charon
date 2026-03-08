package config

import "testing"

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	valid := Config{
		AppEnv:             AppEnvDevelopment,
		APIHTTPAddr:        ":8080",
		PublicBaseURL:      "http://localhost:8080",
		PostgresURL:        "postgres://charon:charon@localhost:5432/charon?sslmode=disable",
		RedisURL:           "redis://localhost:6379/0",
		RabbitMQURL:        "amqp://guest:guest@localhost:5672/",
		AdminWebURL:        "http://localhost:3000",
		AccessTokenSecret:  "01234567890123456789012345678901",
		RefreshTokenPepper: "abcdefghijklmnopqrstuvwxyz123456",
		JWTIssuer:          "charon",
		AccessTokenTTL:     "15m",
		RefreshTokenTTL:    "720h",
		MigrationsDir:      "migrations",
		SeedsDir:           "seeds",
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}

	invalid := valid
	invalid.AppEnv = "staging"
	invalid.PostgresURL = ""

	if err := invalid.Validate(); err == nil {
		t.Fatal("expected validation error for unsupported env and empty postgres url")
	}

	weak := valid
	weak.AccessTokenSecret = "short-secret"
	weak.RefreshTokenPepper = "short-pepper"

	if err := weak.Validate(); err == nil {
		t.Fatal("expected validation error for weak auth secrets")
	}

	production := valid
	production.AppEnv = AppEnvProduction
	production.AccessTokenSecret = "development-access-secret-change-before-shipping"
	production.RefreshTokenPepper = "development-refresh-pepper-change-before-shipping"

	if err := production.Validate(); err == nil {
		t.Fatal("expected validation error for development secrets in production")
	}
}
