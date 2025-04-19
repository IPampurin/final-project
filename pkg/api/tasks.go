package api

import (
	"fmt"
	"net/http"

	"github.com/IPampurin/final-project/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

const RowsLimit = 50 // максимальное число строк к выводу

type AnswerTasks struct {
	Error string `json:"error,omitempty"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerTasks
	var result TasksResp

	tasks, err := db.Tasks(RowsLimit)
	if err != nil {
		ans.Error = fmt.Sprintf("ошибка при получении списка задач %v", err.Error())
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	if tasks == nil {
		result.Tasks = []*db.Task{}
	} else {
		result.Tasks = tasks
	}

	WriterJSON(w, http.StatusOK, result)
}
