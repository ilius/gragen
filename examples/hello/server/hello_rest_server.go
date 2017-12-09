package main

import (
	"net/http"

	"github.com/ilius/gragen/examples/hello"
	"github.com/julienschmidt/httprouter"
)

func main() {
	var server hello.HelloServer = &Server{}
	client := hello.NewHelloClientFromServer(server)
	router := httprouter.New()
	hello.RegisterRestHandlers(client, router)
	err := http.ListenAndServe(":5000", router)
	if err != nil {
		panic(err)
	}
}
