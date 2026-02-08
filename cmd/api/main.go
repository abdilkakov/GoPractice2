package main

import (
	"log"
	"net/http"

	"Practice-2/internal/handlers"
	"Practice-2/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	handler := middleware.APIKeyMiddleware(
		http.HandlerFunc(handlers.TasksHandler),
	)

	mux.Handle("/tasks", handler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
