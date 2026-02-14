package api

import (
	"context"
	"net/http"
	"user-auth-go/internal/models"
)

type ContextKey  string

const UserCtxKey ContextKey = "user"

func AuthGuard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")

		if err != nil {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		session, err := models.GetSessionByToken(cookie.Value)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to validate session")
			return
		}

		if session == nil {
			// invalidate session token
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})

			respondError(w, http.StatusUnauthorized, "Session Expired")
			return
		}

		user, err := models.GetUserByID(session.UserID)

		if err != nil || user == nil {
			respondError(w, http.StatusUnauthorized, "User not found")
		}

		// handle to save user to it's context
		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserFromCtx(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserCtxKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
