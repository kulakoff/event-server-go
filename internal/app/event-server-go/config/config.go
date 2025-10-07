package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Topology     *Topology         `json:"topology"`
	Clickhouse   *ClickhouseConfig `json:"clickhouse"`
	MongoDb      *MongoDbConfig    `json:"mongodb"`
	Postgres     *PostgresConfig   `json:"postgres"`
	Redis        *RedisConfig      `json:"redis"`
	RedisStreams *RedisStreams     `json:"redis_streams"`
	RbtApi       *RbtApi           `json:"rbtApi"`
	FrsApi       *FrsApi           `json:"frsApi"`
	Hw           *HwConfig         `json:"hw"`
}

type Topology struct {
	NAT bool `json:"nat"`
}

type RbtApi struct {
	Internal string `json:"internal"`
}

type FrsApi struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type PanelConfig struct {
	Port        int    `json:"port"`
	APIEndpoint string `json:"api_endpoint,omitempty"`
}

type ClickhouseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

type RedisStreams struct {
	Stream         string `json:"stream"`
	Group          string `json:"group"`
	WorkersCount   int    `json:"workers_count"`
	PendingMinIdle int    `json:"pending_min_idle"`
	BlockTime      int    `json:"block_time"`
}

type MongoDbConfig struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	//SSLMode  string `json:"sslmode"` // disable or require
}

type HwConfig struct {
	Beward       PanelConfig `json:"beward"`
	BewardDS     PanelConfig `json:"beward_ds"`
	Qtech        PanelConfig `json:"qtech"`
	IS           PanelConfig `json:"is"`
	Hikvision    PanelConfig `json:"hikvision"`
	Akuvox       PanelConfig `json:"akuvox"`
	Rubetek      PanelConfig `json:"rubetek"`
	SputnikCloud PanelConfig `json:"sputnik_cloud"`
	Omny         PanelConfig `json:"omny"`
	Ufanet       PanelConfig `json:"ufanet"`
}

type SpamFilters struct {
	Beward       []string `json:"beward"`
	Qtech        []string `json:"qtech"`
	Akuvox       []string `json:"akuvox"`
	Rubetek      []string `json:"rubetek"`
	SputnikCloud []string `json:"sputnik_cloud"`
	Omny         []string `json:"omny"`
	Ufanet       []string `json:"ufanet"`
}

// New parse json config file
func New(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &Config{}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return config, nil
}

// LoadSpamFilters spam words per service
func LoadSpamFilters(filename string) (*SpamFilters, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var filters SpamFilters
	if err := json.NewDecoder(file).Decode(&filters); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &filters, nil
}
