package mqcommon

import (
	"errors"
	"testing"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrPublishTimeout is retryable",
			err:      ErrPublishTimeout,
			expected: true,
		},
		{
			name:     "ErrChannelClosed is retryable",
			err:      ErrChannelClosed,
			expected: true,
		},
		{
			name:     "ErrConnectionClosed is retryable",
			err:      ErrConnectionClosed,
			expected: true,
		},
		{
			name:     "ErrPublishNacked is not retryable",
			err:      ErrPublishNacked,
			expected: false,
		},
		{
			name:     "wrapped ErrPublishTimeout is retryable",
			err:      errors.New("wrapped: " + ErrPublishTimeout.Error()),
			expected: false, // string wrapping doesn't work with errors.Is
		},
		{
			name:     "wrapped with fmt.Errorf %%w is retryable",
			err:      errors.Unwrap(ErrPublishTimeout),
			expected: false, // after unwrap, it's nil
		},
		{
			name:     "random error is not retryable",
			err:      errors.New("some random error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Verify all error types are properly defined
	if ErrPublishTimeout == nil {
		t.Error("ErrPublishTimeout should not be nil")
	}
	if ErrPublishNacked == nil {
		t.Error("ErrPublishNacked should not be nil")
	}
	if ErrChannelClosed == nil {
		t.Error("ErrChannelClosed should not be nil")
	}
	if ErrConnectionClosed == nil {
		t.Error("ErrConnectionClosed should not be nil")
	}
}

func TestErrorMessages(t *testing.T) {
	// Verify error messages are descriptive
	tests := []struct {
		err     error
		wantMsg string
	}{
		{ErrPublishTimeout, "publish confirmation timeout"},
		{ErrPublishNacked, "message nacked by broker"},
		{ErrChannelClosed, "channel closed"},
		{ErrConnectionClosed, "connection closed"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.wantMsg {
			t.Errorf("error message mismatch: got %q, want %q", tt.err.Error(), tt.wantMsg)
		}
	}
}
