package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
	"user-auth-go/internal/config"
)

type Session struct {
	ID int
	UserID int
	Token string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// handle generateToken function
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	
	_, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// handle create session for user after login
func CreateSession(userID int) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	result, err := config.DB.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Session{
		ID:        int(id),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// handle get session logged user by token
func GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	err := config.DB.QueryRow(
		"SELECT id, user_id, token, expires_at FROM sessions WHERE token = ? AND expires_at > NOW()",
		token,
	).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

// handle delete session logged user by token
func DeleteSession(token string) error {
	_, err := config.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// handle delete session logged user by userID
func DeleteUserSessions(userID int) error {
	_, err := config.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}
