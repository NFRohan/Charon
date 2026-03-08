package main

import (
	"context"
	"os/signal"
	"syscall"

	"charon/backend/internal/config"
	"charon/backend/internal/platform/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.AppEnv)

	log.Info("worker process starting", "env", cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Info("worker process stopped")
}
