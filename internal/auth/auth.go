// Package auth provides API key extraction utilities.
// internal/auth/auth.go:
package auth

import (
	"errors"
	"net/http"
	"strings"
)

// ErrNoAuthHeaderIncluded is a custom error returned when the Authorization
// header is missing from the request. This allows callers to handle this
// specific case distinctly.
var ErrNoAuthHeaderIncluded = errors.New("no authorization header included")

// GetAPIKey extracts the API key from the HTTP request headers.
//
// It expects the "Authorization" header in the format "ApiKey <key>".
// If the header is missing, it returns ErrNoAuthHeaderIncluded.
// If the format is invalid (e.g., wrong prefix or missing key), it returns
// a "malformed authorization header" error.
// On success, it returns the extracted key and nil error.
//
// Parameters:
//
//	headers (http.Header): The HTTP headers from the request.
//
// Returns:
//
//	string: The extracted API key.
//	error: Any error encountered during extraction.
func GetAPIKey(headers http.Header) (string, error) {
	// Retrieve the Authorization header value.
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	} // Header is missing; return the predefined error.
	// Split the header by space to separate prefix and key.
	splitAuth := strings.Split(authHeader, " ")
	// Check for exactly two parts and correct "ApiKey" prefix.
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		// Invalid format; return a descriptive error.
		return "", errors.New("malformed authorization header")
	}

	// Valid header; return the key (second part).
	return splitAuth[1], nil
}
