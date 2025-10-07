package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

//sample redis stream worker

var wg = &sync.WaitGroup{}

// Структура для парсинга событий из PHP
type DoorOpenEvent struct {
	Date      int64  `json:"date"`
	IP        string `json:"ip"`
	SubID     *int64 `json:"sub_id"` // может быть null
	EventType int    `json:"event_type"`
	Door      int    `json:"door"`
	Detail    string `json:"detail"`
	Timestamp int64  `json:"timestamp"`
}

// Структура для сообщения в Stream
type StreamMessage struct {
	Payload string `json:"payload"`
}

func Start(ctx context.Context, cancel context.CancelFunc) {
	client := getRedis()
	defer client.Close()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// init consumer group
	initConsumerGroup(ctx, client)

	// start workers
	workersCount, _ := strconv.Atoi(getEnv("WORKERS_COUNT", "3"))
	for i := 1; i <= workersCount; i++ {
		wg.Add(1)
		go worker(ctx, client, fmt.Sprintf("worker_%d", i))
	}

	log.Printf("Started %d workers\n", workersCount)
	wg.Wait()
	log.Println("Door Events Consumer stopped")

}

func getRedis() *redis.Client {
	var (
		host     = getEnv("REDIS_HOST", "172.28.0.4")
		port     = getEnv("REDIS_PORT", "6379")
		password = getEnv("REDIS_PASSWORD", "qqq")
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("✅ Successfully connected to Redis")
	return client
}

func getEnv(envName, valueDefault string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return valueDefault
}

func initConsumerGroup(ctx context.Context, client *redis.Client) {
	stream := "door_open_events_stream"
	group := "door_open_events_workers"

	// Создаем consumer group если не существует
	err := client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil {
		log.Printf("ℹ️ Consumer group already exists: %s\n", group)
	} else {
		log.Printf("✅ Consumer group created: %s\n", group)
	}
}

func worker(ctx context.Context, client *redis.Client, workerName string) {
	defer wg.Done()

	stream := "door_open_events_stream"
	group := "door_events_processor"
	batchSize, _ := strconv.Atoi(getEnv("BATCH_SIZE", "5"))
	blockTime, _ := strconv.Atoi(getEnv("BLOCK_TIME", "5000"))

	log.Printf("👷 Worker %s started\n", workerName)

	processedCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 Worker %s stopping (processed: %d)\n", workerName, processedCount)
			return
		default:
			// Читаем сообщения из stream
			result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: workerName,
				Streams:  []string{stream, ">"},
				Count:    int64(batchSize),
				Block:    time.Duration(blockTime) * time.Millisecond,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// Таймаут - нет новых сообщений
					continue
				}
				log.Printf("❌ Worker %s error reading: %v\n", workerName, err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем полученные сообщения
			for _, stream := range result {
				for _, message := range stream.Messages {
					processedCount++

					if processDoorEvent(message, workerName) {
						// Подтверждаем обработку
						err := client.XAck(ctx, stream.Stream, group, message.ID).Err()
						if err != nil {
							log.Printf("❌ Worker %s failed to ack message %s: %v\n",
								workerName, message.ID, err)
						} else {
							log.Printf("✅ Worker %s acked message %s\n",
								workerName, message.ID)
						}
					} else {
						log.Printf("⚠️ Worker %s processing failed, will retry: %s\n",
							workerName, message.ID)
					}

					// Логируем прогресс каждые 10 сообщений
					if processedCount%10 == 0 {
						log.Printf("📊 Worker %s processed %d messages\n",
							workerName, processedCount)
					}
				}
			}
		}
	}
}

func processDoorEvent(message redis.XMessage, workerName string) bool {
	// Извлекаем payload
	payload, ok := message.Values["payload"].(string)
	if !ok {
		log.Printf("❌ Worker %s: invalid payload format in message %s\n",
			workerName, message.ID)
		return false
	}

	// Парсим JSON с данными события
	var event DoorOpenEvent
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		log.Printf("❌ Worker %s: failed to unmarshal event %s: %v\n",
			workerName, message.ID, err)
		return false
	}

	// Здесь ваша основная логика обработки события
	// Замените эту часть на реальную запись в БД или другую обработку

	log.Printf("🚪 Worker %s processing event: IP=%s, EventType=%d, Door=%d, Detail=%s\n",
		workerName, event.IP, event.EventType, event.Door, event.Detail)

	// Имитируем обработку (замените на реальную логику)
	success := saveToDatabase(event)
	if !success {
		log.Printf("❌ Worker %s: failed to save event to database: %s\n",
			workerName, message.ID)
		return false
	}

	return true
}

// Функция для сохранения в базу данных (замените на вашу реализацию)
func saveToDatabase(event DoorOpenEvent) bool {
	// TODO: Реализуйте запись в вашу базу данных
	// Это замена оригинальному методу addDoorOpenData из PHP

	// Пример логики:
	// db.Exec("INSERT INTO door_events (...) VALUES (?, ?, ?, ?, ?, ?)",
	//     event.Date, event.IP, event.SubID, event.EventType, event.Door, event.Detail)

	// Имитируем работу
	time.Sleep(50 * time.Millisecond)

	// В 95% случаев успешно, в 5% - ошибка (для тестирования)
	if time.Now().UnixNano()%100 < 5 {
		return false
	}

	return true
}

// Worker для обработки pending сообщений (на случай сбоев)
func pendingWorker(ctx context.Context, client *redis.Client) {
	stream := "door_open_events_stream"
	group := "door_events_processor"
	workerName := "pending_recovery"

	log.Printf("🔧 Pending worker %s started\n", workerName)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Обрабатываем pending сообщения
			messages, _, err := client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: workerName,
				MinIdle:  30 * time.Second, // сообщения старше 30 секунд
				Count:    10,
				Start:    "0-0",
			}).Result()

			if err != nil {
				log.Printf("❌ Pending worker error: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(messages) > 0 {
				log.Printf("🔧 Pending worker processing %d messages\n", len(messages))

				for _, msg := range messages {
					if processDoorEvent(msg, workerName) {
						client.XAck(ctx, stream, group, msg.ID)
					}
				}
			}

			time.Sleep(10 * time.Second)
		}
	}
}
