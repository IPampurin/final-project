package db

type Task struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	// и т.д.
}

func AddTask(task *Task) (int64, error) {
	var id int64
	// определите запрос
	query := `INSERT INTO scheduler ...`
	res, err := db.Exec(query /*передайте параметры task.Date, task.Title и т.д.*/)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}
