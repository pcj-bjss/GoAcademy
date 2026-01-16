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
	"time"

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

	// Start the actor goroutine. This runs in the background.
	todo.StartStore(todo.Filename)

	// Set up signal handling to gracefully handle termination signals
	// SIGINT is Ctrl+C. SIGTERM is a generic termination signal (e.g., from a 'kill' command).
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set up HTTP handlers
	http.HandleFunc("/get", api.GetHandler)
	http.HandleFunc("/create", api.CreateHandler)
	http.HandleFunc("/update", api.UpdateHandler)
	http.HandleFunc("/delete", api.DeleteHandler)
	http.HandleFunc("/list", api.ListHandler)
	// serve static files for the web frontend
	http.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("web/static/about"))))

	// redirect "/about" (no trailing slash) -> "/about/" so index.html is returned
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/about/", http.StatusMovedPermanently)
	})

	// create an http.Server that listens on ServerAddr.
	// Handler is the DefaultServeMux wrapped by traceIDMiddleware so each request
	// gets a per-request TraceID placed into r.Context() and an X-Trace-ID header.
	server := &http.Server{Addr: ServerAddr, Handler: traceIDMiddleware(http.DefaultServeMux)}

	// Start the server in a separate goroutine so the main goroutine can continue
	// (for example, to wait for OS signals). ListenAndServe blocks while serving.
	go func() {
		// When ListenAndServe returns, check the error. http.ErrServerClosed is the
		// expected error returned when Shutdown() is called for a graceful stop,
		// so only log errors that are unexpected.
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Default().Log(ctx, slog.LevelError, "HTTP server failed", "error", err)
		}
	}()
	slog.Default().Log(ctx, slog.LevelInfo, "Server started and waiting for requests", "addr", ServerAddr)

	// Block the main goroutine until an interrupt signal is received
	sig := <-sigChan

	slog.Default().Log(
		ctx,
		slog.LevelInfo,
		"Received signal, commencing graceful shutdown.",
		slog.String("signal", sig.String()),
	)

	//Initiate graceful shutdown with a timeout context
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"HTTP server graceful shutdown failed",
			"error", err)
	}

	// Send the shutdown command to the actor. This ensures it processes any remaining items, saves, and exits.
	slog.Default().Log(ctx, slog.LevelInfo, "Sending shutdown command to actor.")
	shutdownCmd := todo.Command{
		Action:  todo.OpShutdown,
		Ctx:     ctx,
		Result:  make(chan any),
		ErrChan: make(chan error),
	}
	todo.Store <- shutdownCmd

	// Wait for the actor to confirm the save is complete.
	select {
	case <-shutdownCmd.Result:
		slog.Default().Log(ctx, slog.LevelInfo, "Actor shut down successfully. Exiting.")
	case err := <-shutdownCmd.ErrChan:
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"Application terminated due to error saving updated to-do data",
			"file", todo.Filename,
			"error", err)
	}
}
