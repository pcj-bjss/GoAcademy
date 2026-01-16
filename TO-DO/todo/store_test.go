package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

// TestConcurrentAccess is a unit test designed to validate the concurrency safety
// of the application. It spawns multiple goroutines that all attempt to write
// to the Store simultaneously.
func TestConcurrentAccess(t *testing.T) {
	// 1. INITIALIZATION: Start the Store actor.
	// We pass a non-existent filename. The actor will detect this and initialize an empty list.
	// Since OpAdd only modifies memory and doesn't save, no file is created on disk.
	// This starts the background goroutine that listens on the 'Store' channel.
	Store = make(chan Command)
	StartStore("memory_only_test.json")
	// Close the channel when the test finishes to stop the actor goroutine.
	t.Cleanup(func() { close(Store) })

	// 3. EXECUTION: Simulate high concurrency.
	// We loop 50 times to create 50 separate sub-tests.
	for i := 0; i < 50; i++ {
		workerID := i // Capture loop variable for parallel execution

		// t.Run defines a sub-test with a unique name (e.g., "Worker-0").
		t.Run(fmt.Sprintf("Worker-%d", i), func(t *testing.T) {
			// t.Parallel() signals to the Go test runner that this sub-test
			// can be executed at the same time as other parallel tests.
			// This causes all 50 iterations of this loop to execute effectively simultaneously.
			t.Parallel()

			// Use slog to match the application's logging pattern
			slog.Info("Worker starting", "worker_id", workerID)

			// Construct a Command to add a new To-Do item.
			cmd := Command{
				Action: OpAdd,
				Item: Item{
					Name: fmt.Sprintf("Concurrent Task %d", workerID),
					Due:  "01-01-2025",
				},
				Ctx:     context.Background(),
				Result:  make(chan any),   // Channel to receive success response
				ErrChan: make(chan error), // Channel to receive error response
			}

			// Send the command to the global Store channel.
			// Because we are in a parallel test, multiple goroutines are hitting this line at once.
			Store <- cmd
			slog.Info("Worker sent command", "worker_id", workerID)

			// Wait for the Actor to process the command and respond.
			select {
			case <-cmd.Result:
				// The operation succeeded. We don't need to check the value for this test.
				slog.Info("Worker received success", "worker_id", workerID)
			case err := <-cmd.ErrChan:
				// If the actor returns an error, fail this specific sub-test.
				t.Errorf("Worker %d failed to add item: %v", workerID, err)
			}
		})
	}
}

func TestConcurrentReads(t *testing.T) {
	// 1. SETUP: Create a temporary file.
	tmpFile, err := os.CreateTemp("", "todo_test_read_*.json")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	// Use t.Cleanup() instead of defer so that cleanup happens after test completes.
	// Otherwise, defer would run before the test goroutines complete because defer runs when the surrounding function returns
	// which the test inteprets as happening when the parallel tests are spawned, not when they finish.
	// This would delete the file before the readers could access it.
	// This wouldn't cause a failure because the store would already have loaded initial data into memory,
	// but it would not be a clean test of reading from disk.
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	// Pre-populate the file with 50 items so we have something to read.
	initialItems := []Item{}
	for i := 0; i < 50; i++ {
		initialItems = append(initialItems, Item{
			ID:   i,
			Name: fmt.Sprintf("Task %d", i),
			Due:  "01-01-2025",
		})
	}
	data, _ := json.Marshal(initialItems)
	_ = os.WriteFile(tmpFile.Name(), data, 0644)

	// 2. INITIALIZATION
	Store = make(chan Command)
	StartStore(tmpFile.Name())
	t.Cleanup(func() { close(Store) })

	// 3. EXECUTION: 50 Concurrent Readers
	for i := 0; i < 50; i++ {
		workerID := i
		t.Run(fmt.Sprintf("Reader-%d", i), func(t *testing.T) {
			t.Parallel()

			slog.Info("Reader starting", "worker_id", workerID)
			cmd := Command{
				Action:  OpGet,
				Ctx:     context.Background(),
				Result:  make(chan any),
				ErrChan: make(chan error),
			}

			Store <- cmd

			select {
			case res := <-cmd.Result:
				items := res.([]Item)
				slog.Info("Reader success", "worker_id", workerID, "count", len(items))
				if len(items) != 50 {
					t.Errorf("Worker %d expected 50 items, got %d", workerID, len(items))
				}
			case err := <-cmd.ErrChan:
				t.Errorf("Worker %d failed: %v", workerID, err)
			}
		})
	}
}

func TestConcurrentUpdates(t *testing.T) {
	// 1. SETUP
	tmpFile, err := os.CreateTemp("", "todo_test_update_*.json")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	// Pre-populate
	initialItems := []Item{}
	for i := 0; i < 50; i++ {
		initialItems = append(initialItems, Item{ID: i, Name: "Original", Due: "01-01-2025"})
	}
	data, _ := json.Marshal(initialItems)
	_ = os.WriteFile(tmpFile.Name(), data, 0644)

	// 2. INITIALIZATION
	Store = make(chan Command)
	StartStore(tmpFile.Name())
	t.Cleanup(func() { close(Store) })

	// 3. EXECUTION: 50 Concurrent Updaters
	for i := 0; i < 50; i++ {
		workerID := i
		t.Run(fmt.Sprintf("Updater-%d", i), func(t *testing.T) {
			t.Parallel()

			slog.Info("Updater starting", "worker_id", workerID)
			newName := fmt.Sprintf("Updated by %d", workerID)
			cmd := Command{
				Action:  OpUpdate,
				ID:      workerID,
				Ctx:     context.Background(),
				Result:  make(chan any),
				ErrChan: make(chan error),
			}
			cmd.UpdatePayload.Name = &newName

			Store <- cmd

			select {
			case <-cmd.Result:
				slog.Info("Updater success", "worker_id", workerID)
			case err := <-cmd.ErrChan:
				t.Errorf("Worker %d failed: %v", workerID, err)
			}
		})
	}
}

func TestStartStore_LoadsExistingData(t *testing.T) {
	// 1. SETUP: Create a file with known data
	tmpFile, err := os.CreateTemp("", "todo_store_existing_*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	initialData := []Item{{ID: 99, Name: "Existing Task", Due: "01-01-2025"}}
	data, _ := json.Marshal(initialData)
	_ = os.WriteFile(tmpFile.Name(), data, 0644)
	tmpFile.Close()

	// 2. START
	Store = make(chan Command)
	StartStore(tmpFile.Name())
	t.Cleanup(func() { close(Store) })

	// 3. VERIFY: Send OpGet to ensure data was loaded
	// If the actor processes this command, it means it started successfully.
	cmd := Command{Action: OpGet, Result: make(chan any)}
	Store <- cmd
	result := <-cmd.Result
	items := result.([]Item)

	if len(items) != 1 || items[0].ID != 99 {
		t.Errorf("Expected loaded item with ID 99, got %v", items)
	}
}

func TestStartStore_HandlesMissingFile(t *testing.T) {
	// 1. SETUP: Define a filename that does not exist
	// We use CreateTemp to get a valid path, then delete it immediately.
	tmpFile, err := os.CreateTemp("", "todo_store_missing_*.json")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()
	tmpFile.Close()
	os.Remove(filename) // Ensure it's gone

	// 2. START
	Store = make(chan Command)
	StartStore(filename)
	t.Cleanup(func() { close(Store) })

	// 3. VERIFY: Send OpGet.
	// If the actor accepts this command, it means it initialized the empty list successfully.
	cmd := Command{Action: OpGet, Result: make(chan any)}
	Store <- cmd
	result := <-cmd.Result
	items := result.([]Item)

	if len(items) != 0 {
		t.Errorf("Expected empty list for missing file, got %d items", len(items))
	}
}

func BenchmarkAddToDo_Direct(b *testing.B) {
	// Benchmark the logic function directly (no actor overhead).
	// This measures the cost of memory allocation and slice appending.
	ctx := context.Background()
	// Start with an empty slice.
	// Note: In a real scenario, this slice would grow very large during the benchmark loop.
	// For this micro-benchmark, we are measuring the amortized cost of append().
	items := []Item{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We re-assign items so the slice grows, simulating real usage.
		items, _ = AddToDo(items, i, "Benchmark Task", "01-01-2025", ctx)
	}
}

func BenchmarkActor_Add(b *testing.B) {
	// Benchmark the full actor round-trip.
	// This measures the cost of the logic PLUS the overhead of Go channels and context switching.

	// 1. Setup
	Store = make(chan Command)
	StartStore("bench_actor.json")
	// Ensure we stop the actor when the benchmark finishes
	b.Cleanup(func() { close(Store) })

	// 2. Reset timer so setup doesn't count towards the score
	b.ResetTimer()

	// 3. The Benchmark Loop
	for i := 0; i < b.N; i++ {
		cmd := Command{
			Action: OpAdd,
			Item: Item{
				Name: "Bench Task",
				Due:  "01-01-2025",
			},
			Ctx:     context.Background(),
			Result:  make(chan any),
			ErrChan: make(chan error),
		}
		Store <- cmd
		<-cmd.Result // Wait for the operation to complete
	}
}
