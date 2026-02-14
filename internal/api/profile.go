package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"user-auth-go/constants"
	"user-auth-go/internal/models"
)

type UpdateProfileRequest struct {
	FullName  string `json:"full_name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

type ProfileResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Telephone    string `json:"telephone"`
	AuthProvider string `json:"auth_provider"`
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user := GetUserFromCtx(r)

	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	respondSuccess(w, "Profile Retrieved", ProfileResponse{
		ID:           user.ID,
		Email:        user.Email,
		FullName:     user.FullName,
		Telephone:    user.Telephone,
		AuthProvider: user.AuthProvider,
	})
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user := GetUserFromCtx(r)

	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
	}

	var req UpdateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: move validation to struct users
	if req.FullName == "" {
		respondError(w, http.StatusBadRequest, "Full name is required")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if user.AuthProvider == constants.AuthProviderGoogle && req.Email != user.Email {
		respondError(w, http.StatusBadRequest, "Cannot change email for Google account")
		return
	}

	// handle update user profile
	err := models.UpdateUserProfile(user.ID, req.FullName, req.Telephone, req.Email)

	if err != nil {
		if errors.Is(err, models.ErrEmailExists) {
			respondError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// handle get newest user id
	// TODO: make sure is my sql support returning value?
	updatedUser, err := models.GetUserByID(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated profile")
		return
	}

	respondSuccess(w, "Profile updated", ProfileResponse{
		ID:           updatedUser.ID,
		Email:        updatedUser.Email,
		FullName:     updatedUser.FullName,
		Telephone:    updatedUser.Telephone,
		AuthProvider: updatedUser.AuthProvider,
	})

}

func Profile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetProfile(w, r)
	case http.MethodPut:
		UpdateProfile(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
