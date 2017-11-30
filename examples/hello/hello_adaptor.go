package hello

import (
	"github.com/ilius/ripo"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func NewRestHandler_SayHello(client HelloClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &HelloRequest{}
		{
			value, err := req.GetString("Message")
			if err != nil {
				return nil, err
			}
			grpcReq.Message = *value
		}
		log.Println("grpcReq =", grpcReq)
		grpcRes, err := client.SayHello(context.Background(), grpcReq)
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

