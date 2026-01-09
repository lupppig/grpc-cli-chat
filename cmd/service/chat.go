package service

import red "github.com/lupppig/grpc-cli-chat/internal/redis"

type ChatServer struct {
	RedisCl *red.Redis
}


