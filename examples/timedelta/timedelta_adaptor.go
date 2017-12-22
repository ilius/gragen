package timedelta

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/ilius/ripo"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
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

func NewRest_TimeDelta(client TimedeltaClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &TimeDeltaRequest{}
		{ // now:
			value, err := req.GetTime("now")
			if err != nil {
				return nil, err
			}
			if value != nil {
				valueProto, err := ptypes.TimestampProto(*value)
				if err != nil {
					return nil, ripo.NewError(ripo.Internal, "", err)
				}
				grpcReq.Now = valueProto
			}
		}
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.TimeDelta(ctx, grpcReq)
		if err != nil {
			return nil, getRestError(err)
		}
		return &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil
	}
}

func NewRest_Sleep(client TimedeltaClient) ripo.Handler {
	return func(req ripo.Request) (*ripo.Response, error) {
		grpcReq := &SleepRequest{}
		{ // duration:
			value, err := req.GetString("duration")
			if err != nil {
				return nil, err
			}
			if value != nil {
				valueGo, err := time.ParseDuration(*value)
				if err != nil {
					return nil, ripo.NewError(ripo.InvalidArgument, "invalid 'duration', must be a valid duration string", err)
				}
				valueProto := ptypes.DurationProto(valueGo)
				grpcReq.Duration = valueProto
			}
		}
		ctx, err := GontextFromRest(req)
		if err != nil {
			return nil, err
		}
		grpcRes, err := client.Sleep(ctx, grpcReq)
		if err != nil {
			return nil, getRestError(err)
		}
		return &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil
	}
}

func RegisterRestHandlers(client TimedeltaClient, router *httprouter.Router) {
	handleRest(router, "GET", "/timedelta", NewRest_TimeDelta(client))
	handleRest(router, "GET", "/sleep", NewRest_Sleep(client))
}

type timedeltaClientByServerImp struct {
	srv TimedeltaServer
}

func (c *timedeltaClientByServerImp) TimeDelta(ctx context.Context, in *TimeDeltaRequest, opts ...grpc.CallOption) (*TimeDeltaResponse, error) {
	return c.srv.TimeDelta(ctx, in)
}

func (c *timedeltaClientByServerImp) Sleep(ctx context.Context, in *SleepRequest, opts ...grpc.CallOption) (*SleepResponse, error) {
	return c.srv.Sleep(ctx, in)
}

func NewTimedeltaClientFromServer(srv TimedeltaServer) TimedeltaClient {
	return &timedeltaClientByServerImp{srv: srv}
}
