package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IPampurin/final-project/pkg/db"
)

// AnswerUpdateErr структура для вывода ошибки в формате по ТЗ
type AnswerUpdateErr struct {
	Error string `json:"error,omitempty"`
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {

	var task db.Task
	var buf bytes.Buffer
	var ans AnswerUpdateErr

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		ans.Error = fmt.Sprintf("невозможно прочитать тело запроса %v", err.Error())
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		ans.Error = fmt.Sprintf("невозможно десериализовать тело запроса %v", err.Error())
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	if task.Title == "" {
		ans.Error = "не указан заголовок задачи"
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	err = checkDate(&task)
	if err != nil {
		ans.Error = err.Error()
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	err = db.UpdateTask(&task)
	if err != nil {
		ans.Error = err.Error()
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	WriterJSON(w, http.StatusCreated, task)
}
