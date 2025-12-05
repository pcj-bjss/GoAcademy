package main

import (
	"GoAcademy/TO-DO/api"
	"GoAcademy/TO-DO/todo"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

const ServerAddr = ":8080"

type contextKey string

const traceIDKey contextKey = "TraceID"

// dynamic extractor struct
type traceIDExtractor struct{}

func (t traceIDExtractor) LogValue() slog.Value {
	// This function runs at log time. It returns a function that takes context
	// and returns the actual log value. This is the idiomatic pattern for this.
	return slog.AnyValue(func(ctx context.Context) slog.Value {
		// Safely retrieve the TraceID from the context
		if id, ok := ctx.Value(traceIDKey).(string); ok {
			return slog.StringValue(id)
		}
		// If not found (e.g., logging outside of a traced call)
		return slog.StringValue("NO_TRACE_ID")
	})
}

func traceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// generate trace id and attach to context
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)

		// expose trace id to clients (optional)
		w.Header().Set("X-Trace-ID", traceID)

		// call next with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {

	// Configure application logger and set it as the global default.
	options := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	appLogger := slog.New(slog.NewTextHandler(os.Stdout, options))

	//Every single log message the logger generates must permanently include a field named "traceID".
	//The value for that field is NOT a static string but the dynamic extractor struct we created.
	appLogger = appLogger.With(
		slog.Attr{
			Key:   "traceID",
			Value: slog.AnyValue(traceIDExtractor{}),
		},
	)

	slog.SetDefault(appLogger)

	ctx := context.WithValue(context.Background(), traceIDKey, uuid.New().String())

	// Set up signal handling to gracefully handle termination signals
	// SIGINT is Ctrl+C. SIGTERM is a generic termination signal (e.g., from a 'kill' command).
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var err error

	todo.ToDos, err = todo.LoadToDos(todo.Filename, ctx)
	if err != nil {
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"Failed to load to-do data, exiting",
			"file", "todos.json",
			"error", err)
		os.Exit(1)
	}

	http.HandleFunc("/get", api.GetHandler)
	http.HandleFunc("/create", api.CreateHandler)
	http.HandleFunc("/update", api.UpdateHandler)
	http.HandleFunc("/delete", api.DeleteHandler)

	server := &http.Server{Addr: ServerAddr, Handler: traceIDMiddleware(http.DefaultServeMux)}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Default().Log(ctx, slog.LevelError, "HTTP server failed", "error", err)
		}
	}()

	// Block the main goroutine until an interrupt signal is received
	sig := <-sigChan

	slog.Default().Log(
		ctx,
		slog.LevelInfo,
		"Received signal, commencing graceful shutdown.",
		slog.String("signal", sig.String()),
	)

	//Initiate graceful shutdown with a timeout context
	shutdownCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"HTTP server graceful shutdown failed",
			"error", err)
	}

	err = todo.SaveToDos(todo.Filename, todo.ToDos, ctx)
	if err != nil {
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"Application terminated due to error saving updated to-do data",
			"file", "todos.json",
			"error", err)
		os.Exit(1)
	}
}
