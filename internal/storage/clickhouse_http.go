package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type ClickhouseClientHttp struct {
	logger     *slog.Logger
	httpClient *http.Client
	config     *config.ClickhouseConfig
}

func NewClickhouseClientHttp(logger *slog.Logger, config *config.ClickhouseConfig) (*ClickhouseClientHttp, error) {
	clickhouseUrl := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	logger.Info("Clickhouse HTTP connection established", "url", clickhouseUrl)

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	return &ClickhouseClientHttp{
		logger:     logger,
		httpClient: client,
		config:     config,
	}, nil
}

func (c *ClickhouseClientHttp) Insert(table, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	clickhouseUrl := fmt.Sprintf("http://%s:%d", c.config.Host, c.config.Port)
	query := fmt.Sprintf("INSERT INTO %s.%s FORMAT JSONEachRow", c.config.Database, table)
	queryUrl := fmt.Sprintf("%s/?async_insert=1&wait_for_async_insert=0&query=%s", clickhouseUrl, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "POST", queryUrl, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send request to Clickhouse", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Clickhouse returned non-OK status", "status", resp.StatusCode)
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	c.logger.Debug("Data inserted success to Clickhouse", "table", table)
	return nil
}
