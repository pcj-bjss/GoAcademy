package todo

import (
	"context"
)

// Define the types of actions our actor can perform.
// Using Global constants to allow other packages to reference these operation types.
type Op int

const (
	OpAdd Op = iota //iota assigns incrementing integers starting from 0
	OpGet
	OpUpdate
	OpDelete
	OpSave
)

// Command is the message we'll send to the actor.
// It includes the operation, the data, and a channel to send the response back.
type Command struct {
	Action Op //holds the operation type, and will be one of the Op constants
	Item   Item
	// UpdatePayload holds pointers for partial updates.
	// This allows us to distinguish between a zero value (e.g., "") and a field that wasn't provided.
	UpdatePayload struct {
		Name      *string
		Due       *string
		Completed *bool
	}
	ID      int
	Ctx     context.Context // Context for managing request-scoped values
	Result  chan any        // Channel to send result back to the caller; the channel is defined as bidirectional so that both sending and receiving are possible; any means any type
	ErrChan chan error      // Channel to send error back to the caller
}

// The Store is our actor. It holds the channels.
// Unbuffered channel for commands meaning the handler will wait until the command is processed; can only send or receive one command at a time; blocking otherwise.
// This is a bidirectional channel, allowing both sending and receiving, but flow will be controlled by the actor's logic.
var Store = make(chan Command)

// StartStore is the function that runs the actor goroutine.
// It will be called once when the application starts.
func StartStore(filename string) {
	go func() {
		// All the actor's logic runs inside the go routine which will execute concurrently, allowing main to continue with executing other functions like initializing the web server.

		// Load initial data. We do this once inside the actor goroutine
		// to ensure no other part of the app can access the list while it's loading.
		var err error
		ToDos, err = LoadToDos(filename, context.Background())
		if err != nil {
			// If loading fails, we can't safely proceed.
			panic(err)
		}
		// If loading succeeds, we can start processing commands.
		// The initial list of ToDos is now available.

		var maxID int
		for _, item := range ToDos {
			if item.ID > maxID {
				maxID = item.ID
			}
		}

		// This is the actor's main loop. It waits for commands on the 'store' channel.
		// Using for and range to continuously listen for incoming commands and also to make sure each
		// command is processed one at a time in the order received.
		for cmd := range Store {
			switch cmd.Action {
			case OpGet:
				// Create a copy to send back, preventing race conditions.
				// The caller gets a snapshot, not a direct reference.
				listCopy := make([]Item, len(ToDos))
				copy(listCopy, ToDos)
				cmd.Result <- listCopy //Sending back on the Result channel that was defined in the Command struct as part of the command message.
			case OpAdd:
				maxID++
				cmd.Item.ID = maxID
				var err error
				ToDos, err = AddToDo(ToDos, cmd.Item.ID, cmd.Item.Name, cmd.Item.Due, cmd.Ctx)
				if err != nil {
					cmd.ErrChan <- err
				} else {
					cmd.Result <- cmd.Item // Acknowledge completion by returning the added item.
				}
			case OpUpdate:
				var err error

				//Need to pass the memory address (&) of the fields to update to prevent situations where a user may not want to
				// update completed (for example) and leaves it blank, which would default to false if not using pointers and addresses.

				ToDos, err = UpdateToDo(ToDos, cmd.ID, cmd.UpdatePayload.Name, cmd.UpdatePayload.Due, cmd.UpdatePayload.Completed, cmd.Ctx)
				if err != nil {
					cmd.ErrChan <- err
				} else {
					var updatedItem Item
					for _, item := range ToDos {
						if item.ID == cmd.ID {
							updatedItem = item
							break
						}
					}
					cmd.Result <- updatedItem
				}
			case OpDelete:
				var err error
				ToDos, err = RemoveToDo(ToDos, cmd.ID, cmd.Ctx)
				if err != nil {
					cmd.ErrChan <- err
				} else {
					cmd.Result <- "success"
				}
			case OpSave:
				// The save operation is now also a message to the actor.
				err := SaveToDos(filename, ToDos, cmd.Ctx)
				if err != nil {
					cmd.ErrChan <- err
				} else {
					cmd.Result <- "success"
				}
			}
		}
	}()
}
