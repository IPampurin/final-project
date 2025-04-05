package main

import (
	"fmt"
	"go1f/pkg/server"
	"os"
)

func main() {

	os.Setenv("TODO_PORT", "8000")

	err := server.Run()
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
