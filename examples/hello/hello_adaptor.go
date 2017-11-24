package hello

import (
	"github.com/ilius/ripo"
	"log"
	"net/http"
)

func SayHelloHandler(req ripo.Request) (*ripo.Response, error) {
	grpcReq := &HelloRequest{}
	{
		message, err := req.GetString("message")
		if err != nil {
			return nil, err
		}
		grpcReq.Message = *message
	}
	log.Println("grpcReq =", grpcReq)
	return nil, nil // FIXME
}

func main() {
	http.HandleFunc("SayHello", ripo.TranslateHandler(SayHelloHandler))
}
