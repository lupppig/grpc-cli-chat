package red

import (
	"context"

	chat "github.com/lupppig/grpc-cli-chat/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type RedisClient interface {
	Red() *redis.Client
	Add(ctx context.Context, key string, message *chat.ChatMessageEvent) error
	Get(ctx context.Context, key string) ([]*chat.ChatMessageEvent, error)
}

func (r *Redis) Add(ctx context.Context, key string, message *chat.ChatMessageEvent) error {
	data, _ := proto.Marshal(message)
	return r.Red().LPush(ctx, key, data).Err()
}

func (r *Redis) Get(ctx context.Context, key string) ([]*chat.ChatMessageEvent, error) {
	vals, _ := r.Red().LRange(ctx, key, 0, 1).Result()
	var out []*chat.ChatMessageEvent
	for _, v := range vals {
		msg := &chat.ChatMessageEvent{}
		proto.Unmarshal([]byte(v), msg)
		out = append(out, msg)
	}
	return out, nil
}
