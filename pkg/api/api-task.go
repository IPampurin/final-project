package api

import "net/http"

// распределяет запросы эндпойнта "/api/task" по типу
func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)

	case http.MethodGet:
		oneTasksHandler(w, r)

	case http.MethodPut:
		updateTaskHandler(w, r)

	case http.MethodDelete:
		deleteTaskHandler(w, r)
	}

}
