package api

import "net/http"

func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)

	case http.MethodGet:
		tasksHandler(w, r)
		//case http.MethodPut:
		//case http.MethodDelete:
		//default:
	}

}
