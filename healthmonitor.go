package main

import (
	"log/slog"
	"sync"
	"time"
)

// AlertFunc is called when a target's health state changes
type AlertFunc func(target HealthTarget, result Result) error

// HealthMonitor manages periodic health checks and alerts
type HealthMonitor struct {
	checker      *HealthChecker
	interval     time.Duration
	alertFuncs   []AlertFunc
	resolveFuncs []AlertFunc
	stopChan     chan struct{}
	stateMap     map[string]monitorState
	stateMu      sync.RWMutex
}

type monitorState struct {
	consecutiveFailures int
	alerted             bool
	lastResult          Result
}

// NewHealthMonitor creates a new HealthMonitor instance
func NewHealthMonitor(
	checker *HealthChecker,
	interval time.Duration,
	alertFuncs []AlertFunc,
	resolveFuncs []AlertFunc,
) *HealthMonitor {
	return &HealthMonitor{
		checker:      checker,
		interval:     interval,
		alertFuncs:   alertFuncs,
		resolveFuncs: resolveFuncs,
		stopChan:     make(chan struct{}),
		stateMap:     make(map[string]monitorState),
	}
}

// Start begins the monitoring process
func (hm *HealthMonitor) Start() {
	ticker := time.NewTicker(hm.interval)
	hm.checkAll()
	go func() {
		for {
			select {
			case <-ticker.C:
				hm.checkAll()
			case <-hm.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop ends the monitoring process
func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
}

func (hm *HealthMonitor) checkAll() {
	results := hm.checker.CheckAll()

	for _, result := range results {
		hm.processResult(result)
	}
}

func (hm *HealthMonitor) processResult(result Result) {
	hm.stateMu.Lock()
	defer hm.stateMu.Unlock()

	state, exists := hm.stateMap[result.Target.ID]
	if !exists {
		state = monitorState{}
	}

	// Update state based on current health check
	if !result.Healthy {
		state.consecutiveFailures++
	} else {
		if state.alerted {
			// If we previously alerted, call resolve functions
			for _, resolveFunc := range hm.resolveFuncs {
				if err := resolveFunc(result.Target, result); err != nil {
					slog.Error("resolveFunc failed", "target", result.Target, "error", err)
				}
			}
			state.alerted = false
		}
		state.consecutiveFailures = 0
	}

	// Check if we need to alert
	if state.consecutiveFailures >= 2 && !state.alerted {
		// Alert on second consecutive failure
		for _, alertFunc := range hm.alertFuncs {
			if err := alertFunc(result.Target, result); err != nil {
				slog.Error("alertFunc failed", "target", result.Target, "error", err)
			}
		}
		state.alerted = true
	}

	state.lastResult = result
	hm.stateMap[result.Target.ID] = state
}

// GetState returns the current state for a target
func (hm *HealthMonitor) GetState(targetID string) (monitorState, bool) {
	hm.stateMu.RLock()
	defer hm.stateMu.RUnlock()

	state, exists := hm.stateMap[targetID]
	return state, exists
}
