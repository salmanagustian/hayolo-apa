package api

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, APIResponse{
		Success: false,
		Message: message,
	})
}

func respondSuccess(w http.ResponseWriter, message string, data interface{}) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
