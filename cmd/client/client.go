package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	chat "github.com/lupppig/grpc-cli-chat/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClient struct {
	Username string
	Stream   chat.ChatService_ChatClient
	Done     chan struct{}

	mu             sync.Mutex
	typing         map[string]bool
	isTyping       bool
	lastTypingTime time.Time
	lastInputLen   int
}

func NewChatClient(addr, username string) (*ChatClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := chat.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		return nil, err
	}

	return &ChatClient{
		Username: username,
		Stream:   stream,
		Done:     make(chan struct{}),
		typing:   make(map[string]bool),
	}, nil
}

func (c *ChatClient) JoinChat() error {
	return c.Stream.Send(&chat.ClientEvent{
		Event: &chat.ClientEvent_UserEvent{
			UserEvent: &chat.UserEvent{
				Type:      chat.UserEventType_USER_JOINED,
				Username:  c.Username,
				Timestamp: time.Now().Unix(),
			},
		},
	})
}

func (c *ChatClient) SendMessage(msg string) error {
	return c.Stream.Send(&chat.ClientEvent{
		Event: &chat.ClientEvent_ChatMessage{
			ChatMessage: &chat.ChatMessageEvent{
				Username:  c.Username,
				Message:   msg,
				Timestamp: time.Now().Unix(),
			},
		},
	})
}

func (c *ChatClient) SendTyping(start bool) error {
	t := chat.TypingEventType_TYPING_STOP
	if start {
		t = chat.TypingEventType_TYPING_START
	}
	return c.Stream.Send(&chat.ClientEvent{
		Event: &chat.ClientEvent_TypingEvent{
			TypingEvent: &chat.TypingEvent{
				Username:  c.Username,
				Type:      t,
				Timestamp: time.Now().Unix(),
			},
		},
	})
}

func (c *ChatClient) updateTypingState(typing bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if typing {
		c.lastTypingTime = time.Now()
	}

	if c.isTyping != typing {
		c.isTyping = typing
		go c.SendTyping(typing)
	}
}

func (c *ChatClient) ReceiveEvents(rl *readline.Instance) {
	for {
		evt, err := c.Stream.Recv()
		if err != nil {
			fmt.Println("\nDisconnected from server.")
			close(c.Done)
			return
		}

		switch e := evt.Event.(type) {
		case *chat.ServerEvent_ChatMessage:
			c.printLine(rl, e.ChatMessage.Username, e.ChatMessage.Message)
		case *chat.ServerEvent_UserEvent:
			msg := ""
			switch e.UserEvent.Type {
			case chat.UserEventType_USER_JOINED:
				msg = fmt.Sprintf("* %s joined the chat", e.UserEvent.Username)
			case chat.UserEventType_USER_LEFT:
				msg = fmt.Sprintf("* %s left the chat", e.UserEvent.Username)
			}
			c.printLine(rl, "", msg)
		case *chat.ServerEvent_TypingEvent:
			c.mu.Lock()
			if e.TypingEvent.Username != c.Username {
				c.typing[e.TypingEvent.Username] = e.TypingEvent.Type == chat.TypingEventType_TYPING_START
				c.printTyping(rl)
			}
			c.mu.Unlock()
		}
	}
}

// printLine prints above prompt
func (c *ChatClient) printLine(rl *readline.Instance, user, msg string) {
	if user == c.Username {
		rl.Write([]byte(fmt.Sprintf("\r\033[K[you] %s\n", msg)))
	} else if user != "" {
		rl.Write([]byte(fmt.Sprintf("\r\033[K[%s] %s\n", user, msg)))
	} else {
		rl.Write([]byte(fmt.Sprintf("\r\033[K%s\n", msg)))
	}
	rl.Refresh()
}

// printTyping shows who is typing
func (c *ChatClient) printTyping(rl *readline.Instance) {
	users := []string{}
	for user, t := range c.typing {
		if t {
			users = append(users, user)
		}
	}
	prompt := "> "
	if len(users) > 0 {
		prompt = fmt.Sprintf("> * %s typing... ", strings.Join(users, ", "))
	}
	rl.SetPrompt(prompt)
	rl.Refresh()
}

func main() {
	// Server address
	address := flag.String("address", "localhost:50051", "gRPC server address")
	flag.Parse()

	// Read username
	reader := bufio.NewReader(os.Stdin)
	var username string
	for {
		fmt.Print("Enter your username: ")
		line, _ := reader.ReadString('\n')
		username = strings.TrimSpace(strings.TrimRight(line, "\r\n"))
		if username != "" {
			break
		}
		fmt.Println("Username cannot be empty.")
	}

	client, err := NewChatClient(*address, username)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.JoinChat(); err != nil {
		log.Fatal(err)
	}

	// Readline instance with listener
	config := &readline.Config{
		Prompt:          "> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Listener:        &ReadlineListener{client: client},
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	go client.ReceiveEvents(rl)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				client.mu.Lock()
				if client.isTyping && time.Since(client.lastTypingTime) > 2*time.Second {
					client.isTyping = false
					go client.SendTyping(false)
				}
				client.mu.Unlock()
			case <-client.Done:
				return
			}
		}
	}()

	fmt.Println("You can start chatting now. Type your messages and press Enter.")

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Println("\nExiting chat...")
			client.updateTypingState(false)
			return
		}

		line = strings.TrimSpace(line)

		client.updateTypingState(false)

		if line == "" {
			continue
		}

		// Send message
		if err := client.SendMessage(line); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}

type ReadlineListener struct {
	client *ChatClient
}

func (l *ReadlineListener) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	currentLen := len(line)

	l.client.mu.Lock()
	lastLen := l.client.lastInputLen
	l.client.lastInputLen = currentLen
	l.client.mu.Unlock()

	if currentLen != lastLen && currentLen > 0 {
		l.client.updateTypingState(true)
	} else if currentLen == 0 {
		l.client.updateTypingState(false)
	}

	return line, pos, true
}
