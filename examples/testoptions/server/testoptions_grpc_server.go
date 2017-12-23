package main

import (
	"fmt"
	"time"

	"github.com/ilius/gragen/examples/testoptions"
	"golang.org/x/net/context"
)

type Server struct{}

func (s *Server) AskName(ctx context.Context, request *testoptions.AskNameRequest) (*testoptions.AskNameResponse, error) {
	return &testoptions.AskNameResponse{
		Name: "John Smith",
	}, nil
}

func (s *Server) SayHello(ctx context.Context, request *testoptions.HelloRequest) (*testoptions.HelloResponse, error) {
	return &testoptions.HelloResponse{
		Message: request.Message,
	}, nil
}

func (s *Server) PostCard(ctx context.Context, request *testoptions.PostCardRequest) (*testoptions.PostCardResponse, error) {
	return &testoptions.PostCardResponse{
		RefId: fmt.Sprintf("%v%v", time.Now().Unix(), request.Card.Message),
	}, nil
}
