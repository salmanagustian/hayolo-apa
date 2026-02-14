package config

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var DB *sql.DB
var GoogleAuthConfig *oauth2.Config

func Init() {
	loadEnvFile()
	initDB()
	initGoogleOAuth()
}

func loadEnvFile() {
	file, err := os.Open(".env")

	if err != nil {
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			os.Setenv(key, val)
		}
	}

}

func initDB() {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASS", ""),
		getEnv("DB_HOST", "127.0.0.1"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "user_auth_db"),
	)

	DB, err = sql.Open("mysql", dsn)

	if err != nil {
		log.Fatalf("Failed to connect DB: %s", err)
	}

	err = DB.Ping()

	if err != nil {
		log.Fatalf("Failed to ping DB: %s", err)
	}

	fmt.Println("DB connected!")
}

func initGoogleOAuth() {
	GoogleAuthConfig = &oauth2.Config{
		ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint}

	fmt.Println("Google OAuth configured!")

}

func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)

	if val == "" {
		return defaultVal
	}

	return val
}
