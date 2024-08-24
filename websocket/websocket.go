package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    fmt.Println("Client connected to WebSocket")

    // Infinite loop to handle incoming messages
    for {
        // Read message from the WebSocket connection
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            fmt.Println("Error reading message:", err)
            return
        }

        // Print the received message to the console
        fmt.Printf("Received message: %s\n", p)

        // Broadcast the received message to all connected clients
        err = conn.WriteMessage(messageType, p)
        if err != nil {
            fmt.Println("Error broadcasting message:", err)
            return
        }
    }
}

func main() {
    // Create a new router
    router := mux.NewRouter()

    // WebSocket endpoint
    router.HandleFunc("/ws", handleWebSocket)

    // Serve static files from the "static" directory
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Start the HTTP server on port 8080 using the router
    fmt.Println("Server is listening on :8080...")
    http.ListenAndServe(":8080", router)
}