package app

import (
	"context"
	"log/slog"

	"charon/backend/internal/config"
	"charon/backend/internal/platform/logger"
)

type Worker struct {
	cfg config.Config
	log *slog.Logger
}

func NewWorker(cfg config.Config) *Worker {
	return &Worker{
		cfg: cfg,
		log: logger.New(cfg.AppEnv),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("worker process starting", "env", w.cfg.AppEnv)
	<-ctx.Done()
	w.log.Info("worker process stopped")
	return nil
}
