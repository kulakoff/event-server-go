package storage

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"log/slog"
)

type ClikhouseHandler struct {
	logger     *slog.Logger
	clickhouse clickhouse.Conn
}

func New(logger *slog.Logger, dsn string) (*ClikhouseHandler, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "qqq",
			Database: "default",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed connect to clickhouse dsn %s: %w ", dsn, err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed ping clickhouse dsn %s: %w ", dsn, err)
	}

	return &ClikhouseHandler{logger: logger, clickhouse: conn}, nil
}
