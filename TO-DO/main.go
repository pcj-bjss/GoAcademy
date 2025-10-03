package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

func main() {

	currentTime := time.Now()
	var err error
	ToDos, err = LoadToDos()
	if err != nil {
		fmt.Printf("An error occurred:%s\n", err)
		os.Exit(1)
	}

	titlePtr := flag.String("t", "", "a title for the to-do item")
	duePtr := flag.String("d", currentTime.Format("02-01-2006"), "due date for the to-do item")
	completedPtr := flag.Bool("c", false, "mark the to-do item as completed")

	if len(os.Args) < 2 {
		fmt.Println("No command provided. Available commands: add")
		os.Exit(1)
	} else if os.Args[1] == "add" {
		flag.CommandLine.Parse(os.Args[2:])
		title := *titlePtr
		due := *duePtr
		_, err := time.Parse("02-01-2006", due)
		if err != nil {
			fmt.Println("Invalid date format - provide format DD-MM-YYYY")
			os.Exit(1) // Exit with status code 1 (indicates an error)
		}
		if title == "" {
			fmt.Println("An error occurred. No title provided. Exiting with status code 1.")
			os.Exit(1) // Exit with status code 1 (indicates an error)
		}
		AddToDo(title, due)
		err = SaveToDos(ToDos)
		if err != nil {
			fmt.Printf("An error occurred while saving: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Item added to list.")
	} else if os.Args[1] == "list" {
		fmt.Println(ToDos)
		os.Exit(0)
	} else if os.Args[1] == "update" {
		if len(os.Args) < 3 {
			fmt.Println("No index provided. Usage: update <index> [options]")
			os.Exit(1)
		}
		hasFlag := false // Track if any flags are provided
		for _, arg := range os.Args[3:] {
			if strings.HasPrefix(arg, "-") {
				hasFlag = true
				break
			}
		}
		if !hasFlag {
			fmt.Println("No flags provided. Usage: update <index> [options]")
			os.Exit(1)
		}
		indexStr := os.Args[2]
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 || index >= len(ToDos) {
			fmt.Println("Invalid index provided.")
			os.Exit(1)
		}
		flag.CommandLine.Parse(os.Args[3:])
		title := *titlePtr
		due := *duePtr
		completed := *completedPtr
		item := ToDos[index]
		if item.Name != title && title != "" {
			item.Name = title
		}
		for _, arg := range os.Args[3:] {
			if strings.HasPrefix(arg, "-d") {
				item.Due = due
			}
		}
		for _, arg := range os.Args[3:] {
			if strings.HasPrefix(arg, "-c") {
				item.Completed = completed
			}
		}
		ToDos[index] = item
		fmt.Println("Item updated.")
		err = SaveToDos(ToDos)
		if err != nil {
			fmt.Printf("An error occurred while saving: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Unknown command. Available commands: add, list, update")
		os.Exit(1)
	}

	err = SaveToDos(ToDos)
	if err != nil {
		fmt.Printf("An error occurred:%s\n", err)
		os.Exit(1)
	}

}
