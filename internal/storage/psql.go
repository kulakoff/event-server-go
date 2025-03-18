package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PSQLStorage struct {
	db *pgxpool.Pool
}

func NewPSQLStorage(dsn string) (*PSQLStorage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DSN: %w", err)
	}
	config.MaxConns = 10 // max connect
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &PSQLStorage{db: db}, nil
}
