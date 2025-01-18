package main

import "log/slog"

func NewLogAlert() AlertFunc {
	return func(target HealthTarget, result Result) error {
		slog.Warn("Target DWON", "target", target, "status", result.Status, "error", result.Error)
		return nil
	}
}

func NewLogResolve() AlertFunc {
	return func(target HealthTarget, result Result) error {
		slog.Warn("Target UP", "target", target, "status", result.Status)
		return nil
	}
}
