package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Task structure for the API
type Task struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

// Database initialization function
func initDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./webapp.db")
    if err != nil {
        return nil, err
    }

    // Create the tasks table if it doesn't exist
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            done BOOLEAN NOT NULL
        )
    `)
    if err != nil {
        return nil, err
    }

    return db, nil
}

// Function to retrieve all tasks from the database
func getTasks(db *sql.DB) ([]Task, error) {
    rows, err := db.Query("SELECT id, title, done FROM tasks")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []Task
    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.ID, &task.Title, &task.Done); err != nil {
            return nil, err
        }
        tasks = append(tasks, task)
    }

    return tasks, nil
}

// Function to create a new task
func createTask(db *sql.DB, title string) (int, error) {
    result, err := db.Exec("INSERT INTO tasks (title, done) VALUES (?, ?)", title, false)
    if err != nil {
        return 0, err
    }

    // Get the ID of the newly created task
    id, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }

    return int(id), nil
}

// Function to update a task's status
func updateTaskStatus(db *sql.DB, id int, done bool) error {
    _, err := db.Exec("UPDATE tasks SET done = ? WHERE id = ?", done, id)
    return err
}

// Function to delete a task
func deleteTask(db *sql.DB, id int) error {
    _, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
    return err
}

func main() {
    // Initialize the database
    db, err := initDB()
    if err != nil {
        fmt.Println("Error initializing database:", err)
        return
    }
    defer db.Close()

    // Create a new router
    router := mux.NewRouter()

    // API endpoint to retrieve all tasks
    router.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
        tasks, err := getTasks(db)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        // Convert tasks to JSON and send the response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(tasks)
    }).Methods(http.MethodGet)

    // API endpoint to create a new task
    router.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
        // Parse the JSON request body
        var task Task
        err := json.NewDecoder(r.Body).Decode(&task)
        if err != nil {
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        // Create the task in the database
        id, err := createTask(db, task.Title)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        // Send the ID of the created task in the response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]int{"id": id})
    }).Methods(http.MethodPost)

    // API endpoint to update the status of a task
    router.HandleFunc("/api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
        // Extract the task ID from the request URL
        vars := mux.Vars(r)
        id := vars["id"]

        // Parse the JSON request body
        var status struct {
            Done bool `json:"done"`
        }
        err := json.NewDecoder(r.Body).Decode(&status)
        if err != nil {
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        // Convert the task ID to an integer
        taskId, err := strconv.Atoi(id)
        if err != nil {
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        // Update the task status in the database
        err = updateTaskStatus(db, taskId, status.Done)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    }).Methods(http.MethodPut)

    // API endpoint to delete a task
    router.HandleFunc("/api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
        // Extract the task ID from the request URL
        vars := mux.Vars(r)
        id := vars["id"]

        // Convert the task ID to an integer
        taskId, err := strconv.Atoi(id)
        if err != nil {
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        // Delete the task from the database
        err = deleteTask(db, taskId)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    }).Methods(http.MethodDelete)

    // Start the HTTP server on port 8080 using the router
    fmt.Println("Server is listening on :8080...")
    http.ListenAndServe(":8080", router)
}