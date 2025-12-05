	// if len(os.Args) < 2 {
	// 	slog.Default().Log(
	// 		ctx,
	// 		slog.LevelError,
	// 		"Application terminated due to missing command",
	// 		"reason", "No command verb provided after executable name.",
	// 		"usage_tip", "Available commands: add, list, update, delete, help")
	// 	os.Exit(1)
	// }

	// titlePtr := flag.String("t", "", "a title for the to-do item")
	// duePtr := flag.String("d", currentTime.Format("02-01-2006"), "due date for the to-do item")
	// completedPtr := flag.Bool("c", false, "mark the to-do item as completed")

	// fmt.Println("Current To-Do List:")
	// fmt.Println(todo.ToDos)
	// fmt.Println("-----")
	// fmt.Println("Available commands: add, list, update, delete,help")

	// switch os.Args[1] {
	// case "add":
	// 	flag.CommandLine.Parse(os.Args[2:])
	// 	title := *titlePtr
	// 	due := *duePtr
	// 	_, err := time.Parse("02-01-2006", due)
	// 	if err != nil {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid date format",
	// 			"reason", "Date format must be DD-MM-YYYY",
	// 			"usage_tip", "Provide date in format DD-MM-YYYY",
	// 			"error", err)
	// 		os.Exit(1) // Exit with status code 1 (indicates an error)
	// 	}
	// 	if title == "" {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,

	// 			"Application terminated due to empty title field",
	// 			"reason", "Title cannot be empty",
	// 			"usage_tip", "When adding a to-do item, provide a title using the -t flag")
	// 		os.Exit(1) // Exit with status code 1 (indicates an error)
	// 	}
	// 	todo.AddToDo(title, due)
	// 	err = todo.SaveToDos(filename, todo.ToDos, ctx)
	// 	if err != nil {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Failed to save to-do data, exiting",
	// 			"file", "todos.json",
	// 			"error", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Println("Item added to list.")
	// 	// Handle "add" command
	// 	// e.g., parse flags and add a to-do item
	// 	//os.Exit(0) - removing to allow graceful shutdown handling

	// case "list":
	// 	fmt.Println(todo.ToDos)
	// 	os.Exit(0)

	// case "update":
	// 	if len(os.Args) < 3 {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid update command",
	// 			"reason", "index argument missing or options not provided",
	// 			"usage_tip", "Provide the index of the to-do item to update (first item is index 0) followed by at least one option to update. Usage: update <index> [options]")
	// 		os.Exit(1)
	// 	}

	// 	indexStr := os.Args[2]
	// index, err := strconv.Atoi(indexStr)
	// 	if err != nil || index < 0 || index >= len(todo.ToDos) {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid update command",
	// 			"reason", "index argument missing or invalid",
	// 			"usage_tip", "Provide the index of the to-do item to update in the form of an integer (first item is index 0). Can't be greater than the number of items in the list.",
	// 			"error", err)
	// 		os.Exit(1)
	// 	}

	// 	flag.CommandLine.Parse(os.Args[3:])
	// 	if flag.NFlag() == 0 {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid update command",
	// 			"reason", "No flags provided.",
	// 			"usage_tip", "Provide at least one option to update. Usage: update <index> [options]")
	// 		os.Exit(1)
	// 	}

	// 	item := todo.ToDos[index]
	// 	flag.Visit(func(f *flag.Flag) {
	// 		switch f.Name {
	// 		case "t":
	// 			// Update Title if -t was explicitly set
	// 			item.Name = *titlePtr
	// 		case "d":
	// 			// Update Due Date if -d was explicitly set
	// 			item.Due = *duePtr
	// 		case "c":
	// 			// Update Completed status if -c was explicitly set
	// 			// Note: You must ensure item is mutable (e.g., a pointer or assigned back to the slice)
	// 			item.Completed = *completedPtr
	// 		}
	// 	})

	// 	todo.ToDos[index] = item
	// 	fmt.Println("Item updated.")
	// 	err = todo.SaveToDos(filename, todo.ToDos, ctx)
	// 	if err != nil {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to error saving updated to-do data",
	// 			"file", "todos.json",
	// 			"error", err)
	// 		os.Exit(1)
	// 	}
	// 	//os.Exit(0) - removing to allow graceful shutdown handling
	// case "delete":
	// 	if len(os.Args) < 3 {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid delete command",
	// 			"reason", "index argument missing",
	// 			"usage_tip", "Provide the index of the to-do item to delete (first item is index 0).")
	// 		os.Exit(1)
	// 	}
	// 	indexStr := os.Args[2]
	// 	index, err := strconv.Atoi(indexStr)
	// 	if err != nil || index < 0 || index >= len(todo.ToDos) {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to invalid delete command",
	// 			"reason", "index argument missing or invalid",
	// 			"usage_tip", "Provide the index of the to-do item to delete in the form of an integer (first item is index 0). Can't be greater than the number of items in the list.",
	// 			"error", err)
	// 		os.Exit(1)
	// 	}
	// 	todo.RemoveToDo(index)
	// 	err = todo.SaveToDos(filename, todo.ToDos, ctx)
	// 	if err != nil {
	// 		slog.Default().Log(
	// 			ctx,
	// 			slog.LevelError,
	// 			"Application terminated due to error saving updated to-do data",
	// 			"file", "todos.json",
	// 			"error", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Println("Item deleted from list.")
	// 	//os.Exit(0) - removing to allow graceful shutdown handling
	// case "help":
	// 	fmt.Println("Available commands:")
	// 	fmt.Println("  add    -t <title> -d <due date>       Add a new to-do item")
	// 	fmt.Println("  list                                 List all to-do items")
	// 	fmt.Println("  update <index> [-t <title>] [-d <due date>] [-c <completed>]   Update a to-do item")
	// 	fmt.Println("  delete <index>                      Delete a to-do item")
	// 	fmt.Println("  help                                Show this help message")
	// 	os.Exit(0)
	// default:
	// 	fmt.Println("No valid command provided. Available commands: add, list, update")
	// 	os.Exit(1)
	// }