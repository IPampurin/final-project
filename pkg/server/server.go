package server

import (
	"fmt"
	"net/http"
	"os"
)

func Run() error {

	// проверяем существует ли переменная окружения TODO_PORT
	// если TODO_PORT не существует или "", значение ok = false
	port, ok := os.LookupEnv("TODO_PORT")
	if !ok {
		port = "7540"
	}

	http.Handle("/", http.FileServer(http.Dir("web")))

	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
