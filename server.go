// server.go
package main

// build: go build -o server server.go
// run: ./server

import (
    "fmt"
    "net/http"
)

func main() {
    fs := http.FileServer(http.Dir("wasm"))
    http.Handle("/", fs)

    fmt.Println("Server running at http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Server error:", err)
    }
}
