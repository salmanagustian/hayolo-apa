package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"user-auth-go/internal/api"
	"user-auth-go/internal/config"
	"user-auth-go/internal/models"
)

func TestMain(m *testing.M) {
	// ggt current working directory
	cwd, _ := os.Getwd()

	// find env file based on root dir
	rootDir := findRootDir(cwd)
	if rootDir != "" {
		os.Chdir(rootDir)
	}

	// init config
	config.Init()

	// run tests
	code := m.Run()

	os.Exit(code)
}

func findRootDir(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// tests login flow
func TestLoginSuccess(t *testing.T) {
	// create user first
	models.CreateUser("login_test@example.com", "password123")

	body := map[string]string{
		"email":    "login_test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.Login)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response api.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Success {
		t.Errorf("Expected success true, got false")
	}

	// cleanup
	config.DB.Exec("DELETE FROM users WHERE email = ?", "login_test@example.com")
}

// tests signup flow
func TestSignupSuccess(t *testing.T) {
	// cleanup
	config.DB.Exec("DELETE FROM users WHERE email = ?", "test_signup@example.com")

	body := map[string]string{
		"email":    "test_signup@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.Signup)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response api.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if !response.Success {
		t.Errorf("Expected success true, got false: %s", response.Message)
	}

	// cleanup
	config.DB.Exec("DELETE FROM users WHERE email = ?", "test_signup@example.com")
}

// tests wrong password
func TestLoginWrongPassword(t *testing.T) {
	// cleanup
	config.DB.Exec("DELETE FROM users WHERE email = ?", "wrong_pass@example.com")

	// create user first
	models.CreateUser("wrong_pass@example.com", "password123")

	body := map[string]string{
		"email":    "wrong_pass@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.Login)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var response api.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if response.Message != "Username or password is incorrect" {
		t.Errorf("Expected 'Password is incorrect', got '%s'", response.Message)
	}

	// cleanup
	config.DB.Exec("DELETE FROM users WHERE email = ?", "wrong_pass@example.com")
}
