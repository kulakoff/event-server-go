package main

import (
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
	"log"
	"log/slog"
)

func main() {
	slog.Info("app started")
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error load config file: %v", err)
	}

	fmt.Printf("Bwward event server used port: %d\n", cfg.Hw.Beward.Port)
	fmt.Printf("Bwward DS event server used port: %d\n", cfg.Hw.BewardDs.Port)
	fmt.Printf("Qtech event server used port: %d\n", cfg.Hw.Qtech.Port)
}
