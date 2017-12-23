package main

import (
	"net/http"

	"github.com/ilius/gragen/examples/customtype"
	"github.com/julienschmidt/httprouter"
)

func main() {
	var server customtype.CustomtypeServer = newServer()
	client := customtype.NewCustomtypeClientFromServer(server)
	router := httprouter.New()
	customtype.RegisterRestHandlers(client, router)
	err := http.ListenAndServe(":5000", router)
	if err != nil {
		panic(err)
	}
}
