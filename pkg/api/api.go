package api

import "net/http"

func Init() {

	http.HandleFunc("/api/nextdate", nextDayHandler)

	http.HandleFunc("/api/task", auth(taskHandler))

	http.HandleFunc("/api/tasks", auth(tasksHandler))

	http.HandleFunc("/api/task/done", auth(doneTaskHandler))

	http.HandleFunc("/api/signin", authHandler)

}
