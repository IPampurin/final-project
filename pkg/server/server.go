package server

import (
	"fmt"
	"go1f/pkg/api"
	"net/http"
	"os"
)

func Run() error {

	port, ok := os.LookupEnv("TODO_PORT")
	if !ok {
		port = "7540"
	}

	api.Init()

	http.Handle("/", http.FileServer(http.Dir("web")))

	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
