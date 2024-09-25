package storage

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"log/slog"
)

type ClikhouseHandler struct {
	logger     *slog.Logger
	clickhouse clickhouse.Conn
}
