package common

import (
	"os"
	"strings"
	"time"
)

// FIXResetWindowEnv is the env var that controls when an inbound Logon with
// ResetSeqNumFlag (141)=Y is allowed. Empty / unset = NEVER allow.
const FIXResetWindowEnv = "FIX_ALLOW_RESET_WINDOW"

// FIXResetWindowAllows returns true when the configured FIX_ALLOW_RESET_WINDOW
// brackets `now`. The env value is parsed as "HH:MM-HH:MM" UTC. A window that
// wraps midnight (e.g. "23:30-00:30") is supported.
//
// Empty config returns false: outside V2 deployments must opt in to allowing
// peer-driven sequence resets — silent acceptance of 141=Y was the V1 default
// and is what FIXREL2 Phase 5 closes.
func FIXResetWindowAllows(now time.Time) bool {
	spec := strings.TrimSpace(os.Getenv(FIXResetWindowEnv))
	if spec == "" {
		return false
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return false
	}
	start, err1 := time.Parse("15:04", strings.TrimSpace(parts[0]))
	end, err2 := time.Parse("15:04", strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return false
	}

	nowUTC := now.UTC()
	nowMins := nowUTC.Hour()*60 + nowUTC.Minute()
	startMins := start.Hour()*60 + start.Minute()
	endMins := end.Hour()*60 + end.Minute()

	if startMins == endMins {
		// Zero-width window — treat as disabled rather than always-open.
		return false
	}
	if startMins < endMins {
		// Same-day window: [start, end)
		return nowMins >= startMins && nowMins < endMins
	}
	// Wraps midnight: [start, 24:00) ∪ [00:00, end)
	return nowMins >= startMins || nowMins < endMins
}

// FIXResetWindowSpec returns the raw env value for logging.
func FIXResetWindowSpec() string {
	return strings.TrimSpace(os.Getenv(FIXResetWindowEnv))
}
