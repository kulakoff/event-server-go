package storage

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"log/slog"
	"time"
)

type ClikhouseHandler struct {
	logger     *slog.Logger
	clickhouse clickhouse.Conn
}

type SyslogStorageMessage struct {
	Date  time.Time `json:"date"`
	Ip    string    `json:"ip"`
	SubId string    `json:"sub_id"`
	Unit  string    `json:"unit"`
	Msg   string    `json:"msg"`
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

func (c *ClikhouseHandler) SendLog(message *SyslogStorageMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := c.clickhouse.Exec(ctx, "INSERT INTO syslog (date, ip, sub_id, unit, msg) VALUES (?, ?, ? ,? ,?)",
		message.Date, message.Ip, message.SubId, message.Unit, message.Msg)
	if err != nil {
		c.logger.Warn("Failed to insert into ClickHouse", "error", err)
	}

	c.logger.Info("Message inserted")
}
