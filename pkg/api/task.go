package api

import "net/http"

func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
		/*
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("тырмандыр"))
		*/ //case http.MethodGet:
		//case http.MethodPut:
		//case http.MethodDelete:
		//default:
	}

}
