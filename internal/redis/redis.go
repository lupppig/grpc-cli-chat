package red

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	red *redis.Client
}

func ConnectRedis(ctx context.Context, redisURL string) (RedisClient, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis server: %w", err)
	}

	cl := redis.NewClient(opts)

	if stat := cl.Ping(ctx); stat.Err() != nil {
		return nil, fmt.Errorf("failed to ping redis server: %w", stat.Err())
	}
	return &Redis{red: cl}, nil
}

func (r *Redis) Red() *redis.Client {
	return r.red
}
