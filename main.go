package main

import (
	"fmt"
	"go1f/pkg/db"
	"go1f/pkg/server"
)

func main() {

	err := db.Init("scheduler.db")
	if err != nil {
		fmt.Printf("ошибка вызова Init: %v\n", err)
		return
	}

	err = server.Run()
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		return
	}

}
