package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	healthCheckDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "url_health_check_duration_seconds",
		Help:    "Duration of health check requests in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"target_id", "url"})

	healthCheckStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "url_health_check_status",
		Help: "Status of health check (1 for healthy, 0 for unhealthy)",
	}, []string{"target_id", "url"})

	healthCheckStatusCode = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "url_health_check_status_code",
		Help: "HTTP status code from health check",
	}, []string{"target_id", "url"})

	healthCheckErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "url_health_check_errors_total",
		Help: "Total number of health check errors",
	}, []string{"target_id", "url", "error_type"})

	registeredTargets = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "url_registered_targets_total",
		Help: "Total number of registered targets",
	})
)
