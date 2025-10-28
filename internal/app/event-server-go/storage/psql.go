package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
	"log/slog"
	"time"
)

type PSQLStorage struct {
	logger *slog.Logger
	DB     *pgxpool.Pool
}

func NewPSQLStorage(logger *slog.Logger, postgresConfig *config.PostgresConfig) (*PSQLStorage, error) {
	// Format connection str
	connStr := formatPostgresURL(postgresConfig)

	// Creat connection pool
	psqlConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DSN: %w", err)
	}
	// configure connection pool
	psqlConf.MaxConns = 10 // max connect
	psqlConf.ConnConfig.ConnectTimeout = 5 * time.Second

	// Connect to db
	db, err := pgxpool.NewWithConfig(context.Background(), psqlConf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Check connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Error("Unable to ping database", "error", err)
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Debug("Successfully connected to PostgreSQL")

	return &PSQLStorage{
		logger: logger,
		DB:     db,
	}, nil
}

// Close db connection
func (s *PSQLStorage) Close() {
	if s.DB != nil {
		s.DB.Close()
		s.logger.Info("🛑 Success closed connection to PostgreSQL")
	}
}

// formatPostgresURL config to connection URI
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
