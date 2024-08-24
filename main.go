package main

import (
    "html/template"
    "net/http"
    "fmt"
    "github.com/gorilla/mux"
)

// Data structure for the greeting template
type GreetingData struct {
    Name string
}

func main() {
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

        // Create a GreetingData instance
        data := GreetingData{Name: name}

        // Parse the HTML template
        tmpl, err := template.ParseFiles("templates/greeting.html")
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        // Execute the template with the data
        err = tmpl.Execute(w, data)
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