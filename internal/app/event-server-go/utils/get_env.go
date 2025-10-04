package utils

import (
	"os"
	"strconv"
)

func GetEnv(envName, defaultValue string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvInt(envName string, defaultValue int) int {
	if value := os.Getenv(envName); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
