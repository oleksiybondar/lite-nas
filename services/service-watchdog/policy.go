package main

import (
	"time"

	"lite-nas/shared/systemd"
)

// unitProgress tracks only Watchdog's own attempts; the systemd restart count
// remains a separate property in each observation.
type unitProgress struct {
	FailureSince       time.Time
	NextAttempt        time.Time
	Attempts           int
	Recovering         bool
	EnablementIncident bool
	RecoveryIncident   bool
}

// shouldAttempt defers to systemd while it is activating or auto-restarting.
// Once the configured delay and backoff have passed, a manually stopped unit
// is eligible too, as requested for protected services.
func shouldAttempt(now time.Time, cfg watchdogConfig, unit protectedUnit, observed systemd.UnitState, progress unitProgress) bool {
	if !unit.Recover || progress.Recovering || observed.ActiveState == "active" {
		return false
	}
	if observed.ActiveState == "activating" || observed.SubState == "auto-restart" {
		return false
	}
	if progress.Attempts >= cfg.MaxAttempts || progress.FailureSince.IsZero() {
		return false
	}
	if now.Before(progress.FailureSince.Add(cfg.RecoveryDelay)) || now.Before(progress.NextAttempt) {
		return false
	}
	return true
}

// nextBackoff applies capped exponential delay after a Watchdog attempt.
func nextBackoff(base time.Duration, attempts int) time.Duration {
	if attempts < 1 {
		return base
	}
	shift := min(attempts-1, 9)
	return base * time.Duration(1<<shift)
}
