package api

import (
	"GoAcademy/TO-DO/todo"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Logic for methods. Note: The methods must have the signature func(w http.ResponseWriter, r *http.Request)
func GetHandler(w http.ResponseWriter, r *http.Request) {
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Received GET request for to-do list.")
	w.Header().Set("Content-Type", "application/json")

	todo.ToDosMutex.RLock()
	defer todo.ToDosMutex.RUnlock()

	data, err := json.MarshalIndent(todo.ToDos, "", "  ")

	if err != nil {
		// If conversion fails, log the error and send a 500 status.
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to encode to-do list to JSON.",
			"error", err,
		)
		// http.Error is a convenience function that sets the status code and writes the message.
		http.Error(w, "Internal server error: Could not format list data.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK) // Status 200 OK
	w.Write(data)

	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"To-do list successfully encoded and sent to client.",
		"items_count", len(todo.ToDos),
	)

}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Recieved CREATE request for to-do item.",
	)

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Invalid HTTP method for /create endpoint.",
			"method", r.Method,
		)
		http.Error(w, "Method Not Allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	var t todo.Item
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to decode request body",
			"error", err)
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if t.Name == "" {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Create request with empty Name.",
		)

		http.Error(w, "Bad Request: Name cannot be empty", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse("02-01-2006", t.Due); err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Create request with invalid Due date format.",
			"due", t.Due,
			"error", err,
		)

		http.Error(w, "Bad Request: Due date must be in DD-MM-YYYY format", http.StatusBadRequest)
		return
	}
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Input data validation successful. Acquiring write lock.",
		"name", t.Name,
		"due", t.Due,
	)

	todo.ToDosMutex.Lock()
	defer todo.ToDosMutex.Unlock()

	todo.AddToDo(t.Name, t.Due, r.Context())
	if err := todo.SaveToDos(todo.Filename, todo.ToDos, r.Context()); err != nil {
		// If saving fails, it is a critical server error (500), not a client error (400).
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to save to-do data after addition.",
			"file", todo.Filename,
			"error", err,
		)

		// Send 500 Internal Server Error
		http.Error(w, "Internal Server Error: Failed to save changes to disk.", http.StatusInternalServerError)
		return
	}

	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"New item successfully added and saved to disk.",
		"name", t.Name,
		"due", t.Due,
	)

	w.WriteHeader(http.StatusCreated) // 201 Created
	w.Write([]byte(`{"status": "success","message":"To-do item created successfully."}`))
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Response sent to client.",
		"status", "201 Created",
	)
}

// Handler skeletons for other endpoints
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Recieved UPDATE request for to-do item.",
	)

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPatch {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Invalid HTTP method for /update endpoint.",
			"method", r.Method,
		)

		http.Error(w, "Method Not Allowed. Use PATCH", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Index     int     `json:"index"`
		Name      *string `json:"name,omitempty"`
		Due       *string `json:"due,omitempty"`
		Completed *bool   `json:"completed,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to decode request body",
			"error", err,
		)
		http.Error(w, "Bad Request: invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate index
	if req.Index < 0 || req.Index >= len(todo.ToDos) {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Update request with out-of-bounds index.",
			"index", req.Index,
		)

		http.Error(w, "Bad Request: Index out of bounds", http.StatusBadRequest)
		return
	}

	todo.ToDosMutex.Lock()
	defer todo.ToDosMutex.Unlock()

	// Perform the update
	todo.UpdateToDo(req.Index, req.Name, req.Due, req.Completed, r.Context())

	// Save the updated list
	if err := todo.SaveToDos(todo.Filename, todo.ToDos, r.Context()); err != nil {
		// If saving fails, it is a critical server error (500), not a client error (400).
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to save to-do data after update.",
			"file", todo.Filename,
			"error", err,
		)

		// Send 500 Internal Server Error
		http.Error(w, "Internal Server Error: Failed to save changes to disk.", http.StatusInternalServerError)
		return
	}

	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"To-do item successfully updated and saved to disk.",
	)

	w.WriteHeader(http.StatusCreated) // 201 Created
	w.Write([]byte(`{"status": "success","message":"To-do item updated successfully."}`))
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Response sent to client.",
		"status", "201 Created",
	)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Recieved DELETE request for to-do item.",
	)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Invalid HTTP method for /delete endpoint.",
			"method", r.Method,
		)

		http.Error(w, "Method Not Allowed. Use DELETE", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("index")
	if q == "" {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Delete request missing index parameter.",
		)
		http.Error(w, "Bad Request: missing index parameter", http.StatusBadRequest)
		return
	}
	idx, err := strconv.Atoi(q)
	if err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Delete request has invalid index parameter.",
		)
		http.Error(w, "Bad Request: invalid index", http.StatusBadRequest)
		return
	}

	todo.ToDosMutex.Lock()
	defer todo.ToDosMutex.Unlock()

	if idx < 0 || idx >= len(todo.ToDos) {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Delete request has out of range index parameter.",
		)
		http.Error(w, "Not Found: index out of range", http.StatusNotFound)
		return
	}

	todo.RemoveToDo(idx, r.Context())
	if err := todo.SaveToDos(todo.Filename, todo.ToDos, r.Context()); err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelError,
			"Failed to save to-do data after delete.",
			"file", todo.Filename,
			"error", err,
		)
		http.Error(w, "Internal Server Error: failed to persist changes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"To-do item successfully deleted and changes saved to disk.",
		"index", idx,
	)

}
