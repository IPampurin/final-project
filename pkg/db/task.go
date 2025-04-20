package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DateFromSearch = "02.01.2006"
	DateOnlyDB     = "20060102"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {

	db, err := sql.Open("sqlite", "scheduler.db")
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

func Tasks(limit int, search string) ([]*Task, error) {

	db, err := sql.Open("sqlite", "scheduler.db")
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

func GetTask(id string) (*Task, error) {

	db, err := sql.Open("sqlite", "scheduler.db")
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

func UpdateTask(task *Task) error {

	db, err := sql.Open("sqlite", "scheduler.db")
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
		return err
	}
	if count == 0 {
		return fmt.Errorf("плохой id для обновления задачи")
	}

	return nil
}
