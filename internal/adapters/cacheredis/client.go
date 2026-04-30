package cacheredis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"restaurants-e2/internal/config"
)

// NewClient construye y valida un *redis.Client listo para usar.
// El caller es responsable de cerrarlo con defer client.Close().
func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     10,
		MinIdleConns: 2,
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
