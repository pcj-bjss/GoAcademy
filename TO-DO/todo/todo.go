package todo

import (
	"encoding/json"
	"log/slog"
	"os"
)

var options = &slog.HandlerOptions{
	// Set the minimum level for logs to be processed.
	// This will include DEBUG, INFO, WARN, and ERROR messages.
	Level: slog.LevelDebug,

	// Tell the Logger to include the file name and line number
	// where the log call was made.
	AddSource: true,
}

var Logger = slog.New(slog.NewTextHandler(os.Stdout, options))

type Item struct { // To-Do item structure: names must be capitalized to be exported
	Name      string
	Completed bool
	Due       string
}

var ToDos []Item

func AddToDo(name string, due string) {
	task := Item{Name: name, Due: due} //Completed defaults to false
	ToDos = append(ToDos, task)
}

func RemoveToDo(index int) {
	ToDos = append(ToDos[:index], ToDos[index+1:]...) //ellipsis to flatten slice
}

func SaveToDos(todos []Item) error { //error is a built-in type
	// Convert the todos slice to JSON
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}
	// Write the JSON data to a file
	err = os.WriteFile("todos.json", data, 0644)
	if err != nil {
		return err
	}
	Logger.Info("To-do data successfully saved to disk",
		"file", "todos.json",
		"items_count", len(todos))
	return nil //must return something of type error
}

func LoadToDos() ([]Item, error) {
	// Read the JSON data from the file
	data, err := os.ReadFile("todos.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil // Return an empty slice if the file doesn't exist
		}
		return nil, err
	}

	// Convert the JSON data to a slice of Item
	var todos []Item
	err = json.Unmarshal(data, &todos)
	if err != nil {
		return nil, err
	}
	return todos, nil
}
