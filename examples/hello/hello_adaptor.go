package hello

import (
	"github.com/ilius/ripo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
)

func GontextFromRest(req ripo.Request) (context.Context, error) {
	headerMap := map[string]string{}
	for _, key := range req.HeaderKeys() {
		value := req.GetHeader(key)
		headerMap[key] = value
	}
	md := metadata.New(headerMap)
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func NewRestHandler_SayHello(client HelloClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &HelloRequest{}
		{
			value, err := req.GetString("message")
			if err != nil {
				return nil, err
			}
			grpcReq.Message = *value
		}
		log.Println("grpcReq =", grpcReq)
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.SayHello(ctx, grpcReq)
		if err != nil {
			return nil, err
		}
		return &ripo.Response{Data: grpcRes}, nil
	}
}

func RegisterRestHandlers(client HelloClient) {
	http.HandleFunc("SayHello", ripo.TranslateHandler(NewRestHandler_SayHello(client)))
}

type helloClientByServerImp struct {
	srv HelloServer
}

func (c *helloClientByServerImp) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	return c.srv.SayHello(ctx, in)
}

func NewHelloClientFromServer(srv HelloServer) HelloClient {
	return &helloClientByServerImp{srv: srv}
}
