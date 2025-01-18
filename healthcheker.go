package main

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"gitlab.com/tozd/go/errors"
)

var (
	ErrTargetNotFound = errors.New("target not found")
)

// HealthTarget represents a URL to be monitored
type HealthTarget struct {
	URL       *url.URL `json:"-"`
	URLString string   `json:"url"`
	ID        string   `json:"id"`
}

// Result represents the health check result
type Result struct {
	Target    HealthTarget
	Status    int
	Healthy   bool
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// HealthChecker manages the health checking process
type HealthChecker struct {
	client  *http.Client
	mu      sync.RWMutex
	targets map[string]HealthTarget
}

// NewHealthChecker creates a new HealthChecker instance
func NewHealthChecker(timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		targets: make(map[string]HealthTarget),
		client:  &http.Client{Timeout: timeout},
	}
}

// AddTarget adds a new target to the health checker
func (hc *HealthChecker) AddTarget(target HealthTarget) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.targets[target.ID] = target
	registeredTargets.Inc()
}

// RemoveTarget removes a target from the health checker
func (hc *HealthChecker) RemoveTarget(id string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.targets, id)
	registeredTargets.Dec()
}

// checkHealth performs the health check for a single target
func (hc *HealthChecker) checkHealth(target HealthTarget) Result {
	startTime := time.Now()
	result := Result{
		Target:    target,
		Timestamp: startTime,
	}

	resp, err := hc.client.Get(target.URL.String())
	result.Duration = time.Since(startTime)

	// Record request duration
	healthCheckDuration.WithLabelValues(target.ID, target.URLString).
		Observe(result.Duration.Seconds())

	if err != nil {
		result.Error = err
		// Record error
		healthCheckErrors.WithLabelValues(
			target.ID,
			target.URLString,
			"connection_error",
		).Inc()
		// Update status gauge to unhealthy
		healthCheckStatus.WithLabelValues(target.ID, target.URLString).Set(0)
		return result
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode
	result.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 300

	// Update Prometheus metrics
	healthCheckStatusCode.WithLabelValues(target.ID, target.URLString).
		Set(float64(result.Status))

	if result.Healthy {
		healthCheckStatus.WithLabelValues(target.ID, target.URLString).Set(1)
	} else {
		healthCheckStatus.WithLabelValues(target.ID, target.URLString).Set(0)
		healthCheckErrors.WithLabelValues(
			target.ID,
			target.URLString,
			"unhealthy_status",
		).Inc()
	}

	return result
}

// CheckTarget performs a health check on a single target
func (hc *HealthChecker) CheckTarget(id string) (Result, error) {
	hc.mu.RLock()
	target, ok := hc.targets[id]
	hc.mu.RUnlock()

	if !ok {
		return Result{}, ErrTargetNotFound
	}

	return hc.checkHealth(target), nil
}

// CheckAll performs health checks on all targets concurrently
func (hc *HealthChecker) CheckAll() []Result {
	hc.mu.RLock()
	targets := MapValues(hc.targets)
	hc.mu.RUnlock()

	results := make([]Result, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(index int, t HealthTarget) {
			defer wg.Done()
			results[index] = hc.checkHealth(t)
		}(i, target)
	}
	wg.Wait()
	return results
}
