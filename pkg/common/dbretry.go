package common

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
)

// retryableDBPatterns contains error message substrings that indicate transient
// PostgreSQL connection failures worth retrying.
var retryableDBPatterns = []string{
	"connection reset by peer",
	"broken pipe",
	"driver: bad connection",
	"i/o timeout",
	"connection refused",
	// lib/pq wire protocol corruption: stale DataRow ('D') or other unexpected
	// message left on a pooled connection from a prior query. The pool discards
	// the bad connection (driver.ErrBadConn is set inside lib/pq), so a retry
	// gets a fresh one.
	"unexpected parse response",
	"unexpected describe response",
	"unexpected commandcomplete",
}

// IsRetryableDBError returns true if the error is a transient database
// connection error that may succeed on retry. It never retries sql.ErrNoRows.
func IsRetryableDBError(err error) bool {
	if err == nil {
		return false
	}

	// Never retry "no rows" — that's a legitimate query result
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	for _, pattern := range retryableDBPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Check for EOF separately — only when it's clearly a connection issue
	// (not sql.ErrNoRows which we already excluded above)
	if strings.Contains(errMsg, "eof") {
		return true
	}

	return false
}

// RetryOnTransient retries fn up to maxRetries times with exponential backoff
// when fn returns a transient database error. Non-retryable errors and context
// cancellation are returned immediately.
func RetryOnTransient[T any](ctx context.Context, maxRetries int, backoff time.Duration, fn func() (T, error)) (T, error) {
	var zero T

	for attempt := 0; ; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return zero, err
		}

		// Don't retry non-transient errors
		if !IsRetryableDBError(err) {
			return zero, err
		}

		// Exhausted retries
		if attempt >= maxRetries {
			return zero, err
		}

		delay := backoff * (1 << uint(attempt)) // exponential: backoff, 2*backoff, 4*backoff, ...
		if L != nil {
			L.Warn("retrying after transient DB error",
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("delay", delay),
				zap.Error(err),
			)
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}
}

// RetryOnTransientNoResult is like RetryOnTransient but for operations that
// don't return a value (e.g., Exec calls).
func RetryOnTransientNoResult(ctx context.Context, maxRetries int, backoff time.Duration, fn func() error) error {
	_, err := RetryOnTransient(ctx, maxRetries, backoff, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}
