package api

import (
	"fmt"
	"net/http"

	"github.com/IPampurin/final-project/pkg/db"
)

// AnswerTaskDone структура для вывода результата или ошибки в формате по ТЗ
type AnswerTaskDone struct {
	Error string `json:"error,omitempty"`
}

// doneTaskHandler служит для удаления выполненной задачи
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerTaskDone

	id := r.FormValue("id")

	if id == "" {
		ans.Error = "Не указан идентификатор задачи"
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		ans.Error = fmt.Sprintf("при получении задачи с id = %v возникла ошибка: %v", id, err.Error())
		WriterJSON(w, http.StatusInternalServerError, ans)
		return
	}

	if task.Repeat == "" {

		err = db.DeleteTask(id)
		if err != nil {
			ans.Error = fmt.Sprintf("при удалении задачи с id = %v возникла ошибка: %v", id, err.Error())
			WriterJSON(w, http.StatusInternalServerError, ans)
			return
		}
	}

	if task.Repeat != "" {

		newDate, err := NextDate("", task.Date, task.Repeat)
		if err != nil {
			ans.Error = fmt.Sprintf("при получении следующей даты для удаляемой задачи с id = %v возникла ошибка: %v", id, err.Error())
			WriterJSON(w, http.StatusBadRequest, ans)
			return
		}

		task.Date = newDate

		err = db.UpdateTask(task)
		if err != nil {
			ans.Error = fmt.Sprintf("при обновлении следующей даты для удаляемой задачи с id = %v возникла ошибка: %v", id, err.Error())
			WriterJSON(w, http.StatusInternalServerError, ans)
			return
		}
	}

	WriterJSON(w, http.StatusOK, ans)
}
