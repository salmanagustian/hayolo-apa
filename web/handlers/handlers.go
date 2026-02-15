package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"user-auth-go/internal/models"
)

var templates map[string]*template.Template

func Init() {
	templates = make(map[string]*template.Template)

	pages := []string{"login", "signup", "profile", "profile_edit"}
	
	for _, page := range pages {
		templates[page] = template.Must(template.ParseFiles(
			filepath.Join("web", "templates", "base.html"),
			filepath.Join("web", "templates", page+".html"),
		))
	}
}

type PageData struct {
	Title string
	Error string
	User  *models.User
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// handle render templates
func render(w http.ResponseWriter, page string, data PageData) {
	tmpl, ok := templates[page]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GET /login
func LoginPage(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	setNoCacheHeaders(w)

	error := r.URL.Query().Get("error")
	render(w, "login", PageData{
		Title: "Login",
		Error: error,
	})
}

// GET /signup
func SignupPage(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	setNoCacheHeaders(w)

	render(w, "signup", PageData{
		Title: "Sign Up",
	})
}

// GET /profile
func ProfilePage(w http.ResponseWriter, r *http.Request) {
	user := getAuthenticatedUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	setNoCacheHeaders(w)

	render(w, "profile", PageData{
		Title: "Profile",
		User:  user,
	})
}

// GET /profile/edit
func ProfileEditPage(w http.ResponseWriter, r *http.Request) {
	user := getAuthenticatedUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	setNoCacheHeaders(w)

	render(w, "profile_edit", PageData{
		Title: "Edit Profile",
		User:  user,
	})
}


// simple helper to check state user authenticated
func isAuthenticated(r *http.Request) bool {
	return getAuthenticatedUser(r) != nil
}

// handle to check current authenticated user
func getAuthenticatedUser(r *http.Request) *models.User {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	session, err := models.GetSessionByToken(cookie.Value)
	if err != nil || session == nil {
		return nil
	}

	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		return nil
	}

	return user
}
