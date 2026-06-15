package watch

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Display handles rendering saga watch status to a Printer
type Display struct {
	printer Printer
	config  Config
}

// NewDisplay creates a new Display with the given printer and config
func NewDisplay(printer Printer, config Config) *Display {
	return &Display{
		printer: printer,
		config:  config,
	}
}

// UpdateStepTimings updates the timing information for steps
func UpdateStepTimings(timing *Timing, steps []StepStatus) {
	now := time.Now()
	for _, step := range steps {
		stepTiming, exists := timing.StepTimings[step.StepTemplateID]
		if !exists {
			stepTiming = &StepTiming{}
			timing.StepTimings[step.StepTemplateID] = stepTiming
		}

		isRunning := IsStepRunning(step.Status)
		isDone := IsStepDone(step.Status)

		// Track start time when step begins running
		if isRunning && stepTiming.StartTime.IsZero() {
			stepTiming.StartTime = now
		}

		// Track end time when step completes
		if isDone && !stepTiming.IsCompleted {
			stepTiming.EndTime = now
			stepTiming.IsCompleted = true
			if !stepTiming.StartTime.IsZero() {
				stepTiming.Duration = stepTiming.EndTime.Sub(stepTiming.StartTime)
			}
		}
	}
}

// UpdateStateTracking updates the state tracking for elapsed time calculation
func UpdateStateTracking(timing *Timing, status string) {
	if timing.LastState != status {
		timing.StateStartTime = time.Now()
		timing.LastState = status
	}
}

// BuildSagaStatus constructs a SagaStatus from raw status and steps
func BuildSagaStatus(sagaInstanceID string, templateID string, status string, steps []StepStatus) SagaStatus {
	sagaStatus := SagaStatus{
		SagaInstanceID: sagaInstanceID,
		SagaTemplateID: templateID,
		Status:         status,
		StepStatuses:   steps,
		StepSummary: StepSummary{
			Total:   len(steps),
			ByState: make(map[string]int),
		},
	}

	for _, step := range steps {
		sagaStatus.StepSummary.ByState[step.Status]++
		if strings.Contains(step.Status, "DONE") {
			sagaStatus.StepSummary.Completed++
		} else if strings.Contains(step.Status, "PENDING") {
			sagaStatus.StepSummary.Pending++
		} else if strings.Contains(step.Status, "ERROR") || strings.Contains(step.Status, "COMPENSATION") {
			sagaStatus.StepSummary.Failed++
		}
	}

	return sagaStatus
}

// BuildStateKey creates a key string for detecting state changes
// Includes running steps to detect progress even when counts don't change
func BuildStateKey(status SagaStatus) string {
	// Find running steps to include in the key
	var runningSteps []string
	for _, step := range status.StepStatuses {
		if IsStepRunning(step.Status) {
			runningSteps = append(runningSteps, ShortenTemplateID(step.StepTemplateID))
		}
	}
	runningKey := ""
	if len(runningSteps) > 0 {
		runningKey = fmt.Sprintf("-running:%v", runningSteps)
	}

	// Include child saga states so display refreshes when children progress
	if len(status.ChildSagasByStep) > 0 {
		var childKeys []string
		for stepID, children := range status.ChildSagasByStep {
			shortStepID := stepID
			if len(shortStepID) > 8 {
				shortStepID = shortStepID[:8]
			}
			for _, child := range children {
				childCompleted := 0
				childTotal := len(child.Steps)
				for _, cs := range child.Steps {
					if IsStepDone(cs.Status) {
						childCompleted++
					}
				}
				childKeys = append(childKeys, fmt.Sprintf("%s:%s:%d/%d", shortStepID, ShortenSagaState(child.State), childCompleted, childTotal))
			}
		}
		sort.Strings(childKeys)
		runningKey += fmt.Sprintf("-children:%v", childKeys)
	}

	return fmt.Sprintf("%s-%d-%d-%d%s", status.Status, status.StepSummary.Completed, status.StepSummary.Pending, status.StepSummary.Failed, runningKey)
}

// PrintStatusOpen prints the opening of a status report (outer border top + header + inner steps box)
// This leaves the outer border open for heartbeats to be added
func (d *Display) PrintStatusOpen(status SagaStatus, timing *Timing) {
	shortState := ShortenSagaState(status.Status)
	timestamp := FormatTimestamp()
	stateElapsed := time.Since(timing.StateStartTime)
	totalElapsed := time.Since(timing.SagaStartTime)

	// Print header with progress bar
	progress := fmt.Sprintf("%d/%d", status.StepSummary.Completed, status.StepSummary.Total)
	progressPct := 0
	if status.StepSummary.Total > 0 {
		progressPct = (status.StepSummary.Completed * 100) / status.StepSummary.Total
	}

	bar := BuildProgressBar(progressPct, d.config.ProgressBarWidth)

	// Start outer bordered report
	d.printer.PrintLine("┌──────────────────────────────────")
	d.printer.PrintLine("│ [%s] [%s +%s] %s [%s] %d%% (total: %s)",
		timestamp, shortState, FormatDuration(stateElapsed), progress, bar, progressPct, FormatDuration(totalElapsed))

	// Print template ID and saga instance ID on its own line if available
	if status.SagaTemplateID != "" {
		d.printer.PrintLine("│ Template: %s  Instance: %s", status.SagaTemplateID, status.SagaInstanceID)
	}

	// Find the current running step index
	currentIdx := d.findCurrentStepIndex(status)

	// Calculate window bounds
	startIdx, endIdx := d.calculateWindow(currentIdx, len(status.StepStatuses))

	// Print inner steps box
	d.printer.PrintLine("│ ┌────")

	// Print ellipsis if skipping steps at start
	if startIdx > 0 {
		d.printer.PrintLine("│ │ ... %d steps before", startIdx)
	}

	// Print step-by-step status for visible window
	for i := startIdx; i < endIdx; i++ {
		var stepChildren []ChildSagaInfo
		if status.ChildSagasByStep != nil {
			stepChildren = status.ChildSagasByStep[status.StepStatuses[i].StepInstanceID]
		}
		d.printStep(i, status.StepStatuses[i], timing, stepChildren)
	}

	// Print ellipsis if skipping steps at end
	if endIdx < len(status.StepStatuses) {
		d.printer.PrintLine("│ │ ... %d steps after", len(status.StepStatuses)-endIdx)
	}

	d.printer.PrintLine("│ └────")
	d.printer.PrintLine("│")
}

// PrintStatusClose prints the closing of the outer border
func (d *Display) PrintStatusClose() {
	d.printer.PrintLine("└──────────────────────────────────")
	d.printer.PrintEmpty()
}

// PrintStatus prints the current saga status with progress bar and timing (complete box, no heartbeats expected)
func (d *Display) PrintStatus(status SagaStatus, timing *Timing) {
	d.PrintStatusOpen(status, timing)
	d.PrintStatusClose()
}

// PrintFinalResults prints all step statuses with full details (for final results)
func (d *Display) PrintFinalResults(steps []StepStatus, timing *Timing, childSagasByStep map[string][]ChildSagaInfo) {
	if len(steps) == 0 {
		return
	}

	d.printer.PrintEmpty()
	d.printer.PrintLine("=== Final Step Results ===")
	d.printer.PrintEmpty()

	for i, step := range steps {
		shortState := ShortenStepState(step.Status)
		shortTemplate := ShortenTemplateID(step.StepTemplateID)
		indicator := GetStepIndicator(step.Status)
		isDone := IsStepDone(step.Status)

		// Get step timing info
		stepDurationStr := ""
		if timing != nil {
			if stepTiming, exists := timing.StepTimings[step.StepTemplateID]; exists {
				if stepTiming.IsCompleted && stepTiming.Duration > 0 {
					stepDurationStr = fmt.Sprintf(" (%s)", FormatDuration(stepTiming.Duration))
				} else if isDone {
					stepDurationStr = " (?ms)"
				}
			} else if isDone {
				stepDurationStr = " (?ms)"
			}
		} else if isDone {
			stepDurationStr = " (?ms)"
		}

		d.printer.PrintLine("  %s Step %2d [%-12s]: %s%s", indicator, i+1, shortState, shortTemplate, stepDurationStr)

		// Show error if present
		if step.ExecutionError != "" {
			d.printer.PrintLine("           ERROR: %s", step.ExecutionError)
		}
		// Show child sagas for this step (skip committed ones - show parent step as normal)
		if childSagasByStep != nil {
			var visibleChildren []ChildSagaInfo
			for _, child := range childSagasByStep[step.StepInstanceID] {
				if !IsSuccessState(child.State) {
					visibleChildren = append(visibleChildren, child)
				}
			}

			for ci, child := range visibleChildren {
				isLastChild := ci == len(visibleChildren)-1
				childShortState := ShortenSagaState(child.State)
				childShortTemplate := ShortenTemplateID(child.SagaTemplateID)
				stepCount := len(child.Steps)

				// Count completed steps for progress
				childCompleted := 0
				for _, cs := range child.Steps {
					if IsStepDone(cs.Status) {
						childCompleted++
					}
				}
				childPct := 0
				if stepCount > 0 {
					childPct = (childCompleted * 100) / stepCount
				}
				childBar := BuildProgressBar(childPct, d.config.ProgressBarWidth)

				// Tree connectors: ├─ for siblings, └─ for last child
				branchConn := "├─"
				if isLastChild {
					branchConn = "└─"
				}
				contPrefix := "    │  "
				if isLastChild {
					contPrefix = "       "
				}

				// Line 1: state + template + step count
				d.printer.PrintLine("    %s [%s] %s (%d steps)", branchConn, childShortState, childShortTemplate, stepCount)
				// Line 2: instance ID + progress bar
				d.printer.PrintLine("%s%s %d/%d [%s] %d%%", contPrefix, child.SagaInstanceID, childCompleted, stepCount, childBar, childPct)

				// Apply same windowing as parent saga
				childStatus := BuildSagaStatus(child.SagaInstanceID, child.SagaTemplateID, child.State, child.Steps)
				childCurrentIdx := d.findCurrentStepIndex(childStatus)
				childStart, childEnd := d.calculateWindow(childCurrentIdx, len(child.Steps))

				if childStart > 0 {
					d.printer.PrintLine("%s... %d steps before", contPrefix, childStart)
				}

				for j := childStart; j < childEnd; j++ {
					childStep := child.Steps[j]
					childIndicator := GetStepIndicator(childStep.Status)
					childStepState := ShortenStepState(childStep.Status)
					childStepTemplate := ShortenTemplateID(childStep.StepTemplateID)
					stepConn := "├─"
					if j == len(child.Steps)-1 {
						stepConn = "└─"
					}
					d.printer.PrintLine("%s%s %s %2d [%-12s]: %s", contPrefix, stepConn, childIndicator, j+1, childStepState, childStepTemplate)
				}

				if childEnd < len(child.Steps) {
					d.printer.PrintLine("%s... %d steps after", contPrefix, len(child.Steps)-childEnd)
				}
			}
		}
	}
	d.printer.PrintEmpty()
}

// PrintSuccess prints the success message
func (d *Display) PrintSuccess(timing *Timing) {
	totalDuration := time.Since(timing.SagaStartTime)
	d.printer.PrintEmpty()
	d.printer.PrintLine("✓ Saga COMMITTED successfully at %s (total: %s)", FormatTimestamp(), FormatDuration(totalDuration))
}

// PrintCompensated prints the compensated (rollback) message with root cause extraction
func (d *Display) PrintCompensated(timing *Timing, steps []StepStatus, compensationReason ...string) {
	totalDuration := time.Since(timing.SagaStartTime)
	d.printer.PrintEmpty()
	d.printer.PrintLine("✗ Saga COMPENSATED (rolled back) at %s (total: %s)", FormatTimestamp(), FormatDuration(totalDuration))
	if len(compensationReason) > 0 && compensationReason[0] != "" {
		d.printer.PrintLine("  Compensation reason: %s", compensationReason[0])
	}
	d.printRootCause(steps)
}

// PrintFailure prints the failure message with root cause extraction
func (d *Display) PrintFailure(state string, timing *Timing, steps []StepStatus) {
	totalDuration := time.Since(timing.SagaStartTime)
	d.printer.PrintEmpty()
	d.printer.PrintLine("✗ Saga FAILED with state %s at %s (total: %s)", ShortenSagaState(state), FormatTimestamp(), FormatDuration(totalDuration))
	d.printRootCause(steps)
}

// printRootCause finds and prints the first step that failed with an execution error
func (d *Display) printRootCause(steps []StepStatus) {
	for _, step := range steps {
		if step.ExecutionError != "" && IsStepExecutionFailed(step.Status) {
			d.printer.PrintLine("  Root cause: step [%s] failed with error:", ShortenTemplateID(step.StepTemplateID))
			d.printer.PrintLine("    %s", step.ExecutionError)
			return
		}
	}
	// Fallback: any step with an execution error
	for _, step := range steps {
		if step.ExecutionError != "" {
			d.printer.PrintLine("  Root cause: step [%s] error:", ShortenTemplateID(step.StepTemplateID))
			d.printer.PrintLine("    %s", step.ExecutionError)
			return
		}
	}
}

// PrintTimeout prints the timeout message
func (d *Display) PrintTimeout(timeout time.Duration, timing *Timing, finalStatus string) {
	totalDuration := time.Since(timing.SagaStartTime)
	d.printer.PrintEmpty()
	d.printer.PrintLine("✗ Timeout after %v (total: %s, final status: %s)", timeout, FormatDuration(totalDuration), ShortenSagaState(finalStatus))
}

// PrintError prints an error message with timestamp
func (d *Display) PrintError(err error) {
	d.printer.PrintLine("  [%s] Error: %v", FormatTimestamp(), err)
}

// PrintNotFound prints a "saga not found" message
func (d *Display) PrintNotFound(notFoundTimeout time.Duration) {
	d.printer.PrintLine("  [%s] Saga not found (404), will timeout after %v if it persists", FormatTimestamp(), notFoundTimeout)
}

// PrintWatchHeader prints the initial watch header
func (d *Display) PrintWatchHeader(sagaInstanceID string, templateID string, timeout, pollInterval time.Duration) {
	d.printer.PrintLine("Watching saga: %s", sagaInstanceID)
	if templateID != "" {
		d.printer.PrintLine("Template: %s", templateID)
	}
	d.printer.PrintLine("Timeout: %v, Poll interval: %v, Window: %d steps", timeout, pollInterval, d.config.WindowSize)
	d.printer.PrintEmpty()
}

// PrintHeartbeat prints a heartbeat message showing elapsed time and current running step
// This is printed inside the outer border (assumes PrintStatusOpen was called)
func (d *Display) PrintHeartbeat(status SagaStatus, timing *Timing) {
	timestamp := FormatTimestamp()
	totalElapsed := time.Since(timing.SagaStartTime)

	// Find the current running step
	runningStep := ""
	runningElapsed := ""
	for _, step := range status.StepStatuses {
		if IsStepRunning(step.Status) {
			runningStep = ShortenTemplateID(step.StepTemplateID)
			if stepTiming, exists := timing.StepTimings[step.StepTemplateID]; exists {
				if !stepTiming.StartTime.IsZero() {
					runningElapsed = fmt.Sprintf(" +%s", FormatDuration(time.Since(stepTiming.StartTime)))
				}
			}
			break
		}
	}

	progress := fmt.Sprintf("%d/%d", status.StepSummary.Completed, status.StepSummary.Total)
	if runningStep != "" {
		d.printer.PrintLine("│ [%s] %s (total:%s) ... %s%s", timestamp, progress, FormatDurationPadded(totalElapsed), runningStep, runningElapsed)
	} else {
		d.printer.PrintLine("│ [%s] %s (total:%s) waiting...", timestamp, progress, FormatDurationPadded(totalElapsed))
	}
}

// findCurrentStepIndex finds the index of the current step (running, or first pending, or last completed)
func (d *Display) findCurrentStepIndex(status SagaStatus) int {
	// First, try to find a running step
	for i, step := range status.StepStatuses {
		if IsStepRunning(step.Status) {
			return i
		}
	}

	// If no running step, find first pending step
	for i, step := range status.StepStatuses {
		if IsStepPending(step.Status) {
			return i
		}
	}

	// Otherwise, use last completed step
	currentIdx := status.StepSummary.Completed - 1
	if currentIdx < 0 {
		currentIdx = 0
	}
	return currentIdx
}

// calculateWindow calculates the window bounds around the current step
func (d *Display) calculateWindow(currentIdx, totalSteps int) (startIdx, endIdx int) {
	windowSize := d.config.WindowSize

	if windowSize == 0 {
		// Show all steps
		return 0, totalSteps
	}

	// Show windowSize steps centered on current
	halfWindow := (windowSize - 1) / 2
	startIdx = currentIdx - halfWindow
	endIdx = currentIdx + halfWindow + 1 + (windowSize-1)%2

	// Adjust window to stay within bounds
	if startIdx < 0 {
		endIdx += -startIdx
		startIdx = 0
	}
	if endIdx > totalSteps {
		startIdx -= endIdx - totalSteps
		endIdx = totalSteps
	}
	if startIdx < 0 {
		startIdx = 0
	}

	return startIdx, endIdx
}

// printStep prints a single step with its status and timing, plus any child sagas
func (d *Display) printStep(index int, step StepStatus, timing *Timing, childSagas []ChildSagaInfo) {
	shortStepState := ShortenStepState(step.Status)
	shortTemplate := ShortenTemplateID(step.StepTemplateID)
	indicator := GetStepIndicator(step.Status)

	// Get step timing info
	stepDurationStr := ""
	isDone := IsStepDone(step.Status)
	if timing != nil {
		if stepTiming, exists := timing.StepTimings[step.StepTemplateID]; exists {
			if stepTiming.IsCompleted && stepTiming.Duration > 0 {
				stepDurationStr = fmt.Sprintf(" (%s)", FormatDuration(stepTiming.Duration))
			} else if !stepTiming.StartTime.IsZero() && !stepTiming.IsCompleted {
				// Currently running - show elapsed time
				stepDurationStr = fmt.Sprintf(" (+%s)", FormatDuration(time.Since(stepTiming.StartTime)))
			} else if isDone {
				// Completed but no timing observed
				stepDurationStr = " (?ms)"
			}
		} else if isDone {
			// No timing entry at all for a completed step
			stepDurationStr = " (?ms)"
		}
	} else if isDone {
		// No timing struct provided for a completed step
		stepDurationStr = " (?ms)"
	}

	d.printer.PrintLine("│ │ %s Step %2d [%-12s]: %s%s", indicator, index+1, shortStepState, shortTemplate, stepDurationStr)

	// Show error if present
	if step.ExecutionError != "" {
		d.printer.PrintLine("│ │          ERROR: %s", step.ExecutionError)
	}
	// Render child saga branches if any (skip committed ones - show parent step as normal)
	// First, collect visible (non-committed) children for correct tree connectors
	var visibleChildren []ChildSagaInfo
	for _, child := range childSagas {
		if !IsSuccessState(child.State) {
			visibleChildren = append(visibleChildren, child)
		}
	}

	for ci, child := range visibleChildren {
		isLastChild := ci == len(visibleChildren)-1
		childShortState := ShortenSagaState(child.State)
		childShortTemplate := ShortenTemplateID(child.SagaTemplateID)
		stepCount := len(child.Steps)

		// Count completed steps for progress
		childCompleted := 0
		for _, cs := range child.Steps {
			if IsStepDone(cs.Status) {
				childCompleted++
			}
		}
		childPct := 0
		if stepCount > 0 {
			childPct = (childCompleted * 100) / stepCount
		}
		childBar := BuildProgressBar(childPct, d.config.ProgressBarWidth)

		// Tree connectors: ├─ for siblings, └─ for last child
		branchConn := "├─"
		if isLastChild {
			branchConn = "└─"
		}
		// Continuation line: │ if more siblings follow, space if last
		contPrefix := "│ │   │  "
		if isLastChild {
			contPrefix = "│ │      "
		}

		// Line 1: state + template + step count
		d.printer.PrintLine("│ │   %s [%s] %s (%d steps)", branchConn, childShortState, childShortTemplate, stepCount)
		// Line 2: instance ID + progress bar
		d.printer.PrintLine("%s%s %d/%d [%s] %d%%", contPrefix, child.SagaInstanceID, childCompleted, stepCount, childBar, childPct)

		// Apply same windowing as parent saga
		childStatus := BuildSagaStatus(child.SagaInstanceID, child.SagaTemplateID, child.State, child.Steps)
		childCurrentIdx := d.findCurrentStepIndex(childStatus)
		childStart, childEnd := d.calculateWindow(childCurrentIdx, len(child.Steps))

		if childStart > 0 {
			d.printer.PrintLine("%s... %d steps before", contPrefix, childStart)
		}

		for j := childStart; j < childEnd; j++ {
			childStep := child.Steps[j]
			childIndicator := GetStepIndicator(childStep.Status)
			childStepState := ShortenStepState(childStep.Status)
			childStepTemplate := ShortenTemplateID(childStep.StepTemplateID)
			stepConn := "├─"
			if j == len(child.Steps)-1 {
				stepConn = "└─"
			}
			d.printer.PrintLine("%s%s %s %2d [%-12s]: %s", contPrefix, stepConn, childIndicator, j+1, childStepState, childStepTemplate)
		}

		if childEnd < len(child.Steps) {
			d.printer.PrintLine("%s... %d steps after", contPrefix, len(child.Steps)-childEnd)
		}
	}
}
