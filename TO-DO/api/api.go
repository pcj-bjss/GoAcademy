package api

import (
	"GoAcademy/TO-DO/todo"
	"encoding/json"
	"html/template"
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

	// Create the command to send to the actor
	cmd := todo.Command{
		Action:  todo.OpGet,
		Ctx:     r.Context(),
		Result:  make(chan any), // The Command creates a new channel specific to the request to receive the response from the actor
		ErrChan: make(chan error),
	}
	slog.Default().Log(r.Context(), slog.LevelInfo, "Sending 'get' command to actor.")
	//cmd is sent to the actor via the Store channel
	todo.Store <- cmd

	// Wait for the response from the actor
	select {
	// The select statement waits on multiple channel operations, allowing us to handle whichever one completes first.
	// Here, we wait for either a result or an error from the actor.
	// If the actor sends a result via the Result channel, we handle it.
	// In the case of OpGet, the result will be the list of to-do items.
	case result := <-cmd.Result:
		slog.Default().Log(r.Context(), slog.LevelInfo, "Received successful result from actor.")
		// Type assertion to convert the result back to the expected type ([]todo.Item)
		todos, ok := result.([]todo.Item)
		if !ok {
			slog.Default().Log(r.Context(), slog.LevelError, "Actor returned invalid result type.")
			http.Error(w, "Internal server error: invalid result type", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(todos)
		slog.Default().Log(r.Context(), slog.LevelInfo, "Successfully sent to-do list to client.", "items_count", len(todos))

	case err := <-cmd.ErrChan:
		slog.Default().Log(r.Context(), slog.LevelError, "Actor returned an error.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		"Input data validation successful.",
		"name", t.Name,
		"due", t.Due,
	)

	cmd := todo.Command{
		Action:  todo.OpAdd,
		Ctx:     r.Context(),
		Item:    t,
		Result:  make(chan any), // The Command creates a new channel specific to the request to receive the response from the actor
		ErrChan: make(chan error),
	}
	slog.Default().Log(r.Context(), slog.LevelInfo, "Sending 'add' command to actor.")
	//cmd is sent to the actor via the Store channel
	todo.Store <- cmd

	// Wait for the response from the actor
	select {
	// The select statement waits on multiple channel operations, allowing us to handle whichever one completes first.
	// Here, we wait for either a result or an error from the actor.
	// If the actor sends a result via the Result channel, we handle it.
	// In the case of OpAdd, the result will be the Item that was added.
	case result := <-cmd.Result:
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Received successful result from actor.",
			"name", result.(todo.Item).Name,
			"due", result.(todo.Item).Due,
		)

		w.WriteHeader(http.StatusCreated) // 201 Created
		w.Write([]byte(`{"status": "success","message":"To-do item created successfully."}`))
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Response sent to client.",
			"status", "201 Created",
		)

	case err := <-cmd.ErrChan:
		slog.Default().Log(r.Context(), slog.LevelError, "Actor returned an error.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
		ID        int     `json:"id"`
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

	cmd := todo.Command{
		Action: todo.OpUpdate,
		Ctx:    r.Context(),
		UpdatePayload: struct {
			Name      *string
			Due       *string
			Completed *bool
		}{Name: req.Name, Due: req.Due, Completed: req.Completed},
		ID:      req.ID,
		Result:  make(chan any), // The Command creates a new channel specific to the request to receive the response from the actor
		ErrChan: make(chan error),
	}
	slog.Default().Log(r.Context(), slog.LevelInfo, "Sending 'update' command to actor.")
	//cmd is sent to the actor via the Store channel
	todo.Store <- cmd

	// Wait for the response from the actor
	select {
	// The select statement waits on multiple channel operations, allowing us to handle whichever one completes first.
	// Here, we wait for either a result or an error from the actor.
	// If the actor sends a result via the Result channel, we handle it.
	// In the case of OpUpdate, the result will be the Item that was updated.
	case result := <-cmd.Result:
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Received successful result from actor.",
			"id", result.(todo.Item).ID,
			"name", result.(todo.Item).Name,
			"due", result.(todo.Item).Due,
		)

		w.WriteHeader(http.StatusCreated) // 201 Created
		w.Write([]byte(`{"status": "success","message":"To-do item updated successfully."}`))
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Response sent to client.",
			"status", "201 Created",
		)

	case err := <-cmd.ErrChan:
		slog.Default().Log(r.Context(), slog.LevelError, "Actor returned an error.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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

	q := r.URL.Query().Get("id")
	if q == "" {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Delete request missing id parameter.",
		)
		http.Error(w, "Bad Request: missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(q)
	if err != nil {
		slog.Default().Log(
			r.Context(),
			slog.LevelWarn,
			"Delete request has invalid id parameter.",
		)
		http.Error(w, "Bad Request: invalid id", http.StatusBadRequest)
		return
	}

	cmd := todo.Command{
		Action:  todo.OpUpdate,
		Ctx:     r.Context(),
		ID:      id,
		Result:  make(chan any), // The Command creates a new channel specific to the request to receive the response from the actor
		ErrChan: make(chan error),
	}
	slog.Default().Log(r.Context(), slog.LevelInfo, "Sending 'update' command to actor.")
	//cmd is sent to the actor via the Store channel
	todo.Store <- cmd

	// Wait for the response from the actor
	select {
	// The select statement waits on multiple channel operations, allowing us to handle whichever one completes first.
	// Here, we wait for either a result or an error from the actor.
	// If the actor sends a result via the Result channel, we handle it.
	// In the case of OpUpdate, the result will be the Item that was updated.
	case result := <-cmd.Result:
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Received successful result from actor.",
			"id", result.(todo.Item).ID,
			"name", result.(todo.Item).Name,
			"due", result.(todo.Item).Due,
		)

		w.WriteHeader(http.StatusCreated) // 201 Created
		w.Write([]byte(`{"status": "success","message":"To-do item updated successfully."}`))
		slog.Default().Log(
			r.Context(),
			slog.LevelInfo,
			"Response sent to client.",
			"status", "201 Created",
		)

	case err := <-cmd.ErrChan:
		slog.Default().Log(r.Context(), slog.LevelError, "Actor returned an error.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

var listTmpl = template.Must(template.ParseFiles("web/templates/list.html"))

func ListHandler(w http.ResponseWriter, r *http.Request) {
	slog.Default().Log(
		r.Context(),
		slog.LevelInfo,
		"Received LIST request for to-do list page.",
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create the command to send to the actor to get the list of items
	cmd := todo.Command{
		Action:  todo.OpGet,
		Ctx:     r.Context(),
		Result:  make(chan any),
		ErrChan: make(chan error),
	}
	slog.Default().Log(r.Context(), slog.LevelInfo, "Sending 'get' command to actor for list page.")
	todo.Store <- cmd

	// Wait for the response from the actor
	select {
	case result := <-cmd.Result:
		slog.Default().Log(r.Context(), slog.LevelInfo, "Received successful result from actor for list page.")
		items, ok := result.([]todo.Item)
		if !ok {
			slog.Default().Log(r.Context(), slog.LevelError, "Actor returned invalid result type for list page.")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Default().Log(r.Context(), slog.LevelInfo, "Rendering list page", "items_count", len(items))
		if err := listTmpl.Execute(w, items); err != nil {
			slog.Default().Log(r.Context(), slog.LevelError, "Failed to render to-do list template.", "error", err)
			http.Error(w, "Internal Server Error: could not render page", http.StatusInternalServerError)
			return
		}
		slog.Default().Log(r.Context(), slog.LevelInfo, "To-do list page successfully rendered and sent to client.", "items_count", len(items))

	case err := <-cmd.ErrChan:
		slog.Default().Log(r.Context(), slog.LevelError, "Actor returned an error for list page.", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
