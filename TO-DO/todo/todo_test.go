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

	tmpStruct := []Item{}
	originalToDos := ToDos
	ToDos = tmpStruct
	defer func() { ToDos = originalToDos }()
	// Add sample todos
	AddToDo("ToDo 1", "2024-12-01T10:00:00Z")
	AddToDo("ToDo 2", "2024-12-02T11:00:00Z")

	ctx := context.Background()
	// Test 1 (No Error):
	saveErr := SaveToDos(tmpFileName, ToDos, ctx)
	if saveErr != nil {
		t.Fatalf("SaveToDos failed unexpectedly: %v", saveErr)
	}
	// Test 2 (File Content):
	data, readErr := os.ReadFile(tmpFileName)
	if readErr != nil {
		t.Fatalf("Failed to read saved file: %v", readErr)
	}
	var expectedContent []Item
	task1 := Item{Name: "ToDo 1", Due: "2024-12-01T10:00:00Z"} //Completed defaults to false
	task2 := Item{Name: "ToDo 2", Due: "2024-12-02T11:00:00Z"}
	expectedContent = append(expectedContent, task1)
	expectedContent = append(expectedContent, task2)
	expectedJSON, err := json.Marshal(expectedContent)
	if err != nil {
		t.Fatalf("Failed to marshal expected data: %v", err)
	}

	if string(data) != string(expectedJSON) {
		t.Errorf("Saved file content does not match expected.\nGot: %s\nWant: %s", string(data), string(expectedJSON))
	}
}

func TestAddToDo(t *testing.T) {
	// Set up a temporary struct to hold todos
	tmpStruct := []Item{}
	// Override the global ToDos slice for testing
	originalToDos := ToDos
	ToDos = tmpStruct
	defer func() { ToDos = originalToDos }() // Restore original after test
	// Test adding a new to-do
	todoName := "New Test ToDo"
	todoDue := "2024-11-30T18:00:00Z"
	AddToDo(todoName, todoDue)

	// Test 1 (Item Added):
	if len(ToDos) != 1 {
		t.Fatalf("Expected 1 todo after addition, got %d", len(ToDos))
	}
	// Test 2 (Correct Data):
	if ToDos[0].Name != todoName || ToDos[0].Due != todoDue || ToDos[0].Completed != false {
		t.Errorf("Added todo does not match expected values. Got: %+v", ToDos[0])
	}
}

func TestRemoveToDo(t *testing.T) {
	tmpStruct := []Item{}
	originalToDos := ToDos
	ToDos = tmpStruct
	defer func() { ToDos = originalToDos }()
	// Add sample todos
	AddToDo("ToDo 1", "2024-12-01T10:00:00Z")
	AddToDo("ToDo 2", "2024-12-02T11:00:00Z")
	AddToDo("ToDo 3", "2024-12-03T12:00:00Z")
	// Test removing the second to-do (index 1)
	RemoveToDo(1)
	// Test 1 (Item Removed):
	if len(ToDos) != 2 {
		t.Fatalf("Expected 2 todos after removal, got %d", len(ToDos))
	}
	// Test 2 (Correct Items Remaining):
	if ToDos[0].Name != "ToDo 1" || ToDos[1].Name != "ToDo 3" {
		t.Errorf("Remaining todos do not match expected values. Got: %+v", ToDos)
	}
}
