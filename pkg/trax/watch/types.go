// Package watch provides saga progress watching and display functionality.
// It can be used by CLI tools and E2E tests to monitor saga execution
// with a consistent, rich progress display.
package watch

import (
	"time"
)

// StepStatus represents the status of a saga step
type StepStatus struct {
	StepInstanceID     string      `json:"instance_id"`
	StepTemplateID     string      `json:"saga_step_template_id"`
	Status             string      `json:"state"`
	Result             interface{} `json:"result_data"`
	CompensationResult interface{} `json:"compensation_result_data,omitempty"`
	ExecutionError     string      `json:"execution_error,omitempty"`
}

// ChildSagaInfo holds display information for a child/sub-saga
type ChildSagaInfo struct {
	SagaInstanceID       string       `json:"saga_instance_id"`
	SagaTemplateID       string       `json:"saga_template_id"`
	State                string       `json:"state"`
	ParentStepInstanceID string       `json:"parent_saga_step_instance_id"`
	Steps                []StepStatus `json:"steps,omitempty"`
}

// SagaStatus represents the full saga status for watch-style display
type SagaStatus struct {
	SagaInstanceID     string                     `json:"saga_instance_id"`
	TraceID            string                     `json:"trace_id"`
	Status             string                     `json:"status"`
	SagaTemplateID     string                     `json:"saga_template_id"`
	CompensationReason string                     `json:"compensation_reason,omitempty"`
	StepSummary        StepSummary                `json:"step_summary"`
	StepStatuses       []StepStatus               `json:"step_statuses,omitempty"`
	ChildSagasByStep   map[string][]ChildSagaInfo `json:"child_sagas_by_step,omitempty"`
}

// StepSummary provides a summary of saga step states
type StepSummary struct {
	Total     int            `json:"total"`
	Completed int            `json:"completed"`
	Pending   int            `json:"pending"`
	Failed    int            `json:"failed"`
	ByState   map[string]int `json:"by_state"`
}

// StepTiming tracks timing information for a step
type StepTiming struct {
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	IsCompleted bool
}

// Timing tracks timing information for saga watch display
type Timing struct {
	SagaStartTime  time.Time
	StateStartTime time.Time
	LastState      string
	StepTimings    map[string]*StepTiming // keyed by step template ID
}

// NewTiming creates a new Timing tracker
func NewTiming() *Timing {
	return &Timing{
		SagaStartTime:  time.Now(),
		StateStartTime: time.Now(),
		StepTimings:    make(map[string]*StepTiming),
	}
}

// Config holds configuration for the saga watcher
type Config struct {
	// WindowSize controls how many steps to show around the current step.
	// 0 means show all steps.
	WindowSize int

	// ProgressBarWidth is the width of the progress bar in characters.
	ProgressBarWidth int
}

// DefaultConfig returns the default watcher configuration
func DefaultConfig() Config {
	return Config{
		WindowSize:       5,
		ProgressBarWidth: 20,
	}
}
