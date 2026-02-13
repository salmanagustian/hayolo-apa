package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var DB *sql.DB
var GoogleAuthConfig *oauth2.Config

func Init() {
	initDB()
	initGoogleOAuth()
}

func initDB() {
	var err error

	dsn := "ROOT:SOMEPASSWORD@tcp(127.0.0.1:3306)/user_auth_db?parseTime=true"
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
		ClientID:     "SOME CLIENT ID",
		ClientSecret: "SOME CLIENT SECRET",
		RedirectURL:  "SOME URL CALLBACK",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint}

	fmt.Println("Google OAuth configured!")

}
