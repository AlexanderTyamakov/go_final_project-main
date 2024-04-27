package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func manageDatabase() {
	var dbFile string

	dbFilePath := os.Getenv("TODO_DBFILE")
	if dbFilePath != "" {
		dbFile = dbFilePath
	} else {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Println("Файл базы данных не найден. Создание новой базы данных.")

		db, err := sql.Open("sqlite", dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		_, err = db.Exec(`CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT,
			title TEXT,
			comment TEXT,
			repeat TEXT
		)`)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("CREATE INDEX idx_date ON scheduler(date)")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("База данных успешно создана.")
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Файл базы данных найден.")
	}
}
