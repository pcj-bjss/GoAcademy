# GoAcademy TO-DO App

A concurrent, RESTful To-Do list application built with Go. This project demonstrates key Go concepts including the Actor model (CSP) for state management, structured logging, and graceful shutdown.

## Features

*   **REST API**: JSON endpoints for Create, Read, Update, and Delete operations.
*   **Web Interface**:
    *   **List View**: A dynamic HTML page to view and complete tasks (`/list`).
    *   **About Page**: A static page with a form to add new tasks (`/about/`).
*   **Concurrency**: Uses the Actor pattern (Communicating Sequential Processes) via channels to manage data access safely without explicit mutex locks in the business logic.
*   **Observability**: Implements structured logging using `log/slog` with a custom middleware that attaches a unique `TraceID` to every request and log entry.
*   **Persistence**: Automatically saves tasks to a local JSON file (`todos.json`) upon modification and shutdown.
*   **Graceful Shutdown**: Listens for OS signals (SIGINT/SIGTERM) to close the HTTP server and ensure data is flushed to disk before exiting.

## Getting Started

### Prerequisites

*   Go 1.21 or higher.

### Running the Application

1.  Navigate to the project directory.
2.  Run the application:
    ```bash
    go run main.go
    ```
3.  The server will start on `http://localhost:8080`.

## Usage

### Web Interface

*   **View List**: Open http://localhost:8080/list to see your tasks.
*   **Add Tasks**: Open http://localhost:8080/about/ to access the "Add To-Do item" form.

### API Endpoints

You can interact with the API directly using `curl` or other HTTP clients.

#### 1. Create a Task
**POST** `/create`

Requires a JSON body with `Name` and `Due` (format: DD-MM-YYYY).

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"Name": "Book taxi", "Due": "27-12-2025"}' \
     http://localhost:8080/create
```

#### 2. Get All Tasks
**GET** `/get`

Returns a JSON list of all items.

```bash
curl http://localhost:8080/get
```

#### 3. Update a Task
**PATCH** `/update`

Accepts a JSON body with the `id` and fields to update (`name`, `due`, or `completed`).

```bash
curl -X PATCH -H "Content-Type: application/json" \
     -d '{"id": 1, "completed": true}' \
     http://localhost:8080/update
```

#### 4. Delete a Task
**DELETE** `/delete`

Requires the `id` as a query parameter.

```bash
curl -X DELETE "http://localhost:8080/delete?id=1"
```

## Testing

Unit tests are included for the core logic. Run them using:

```bash
go test ./todo/...
```

### Concurrent Safety Tests 
The application includes specific tests to validate the Actor model's ability to handle concurrent access safely. These tests use t.Parallel() to spawn multiple goroutines that simultaneously interact with the Store channel. 
* TestConcurrentAccess: Spawns 50 workers to add items simultaneously. 
* TestConcurrentReads: Spawns 50 workers to read the list simultaneously. 
* TestConcurrentUpdates: Spawns 50 workers to update specific items simultaneously. 

To show concurrency pattern in the test, run with verbose enabled:

```bash
go test -v ./todo/... 
```

### Benchmarking 
Benchmark tests are included to measure the performance of adding items directly vs. via the Actor model. 

To run the benchmarks: 
```bash 
go test -bench=. -benchmem ./todo/... 
```

### Load Testing 
A load testing application is included in the todo_loadtest directory to stress test the server. 

1. Start the server:

```bash
go run main.go
```
2. Run the load tester (in a separate terminal):

```bash
go run loadtest/main.go
```