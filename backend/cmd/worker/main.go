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

	worker := app.NewWorker(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return worker.Run(ctx)
}
