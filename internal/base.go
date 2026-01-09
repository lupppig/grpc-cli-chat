package internal

import (
	"context"
	"os"
	"time"

	red "github.com/lupppig/grpc-cli-chat/internal/redis"
)

func DB() (red.RedisClient, error) {
	redAddr := os.Getenv("REDIS_ADDR")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := red.ConnectRedis(ctx, redAddr)
	if err != nil {
		return nil, err
	}

	return r, nil
}
