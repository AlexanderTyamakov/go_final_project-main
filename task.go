package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Task struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTask(w, r)
	case http.MethodPost:
		AddTask(w, r)
	case http.MethodPut:
		UpdateTask(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTasks(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		MarkTaskDone(w, r)
	case http.MethodDelete:
		DeleteTask(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	} else {
		_, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
	}

	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		ID int `json:"id"`
	}{int(id)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= ? ORDER BY date ASC LIMIT 50", time.Now().Format("20060102"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Tasks []Task `json:"tasks"`
	}{tasks}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var task Task
	err = db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if task.Date != "" {
		_, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
	}

	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	_, err = db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func MarkTaskDone(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	now := time.Now()

	var task Task
	err = db.QueryRow("SELECT date, repeat FROM scheduler WHERE id = ?", id).Scan(&task.Date, &task.Repeat)
	if err != nil {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	if task.Repeat != "" {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
