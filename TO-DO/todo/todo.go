package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

var Filename string = "todos.json"

type Item struct { // To-Do item structure: names must be capitalized to be exported
	ID        int
	Name      string
	Completed bool
	Due       string
}

var ToDos []Item
var ToDosMutex sync.RWMutex //

func AddToDo(toDos []Item, id int, name string, due string, ctx context.Context) ([]Item, error) {
	task := Item{ID: id, Name: name, Due: due} //Completed defaults to false
	toDos = append(toDos, task)
	slog.Default().Log(
		ctx,
		slog.LevelInfo,
		"To-do data successfully added",
		"name", name,
		"due", due)
	return toDos, nil
}

func RemoveToDo(toDos []Item, id int, ctx context.Context) ([]Item, error) {
	for i, item := range toDos {
		if item.ID == id {
			toDos = append(toDos[:i], toDos[i+1:]...) //ellipsis to flatten slice
			slog.Default().Log(
				ctx,
				slog.LevelInfo,
				"To-do data successfully removed",
				"id", id)
			return toDos, nil // Return successfully after removing the item.
		}
	}
	// If the loop completes without finding the ID, return an error.
	return toDos, fmt.Errorf("item with id %d not found", id)
}

func UpdateToDo(toDos []Item, id int, name *string, due *string, completed *bool, ctx context.Context) ([]Item, error) {
	slog.Default().Log(ctx, slog.LevelInfo, "updating with the following values", "id", id, "name", name, "due", due, "completed", completed)
	for i, item := range toDos {
		if item.ID == id {
			if name != nil {
				toDos[i].Name = *name
			}
			if due != nil {
				toDos[i].Due = *due
			}
			if completed != nil {
				toDos[i].Completed = *completed
			}
			slog.Default().Log(ctx, slog.LevelInfo, "To-do data successfully updated", "id", id)
			return toDos, nil // Return successfully after updating.
		}
	}
	// If the loop completes without finding the ID, return an error.
	return toDos, fmt.Errorf("item with id %d not found", id)
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
		"file", filename,
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
