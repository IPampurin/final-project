package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
CREATE TABLE scheduler (
id INTEGER PRIMARY KEY AUTOINCREMENT,
date CHAR(8) NOT NULL DEFAULT "",
title VARCHAR(256) NOT NULL DEFAULT "",
comment TEXT NOT NULL DEFAULT "",
repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX scheduler_date ON scheduler (date);`

func Init(dbFile string) error {

	envDbFile, ok := os.LookupEnv("TODO_DBFILE")
	if !ok && len(envDbFile) > 0 {
		dbFile = envDbFile
	}

	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Printf("ошибка открытия %s: ", dbFile)
		return err
	}
	defer db.Close()

	if install {
		_, err = db.Exec(schema)
		if err != nil {
			fmt.Printf("ошибка создания таблицы в %v или добавления индекса: ", dbFile)
			return err
		}
	}

	return nil
}
