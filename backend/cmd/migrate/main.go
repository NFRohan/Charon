package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"charon/backend/internal/config"
	"charon/backend/internal/platform/postgres"

	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	goose.SetDialect("postgres")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := postgres.OpenSQL(ctx, postgres.Config{URL: cfg.PostgresURL})
	if err != nil {
		return err
	}
	defer db.Close()

	command := "up"
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "up":
		return goose.Up(db, cfg.MigrationsDir)
	case "down":
		return goose.Down(db, cfg.MigrationsDir)
	case "reset":
		return goose.Reset(db, cfg.MigrationsDir)
	case "status":
		return goose.Status(db, cfg.MigrationsDir)
	default:
		return fmt.Errorf("unsupported migrate command %q, expected one of: up, down, reset, status", command)
	}
}
