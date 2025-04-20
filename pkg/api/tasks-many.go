package api

import (
	"fmt"
	"net/http"

	"github.com/IPampurin/final-project/pkg/db"
)

const DateFromSearch = "02.01.2006"

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// максимальное число строк к выводу
const RowsLimit = 50

// AnswerTasksErr структура для вывода ошибки в формате по ТЗ
type AnswerTasksErr struct {
	Error string `json:"error,omitempty"`
}

// tasksHandler выводит список всех задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerTasksErr
	var result TasksResp
	var tasks []*db.Task

	search := r.FormValue("search")

	tasks, err := db.Tasks(RowsLimit, search)
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
