package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"user-auth-go/constants"
	"user-auth-go/internal/config"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	Password     string
	FullName     string
	Telephone    string
	AuthProvider string
	GoogleID     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

var (
	ErrEmailExists = errors.New("email already exists")
)

// handle create user with provider type is `local`
func CreateUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	result, err := config.DB.Exec("INSERT INTO users (email, password, auth_provider) VALUES (?, ? , ?)",
		email, string(hashedPassword), constants.AuthProviderLocal)

	if err != nil {
		if isDuplicateEntryError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &User{ID: int(id), Email: email, AuthProvider: constants.AuthProviderLocal}, nil
}

// handle create user with provider type is `google`
func CreateUserWithGoogle(email, googleID, fullName string) (*User, error) {
	result, err := config.DB.Exec("INSERT INTO users (email, google_id, fullname, auth_provider)",
		email, googleID, fullName, constants.AuthProviderGoogle)

	if err != nil {
		if isDuplicateEntryError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &User{ID: int(id), Email: email, AuthProvider: constants.AuthProviderGoogle}, nil
}

// handle get user from theirs ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	err := config.DB.QueryRow(
		"SELECT id, email, COALESCE(password, ''), COALESCE(full_name, ''), COALESCE(telephone, ''), auth_provider, COALESCE(google_id, '') FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.FullName, &user.Telephone, &user.AuthProvider, &user.GoogleID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// handle get user based on identifier
// identifier can be email or googleID
func GetUserByIdentifier(email, googleID string) (*User, error) {
	var condition string
	var arg interface{}

	if googleID != "" {
		condition = "google_id = ?"
		arg = googleID
	} else {
		condition = "email = ?"
		arg = email
	}

	user := &User{}
	err := config.DB.QueryRow(
		"SELECT id, email, COALESCE(password, ''), COALESCE(full_name, ''), COALESCE(telephone, ''), auth_provider, COALESCE(google_id, '') FROM users WHERE "+condition,
		arg,
	).Scan(&user.ID, &user.Email, &user.Password, &user.FullName, &user.Telephone, &user.AuthProvider, &user.GoogleID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// handle update user profile
func UpdateUserProfile(id int, fullName, telephone, email string) error {
	_, err := config.DB.Exec(
		"UPDATE users SET full_name = ?, telephone = ?, email = ? WHERE id = ?",
		fullName, telephone, email, id,
	)
	if err != nil {
		if isDuplicateEntryError(err) {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

// handle catch the error, with duplicate constraint error
// with this we don't need to check manually is email is already exists or not
func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError

	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	return strings.Contains(err.Error(), "Duplicate entry")
}
