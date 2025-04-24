package api

import (
	"fmt"
	"net/http"

	"github.com/IPampurin/final-project/pkg/db"
)

// oneTasksHandler выводит одну задачу по её id при GET-запросе
func oneTasksHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerTasksErr

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

	WriterJSON(w, http.StatusOK, task)
}
