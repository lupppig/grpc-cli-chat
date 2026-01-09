package red

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Red() *redis.Client
	Add(ctx context.Context, key string, message string) error
	Get(ctx context.Context, key string) (interface{}, error)
}

func (r *Redis) Add(ctx context.Context, key string, message string) error {
	fmt.Println("add data to redis lmaooooo")
	return nil
}

// get user messages from redis
func (r *Redis) Get(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}
