package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"user-auth-go/internal/api"
	"user-auth-go/internal/config"
)

func main() {
	// initialize config
	config.Init()

	// register public auth handlers
	http.HandleFunc("/api/signup", api.Signup)
	http.HandleFunc("/api/login", api.Login)
	http.HandleFunc("/api/logout", api.Logout)
	http.HandleFunc("/api/auth/google", api.GoogleLogin)
	http.HandleFunc("/api/auth/google/callback", api.GoogleCallback)

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
