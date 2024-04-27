package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %s", err.Error())
	}
	port := os.Getenv("TODO_PORT")
	if port == "" {
		// Если переменная окружения не задана, используем порт по умолчанию
		port = "7540"
	}

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/tasks", getTasksHandler)
	http.HandleFunc("/api/task/done", DoneTaskHandler)

	manageDatabase()

	fmt.Printf("Сервер запущен на http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
