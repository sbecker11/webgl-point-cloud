// server.go
package main

// build: go build -o server server.go
// run: ./server

import (
    "fmt"
    "net/http"
)

func main() {
    // configure the server to serve files from the current directory
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)

    // server configured to listen on port 8080
    fmt.Println("Server running at http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Server error:", err)
    }
}
