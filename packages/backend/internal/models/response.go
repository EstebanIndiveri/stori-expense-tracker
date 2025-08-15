package models

import (
	"net/http"
	"time"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an error in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIMeta contains metadata for API responses
type APIMeta struct {
	Total      int    `json:"total,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"perPage,omitempty"`
	TotalPages int    `json:"totalPages,omitempty"`
	HasMore    bool   `json:"hasMore,omitempty"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Standard error codes
const (
	ErrorCodeValidation      = "VALIDATION_ERROR"
	ErrorCodeNotFound        = "NOT_FOUND"
	ErrorCodeInternalServer  = "INTERNAL_SERVER_ERROR"
	ErrorCodeUnauthorized    = "UNAUTHORIZED"
	ErrorCodeForbidden       = "FORBIDDEN"
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeTooManyRequests = "TOO_MANY_REQUESTS"
)

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}, meta *APIMeta) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC(),
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}
}

// HTTPStatusFromErrorCode returns appropriate HTTP status code for error codes
func HTTPStatusFromErrorCode(code string) int {
	switch code {
	case ErrorCodeValidation, ErrorCodeBadRequest:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrorCodeInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
