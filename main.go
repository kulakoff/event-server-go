package main

import (
	"github.com/kulakoff/event-server-go/internal/config"
	"github.com/kulakoff/event-server-go/internal/handlers"
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"os"
)

func main() {
	// TODO: added log level from ENV
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("app started")

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// clickhouse init
	chDsn := cfg.Clickhouse
	ch, err := storage.New(logger, &chDsn)
	if err != nil {
		logger.Error("Error init Clickhouse", "error", err)
		os.Exit(1)
	}

	// mongodb init
	mongo, err := storage.NewMongoDb(logger, cfg.MongoDb.URI, cfg.MongoDb.Database)

	// load spam filter
	spamFilers, err := config.LoadSpamFilters("spamwords.json")
	if err != nil {
		logger.Warn("Error loading spam filters", "error", err)
	}

	// ----- Beward syslog_custom server
	bewardHandler := handlers.NewBewardHandler(logger, spamFilers.Beward, ch)
	bewardServer := syslog_custom.New(cfg.Hw.Beward.Port, "Beward", logger, bewardHandler)
	go bewardServer.Start()

	// ----- Qtech syslog_custom server
	qtechHandler := handlers.NewQtechHandler(logger, spamFilers.Qtech, ch)
	qtechServer := syslog_custom.New(cfg.Hw.Qtech.Port, "Qtech", logger, qtechHandler)
	go qtechServer.Start()

	//utils.GetBestQuality(8, "2024-09-27 16:26:23")
	// Block main thread
	select {}
}
