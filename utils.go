package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func env(key, fallback string) string {
	if val, hasVal := os.LookupEnv(key); hasVal {
		return val
	}
	return fallback
}

func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w\n", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w\n", err)
	}
	return v, nil
}
