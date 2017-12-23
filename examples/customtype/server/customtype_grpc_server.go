package main

import (
	"github.com/ilius/gragen/examples/customtype"
	"github.com/ilius/gragen/examples/customtype/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newServer() *Server {
	return &Server{
		userMap: map[string]*types.User{},
	}
}

type Server struct {
	userMap map[string]*types.User
}

func (s *Server) GetUserInfo(ctx context.Context, request *customtype.GetUserInfoRequest) (*customtype.GetUserInfoResponse, error) {
	if request.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing 'userId'")
	}
	info, ok := s.userMap[request.UserId]
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &customtype.GetUserInfoResponse{
		UserInfo: info,
	}, nil
}

func (s *Server) UpdateUserInfo(ctx context.Context, request *customtype.UpdateUserInfoRequest) (*customtype.UpdateUserInfoResponse, error) {
	if request.UserInfo.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing userInfo.id")
	}
	s.userMap[request.UserInfo.Id] = request.UserInfo
	return &customtype.UpdateUserInfoResponse{}, nil
}
