package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Data structure for the greeting template
type GreetingData struct {
	Name string
}

// Data structure for the login template
type LoginData struct {
	Message string
}

// Database initialization function
func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./webapp.db")
	if err != nil {
		return nil, err
	}

	// Create the users table if it doesn't exist
	_, err = db.Exec(`
		drop table if exists users;
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        )
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Function to insert a user into the database
func insertUser(db *sql.DB, username, password string) error {
	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
	// fmt.Printf("err:%s", err)
	return err
}

// Function to query all users from the database
func queryUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT username FROM users")
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

// Function to verify login credentials
func verifyLogin(db *sql.DB, username, password string) bool {
	// Retrieve the hashed password from the database
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
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
		err := insertUser(db, name, "password") // Default password for demonstration
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

	// Define a handler function for the login page
	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is a POST request
		if r.Method == http.MethodPost {
			// Parse the form data
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Get the username and password from the form
			username := strings.TrimSpace(r.Form.Get("username"))
			password := r.Form.Get("password")

			// Verify login credentials
			if verifyLogin(db, username, password) {
				// Redirect to the greeting page with the username as a query parameter
				http.Redirect(w, r, "/greet?name="+username, http.StatusFound)
				return
			}

			// Display an error message on the login page
			tmpl, err := template.ParseFiles("templates/login.html")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Execute the template with the error message
			err = tmpl.Execute(w, LoginData{Message: "Invalid username or password"})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

		// Parse the HTML template for the login page
		tmpl, err := template.ParseFiles("templates/login.html")
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

	// Register the handlers for the home, greet, and login pages
	router.HandleFunc("/", homeHandler).Methods(http.MethodGet)
	router.HandleFunc("/greet", greetHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", loginHandler).Methods(http.MethodGet, http.MethodPost)

	// Serve static files from the "static" directory
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the HTTP server on port 8080 using the router
	fmt.Println("Server is listening on :8080...")
	http.ListenAndServe(":8080", router)
}
