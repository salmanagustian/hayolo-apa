package main

import (
	"user-auth-go/internal/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// initialize config
	config.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server is running!")
	})

	// initialize server
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
