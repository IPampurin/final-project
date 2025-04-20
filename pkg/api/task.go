package api

import "net/http"

func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)

	case http.MethodGet:
		oneTasksHandler(w, r)

	case http.MethodPut:
		updateTaskHandler(w, r)
		//case http.MethodDelete:
		//default:
	}

}
