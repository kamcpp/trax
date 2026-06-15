package watch

import (
	"encoding/json"
	"fmt"
	"time"
)

// StatusFetcher is a function that fetches saga status and steps.
// It should return the saga state string, template ID, and step statuses.
type StatusFetcher func() (status string, templateID string, steps []StepStatus, compensationReason string, err error)

// ChildSagaFetcher fetches child sagas for a parent saga instance.
// Returns a list of ChildSagaInfo (with their steps populated).
// This is optional - if nil, child sagas are not displayed.
type ChildSagaFetcher func(parentSagaInstanceID string) ([]ChildSagaInfo, error)

// TraxctrlPoster is a function that POSTs JSON to a traxctrl API path and returns the response body.
// The path is relative to the traxctrl base URL (e.g., "/saga-instances/{id}/children").
type TraxctrlPoster func(path string, body interface{}) ([]byte, error)

// NewChildSagaFetcher creates a reusable ChildSagaFetcher that queries traxctrl
// for child sagas and their step statuses. Both CLI tools and E2E tests should
// use this instead of implementing their own fetching logic.
func NewChildSagaFetcher(poster TraxctrlPoster, clusterID string) ChildSagaFetcher {
	return func(parentSagaInstanceID string) ([]ChildSagaInfo, error) {
		// Fetch child saga instances
		childrenBody, err := poster(
			fmt.Sprintf("/saga-instances/%s/children", parentSagaInstanceID),
			map[string]string{"cluster_id": clusterID},
		)
		if err != nil {
			return nil, err
		}

		var childrenResp struct {
			SagaInstances []struct {
				InstanceID               string `json:"instance_id"`
				SagaTemplateID           string `json:"saga_template_id"`
				State                    string `json:"state"`
				ParentSagaStepInstanceID string `json:"parent_saga_step_instance_id"`
			} `json:"saga_instances"`
		}
		if err := json.Unmarshal(childrenBody, &childrenResp); err != nil {
			return nil, fmt.Errorf("failed to parse children response: %w", err)
		}

		var result []ChildSagaInfo
		for _, child := range childrenResp.SagaInstances {
			// Fetch steps for each child saga
			stepsBody, err := poster("/saga-step-instances/list", map[string]string{
				"cluster_id":       clusterID,
				"saga_instance_id": child.InstanceID,
			})

			var childSteps []StepStatus
			if err == nil {
				var stepsResp struct {
					SagaStepInstances []StepStatus `json:"saga_step_instances"`
				}
				if json.Unmarshal(stepsBody, &stepsResp) == nil {
					childSteps = stepsResp.SagaStepInstances
				}
			}

			result = append(result, ChildSagaInfo{
				SagaInstanceID:       child.InstanceID,
				SagaTemplateID:       child.SagaTemplateID,
				State:                child.State,
				ParentStepInstanceID: child.ParentSagaStepInstanceID,
				Steps:                childSteps,
			})
		}
		return result, nil
	}
}

// WatchResult contains the final result of watching a saga
type WatchResult struct {
	// FinalStatus is the saga's final state
	FinalStatus string

	// Steps contains the final step statuses with their results
	Steps []StepStatus

	// Timing contains timing information for the saga and steps
	Timing *Timing

	// CompensationReason is the saga-level reason why compensation was triggered
	CompensationReason string

	// Success indicates if the saga completed successfully (COMMITTED)
	Success bool

	// Error contains any error that occurred during watching (timeout, saga failure, etc.)
	Error error
}

// WatcherConfig holds configuration for the Watcher
type WatcherConfig struct {
	// Timeout is the maximum time to wait for saga completion
	Timeout time.Duration

	// PollInterval is how often to check saga status
	PollInterval time.Duration

	// NotFoundTimeout is how long to tolerate 404s before giving up
	NotFoundTimeout time.Duration

	// HeartbeatInterval is how often to print a heartbeat when no state change occurs
	// Set to 0 to disable heartbeat
	HeartbeatInterval time.Duration

	// DisplayConfig controls the display output
	DisplayConfig Config

	// ChildSagaFetcher fetches child sagas for display as tree branches.
	// If nil, no child sagas are shown.
	ChildSagaFetcher ChildSagaFetcher
}

// DefaultWatcherConfig returns sensible defaults for watching
func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{
		Timeout:           10 * time.Minute,
		PollInterval:      2 * time.Second,
		NotFoundTimeout:   60 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		DisplayConfig:     DefaultConfig(),
	}
}

// Watcher watches a saga and returns the result when complete
type Watcher struct {
	sagaInstanceID string
	config         WatcherConfig
	display        *Display
	timing         *Timing
	fetcher        StatusFetcher
}

// NewWatcher creates a new Watcher
func NewWatcher(sagaInstanceID string, printer Printer, config WatcherConfig, fetcher StatusFetcher) *Watcher {
	return &Watcher{
		sagaInstanceID: sagaInstanceID,
		config:         config,
		display:        NewDisplay(printer, config.DisplayConfig),
		timing:         NewTiming(),
		fetcher:        fetcher,
	}
}

// Watch polls the saga until it reaches a terminal state or times out.
// Returns a WatchResult with the final status and all step data.
func (w *Watcher) Watch() *WatchResult {
	result := &WatchResult{
		Timing: w.timing,
	}

	deadline := time.Now().Add(w.config.Timeout)
	var firstNotFoundTime time.Time
	var lastStateKey string
	var lastPrintTime time.Time
	var lastSagaStatus SagaStatus
	var headerPrinted bool
	var borderOpen bool // Track if outer border is currently open

	for time.Now().Before(deadline) {
		status, templateID, steps, compensationReason, err := w.fetcher()
		if err != nil {
			// Handle 404 errors with timeout
			if isNotFoundError(err) {
				if firstNotFoundTime.IsZero() {
					firstNotFoundTime = time.Now()
					if !headerPrinted {
						w.display.PrintWatchHeader(w.sagaInstanceID, "", w.config.Timeout, w.config.PollInterval)
						headerPrinted = true
					}
					w.display.PrintNotFound(w.config.NotFoundTimeout)
				}

				if time.Since(firstNotFoundTime) > w.config.NotFoundTimeout {
					if borderOpen {
						w.display.PrintStatusClose()
					}
					result.FinalStatus = "NOT_FOUND"
					result.Error = fmt.Errorf("saga not found (404) for more than %v", w.config.NotFoundTimeout)
					return result
				}
			} else {
				firstNotFoundTime = time.Time{}
			}

			w.display.PrintError(err)
			time.Sleep(w.config.PollInterval)
			continue
		}

		// Print header on first successful fetch (includes template ID)
		if !headerPrinted {
			w.display.PrintWatchHeader(w.sagaInstanceID, templateID, w.config.Timeout, w.config.PollInterval)
			headerPrinted = true
		}

		// Reset 404 timer on success
		firstNotFoundTime = time.Time{}

		// Update timing
		UpdateStepTimings(w.timing, steps)
		UpdateStateTracking(w.timing, status)

		// Build status for display
		sagaStatus := BuildSagaStatus(w.sagaInstanceID, templateID, status, steps)

		// Fetch child sagas if fetcher is configured
		if w.config.ChildSagaFetcher != nil {
			children, fetchErr := w.config.ChildSagaFetcher(w.sagaInstanceID)
			if fetchErr != nil {
				w.display.PrintError(fmt.Errorf("child saga fetch failed: %w", fetchErr))
			} else if len(children) > 0 {
				sagaStatus.ChildSagasByStep = make(map[string][]ChildSagaInfo)
				for _, child := range children {
					key := child.ParentStepInstanceID
					sagaStatus.ChildSagasByStep[key] = append(sagaStatus.ChildSagasByStep[key], child)
				}
			}
		}

		lastSagaStatus = sagaStatus

		// Print if state changed, or heartbeat if enough time passed
		stateKey := BuildStateKey(sagaStatus)
		if stateKey != lastStateKey {
			// Close previous border if open, then open new one
			if borderOpen {
				w.display.PrintStatusClose()
			}
			w.display.PrintStatusOpen(sagaStatus, w.timing)
			borderOpen = true
			lastStateKey = stateKey
			lastPrintTime = time.Now()
		} else if w.config.HeartbeatInterval > 0 && time.Since(lastPrintTime) >= w.config.HeartbeatInterval {
			// Heartbeat goes inside the open border
			w.display.PrintHeartbeat(sagaStatus, w.timing)
			lastPrintTime = time.Now()
		}

		// Check for terminal state
		if IsTerminalState(status) {
			// Close the border before printing final results
			if borderOpen {
				w.display.PrintStatusClose()
				borderOpen = false
			}

			result.FinalStatus = status
			result.Steps = steps
			result.CompensationReason = compensationReason

			if IsSuccessState(status) {
				result.Success = true
				w.display.PrintSuccess(w.timing)
			} else if ShortenSagaState(status) == "COMPENSATED" {
				w.display.PrintCompensated(w.timing, steps, compensationReason)
				result.Error = fmt.Errorf("saga compensated")
			} else {
				w.display.PrintFailure(status, w.timing, steps)
				result.Error = fmt.Errorf("saga failed: %s", status)
			}

			w.display.PrintFinalResults(steps, w.timing, sagaStatus.ChildSagasByStep)
			return result
		}

		time.Sleep(w.config.PollInterval)
	}
	_ = lastSagaStatus // used for potential future enhancements

	// Timeout - close border if open
	if borderOpen {
		w.display.PrintStatusClose()
	}

	// Timeout - do one final check
	status, _, steps, _, err := w.fetcher()
	if err == nil {
		result.FinalStatus = status
		result.Steps = steps
		if IsTerminalState(status) && IsSuccessState(status) {
			result.Success = true
			return result
		}
	}

	w.display.PrintTimeout(w.config.Timeout, w.timing, result.FinalStatus)
	result.Error = fmt.Errorf("timeout after %v (final status: %s)", w.config.Timeout, ShortenSagaState(result.FinalStatus))
	return result
}

// WatchUntil polls until the saga reaches the expected status or a terminal state
func (w *Watcher) WatchUntil(expectedStatus string) *WatchResult {
	result := &WatchResult{
		Timing: w.timing,
	}

	deadline := time.Now().Add(w.config.Timeout)
	var firstNotFoundTime time.Time
	var lastStateKey string
	var lastPrintTime time.Time
	var headerPrinted bool
	var borderOpen bool // Track if outer border is currently open

	for time.Now().Before(deadline) {
		status, templateID, steps, compensationReason, err := w.fetcher()
		if err != nil {
			if isNotFoundError(err) {
				if firstNotFoundTime.IsZero() {
					firstNotFoundTime = time.Now()
					if !headerPrinted {
						w.display.PrintWatchHeader(w.sagaInstanceID, "", w.config.Timeout, w.config.PollInterval)
						headerPrinted = true
					}
					w.display.PrintNotFound(w.config.NotFoundTimeout)
				}
				if time.Since(firstNotFoundTime) > w.config.NotFoundTimeout {
					if borderOpen {
						w.display.PrintStatusClose()
					}
					result.FinalStatus = "NOT_FOUND"
					result.Error = fmt.Errorf("saga not found (404) for more than %v", w.config.NotFoundTimeout)
					return result
				}
			} else {
				firstNotFoundTime = time.Time{}
			}
			w.display.PrintError(err)
			time.Sleep(w.config.PollInterval)
			continue
		}

		// Print header on first successful fetch (includes template ID)
		if !headerPrinted {
			w.display.PrintWatchHeader(w.sagaInstanceID, templateID, w.config.Timeout, w.config.PollInterval)
			headerPrinted = true
		}

		firstNotFoundTime = time.Time{}

		UpdateStepTimings(w.timing, steps)
		UpdateStateTracking(w.timing, status)

		sagaStatus := BuildSagaStatus(w.sagaInstanceID, templateID, status, steps)

		// Fetch child sagas if fetcher is configured
		if w.config.ChildSagaFetcher != nil {
			children, fetchErr := w.config.ChildSagaFetcher(w.sagaInstanceID)
			if fetchErr != nil {
				w.display.PrintError(fmt.Errorf("child saga fetch failed: %w", fetchErr))
			} else if len(children) > 0 {
				sagaStatus.ChildSagasByStep = make(map[string][]ChildSagaInfo)
				for _, child := range children {
					key := child.ParentStepInstanceID
					sagaStatus.ChildSagasByStep[key] = append(sagaStatus.ChildSagasByStep[key], child)
				}
			}
		}

		stateKey := BuildStateKey(sagaStatus)
		if stateKey != lastStateKey {
			// Close previous border if open, then open new one
			if borderOpen {
				w.display.PrintStatusClose()
			}
			w.display.PrintStatusOpen(sagaStatus, w.timing)
			borderOpen = true
			lastStateKey = stateKey
			lastPrintTime = time.Now()
		} else if w.config.HeartbeatInterval > 0 && time.Since(lastPrintTime) >= w.config.HeartbeatInterval {
			// Heartbeat goes inside the open border
			w.display.PrintHeartbeat(sagaStatus, w.timing)
			lastPrintTime = time.Now()
		}

		// Check if we reached expected status
		if status == expectedStatus {
			// Close the border before printing final results
			if borderOpen {
				w.display.PrintStatusClose()
				borderOpen = false
			}
			result.FinalStatus = status
			result.Steps = steps
			result.CompensationReason = compensationReason
			result.Success = true
			w.display.PrintSuccess(w.timing)
			w.display.PrintFinalResults(steps, w.timing, sagaStatus.ChildSagasByStep)
			return result
		}

		// Check for other terminal states (failure)
		if IsTerminalState(status) {
			// Close the border before printing final results
			if borderOpen {
				w.display.PrintStatusClose()
				borderOpen = false
			}
			result.FinalStatus = status
			result.Steps = steps

			if ShortenSagaState(status) == "COMPENSATED" {
				w.display.PrintCompensated(w.timing, steps, compensationReason)
				result.Error = fmt.Errorf("saga compensated")
			} else {
				w.display.PrintFailure(status, w.timing, steps)
				result.Error = fmt.Errorf("saga failed: %s", status)
			}
			w.display.PrintFinalResults(steps, w.timing, sagaStatus.ChildSagasByStep)
			return result
		}

		time.Sleep(w.config.PollInterval)
	}

	// Timeout - close border if open
	if borderOpen {
		w.display.PrintStatusClose()
	}

	// Timeout
	status, _, steps, _, _ := w.fetcher()
	result.FinalStatus = status
	result.Steps = steps
	if status == expectedStatus {
		result.Success = true
		return result
	}

	w.display.PrintTimeout(w.config.Timeout, w.timing, result.FinalStatus)
	result.Error = fmt.Errorf("timeout after %v (final status: %s)", w.config.Timeout, ShortenSagaState(result.FinalStatus))
	return result
}

// GetStepResult returns the Result field for a specific step by template ID
func (r *WatchResult) GetStepResult(stepTemplateID string) interface{} {
	for _, step := range r.Steps {
		if step.StepTemplateID == stepTemplateID || ShortenTemplateID(step.StepTemplateID) == stepTemplateID {
			return step.Result
		}
	}
	return nil
}

// GetStepCompensationResult returns the CompensationResult field for a specific step by template ID
func (r *WatchResult) GetStepCompensationResult(stepTemplateID string) interface{} {
	for _, step := range r.Steps {
		if step.StepTemplateID == stepTemplateID || ShortenTemplateID(step.StepTemplateID) == stepTemplateID {
			return step.CompensationResult
		}
	}
	return nil
}

// GetCompensationReason returns the saga-level CompensationReason
func (r *WatchResult) GetCompensationReason() string {
	return r.CompensationReason
}

// GetStepError returns the ExecutionError for a specific step by template ID
func (r *WatchResult) GetStepError(stepTemplateID string) string {
	for _, step := range r.Steps {
		if step.StepTemplateID == stepTemplateID || ShortenTemplateID(step.StepTemplateID) == stepTemplateID {
			return step.ExecutionError
		}
	}
	return ""
}

// isNotFoundError checks if an error is a 404/not found error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "404") || contains(errStr, "not found")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
