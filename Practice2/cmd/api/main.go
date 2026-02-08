package main

import (
	"log"
	"net/http"
	"kbtu.practice2.alikh/internal/handlers"
	"kbtu.practice2.alikh/internal/middleware"
	"kbtu.practice2.alikh/internal/storage"
)

func main() {
	store := storage.NewTaskStore()
	taskHandler := handlers.NewTaskHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", taskHandler.Tasks)

	handler := middleware.LoggingMiddleware("")(
		middleware.APIKeyMiddleware("secret12345")(mux),
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Server running on :8080")
	log.Fatal(server.ListenAndServe())
}
