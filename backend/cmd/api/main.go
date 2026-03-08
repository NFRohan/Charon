package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"charon/backend/internal/app"
	"charon/backend/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	api, err := app.NewAPI(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return api.Run(ctx)
}
