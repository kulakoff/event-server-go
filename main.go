package main

import (
	"context"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/feature"
	handlers2 "github.com/kulakoff/event-server-go/internal/app/event-server-go/handlers"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository"
	storage2 "github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/syslog_custom"
	"os/signal"
	"sync"
	"syscall"
	"time"

	//"github.com/kulakoff/event-server-go/internal/app/event-server-go/utils"
	"log/slog"
	"os"
	//"time"

	//"github.com/google/uuid"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
)

func main() {
	startServer()
}

// test implementation
func startServer() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("app started")

	// context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// clickhouse init
	ch, err := storage2.NewClickhouseHttpClient(logger, cfg.Clickhouse)
	if err != nil {
		logger.Error("Error init Clickhouse", "error", err)
		os.Exit(1)
	}

	// mongodb init
	mongo, err := storage2.NewMongoDb(logger, cfg.MongoDb)
	if err != nil {
		logger.Error("Error init MongoDB", "error", err)
		os.Exit(1)
	}

	// postgres init
	psqlStorage, err := storage2.NewPSQLStorage(logger, cfg.Postgres)
	if err != nil {
		logger.Error("Error init PSQLStorage", "error", err)
		os.Exit(1)
	}
	defer psqlStorage.Close()

	// init postgres storage
	repo, err := repository.NewPostgresRepository(psqlStorage.DB, logger)

	// load spam filter
	spamFilers, err := config.LoadSpamFilters("spamwords.json")
	if err != nil {
		logger.Warn("Error loading spam filters", "error", err)
	}

	// start redis stream process
	redis, err := storage2.NewRedisStorage(logger, cfg.Redis)
	if err != nil {
		logger.Error("Error init Redis", "error", err)
	}
	redis.Ping(ctx)

	// ----- Beward syslog_custom server
	bewardHandler := handlers2.NewBewardHandler(logger, spamFilers.Beward, ch, mongo, repo, cfg.RbtApi, cfg.FrsApi, redis.Client)
	bewardServer := syslog_custom.New(cfg.Hw.Beward.Port, "Beward", logger, bewardHandler)

	// start servers
	go startServerWithWG(bewardServer, ctx, &wg)

	// TODO: refactor config
	streamProcessConfig := feature.StreamProcessorConfig{
		StreamName:     "door_open_events_stream",
		GroupName:      "door_events_processor",
		WorkersCount:   3,
		BatchSize:      5,
		BlockTime:      5 * time.Second,
		PendingMinIdle: 30 * time.Second,
	}
	streamProcess := feature.NewStreamProcessor(logger, redis, mongo, ch, streamProcessConfig, repo, cfg.FrsApi)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := streamProcess.Start(ctx); err != nil {
			logger.Error("Error starting stream", "error", err)
		}
	}()

	logger.Info("✅ All services started")

	// Graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	logger.Info("🚀 Application is running. Press Ctrl+C to stop.")
	<-signalCh

	logger.Info("🛑 Shutting down ...")
	cancel()  // cancel context -  all services receive signal
	wg.Wait() // waiting for all servers to complete
}

// wrapper for usage wg sync
func startServerWithWG(server *syslog_custom.SyslogServer, ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Start(ctx)
	}()
}
