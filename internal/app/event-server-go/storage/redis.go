package storage

import (
	"context"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type RedisStorage struct {
	logger *slog.Logger
	Client *redis.Client
}

func NewRedisStorage(logger *slog.Logger, redisConfig *config.RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis",
		"host", redisConfig.Host,
		"port", redisConfig.Port,
		"db", redisConfig.DB)

	return &RedisStorage{
		logger: logger,
		Client: client,
	}, nil
}

// Close
func (s *RedisStorage) Close() {
	if s.Client != nil {
		if err := s.Client.Close(); err != nil {
			s.logger.Error("Error closing Redis connection", "error", err)
		} else {
			s.logger.Info("Successfully closed connection to Redis")
		}
	}
}

// Ping check available connection to Redis
func (s *RedisStorage) Ping(ctx context.Context) error {
	return s.Client.Ping(ctx).Err()
}
