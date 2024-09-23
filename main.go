package main

import (
	"github.com/kulakoff/event-server-go/internal/config"
	"github.com/kulakoff/event-server-go/internal/handlers"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("app started")

	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// ----- Beward syslog_custom server
	bewardHandler := handlers.NewBewardHandler(logger)
	bewardServer := syslog_custom.New(cfg.Hw.Beward.Port, "Beward", logger, bewardHandler)
	go bewardServer.Start()

	// ----- Qtech syslog_custom server
	qtechHandler := handlers.NewQtechHandler(logger)
	qtechServer := syslog_custom.New(cfg.Hw.Qtech.Port, "Qtech", logger, qtechHandler)
	go qtechServer.Start()

	// Block main thread
	select {}
}
