package main

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/ilius/gragen/examples/timedelta"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct{}

func (s *Server) TimeDelta(ctx context.Context, request *timedelta.TimeDeltaRequest) (*timedelta.TimeDeltaResponse, error) {
	reqTime, err := ptypes.Timestamp(request.Now)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in parsing timestamp")
	}
	now := time.Now()
	nowProto, err := ptypes.TimestampProto(now)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in generating timestamp")
	}

	delta := now.Sub(reqTime)
	deltaProto := ptypes.DurationProto(delta)
	return &timedelta.TimeDeltaResponse{
		Now:   nowProto,
		Delta: deltaProto,
	}, nil
}
