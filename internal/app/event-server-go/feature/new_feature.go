package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"sync"
	"time"
)

// DoorOpenEvent - parse structure from php backend
type DoorOpenEvent struct {
	Date      int64  `json:"date"`
	IP        string `json:"ip"`
	SubID     *int64 `json:"sub_id"` // может быть null
	EventType int    `json:"event_type"`
	Door      int    `json:"door"`
	Detail    string `json:"detail"`
	Timestamp int64  `json:"timestamp"`
}

type StreamProcessorConfig struct {
	StreamName     string
	GroupName      string
	WorkersCount   int
	BatchSize      int
	BlockTime      time.Duration
	PendingMinIdle time.Duration
}

type StreamProcessor struct {
	logger *slog.Logger
	redis  *storage.RedisStorage
	config StreamProcessorConfig
	wg     sync.WaitGroup
}

func NewStreamProcessor(
	logger *slog.Logger,
	redisStorage *storage.RedisStorage,
	config StreamProcessorConfig,
) *StreamProcessor {
	return &StreamProcessor{
		logger: logger,
		redis:  redisStorage,
		config: config,
	}
}

// Start - process stream messages
func (s *StreamProcessor) Start(ctx context.Context) error {
	// Init consumer group
	if err := s.initConsumerGroup(ctx); err != nil {
		return fmt.Errorf("failed to init consumer group: %w", err)
	}

	// Start workers
	for i := 1; i <= s.config.WorkersCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, fmt.Sprintf("worker_%d", i))
	}

	// Start pending worker, process failed task
	s.wg.Add(1)
	go s.pendingWorker(ctx)

	s.logger.Info("Stream processor started",
		"workers", s.config.WorkersCount,
		"stream", s.config.StreamName,
		"group", s.config.GroupName)

	return nil
}

// initConsumerGroup make consumer group
func (s *StreamProcessor) initConsumerGroup(ctx context.Context) error {
	err := s.redis.Client.XGroupCreateMkStream(
		ctx,
		s.config.StreamName,
		s.config.GroupName,
		"0",
	).Err()

	if err != nil {
		// Consumer group already exist, normal!
		s.logger.Debug("Consumer group already exists", "group", s.config.GroupName)
	} else {
		s.logger.Info("Consumer group created", "group", s.config.GroupName)
	}

	return nil
}

// worker - main worker for message process
func (s *StreamProcessor) worker(ctx context.Context, workerName string) {
	defer s.wg.Done()

	s.logger.Info("Worker started", "worker", workerName)
	processedCount := 0

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Worker stopping",
				"worker", workerName,
				"processed", processedCount)
			return
		default:
			// Читаем сообщения из stream
			result, err := s.redis.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    s.config.GroupName,
				Consumer: workerName,
				Streams:  []string{s.config.StreamName, ">"},
				Count:    int64(s.config.BatchSize),
				Block:    s.config.BlockTime,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// Таймаут - нет новых сообщений
					continue
				}
				s.logger.Error("Error reading from stream",
					"worker", workerName,
					"error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем полученные сообщения
			for _, stream := range result {
				for _, message := range stream.Messages {
					processedCount++

					if s.processDoorEvent(message, workerName) {
						// Подтверждаем обработку
						err := s.redis.Client.XAck(
							ctx,
							stream.Stream,
							s.config.GroupName,
							message.ID,
						).Err()

						if err != nil {
							s.logger.Error("Failed to ack message",
								"worker", workerName,
								"message_id", message.ID,
								"error", err)
						} else {
							s.logger.Debug("Message acknowledged",
								"worker", workerName,
								"message_id", message.ID)
						}
					} else {
						s.logger.Warn("Processing failed, will retry",
							"worker", workerName,
							"message_id", message.ID)
					}

					// Логируем прогресс каждые 10 сообщений
					if processedCount%10 == 0 {
						s.logger.Info("Worker progress",
							"worker", workerName,
							"processed", processedCount)
					}
				}
			}
		}
	}
}

// processDoorEvent - process single event
func (s *StreamProcessor) processDoorEvent(message redis.XMessage, workerName string) bool {
	// Извлекаем payload
	payload, ok := message.Values["payload"].(string)
	if !ok {
		s.logger.Error("Invalid payload format",
			"worker", workerName,
			"message_id", message.ID)
		return false
	}

	// Парсим JSON с данными события
	var event DoorOpenEvent
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		s.logger.Error("Failed to unmarshal event",
			"worker", workerName,
			"message_id", message.ID,
			"error", err)
		return false
	}

	// Основная логика обработки события
	s.logger.Debug("Processing door event",
		"worker", workerName,
		"ip", event.IP,
		"event_type", event.EventType,
		"door", event.Door,
		"detail", event.Detail)

	// TODO: Замените на вашу реальную логику сохранения в БД
	success := s.saveToDatabase(event)
	if !success {
		s.logger.Error("Failed to save event to database",
			"worker", workerName,
			"message_id", message.ID)
		return false
	}

	return true
}

// saveToDatabase - storage data
func (s *StreamProcessor) saveToDatabase(event DoorOpenEvent) bool {
	// TODO: work imitation
	time.Sleep(50 * time.Millisecond)

	// TODO: 95% true, 5% false for test
	if time.Now().UnixNano()%100 < 5 {
		return false
	}

	return true
}

// pendingWorker - process failed tasks
func (s *StreamProcessor) pendingWorker(ctx context.Context) {
	defer s.wg.Done()

	workerName := "pending_recovery"
	s.logger.Info("Pending worker started", "worker", workerName)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Pending worker stopping", "worker", workerName)
			return
		default:
			// process pending tasks
			messages, _, err := s.redis.Client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   s.config.StreamName,
				Group:    s.config.GroupName,
				Consumer: workerName,
				MinIdle:  s.config.PendingMinIdle,
				Count:    10,
				Start:    "0-0",
			}).Result()

			if err != nil {
				s.logger.Error("Pending worker error",
					"worker", workerName,
					"error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(messages) > 0 {
				s.logger.Info("Pending worker processing messages",
					"worker", workerName,
					"count", len(messages))

				for _, msg := range messages {
					if s.processDoorEvent(msg, workerName) {
						s.redis.Client.XAck(ctx, s.config.StreamName, s.config.GroupName, msg.ID)
					}
				}
			}

			time.Sleep(10 * time.Second)
		}
	}
}
