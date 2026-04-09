// Package dto contains data transfer objects for the User Service API.
//
// DTOs define the structure of request and response payloads for API endpoints.
// They handle serialization/deserialization with JSON and provide type safety
// for HTTP communication.
package dto

// ErrorResponse represents an error response from the API.
// Used for all error HTTP responses to provide consistent error formatting.
//
// swagger:model
type ErrorResponse struct {
	// Error message or description
	Errors string `json:"errors" example:"unauthorized"`
}
