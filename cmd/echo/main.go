package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Got request", "from", r.RemoteAddr)
	rBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received: %s\n", rBody)

	w.Write(rBody)
}

func main() {
	http.HandleFunc("/", echoHandler) // Set the handler function for the root path

	log.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil) // Start the server on port 8080
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
