package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"charon/backend/internal/config"
	"charon/backend/internal/httpapi"
	"charon/backend/internal/platform/logger"
)

type API struct {
	cfg    config.Config
	log    *slog.Logger
	server *http.Server
}

func NewAPI(cfg config.Config) *API {
	log := logger.New(cfg.AppEnv)

	return &API{
		cfg: cfg,
		log: log,
		server: &http.Server{
			Addr:              cfg.APIHTTPAddr,
			Handler:           httpapi.NewRouter(cfg),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (a *API) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.log.Info("api server starting", "addr", a.cfg.APIHTTPAddr, "env", a.cfg.AppEnv)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		a.log.Info("api server shutting down")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown api server: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	a.log.Info("api server stopped")
	return nil
}
