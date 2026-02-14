package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"user-auth-go/internal/config"
)

func main() {
	// initialize config
	config.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// initialize server
	fmt.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
