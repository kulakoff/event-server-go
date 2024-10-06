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

type ClickhouseHttpClient struct {
	logger     *slog.Logger
	httpClient *http.Client
	config     *config.ClickhouseConfig
}

func NewClickhouseHttpClient(logger *slog.Logger, config *config.ClickhouseConfig) (*ClickhouseHttpClient, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	chClient := &ClickhouseHttpClient{
		logger:     logger,
		httpClient: client,
		config:     config,
	}

	if err := chClient.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Clickhouse: %w", err)
	}

	logger.Info("Clickhouse HTTP connection established")

	return chClient, nil
}

func (c *ClickhouseHttpClient) Insert(table, data string) error {
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

func (c *ClickhouseHttpClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	clickhouseUrl := fmt.Sprintf("http://%s:%d", c.config.Host, c.config.Port)
	pingQuery := "SELECT 1"
	queryUrl := fmt.Sprintf("%s/?query=%s", clickhouseUrl, url.QueryEscape(pingQuery))

	req, err := http.NewRequestWithContext(ctx, "GET", queryUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to ping Clickhouse", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Clickhouse ping returned non-OK status", "status", resp.StatusCode)
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	c.logger.Debug("Ping to Clickhouse successful")
	return nil
}
