package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Topology struct {
		NAT bool `json:"nat"`
	}
	Clickhouse ClickhouseConfig `json:"clickhouse"`
	MongoDb    *MongoDbConfig   `json:"mongodb"`
	API        struct {
		Internal string `json:"internal"`
	}
	Hw HwConfig `json:"hw"`
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

type MongoDbConfig struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
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
