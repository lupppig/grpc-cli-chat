package main

import (
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/lupppig/grpc-cli-chat/cmd/service"
	"github.com/lupppig/grpc-cli-chat/internal"
	chat "github.com/lupppig/grpc-cli-chat/pb"
	"github.com/stackus/dotenv"
	"google.golang.org/grpc"
)

func main() {
	dotenv.Load()
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %d", *port)

	red, err := internal.DB()
	if err != nil {
		log.Fatal(err)
	}
	chatServer := service.NewChatServer(red)
	portStr := strconv.Itoa(*port)
	address := net.JoinHostPort("0.0.0.0", portStr)

	listner, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, chatServer)

	err = grpcServer.Serve(listner)
	if err != nil {
		log.Fatalf("cannot listen to server on: %v", err)
	}

}
