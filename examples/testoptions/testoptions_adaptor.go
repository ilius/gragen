package testoptions

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/ilius/ripo"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"reflect"
)

func init() {
	ripo.SetDefaultParamSources(
		ripo.FromBody,
		ripo.FromForm,
		// ripo.FromContext,
		ripo.FromEmpty,
	)
}

var restJsonMarshaler = jsonpb.Marshaler{}

type restResponseWrapper struct {
	grpcRes interface{}
}

func (rw *restResponseWrapper) MarshalJSON() ([]byte, error) {
	protoMsg, ok := rw.grpcRes.(proto.Message)
	if !ok {
		return json.Marshal(rw.grpcRes)
	}
	buf := bytes.NewBuffer(nil)
	err := restJsonMarshaler.Marshal(buf, protoMsg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const restHeaderToContextPrefix = "rest-header-"

func GontextFromRest(req ripo.Request) (context.Context, error) {
	headerMap := map[string]string{}
	for _, key := range req.HeaderKeys() {
		value := req.Header(key)
		headerMap[restHeaderToContextPrefix+key] = value
	}
	md := metadata.New(headerMap)
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

// getRestError: convert grpc error to rest
func getRestError(err error) ripo.RPCError {
	st, ok := status.FromError(err)
	if !ok {
		return ripo.NewError(ripo.Unknown, "", err)
	}
	return ripo.NewError(ripo.Code(int32(st.Code())), st.Message(), err)
}

func handleRest(router *httprouter.Router, method string, path string, handler ripo.Handler) {
	handlerFunc := ripo.TranslateHandler(handler)
	router.Handle(
		method,
		path,
		func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			r.ParseForm()
			for _, p := range params {
				r.Form.Add(p.Key, p.Value)
			}
			handlerFunc(w, r)
		},
	)
}

func NewRest_AskName(client TestoptionsClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &AskNameRequest{}
		{ // message:
			value, err := req.GetString("message")
			if err != nil {
				return nil, err
			}
			if value != nil && *value != "" {
				grpcReq.Message = *value
			}
		}
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.AskName(ctx, grpcReq)
		if err != nil {
			return nil, getRestError(err)
		}
		return &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil
	}
}

func NewRest_SayHello(client TestoptionsClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &HelloRequest{}
		{ // message:
			value, err := req.GetString("message")
			if err != nil {
				return nil, err
			}
			if value != nil && *value != "" {
				grpcReq.Message = *value
			}
		}
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.SayHello(ctx, grpcReq)
		if err != nil {
			return nil, getRestError(err)
		}
		return &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil
	}
}

func NewRest_PostCard(client TestoptionsClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &PostCardRequest{}
		{ // card:
			var valueNil *Card
			value, err := req.GetObject("card", reflect.TypeOf(valueNil))
			if err != nil {
				return nil, err
			}
			if value != nil {
				grpcReq.Card = value.(*Card)
			}
		}
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.PostCard(ctx, grpcReq)
		if err != nil {
			return nil, getRestError(err)
		}
		return &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil
	}
}

func RegisterRestHandlers(client TestoptionsClient, router *httprouter.Router) {
	handleRest(router, "GET", "/askname", NewRest_AskName(client))
	handleRest(router, "GET", "/sayhello", NewRest_SayHello(client))
	handleRest(router, "POST", "/postcard", NewRest_PostCard(client))
}

type testoptionsClientByServerImp struct {
	srv TestoptionsServer
}

func (c *testoptionsClientByServerImp) AskName(ctx context.Context, in *AskNameRequest, opts ...grpc.CallOption) (*AskNameResponse, error) {
	return c.srv.AskName(ctx, in)
}

func (c *testoptionsClientByServerImp) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	return c.srv.SayHello(ctx, in)
}

func (c *testoptionsClientByServerImp) PostCard(ctx context.Context, in *PostCardRequest, opts ...grpc.CallOption) (*PostCardResponse, error) {
	return c.srv.PostCard(ctx, in)
}

func NewTestoptionsClientFromServer(srv TestoptionsServer) TestoptionsClient {
	return &testoptionsClientByServerImp{srv: srv}
}
