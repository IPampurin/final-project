package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/IPampurin/final-project/pkg/db"
)

// AnswerDeleteTask структура для вывода результата или ошибки в формате по ТЗ
type AnswerDeleteTask struct {
	Error string `json:"error,omitempty"`
}

// deleteTaskHandler служит для удаления ненужной задачи (метод Delete)
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerDeleteTask

	id := r.FormValue("id")

	if id == "" {
		ans.Error = "не указан идентификатор задачи"
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	_, err := strconv.Atoi(id)
	if err != nil {
		ans.Error = "идентификатор задачи должен быть числом"
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		ans.Error = fmt.Sprintf("при удалении задачи с id = %v возникла ошибка: %v", id, err.Error())
		WriterJSON(w, http.StatusInternalServerError, ans)
		return
	}

	WriterJSON(w, http.StatusOK, ans)
}
