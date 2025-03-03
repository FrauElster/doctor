package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=openapi.config.yml openapi.yml
var _ ServerInterface = &Server{}

type Server struct {
	checker *HealthChecker
}

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) RegisterTarget(w http.ResponseWriter, r *http.Request) {
	var target Target
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(target.Url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	s.checker.AddTarget(HealthTarget{
		URL:       parsedURL,
		URLString: target.Url,
		ID:        target.Id,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetStatus(w http.ResponseWriter, r *http.Request) {
	results := s.checker.CheckAll()

	type JSONResult struct {
		ID              string    `json:"id"`
		URL             string    `json:"url"`
		Status          int       `json:"status"`
		Healthy         bool      `json:"healthy"`
		Timestamp       time.Time `json:"timestamp"`
		DurationSeconds float64   `json:"duration_seconds"`
		Error           *string   `json:"error,omitempty"`
	}

	jsonResults := make([]JSONResult, len(results))
	for i, result := range results {
		jsonResult := JSONResult{
			ID:              result.Target.ID,
			URL:             result.Target.URL.String(),
			Status:          result.Status,
			Healthy:         result.Healthy,
			Timestamp:       result.Timestamp,
			DurationSeconds: result.Duration.Seconds(),
		}

		if result.Error != nil {
			errStr := result.Error.Error()
			jsonResult.Error = &errStr
		}

		jsonResults[i] = jsonResult
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonResults); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
