// package api

// //This was a more complex solution to handle routing and dependency injection
// //However, for simplicity, we moved to a more straightforward approach in main.go

// import (
// 	"GoAcademy/TO-DO/todo"
// 	"encoding/json"
// 	"log/slog"
// 	"net/http"
// 	"sync"
// 	"time"
// )

// type Server struct {
// 	Filename string
// 	Logger   *slog.Logger
// 	router   http.Handler
// 	once     sync.Once
// }

// // Constructor to initialise the Server struct and inject dependencies
// // This will be called from main.go
// func NewServer(filename string, logger *slog.Logger) *Server {
// 	if logger == nil {
// 		logger = slog.Default()
// 	}
// 	return &Server{
// 		Filename: filename,
// 		Logger:   logger,
// 	}
// }

// // A function on Server to return the router
// // This will be called from main.go to get the handler for http.Server
// // Router creates and configures the ServeMux (the router) and registers all endpoints
// // It then returns a configured http.Handler which can be passed to http.Server.

// func (s *Server) Router() http.Handler {
// 	s.once.Do(func() { // Ensure this block runs only once so that handlers aren't registered multiple times
// 		s.Logger.Info("Initializing API router.")

// 		// The ServeMux is created here
// 		mux := http.NewServeMux()

// 		// Register endpoints as methods on the Server struct (s.GetHandler, etc.)
// 		mux.HandleFunc("/create", s.CreateHandler)
// 		mux.HandleFunc("/get", s.GetHandler)
// 		mux.HandleFunc("/update", s.UpdateHandler)
// 		mux.HandleFunc("/delete", s.DeleteHandler)

// 		// Assign the configured mux to the Server struct's router field so that it can be returned and reused
// 		s.router = mux
// 	})

// 	// Return the configured router
// 	return s.router
// }

// // Logic for methods. Note: The methods must have the signature func(w http.ResponseWriter, r *http.Request)
// func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
// 	s.Logger.Info("Received GET request for to-do list.")
// 	w.Header().Set("Content-Type", "application/json")

// 	todo.ToDosMutex.RLock()
// 	defer todo.ToDosMutex.RUnlock()

// 	data, err := json.MarshalIndent(todo.ToDos, "", "  ")

// 	if err != nil {
// 		// If conversion fails, log the error and send a 500 status.
// 		s.Logger.Error("Failed to encode to-do list to JSON", "error", err)
// 		// http.Error is a convenience function that sets the status code and writes the message.
// 		http.Error(w, "Internal server error: Could not format list data.", http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK) // Status 200 OK
// 	w.Write(data)

// 	s.Logger.Info("Successfully served to-do list.", "items_count", len(todo.ToDos))
// }

// // Handler skeletons for other endpoints
// func (s *Server) CreateHandler(w http.ResponseWriter, r *http.Request) {
// 	s.Logger.Info("Recieved CREATE request for to-do item.")
// 	w.Header().Set("Content-Type", "application/json")

// 	if r.Method != http.MethodPost {
// 		s.Logger.Warn("Invalid HTTP method for /create endpoint.", "method", r.Method)
// 		http.Error(w, "Method Not Allowed. Use POST", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var t todo.Item
// 	err := json.NewDecoder(r.Body).Decode(&t)
// 	if err != nil {
// 		s.Logger.Error("Failed to decode request body", "error", err)
// 		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	// Validate input
// 	if t.Name == "" {
// 		s.Logger.Warn("Create request with empty Name.")
// 		http.Error(w, "Bad Request: Name cannot be empty", http.StatusBadRequest)
// 		return
// 	}
// 	_, err = time.Parse("02-01-2006", t.Due)
// 	if err != nil {
// 		s.Logger.Warn("Create request with invalid Due date format.", "due", t.Due)
// 		http.Error(w, "Bad Request: Due date must be in DD-MM-YYYY format", http.StatusBadRequest)
// 		return
// 	}
// 	s.Logger.Info("Data validation successful. Acquiring write lock.")

// 	todo.ToDosMutex.Lock()
// 	defer todo.ToDosMutex.Unlock()

// 	todo.AddToDo(t.Name, t.Due)
// 	err = todo.SaveToDos(s.Filename, todo.ToDos, r.Context())
// 	if err != nil {
// 		// If saving fails, it is a critical server error (500), not a client error (400).
// 		s.Logger.Error("Failed to save to-do data after addition.", "file", s.Filename, "error", err)

// 		// Send 500 Internal Server Error
// 		http.Error(w, "Internal Server Error: Failed to save changes to disk.", http.StatusInternalServerError)
// 		return
// 	}

// 	s.Logger.Info("New item successfully added and saved to disk.")

// 	w.WriteHeader(http.StatusCreated) // 201 Created
// 	w.Write([]byte(`{"status": "success","message":"To-do item created successfully."}`))
// 	s.Logger.Info("Response sent to client.", "status", "201 Created")
// }

// func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
// 	s.Logger.Warn("Endpoint /update called but not implemented yet.")
// 	http.Error(w, "Not Implemented", http.StatusNotImplemented)
// }

// func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
// 	s.Logger.Warn("Endpoint /delete called but not implemented yet.")
// 	http.Error(w, "Not Implemented", http.StatusNotImplemented)
// }
