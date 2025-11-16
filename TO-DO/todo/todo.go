package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

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

func SaveToDos(filename string, todos []Item, ctx context.Context) error { //error is a built-in type
	// Convert the todos slice to JSON
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}
	// Write the JSON data to a file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	slog.Default().Log(
		ctx,
		slog.LevelInfo,
		"To-do data successfully saved to disk",
		"file", "todos.json",
		"items_count", len(todos))
	return nil //must return something of type error
}

func LoadToDos(filename string, ctx context.Context) ([]Item, error) {
	// Read the JSON data from the file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Default().Log(
				ctx,
				slog.LevelInfo,
				"Data file not found, initializing empty list.",
				"file", filename)
			return []Item{}, nil // Return an empty slice if the file doesn't exist
		}
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"Failed to read data file",
			"file", filename,
			"error", err)
		return nil, fmt.Errorf("could not read file %s: %w", filename, err)
	}

	if len(data) == 0 {
		slog.Default().Log(
			ctx,
			slog.LevelInfo,
			"Empty file, initializing empty list.",
			"file", filename)
		// If the file is empty, return an empty list.
		return []Item{}, nil
	}

	// Convert the JSON data to a slice of Item
	var todos []Item
	err = json.Unmarshal(data, &todos)
	if err != nil {
		slog.Default().Log(
			ctx,
			slog.LevelError,
			"Failed to decode data file contents",
			"file", filename,
			"error", err)
		return nil, fmt.Errorf("could not unmarshal data: %w", err)
	}
	return todos, nil
}
