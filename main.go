package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Data structure for the greeting template
type GreetingData struct {
	Name string
}

// Database initialization function
func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./webapp.db")
	if err != nil {
		return nil, err
	}

	// Create the users table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Function to insert a user into the database
func insertUser(db *sql.DB, name string) error {
	_, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)
	return err
}

// Function to query all users from the database
func queryUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
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

	// Define a handler function for the home page
	homeHandler := func(w http.ResponseWriter, r *http.Request) {
		// Parse the HTML template
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Execute the template
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Define a handler function for the greeting page
	greetHandler := func(w http.ResponseWriter, r *http.Request) {
		// Get the "name" query parameter from the URL
		name := r.URL.Query().Get("name")

		// If the name is empty, set a default name
		if name == "" {
			name = "Guest"
		}

		// Insert the user into the database
		err := insertUser(db, name)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Query all users from the database
		users, err := queryUsers(db)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create a GreetingData instance
		data := GreetingData{Name: name}

		// Parse the HTML template
		tmpl, err := template.ParseFiles("templates/greeting.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Execute the template with the data
		err = tmpl.Execute(w, struct {
			GreetingData
			Users []string
		}{data, users})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Register the handlers for the home and greet pages
	router.HandleFunc("/", homeHandler).Methods(http.MethodGet)
	router.HandleFunc("/greet", greetHandler).Methods(http.MethodGet)

	// Serve static files from the "static" directory
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the HTTP server on port 8080 using the router
	fmt.Println("Server is listening on :8080...")
	http.ListenAndServe(":8080", router)
}
