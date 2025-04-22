package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriterJSON служит для формирования и направления ответа в формате JSON
func WriterJSON(w http.ResponseWriter, status int, data interface{}) {

	emergencyError := `{"fatal error":"%q"}`

	js, err := json.Marshal(data)
	if err != nil {
		status = http.StatusInternalServerError
		js = []byte(fmt.Sprintf(emergencyError, err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}
