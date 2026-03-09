package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"charon/backend/internal/config"
	"charon/backend/internal/domain/auth"
	"charon/backend/internal/domain/wallet"
	"charon/backend/internal/httpapi"
	"charon/backend/internal/platform/logger"
	"charon/backend/internal/platform/postgres"
)

type API struct {
	cfg    config.Config
	log    *slog.Logger
	db     *sql.DB
	server *http.Server
}

func NewAPI(cfg config.Config) (*API, error) {
	log := logger.New(cfg.AppEnv)

	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.OpenSQL(startupCtx, postgres.Config{URL: cfg.PostgresURL})
	if err != nil {
		return nil, fmt.Errorf("open api postgres connection: %w", err)
	}

	accessTTL, err := time.ParseDuration(cfg.AccessTokenTTL)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("parse access token ttl: %w", err)
	}

	refreshTTL, err := time.ParseDuration(cfg.RefreshTokenTTL)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("parse refresh token ttl: %w", err)
	}

	tokenManager, err := auth.NewTokenManager(auth.TokenConfig{
		AccessSecret:  cfg.AccessTokenSecret,
		RefreshPepper: cfg.RefreshTokenPepper,
		Issuer:        cfg.JWTIssuer,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create token manager: %w", err)
	}

	authService, err := auth.NewService(
		auth.NewPostgresRepository(db),
		auth.NewArgon2idHasher(auth.RecommendedArgon2idParams(cfg.AppEnv)),
		tokenManager,
	)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create auth service: %w", err)
	}

	walletService, err := wallet.NewService(wallet.NewPostgresRepository(db))
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create wallet service: %w", err)
	}

	router, err := httpapi.NewRouter(cfg, httpapi.Dependencies{
		Auth:   authService,
		Wallet: walletService,
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create api router: %w", err)
	}

	return &API{
		cfg: cfg,
		log: log,
		db:  db,
		server: &http.Server{
			Addr:              cfg.APIHTTPAddr,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
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
		return errors.Join(fmt.Errorf("shutdown api server: %w", err), a.closeDB())
	}

	if err := <-errCh; err != nil {
		return errors.Join(err, a.closeDB())
	}

	a.log.Info("api server stopped")
	return a.closeDB()
}

func (a *API) closeDB() error {
	if a.db == nil {
		return nil
	}

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("close api postgres connection: %w", err)
	}

	a.db = nil
	return nil
}
