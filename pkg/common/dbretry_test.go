package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestIsRetryableDBError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"sql.ErrNoRows", sql.ErrNoRows, false},
		{"wrapped sql.ErrNoRows", fmt.Errorf("query failed: %w", sql.ErrNoRows), false},
		{"connection reset by peer", errors.New("read tcp 172.18.0.6:55918->172.18.0.3:5432: read: connection reset by peer"), true},
		{"broken pipe", errors.New("write: broken pipe"), true},
		{"driver bad connection", errors.New("driver: bad connection"), true},
		{"i/o timeout", errors.New("dial tcp 172.18.0.3:5432: i/o timeout"), true},
		{"connection refused", errors.New("dial tcp 172.18.0.3:5432: connect: connection refused"), true},
		{"EOF", errors.New("unexpected EOF"), true},
		{"wrapped transient", fmt.Errorf("failed to get slot: %w", errors.New("read tcp: connection reset by peer")), true},
		{"unique constraint", errors.New("pq: duplicate key value violates unique constraint"), false},
		{"syntax error", errors.New("pq: syntax error at or near \"SELECT\""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableDBError(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryableDBError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestRetryOnTransient_Success(t *testing.T) {
	calls := 0
	result, err := RetryOnTransient(context.Background(), 3, time.Millisecond, func() (string, error) {
		calls++
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: %v", result)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryOnTransient_RetriesThenSucceeds(t *testing.T) {
	calls := 0
	result, err := RetryOnTransient(context.Background(), 3, time.Millisecond, func() (int, error) {
		calls++
		if calls < 3 {
			return 0, errors.New("connection reset by peer")
		}
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryOnTransient_ExhaustsRetries(t *testing.T) {
	calls := 0
	_, err := RetryOnTransient(context.Background(), 3, time.Millisecond, func() (string, error) {
		calls++
		return "", errors.New("connection reset by peer")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// 1 initial + 3 retries = 4 calls
	if calls != 4 {
		t.Fatalf("expected 4 calls (1 initial + 3 retries), got %d", calls)
	}
}

func TestRetryOnTransient_NonRetryableError(t *testing.T) {
	calls := 0
	_, err := RetryOnTransient(context.Background(), 3, time.Millisecond, func() (string, error) {
		calls++
		return "", errors.New("pq: unique constraint violation")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (no retry for non-transient), got %d", calls)
	}
}

func TestRetryOnTransient_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	calls := 0
	_, err := RetryOnTransient(ctx, 3, time.Millisecond, func() (string, error) {
		calls++
		return "", errors.New("connection reset by peer")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should return after first attempt since context is already cancelled
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryOnTransientNoResult(t *testing.T) {
	calls := 0
	err := RetryOnTransientNoResult(context.Background(), 3, time.Millisecond, func() error {
		calls++
		if calls < 2 {
			return errors.New("broken pipe")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}
