package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charon/backend/internal/config"
	"charon/backend/internal/platform/postgres"
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

	seedSet := cfg.AppEnv
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		seedSet = strings.TrimSpace(args[0])
	}

	seedDir := filepath.Join(cfg.SeedsDir, seedSet)
	files, err := seedFiles(seedDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Printf("no seed files found in %s\n", seedDir)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.OpenSQL(ctx, postgres.Config{URL: cfg.PostgresURL})
	if err != nil {
		return err
	}
	defer db.Close()

	for _, file := range files {
		if err := applySeedFile(ctx, db, file); err != nil {
			return fmt.Errorf("apply seed %s: %w", file, err)
		}
		fmt.Printf("applied seed %s\n", file)
	}

	return nil
}

func seedFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		files = append(files, filepath.Join(dir, entry.Name()))
	}

	slices.Sort(files)
	return files, nil
}

func applySeedFile(ctx context.Context, db *sql.DB, path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
