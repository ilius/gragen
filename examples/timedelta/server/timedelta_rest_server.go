package main

import (
	"net/http"

	"github.com/ilius/gragen/examples/timedelta"
	"github.com/julienschmidt/httprouter"
)

func main() {
	var server timedelta.TimedeltaServer = &Server{}
	client := timedelta.NewTimedeltaClientFromServer(server)
	router := httprouter.New()
	timedelta.RegisterRestHandlers(client, router)
	err := http.ListenAndServe(":5000", router)
	if err != nil {
		panic(err)
	}
}
