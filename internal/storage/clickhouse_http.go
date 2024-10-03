package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ClickhouseClientHttp struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func NewClickhouseClientHttp(host string, port int, username, password, database string) *ClickhouseClientHttp {
	return &ClickhouseClientHttp{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}
}

func (c *ClickhouseClientHttp) Insert(table string, data []map[string]interface{}) (bool, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Prepare the data in JSONEachRow format
	var requestBody bytes.Buffer
	for _, line := range data {
		jsonLine, err := json.Marshal(line)
		if err != nil {
			return false, fmt.Errorf("failed to marshal json: %w", err)
		}
		requestBody.Write(jsonLine)
		requestBody.WriteString("\n")
	}

	// Create the query string
	query := fmt.Sprintf("INSERT INTO %s.%s FORMAT JSONEachRow", c.Database, table)
	encodedQuery := url.QueryEscape(query)

	// Construct the full URL
	chUrl := fmt.Sprintf("http://%s:%d/?async_insert=1&wait_for_async_insert=0&query=%s", c.Host, c.Port, encodedQuery)

	// Create the request
	req, err := http.NewRequest("POST", chUrl, &requestBody)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("X-ClickHouse-User", c.Username)
	req.Header.Set("X-ClickHouse-Key", c.Password)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for ClickHouse errors in the response headers
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("ClickHouse error, status code: %d", resp.StatusCode)
	}

	return true, nil
}
