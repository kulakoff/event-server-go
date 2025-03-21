package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kulakoff/event-server-go/internal/config"
	"log/slog"
	"time"
)

type PSQLStorage struct {
	logger *slog.Logger
	db     *pgxpool.Pool
}

func NewPSQLStorage(logger *slog.Logger, postgresConfig *config.PostgresConfig) (*PSQLStorage, error) {
	connStr := formatPostgresURL(postgresConfig)
	psqlConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DSN: %w", err)
	}
	psqlConf.MaxConns = 10 // max connect
	psqlConf.ConnConfig.ConnectTimeout = 5 * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), psqlConf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// check connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Error("Unable to ping database", "error", err)
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PSQLStorage{
		logger: logger,
		db:     db}, nil
}

func (s *PSQLStorage) Close() {
	if s.db != nil {
		s.db.Close()
		s.logger.Info("Success closed connection to PostgreSQL")
	}
}

func formatPostgresURL(cfg *config.PostgresConfig) string {
	//urlExample := "postgres://username:password@localhost:5432/database_name"
	//return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		//cfg.SSLMode,
	)
}
