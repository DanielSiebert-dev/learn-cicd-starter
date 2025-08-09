// This file provides helper functions for sending JSON responses in a web app. It handles success responses (respondWithJSON) and error responses (respondWithError), with logging for server-side errors. The overall flow is:
// 1. For errors: Log if needed, create an error JSON, and send it.
// 2. For success: Marshal data to JSON, set headers, write response, handle any errors.
// This is used in the main app to return API data or errors securely.

package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string, logErr error) {
	if logErr != nil {
		log.Println(logErr) // Log any incoming error.
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg) // Log server-side errors (5XX).
	}
	type errorResponse struct {
		Error string `json:"error"` // Structure for JSON error response.
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json") // Set JSON header.
	dat, err := json.Marshal(payload)                  // Convert payload to JSON.
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err) // Log marshalling error.
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code) // Set HTTP status code.
	if _, err := w.Write(dat); err != nil {
		log.Printf("Error writing response: %v", err) // Fix G104: Handle write error with detailed log (%v for full err).
	}
}
