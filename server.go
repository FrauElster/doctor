package main

import (
	"encoding/json"
	"log/slog"
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
		respondError(w, r, ErrParseJsonBody(err.Error(), err))
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(target.Url)
	if err != nil {
		respondError(w, r, ErrInvalidUrl("", err))
		return
	}

	apiErr := s.checker.AddTarget(HealthTarget{
		URL:       parsedURL,
		URLString: target.Url,
		ID:        target.Id,
	})
	if apiErr != nil {
		respondError(w, r, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) UnregisterTarget(w http.ResponseWriter, r *http.Request, id string) {
	apiErr := s.checker.RemoveTarget(id)
	if apiErr != nil {
		respondError(w, r, apiErr)
		return
	}

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

	respondJSON(w, r, http.StatusOK, jsonResults)
}

func respondError(w http.ResponseWriter, r *http.Request, error *ApiError) {
	slog.Error("unhandled error", "method", r.Method, "url", r.URL, "error", error.Error, "origin", error.Origin)
	w.WriteHeader(error.Status)
	_ = json.NewEncoder(w).Encode(Error{Code: error.Code, Message: error.Message})
}

func respondJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		respondError(w, r, ErrEncodeJsonBody("", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}
