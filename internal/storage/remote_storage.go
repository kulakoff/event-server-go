package storage

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/kulakoff/event-server-go/internal/config"
	"log/slog"
	"strconv"
	"time"
)

type ClikhouseHandler struct {
	logger     *slog.Logger
	clickhouse clickhouse.Conn
}

type SyslogStorageMessage struct {
	Date  string `json:"date"`
	Ip    string `json:"ip"`
	SubId string `json:"sub_id"`
	Unit  string `json:"unit"`
	Msg   string `json:"msg"`
}

func New(logger *slog.Logger, config *config.ClickhouseConfig) (*ClikhouseHandler, error) {
	dsn := config.Host + ":" + strconv.Itoa(config.Port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Username: config.Username,
			Password: config.Password,
			Database: config.Database,
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

func (c *ClikhouseHandler) SendLog(message SyslogStorageMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := c.clickhouse.Exec(ctx, "INSERT INTO syslog (date, ip, sub_id, unit, msg) VALUES (?, ?, ? ,? ,?)",
		message.Date, message.Ip, message.SubId, message.Unit, message.Msg)
	if err != nil {
		c.logger.Error("Failed to insert into ClickHouse", "error", err)
		return
	}

	c.logger.Debug("Message inserted")
}
