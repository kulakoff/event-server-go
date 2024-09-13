package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Hw struct {
		Beward struct {
			Port int `json:"port"`
		} `json:"beward"`
		BewardDs struct {
			Port int `json:"port"`
		} `json:"beward_ds"`
		Qtech struct {
			Port int `json:"port"`
		} `json:"qtech"`
	} `json:"hw"`
}

func LoadConfig(fileName string) (*Config, error) {
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
