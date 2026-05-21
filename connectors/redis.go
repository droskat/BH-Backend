package connectors

import (
	"context"
	"fmt"

	"github.com/droskat/BH-Backend/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,
		MinIdleConns: 20,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return client, nil
}
