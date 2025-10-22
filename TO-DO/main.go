package main

import (
	"GoAcademy/TO-DO/todo"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	currentTime := time.Now()
	var err error
	todo.ToDos, err = todo.LoadToDos()
	if err != nil {
		fmt.Printf("An error occurred:%s\n", err)
		os.Exit(1)
	}
	fmt.Println("Current To-Do List:")
	fmt.Println(todo.ToDos)
	fmt.Println("-----")
	fmt.Println("Available commands: add, list, update, delete,help")

	titlePtr := flag.String("t", "", "a title for the to-do item")
	duePtr := flag.String("d", currentTime.Format("02-01-2006"), "due date for the to-do item")
	completedPtr := flag.Bool("c", false, "mark the to-do item as completed")

	if len(os.Args) < 2 {
		fmt.Println("No command provided. Available commands: add")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
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
		todo.AddToDo(title, due)
		err = todo.SaveToDos(todo.ToDos)
		if err != nil {
			fmt.Printf("An error occurred while saving: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Item added to list.")
		// Handle "add" command
		// e.g., parse flags and add a to-do item
		os.Exit(0)

	case "list":
		fmt.Println(todo.ToDos)
		os.Exit(0)
	case "update":
		if len(os.Args) < 3 {
			fmt.Println("Invalid command. Usage: update <index> [options]")
			os.Exit(1)
		}

		indexStr := os.Args[2]
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 || index >= len(todo.ToDos) {
			fmt.Println("Invalid index provided.")
			os.Exit(1)
		}

		flag.CommandLine.Parse(os.Args[3:])
		if flag.NFlag() == 0 {
			fmt.Println("No flags provided. Usage: update <index> [options]")
			os.Exit(1)
		}

		title := *titlePtr
		due := *duePtr
		completed := *completedPtr
		item := todo.ToDos[index]
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
		todo.ToDos[index] = item
		fmt.Println("Item updated.")
		err = todo.SaveToDos(todo.ToDos)
		if err != nil {
			fmt.Printf("An error occurred while saving: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Invalid command. Usage: delete <index>")
			os.Exit(1)
		}
		indexStr := os.Args[2]
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 || index >= len(todo.ToDos) {
			fmt.Println("Invalid index provided.")
			os.Exit(1)
		}
		todo.RemoveToDo(index)
		err = todo.SaveToDos(todo.ToDos)
		if err != nil {
			fmt.Printf("An error occurred while saving: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Item deleted from list.")
		os.Exit(0)
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  add    -t <title> -d <due date>       Add a new to-do item")
		fmt.Println("  list                                 List all to-do items")
		fmt.Println("  update <index> [-t <title>] [-d <due date>] [-c <completed>]   Update a to-do item")
		fmt.Println("  delete <index>                      Delete a to-do item")
		fmt.Println("  help                                Show this help message")
		os.Exit(0)
	default:
		fmt.Println("No valid command provided. Available commands: add, list, update")
		os.Exit(1)
	}

	err = todo.SaveToDos(todo.ToDos)
	if err != nil {
		fmt.Printf("An error occurred:%s\n", err)
		os.Exit(1)
	}

}
