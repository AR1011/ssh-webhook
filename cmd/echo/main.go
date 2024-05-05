package main

import (
	"io"
	"log"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	// Copy the request body directly to the response body
	_, err := io.Copy(w, r.Body)
	if err != nil {
		http.Error(w, "Failed to echo request body", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", echoHandler) // Set the handler function for the root path

	log.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil) // Start the server on port 8080
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
