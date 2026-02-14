package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"user-auth-go/constants"
	"user-auth-go/internal/config"
	"user-auth-go/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest = LoginRequest

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Telephone    string `json:"telephone"`
	AuthProvider string `json:"auth_provider"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: add email validation
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// TODO: add another various checking for better password validation
	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	user, err := models.CreateUser(req.Email, req.Password)

	if err != nil {
		if errors.Is(err, models.ErrEmailExists) {
			respondError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	session, err := models.CreateSession(user.ID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sess_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	respondSuccess(w, "Signup Successfull", AuthResponse{
		Token: session.Token,
		User: UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			AuthProvider: user.AuthProvider,
		},
	})
}

// handler login `POST /api/login`
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := models.GetUserByIdentifier(req.Email, "")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		respondError(w, http.StatusUnauthorized, "Username or password is incorrect")
		return
	}

	if user.AuthProvider == constants.AuthProviderGoogle {
		respondError(w, http.StatusBadRequest, "Please login with Google")
		return
	}

	if !user.CheckPassword(req.Password) {
		respondError(w, http.StatusUnauthorized, "Username or password is incorrect")
		return
	}

	session, err := models.CreateSession(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	respondSuccess(w, "Login successful", AuthResponse{
		Token: session.Token,
		User: UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			FullName:     user.FullName,
			Telephone:    user.Telephone,
			AuthProvider: user.AuthProvider,
		},
	})
}

// handler login `POST /api/logout`
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No session found")
		return
	}

	models.DeleteSession(cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	respondSuccess(w, "Logout successful", nil)
}

// handler login `GET /api/auth/google`
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := config.GoogleOAuthConfig.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handler login `GET /api/auth/google/callback`
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error=Code not found", http.StatusTemporaryRedirect)
		return
	}

	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/login?error=Failed to exchange token", http.StatusTemporaryRedirect)
		return
	}

	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Redirect(w, r, "/login?error=Failed to get user info", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	json.Unmarshal(body, &googleUser)

	user, err := models.GetUserByIdentifier("", googleUser.ID)
	if err != nil {
		http.Redirect(w, r, "/login?error=Failed to get user", http.StatusTemporaryRedirect)
		return
	}

	if user == nil {
		user, err = models.CreateUserWithGoogle(googleUser.Email, googleUser.ID, googleUser.Name)
		if err != nil {
			if errors.Is(err, models.ErrEmailExists) {
				http.Redirect(w, r, "/login?error=Email already registered", http.StatusTemporaryRedirect)
				return
			}
			http.Redirect(w, r, "/login?error=Failed to create user", http.StatusTemporaryRedirect)
			return
		}
	}

	session, err := models.CreateSession(user.ID)
	if err != nil {
		http.Redirect(w, r, "/login?error=Failed to create session", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}
