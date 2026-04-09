package handlers

import (
	"encoding/json"
	"net/http"
	"user-service/internal/dto"
)

// writeJSON writes a JSON response with the given status code and payload.
// It sets the Content-Type header to application/json before writing the status and body.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes an error response as JSON with the given status code and error message.
// It uses the ErrorResponse DTO to format the error in a consistent manner.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Errors: message})
}
