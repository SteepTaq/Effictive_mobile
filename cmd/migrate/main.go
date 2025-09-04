package main

import (
	"flag"
	"fmt"
	"log/slog"

	"rest-service/internal/config"
	"rest-service/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		panic(err)
	}
}

func run() error {
	action := flag.String("action", "up", "")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	slog.SetDefault(logger.Setup(cfg.Logger.Level))
	slog.Info("starting migrations", "action", *action)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("failed to run migrate: %w", err)
	}
	switch *action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("up migration failed: %w", err)
		}
		slog.Info("migrations up completed")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("down migration failed: %w", err)
		}
		slog.Info("migrations down completed")
	default:
		return fmt.Errorf("unknown action: %s", *action)
	}
	return nil
}
