package stream_publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/redis/go-redis/v9"
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

type StreamPublisherConfig struct {
	StreamName        string
	PublishRate       int           // сообщений в пачке (для batch режима)
	BurstInterval     time.Duration // интервал между пачками (для batch режима)
	MessagesPerSecond int           // сообщений в секунду (для steady режима)
	IPPool            []string      // пул IP адресов для генерации
	Mode              string        // "batch" или "steady"
}

type StreamPublisher struct {
	logger *slog.Logger
	redis  *storage.RedisStorage
	config StreamPublisherConfig
}

func NewStreamPublisher(
	logger *slog.Logger,
	redisStorage *storage.RedisStorage,
	config StreamPublisherConfig,
) *StreamPublisher {
	// Если IPPool не задан, используем дефолтные значения
	if len(config.IPPool) == 0 {
		config.IPPool = []string{
			"192.168.1.100", "192.168.1.101", "192.168.1.102",
			"192.168.1.103", "192.168.1.104", "192.168.1.105",
		}
	}

	// Определяем режим по умолчанию
	if config.Mode == "" {
		if config.MessagesPerSecond > 0 {
			config.Mode = "steady"
		} else {
			config.Mode = "batch"
		}
	}

	return &StreamPublisher{
		logger: logger,
		redis:  redisStorage,
		config: config,
	}
}

// Start запускает генерацию тестовых данных
func (s *StreamPublisher) Start(ctx context.Context) error {
	switch s.config.Mode {
	case "steady":
		return s.startSteadyMode(ctx)
	case "batch":
		return s.startBatchMode(ctx)
	default:
		return fmt.Errorf("unknown mode: %s", s.config.Mode)
	}
}

// startSteadyMode запускает режим с постоянной скоростью (одно сообщение в секунду)
func (s *StreamPublisher) startSteadyMode(ctx context.Context) error {
	s.logger.Info("Starting stream publisher in steady mode",
		"stream", s.config.StreamName,
		"rate", s.config.MessagesPerSecond,
		"unit", "msg/s")

	// Интервал между сообщениями
	interval := time.Second / time.Duration(s.config.MessagesPerSecond)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	messageCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stream publisher stopping",
				"mode", "steady",
				"total_messages", messageCount)
			return nil
		case <-ticker.C:
			// Отправляем одно сообщение
			if err := s.sendSingleMessage(ctx, messageCount); err != nil {
				s.logger.Error("Failed to send message", "error", err)
				continue
			}

			messageCount++

			// Логируем прогресс каждые 10 сообщений
			if messageCount%10 == 0 {
				elapsed := time.Since(startTime).Seconds()
				rate := float64(messageCount) / elapsed
				s.logger.Info("Progress",
					"mode", "steady",
					"sent", messageCount,
					"elapsed", fmt.Sprintf("%.1fs", elapsed),
					"actual_rate", fmt.Sprintf("%.1f msg/s", rate))
			}
		}
	}
}

// startBatchMode запускает режим пачками (оригинальный режим)
func (s *StreamPublisher) startBatchMode(ctx context.Context) error {
	s.logger.Info("Starting stream publisher in batch mode",
		"stream", s.config.StreamName,
		"rate", s.config.PublishRate,
		"interval", s.config.BurstInterval)

	ticker := time.NewTicker(s.config.BurstInterval)
	defer ticker.Stop()

	messageCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stream publisher stopping",
				"mode", "batch",
				"total_messages", messageCount)
			return nil
		case <-ticker.C:
			// Отправляем пачку сообщений
			batchCount, err := s.sendBatch(ctx, messageCount)
			if err != nil {
				s.logger.Error("Failed to send batch", "error", err)
				continue
			}

			messageCount += batchCount

			// Логируем статистику
			elapsed := time.Since(startTime).Seconds()
			rate := float64(messageCount) / elapsed
			s.logger.Info("Batch sent",
				"mode", "batch",
				"messages", batchCount,
				"total", messageCount,
				"elapsed", fmt.Sprintf("%.1fs", elapsed),
				"rate", fmt.Sprintf("%.1f msg/s", rate))
		}
	}
}

// sendSingleMessage отправляет одно тестовое сообщение
func (s *StreamPublisher) sendSingleMessage(ctx context.Context, sequence int) error {
	event := s.generateDoorOpenEvent(sequence)

	// Сериализуем событие в JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Отправляем в Redis Stream
	err = s.redis.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.config.StreamName,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	}).Err()

	if err != nil {
		return fmt.Errorf("send to redis: %w", err)
	}

	return nil
}

// sendBatch отправляет пачку тестовых сообщений
func (s *StreamPublisher) sendBatch(ctx context.Context, baseCount int) (int, error) {
	batchCount := 0

	for i := 0; i < s.config.PublishRate; i++ {
		event := s.generateDoorOpenEvent(baseCount + i)

		// Сериализуем событие в JSON
		payload, err := json.Marshal(event)
		if err != nil {
			s.logger.Error("Failed to marshal event", "error", err)
			continue
		}

		// Отправляем в Redis Stream
		err = s.redis.Client.XAdd(ctx, &redis.XAddArgs{
			Stream: s.config.StreamName,
			Values: map[string]interface{}{
				"payload": string(payload),
			},
		}).Err()

		if err != nil {
			s.logger.Error("Failed to send to Redis stream", "error", err)
			continue
		}

		batchCount++
	}

	return batchCount, nil
}

// generateDoorOpenEvent генерирует тестовое событие открытия двери
func (s *StreamPublisher) generateDoorOpenEvent(sequence int) DoorOpenEvent {
	now := time.Now()

	// Случайный IP из пула
	ip := s.config.IPPool[rand.Intn(len(s.config.IPPool))]

	// Случайный тип события (1-10)
	eventType := rand.Intn(10) + 1

	// Случайная дверь (0-1)
	door := rand.Intn(2)

	// Случайный sub_id (иногда nil)
	var subID *int64
	if rand.Float32() > 0.3 { // 70% случаев с sub_id
		subIDValue := int64(rand.Intn(1000) + 1)
		subID = &subIDValue
	}

	// Различные детали в зависимости от типа события
	var detail string
	switch eventType {
	case 1:
		detail = "Open by code"
	case 2:
		detail = "Open by RFID"
	case 3:
		detail = "Open by button"
	case 4:
		detail = "Open by mobile"
	case 5:
		detail = "Open by call"
	case 6:
		detail = "Emergency open"
	case 7:
		detail = "Schedule open"
	case 8:
		detail = "Open by key"
	case 9:
		detail = "Auto open"
	case 10:
		detail = "Manual open"
	default:
		detail = "Unknown open"
	}

	// Добавляем немного вариативности во времени
	eventTime := now.Add(-time.Duration(rand.Intn(60)) * time.Second)

	return DoorOpenEvent{
		Date:      eventTime.Unix(),
		IP:        ip,
		SubID:     subID,
		EventType: eventType,
		Door:      door,
		Detail:    detail,
		Timestamp: eventTime.Unix(),
	}
}
