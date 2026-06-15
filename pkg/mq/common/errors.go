package mqcommon

import (
	"errors"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Sentinel errors for RabbitMQ operations
var (
	ErrPublishTimeout   = errors.New("publish confirmation timeout")
	ErrPublishNacked    = errors.New("message nacked by broker")
	ErrChannelClosed    = errors.New("channel closed")
	ErrConnectionClosed = errors.New("connection closed")
)

// IsRetryableError checks if error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check sentinel errors
	if errors.Is(err, ErrPublishTimeout) ||
		errors.Is(err, ErrChannelClosed) ||
		errors.Is(err, ErrConnectionClosed) {
		return true
	}

	// Check AMQP library errors
	if errors.Is(err, amqp.ErrClosed) {
		return true
	}

	// Check AMQP error codes (typed error checking)
	var amqpErr *amqp.Error
	if errors.As(err, &amqpErr) {
		switch amqpErr.Code {
		case 320: // CONNECTION_FORCED
			return true
		case 504: // CHANNEL_ERROR - "channel/connection is not open"
			return true
		case 505: // UNEXPECTED_FRAME
			return true
		case 506: // RESOURCE_ERROR
			return true
		}
	}

	// Fallback: check error message for known patterns (last resort)
	// This handles cases where error is wrapped or not typed
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"channel/connection is not open",
		"connection closed",
		"channel closed",
		"broken pipe",
		"connection reset",
		"eof",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
