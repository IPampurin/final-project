package main

import (
	"fmt"
	"go1f/pkg/server"
)

func main() {
	err := server.Run()
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
