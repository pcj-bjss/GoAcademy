package todo

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

// The Go testing framework (go test) looks for any public function that starts with Test followed by an uppercase letter.
// This signals that the function is a test case.
// (t *testing.T) is the mandatory parameter.
// It accepts a pointer to a testing.T struct from the testing package.
// The t variable provides all the methods you use to manage your test: logging messages (t.Log()), marking a test as failed (t.Errorf()), or stopping a test immediately (t.Fatalf()).

func TestLoadToDos_EmptyFile(t *testing.T) {
	// Set up a temporary file to simulate an empty todos.json
	// <file>_* is the filename pattern. Go replaces the * with a random, unique string (e.g., todos_123456.json) to guarantee no two tests will overwrite each other.
	tmpFile, err := os.CreateTemp("", "todos_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()              // Close the handle immediately after creation
	defer os.Remove(tmpFileName) // Ensure the file is deleted when the test exits.
	// defer statements run at the end before returning, even if the test fails.

	ctx := context.Background() // Create a dummy context to satisfy logging requirements
	todos, loadErr := LoadToDos(tmpFileName, ctx)

	//Test 1 (No Error):
	// Asserts that reading an empty file should not result in a critical error that halts execution.
	// If an error occurs, the test fails immediately with t.Fatalf.
	if loadErr != nil {
		t.Fatalf("LoadToDos failed unexpectedly: %v", loadErr)
	}
	//Test 2 (Empty Result):
	// Asserts that reading an empty file should return a list with zero items ([]Item{}).
	if len(todos) != 0 {
		t.Errorf("Expected 0 todos for empty file, got %d", len(todos))
	}

}

func TestLoadToDos_NonExistentFile(t *testing.T) {
	// Use a filename that doesn't exist
	nonExistentFile := "non_existent_todos.json"
	ctx := context.Background()
	todos, err := LoadToDos(nonExistentFile, ctx)

	// Test 1 (No Error):
	// Asserts that attempting to load from a non-existent file should not result in a critical error.
	// This is because the application is designed to handle this case gracefully by initializing an empty list.
	// If an error occurs, the test fails immediately with t.Fatalf.
	if err != nil {
		t.Fatalf("LoadToDos failed unexpectedly: %v", err)
	}
	// Test 2 (Empty Result):
	if len(todos) != 0 {
		t.Errorf("Expected 0 todos for non-existent file, got %d", len(todos))
	}
}

func TestLoadToDos_HappyPath(t *testing.T) {

	// Set up a temporary file with valid JSON data
	tmpFile, err := os.CreateTemp("", "todos_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName) // Ensure the file is deleted when the test exits
	validJSON := `[{"Name":"Test ToDo","DueDate":"2024-12-31T23:59:59Z","Completed":false}]`
	_, writeErr := tmpFile.WriteString(validJSON)
	if writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
	tmpFile.Close() // Close the file to ensure data is flushed

	ctx := context.Background()
	todos, loadErr := LoadToDos(tmpFileName, ctx)
	// Test 1 (No Error):
	if loadErr != nil {
		t.Fatalf("LoadToDos failed unexpectedly: %v", loadErr)
	}
	// Test 2 (Correct Data):
	if len(todos) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(todos))
	}
	expectedName := "Test ToDo"
	if todos[0].Name != expectedName {
		t.Errorf("Expected todo name %q, got %q", expectedName, todos[0].Name)
	}
}

func TestLoadToDos_MalformedJSON(t *testing.T) {
	// Set up a temporary file with malformed JSON data
	tmpFile, err := os.CreateTemp("", "todos_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName)
	malformedJSON := `{"Name":"Test ToDo","DueDate":"2024-12-31T23:59:59Z","Completed":false` // Missing closing brace
	_, writeErr := tmpFile.WriteString(malformedJSON)
	if writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
	tmpFile.Close() // Close the file to ensure data is flushed

	ctx := context.Background()
	_, loadErr := LoadToDos(tmpFileName, ctx)
	// Test 1 (Error Expected):
	if loadErr == nil {
		t.Fatal("Expected LoadToDos to fail due to malformed JSON, but it succeeded")
	}
	// Test 2 (Error Type):
	if !os.IsNotExist(loadErr) && loadErr.Error() == "" {
		t.Errorf("Expected a JSON unmarshal error, got: %v", loadErr)
	}
}

func TestSaveToDos(t *testing.T) {
	// Set up a temporary file to save todos
	tmpFile, err := os.CreateTemp("", "todos_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName)
	tmpFile.Close()

	ctx := context.Background() // Create a dummy context to satisfy logging requirements

	// Create a local slice for testing, not using the global one.
	todosToSave := []Item{
		{ID: 1, Name: "ToDo 1", Due: "2024-12-01T10:00:00Z", Completed: false},
		{ID: 2, Name: "ToDo 2", Due: "2024-12-02T11:00:00Z", Completed: false},
	}

	// Test 1 (No Error):
	saveErr := SaveToDos(tmpFileName, todosToSave, ctx)
	if saveErr != nil {
		t.Fatalf("SaveToDos failed unexpectedly: %v", saveErr)
	}

	// Test 2 (File Content):
	data, readErr := os.ReadFile(tmpFileName)
	if readErr != nil {
		t.Fatalf("Failed to read saved file: %v", readErr)
	}
	// Create the expected data structure, matching what was passed to SaveToDos.
	var expectedContent []Item
	task1 := Item{ID: 1, Name: "ToDo 1", Due: "2024-12-01T10:00:00Z", Completed: false}
	task2 := Item{ID: 2, Name: "ToDo 2", Due: "2024-12-02T11:00:00Z", Completed: false}
	expectedContent = append(expectedContent, task1)
	expectedContent = append(expectedContent, task2)

	// Marshal the expected data to JSON for comparison.
	expectedJSON, err := json.Marshal(expectedContent)
	if err != nil {
		t.Fatalf("Failed to marshal expected data: %v", err)
	}

	// Compare the actual file content with the expected JSON.
	if string(data) != string(expectedJSON) {
		t.Errorf("Saved file content does not match expected.\nGot: %s\nWant: %s", string(data), string(expectedJSON))
	}
}

func TestAddToDo(t *testing.T) {
	// Start with an empty local slice for a clean test environment.
	todos := []Item{}
	ctx := context.Background() // Create a dummy context to satisfy logging requirements

	todoName := "New Test ToDo"
	todoDue := "2024-11-30T18:00:00Z"

	// Call the new AddToDo function, passing the local slice and capturing the returned slice.
	updatedTodos, err := AddToDo(todos, 1, todoName, todoDue, ctx)
	if err != nil {
		t.Fatalf("AddToDo failed unexpectedly: %v", err)
	}

	// Test 1 (Item Added):
	if len(updatedTodos) != 1 {
		t.Fatalf("Expected 1 todo after addition, got %d", len(updatedTodos))
	}
	// Test 2 (Correct Data):
	if updatedTodos[0].ID != 1 || updatedTodos[0].Name != todoName || updatedTodos[0].Due != todoDue || updatedTodos[0].Completed != false {
		t.Errorf("Added todo does not match expected values. Got: %+v", updatedTodos[0])
	}
}

func TestRemoveToDo(t *testing.T) {
	// Start with a local slice with sample data to test against.
	todos := []Item{
		{ID: 1, Name: "ToDo 1"},
		{ID: 5, Name: "ToDo 2 to be removed"}, // Use a non-sequential ID to test ID-based removal
		{ID: 10, Name: "ToDo 3"},
	}
	ctx := context.Background() // Create a dummy context to satisfy logging requirements

	// Test removing the item with ID 5
	updatedTodos, err := RemoveToDo(todos, 5, ctx)
	if err != nil {
		t.Fatalf("RemoveToDo failed unexpectedly: %v", err)
	}

	// Test 1 (Item Removed):
	if len(updatedTodos) != 2 {
		t.Fatalf("Expected 2 todos after removal, got %d", len(updatedTodos))
	}
	// Test 2 (Correct Items Remaining):
	if updatedTodos[0].ID != 1 || updatedTodos[1].ID != 10 {
		t.Errorf("Remaining todos do not match expected values. Got: %+v", updatedTodos)
	}
}
