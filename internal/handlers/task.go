package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var (
	tasks  = make(map[int]Task)
	nextID = 1
	mu     sync.Mutex
)

const maxTitleLength = 100

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		idStr := r.URL.Query().Get("id")
		doneStr := r.URL.Query().Get("done")

		// GET by ID
		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
				return
			}

			task, ok := tasks[id]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
				return
			}

			json.NewEncoder(w).Encode(task)
			return
		}

		// GET with filtering
		var result []Task
		for _, task := range tasks {
			if doneStr != "" {
				done, err := strconv.ParseBool(doneStr)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": "invalid done value"})
					return
				}
				if task.Done == done {
					result = append(result, task)
				}
			} else {
				result = append(result, task)
			}
		}

		json.NewEncoder(w).Encode(result)

	case http.MethodPost:
		var input struct {
			Title string `json:"title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
			return
		}

		if input.Title == "" || len(input.Title) > maxTitleLength {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "title must be non-empty and <= 100 chars",
			})
			return
		}

		mu.Lock()
		task := Task{
			ID:    nextID,
			Title: input.Title,
			Done:  false,
		}
		tasks[nextID] = task
		nextID++
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	case http.MethodPatch:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		var body struct {
			Done bool `json:"done"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
			return
		}

		mu.Lock()
		task, ok := tasks[id]
		if !ok {
			mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		}

		task.Done = body.Done
		tasks[id] = task
		mu.Unlock()

		json.NewEncoder(w).Encode(map[string]bool{"updated": true})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		mu.Lock()
		if _, ok := tasks[id]; !ok {
			mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		}

		delete(tasks, id)
		mu.Unlock()

		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
