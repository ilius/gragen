package main

import (
	"net/http"

	"github.com/ilius/gragen/examples/testoptions"
	"github.com/julienschmidt/httprouter"
)

func main() {
	var server testoptions.TestoptionsServer = &Server{}
	client := testoptions.NewTestoptionsClientFromServer(server)
	router := httprouter.New()
	testoptions.RegisterRestHandlers(client, router)
	err := http.ListenAndServe(":5000", router)
	if err != nil {
		panic(err)
	}
}
