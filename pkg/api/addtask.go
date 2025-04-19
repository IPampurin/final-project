package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/IPampurin/final-project/pkg/db"
)

type AnswerAddTask struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {

	var task db.Task
	var buf bytes.Buffer
	var ans AnswerAddTask

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

	id, err := db.AddTask(&task)
	if err != nil {
		ans.Error = err.Error()
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	ans.ID = strconv.Itoa(int(id))

	WriterJSON(w, http.StatusCreated, ans)
}

func checkDate(task *db.Task) error {

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(DateOnlyApi)
	}
	t, err := time.Parse(DateOnlyApi, task.Date)
	if err != nil {
		return fmt.Errorf("ошибка парсинга даты задачи (поле 'task.Date'): %v", err)
	}

	var next string

	if len(task.Repeat) != 0 {
		next, err = NextDate("", task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if now.After(t) {
		if task.Repeat == "" {
			task.Date = now.Format(DateOnlyApi)
		} else {
			task.Date = next
		}
	}

	if now.Format(DateOnlyApi) == t.Format(DateOnlyApi) {
		task.Date = now.Format(DateOnlyApi)
	}

	return nil
}

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
