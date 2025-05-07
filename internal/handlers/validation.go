package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
)

// encode encodes a response to JSON
func encode(w http.ResponseWriter, r *http.Request, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// decode decodes a request body into the provided type
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// decodeValid decodes and validates the request body
func decodeValid[T dto.Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	
	problems := v.Valid(r.Context())
	if len(problems) > 0 {
		return v, problems, fmt.Errorf("validation failed with %d problems", len(problems))
	}
	
	return v, nil, nil
}