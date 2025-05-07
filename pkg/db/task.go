package db

// В данном файле собраны функции AddTask, Tasks, GetTask, UpdateTask, DeleteTask

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DateFromSearch = "02.01.2006"
	DateOnlyDB     = "20060102"
)

// Task структура задания
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в БД
func AddTask(task *Task) (int64, error) {

	dbFile := "scheduler.db"
	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var id int64

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)"
	res, err := db.Exec(query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err == nil {
		id, err = res.LastInsertId()
	}

	return id, err
}

// Tasks выводит все задачи из БД
func Tasks(limit int, search string) ([]*Task, error) {

	dbFile := "scheduler.db"
	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var allTasks []*Task
	var rows *sql.Rows

	switch {

	case search == "":

		rows, err = db.Query("SELECT * FROM scheduler ORDER BY date LIMIT :limit",
			sql.Named("limit", limit))
		if err != nil {
			return []*Task{}, err
		}
		defer rows.Close()

	case search != "":

		searchTime, errFromParse := time.Parse(DateFromSearch, search)

		if errFromParse == nil {
			rows, err = db.Query("SELECT * FROM scheduler WHERE date = :search_date ORDER BY date LIMIT :limit",
				sql.Named("search_date", searchTime.Format(DateOnlyDB)),
				sql.Named("limit", limit))
			if err != nil {
				return []*Task{}, err
			}
			defer rows.Close()
		}

		if errFromParse != nil {
			search = "%" + search + "%"
			rows, err = db.Query("SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit",
				sql.Named("search", search),
				sql.Named("limit", limit))
			if err != nil {
				return []*Task{}, err
			}
			defer rows.Close()
		}
	}

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return []*Task{}, err
		}

		allTasks = append(allTasks, &task)
	}

	if err := rows.Err(); err != nil {
		return []*Task{}, err
	}

	return allTasks, nil
}

// GetTask выводит задачу по id из БД
func GetTask(id string) (*Task, error) {

	dbFile := "scheduler.db"
	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	task := Task{}

	row := db.QueryRow("SELECT * FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return &Task{}, err
	}

	return &task, nil
}

// UpdateTask обновляет данные о задаче в БД
func UpdateTask(task *Task) error {

	dbFile := "scheduler.db"
	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("date", &task.Date),
		sql.Named("title", &task.Title),
		sql.Named("comment", &task.Comment),
		sql.Named("repeat", &task.Repeat),
		sql.Named("id", &task.ID))
	if err != nil {
		return fmt.Errorf("не удалось обновить данные задачи")
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось обновить данные задачи, ошибка: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("плохой id для обновления задачи")
	}

	return nil
}

// DeleteTask удаляет задачу по id из БД
func DeleteTask(id string) error {

	dbFile := "scheduler.db"
	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("не удалось удалить данные задачи, ошибка: %v", err)
	}

	return nil
}
