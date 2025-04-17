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

func addTaskHandler(w http.ResponseWriter, r *http.Request) {

	var task *db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		message := "не указан заголовок задачи"
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	err = checkDate(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.ID = strconv.Itoa(int(id))

	_, err = json.Marshal(task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(task.ID))

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
	if task.Repeat != "" {
		next, err = NextDate(now.Format(DateOnlyApi), task.Date, task.Repeat)
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

	return nil
}
