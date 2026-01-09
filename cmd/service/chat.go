package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	red "github.com/lupppig/grpc-cli-chat/internal/redis"
	chat "github.com/lupppig/grpc-cli-chat/pb"
	"google.golang.org/grpc"
)

type ChatServer struct {
	RedisCl red.RedisClient
	Clients map[string]*ClientConn // active clients
	Mutex   sync.Mutex
	chat.UnimplementedChatServiceServer
}

func NewChatServer(red red.RedisClient) *ChatServer {
	return &ChatServer{
		RedisCl: red,
		Clients: make(map[string]*ClientConn),
	}
}

type ClientConn struct {
	Stream      grpc.BidiStreamingServer[chat.ClientEvent, chat.ServerEvent]
	Username    string
	IsTyping    bool
	RateLimiter *RateLimiter
}

func (s *ChatServer) Chat(stream grpc.BidiStreamingServer[chat.ClientEvent, chat.ServerEvent]) error {
	firstEvt, err := stream.Recv()
	if err != nil {
		return err
	}

	var username string
	if ue, ok := firstEvt.Event.(*chat.ClientEvent_UserEvent); ok && ue.UserEvent.Type == chat.UserEventType_USER_JOINED {
		username = ue.UserEvent.Username
	} else {
		return fmt.Errorf("first event must be USER_JOINED with username")
	}

	client := &ClientConn{
		Stream:      stream,
		Username:    username,
		RateLimiter: NewRateLimiter(),
	}

	s.Mutex.Lock()
	s.Clients[username] = client
	s.Mutex.Unlock()

	s.broadcastUserEvent(chat.UserEventType_USER_JOINED, username)

	// history, _ := s.RedisCl.Get(context.Background(), "chat:messages")
	// for _, msg := range history {
	// 	stream.Send(&chat.ServerEvent{
	// 		Event: &chat.ServerEvent_ChatMessage{ChatMessage: msg},
	// 	})
	// }

	go s.handleIncomingEvents(client)

	<-stream.Context().Done()

	s.removeClient(username)
	return nil
}
func (s *ChatServer) HealthCheck(ctx context.Context, req *chat.HealthCheckRequest) (*chat.HealthCheckResponse, error) {
	return nil, nil
}

func (s *ChatServer) handleIncomingEvents(client *ClientConn) {
	for {
		evt, err := client.Stream.Recv()
		if err != nil {
			log.Printf("error: %v", err.Error())
			break
		}

		switch e := evt.Event.(type) {
		case *chat.ClientEvent_ChatMessage:
			if !client.RateLimiter.Allow() {
				log.Printf("rate limit hit")
				continue
			}

			msg := &chat.ChatMessageEvent{
				Message:   e.ChatMessage.Message,
				Username:  client.Username,
				Timestamp: time.Now().Unix(),
			}

			s.RedisCl.Add(context.Background(), "chat:messages", msg)

			s.broadcastChatMessage(msg)

		case *chat.ClientEvent_TypingEvent:
			client.IsTyping = e.TypingEvent.Type == chat.TypingEventType_TYPING_START

			s.broadcastTypingEvent(client.Username, e.TypingEvent.Type)
		}
	}
}

func (s *ChatServer) broadcastChatMessage(msg *chat.ChatMessageEvent) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for _, c := range s.Clients {
		c.Stream.Send(&chat.ServerEvent{
			Event: &chat.ServerEvent_ChatMessage{ChatMessage: msg},
		})
	}
}

func (s *ChatServer) broadcastTypingEvent(username string, tType chat.TypingEventType) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for _, c := range s.Clients {
		if c.Username == username {
			continue
		}
		c.Stream.Send(&chat.ServerEvent{
			Event: &chat.ServerEvent_TypingEvent{
				TypingEvent: &chat.TypingEvent{
					Username:  username,
					Type:      tType,
					Timestamp: time.Now().Unix(),
				},
			},
		})
	}
}

func (s *ChatServer) broadcastUserEvent(eventType chat.UserEventType, username string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	evt := &chat.ServerEvent{
		Event: &chat.ServerEvent_UserEvent{
			UserEvent: &chat.UserEvent{
				Type:      eventType,
				Username:  username,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	for _, c := range s.Clients {
		c.Stream.Send(evt)
	}
}

func (s *ChatServer) removeClient(username string) {
	s.Mutex.Lock()
	delete(s.Clients, username)
	s.Mutex.Unlock()

	s.broadcastUserEvent(chat.UserEventType_USER_LEFT, username)
}
