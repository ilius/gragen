package main

import (
	"github.com/ilius/gragen/examples/hello"
	"golang.org/x/net/context"
)

type Server struct{}

func (s *Server) SayHello(ctx context.Context, request *hello.HelloRequest) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		Message: request.Message,
	}, nil
}
