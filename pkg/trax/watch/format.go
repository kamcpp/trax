package watch

import (
	"fmt"
	"strings"
	"time"
)

// FormatTimestamp formats the current time in UTC with milliseconds and Z suffix
func FormatTimestamp() string {
	return time.Now().UTC().Format("15:04:05.000Z")
}

// FormatDuration formats a duration in a human-readable way, always including milliseconds
func FormatDuration(d time.Duration) string {
	ms := d.Milliseconds() % 1000
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		secs := int(d.Seconds())
		return fmt.Sprintf("%d.%03ds", secs, ms)
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%d.%03ds", minutes, seconds, ms)
}

// FormatDurationPadded formats a duration with left padding to ensure consistent width (10 chars)
// e.g., "  1m9.306s" or "10m25.123s"
func FormatDurationPadded(d time.Duration) string {
	return fmt.Sprintf("%10s", FormatDuration(d))
}

// ShortenID shortens a UUID/ID to first 4 chars + ".." + last 3 chars
// e.g., "12345678-abcd-efgh" -> "1234..fgh"
func ShortenID(id string) string {
	if len(id) <= 7 {
		return id
	}
	return id[:4] + ".." + id[len(id)-3:]
}

// ShortenSagaState extracts short state name from full enum
// e.g., "SAGA_STATE_ENUM_RUNNING" -> "RUNNING"
func ShortenSagaState(state string) string {
	const prefix = "SAGA_STATE_ENUM_"
	if len(state) > len(prefix) && state[:len(prefix)] == prefix {
		return state[len(prefix):]
	}
	return state
}

// ShortenStepState extracts short state name from full enum
// e.g., "SAGA_STEP_STATE_ENUM_EXECUTION_DONE" -> "EXEC_DONE"
func ShortenStepState(state string) string {
	const prefix = "SAGA_STEP_STATE_ENUM_"
	if len(state) > len(prefix) && state[:len(prefix)] == prefix {
		shortState := state[len(prefix):]
		if len(shortState) > 10 && shortState[:10] == "EXECUTION_" {
			return "EXEC_" + shortState[10:]
		}
		if len(shortState) > 13 && shortState[:13] == "COMPENSATION_" {
			return "COMP_" + shortState[13:]
		}
		return shortState
	}
	return state
}

// ShortenTemplateID extracts the last part of a template ID after ":"
// e.g., "transfer_authorized_instrument:validate_balances" -> "validate_balances"
func ShortenTemplateID(templateID string) string {
	for i := len(templateID) - 1; i >= 0; i-- {
		if templateID[i] == ':' {
			return templateID[i+1:]
		}
	}
	return templateID
}

// BuildProgressBar creates a progress bar string
// e.g., "████████░░░░░░░░░░░░" for 40% with width 20
func BuildProgressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := (percent * width) / 100
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// GetStepIndicator returns the visual indicator for a step state
func GetStepIndicator(status string) string {
	switch {
	case strings.Contains(status, "DONE"):
		return "✓"
	case strings.Contains(status, "RUNNING") || strings.Contains(status, "STARTED"):
		return "▶"
	case strings.Contains(status, "PENDING"):
		return "○"
	case strings.Contains(status, "ERROR") || strings.Contains(status, "COMPENSATION"):
		return "✗"
	default:
		return "?"
	}
}

// IsStepRunning checks if a step is currently running
func IsStepRunning(status string) bool {
	return strings.Contains(status, "RUNNING") || strings.Contains(status, "STARTED")
}

// IsStepDone checks if a step is completed (success or failure)
func IsStepDone(status string) bool {
	return strings.Contains(status, "DONE") || strings.Contains(status, "ERROR") || strings.Contains(status, "COMPENSATION")
}

// IsStepPending checks if a step is pending
func IsStepPending(status string) bool {
	return strings.Contains(status, "PENDING")
}

// IsStepExecutionFailed checks if a step failed during execution (not compensation)
func IsStepExecutionFailed(status string) bool {
	return strings.Contains(status, "EXECUTION_FAILED")
}

// IsTerminalState checks if a saga state is terminal (no more progress expected)
func IsTerminalState(state string) bool {
	shortState := ShortenSagaState(state)
	switch shortState {
	case "COMMITTED", "COMPENSATED", "BLOCKED", "INVALID_STATE":
		return true
	}
	return false
}

// IsSuccessState checks if a saga state indicates success
func IsSuccessState(state string) bool {
	return ShortenSagaState(state) == "COMMITTED"
}

// IsFailureState checks if a saga state indicates failure
func IsFailureState(state string) bool {
	shortState := ShortenSagaState(state)
	switch shortState {
	case "COMPENSATED", "BLOCKED", "INVALID_STATE":
		return true
	}
	return false
}
