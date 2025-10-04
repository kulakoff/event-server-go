package main

import (
	"context"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/feature/stream_publisher"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("Starting test publisher")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	var wg sync.WaitGroup

	// Загрузка конфигурации
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}

	redisStorage, err := storage.NewRedisStorage(logger, cfg.Redis)
	if err != nil {
		logger.Error("Failed to initialize Redis", "error", err)
		return
	}
	defer redisStorage.Close()

	// make fake publisher
	//publisger := stream_publisher.NewStreamPublisher(logger, redisStorage, stream_publisher.StreamPublisherConfig{
	//	StreamName:        "door_open_events_stream",
	//	MessagesPerSecond: 1,
	//	Mode:              "steady",
	//})

	publisher := stream_publisher.NewStreamPublisher(logger, redisStorage, stream_publisher.StreamPublisherConfig{
		StreamName:    "door_open_events_stream",
		PublishRate:   50,              // 50 сообщений в пачке
		BurstInterval: 5 * time.Second, // каждые 5 секунд
		Mode:          "batch",         // явно указываем режим
	})

	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := publisher.Start(ctx); err != nil {
			logger.Error("Failed to start publisher", "error", err)
		}
	}()

	wg.Wait()
	logger.Info("Test publisher stopped")
}
