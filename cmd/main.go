package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"user-auth-go/internal/api"
	"user-auth-go/internal/config"
	"user-auth-go/web/handlers"
)

func main() {
	// initialize config
	config.Init()
	handlers.Init()

	// static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// register web pages routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	http.HandleFunc("/login", handlers.LoginPage)
	http.HandleFunc("/signup", handlers.SignupPage)
	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/profile/edit", handlers.ProfileEditPage)

	// register public and protected API routes
	http.HandleFunc("/api/signup", api.Signup)
	http.HandleFunc("/api/login", api.Login)
	http.HandleFunc("/api/auth/google", api.GoogleLogin)
	http.HandleFunc("/api/auth/google/callback", api.GoogleCallback)

	// protected handlers
	http.HandleFunc("/api/logout", api.AuthGuard(api.Logout))
	http.HandleFunc("/api/profile", api.AuthGuard(api.Profile))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// initialize server
	fmt.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
