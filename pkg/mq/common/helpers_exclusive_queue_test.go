package mqcommon

import (
	"context"
	"strings"
	"testing"
)

// TestDeclareExclusiveQueueWithTopicBindings_NilPool verifies that
// DeclareExclusiveQueueWithTopicBindings returns an error when
// the RabbitMQ channel pool has not been initialized.
func TestDeclareExclusiveQueueWithTopicBindings_NilPool(t *testing.T) {
	origPool := RabbitMQChannelPool
	defer func() { RabbitMQChannelPool = origPool }()
	RabbitMQChannelPool = nil

	ch, err := DeclareExclusiveQueueWithTopicBindings(
		context.Background(),
		"test-exchange",
		"test-queue",
		[]string{"routing.key.#"},
	)

	if err == nil {
		t.Fatal("expected error when channel pool is nil, got nil")
	}
	if ch != nil {
		t.Fatal("expected nil channel when pool is nil, got non-nil")
	}
	if !strings.Contains(err.Error(), "channel pool not initialized") {
		t.Errorf("expected error to mention 'channel pool not initialized', got: %s", err.Error())
	}
}

// TestConsumeExclusiveQueueAsync_NilPool verifies that ConsumeExclusiveQueueAsync
// returns an error and nil channel/cancel when the RabbitMQ channel pool is nil.
func TestConsumeExclusiveQueueAsync_NilPool(t *testing.T) {
	origPool := RabbitMQChannelPool
	defer func() { RabbitMQChannelPool = origPool }()
	RabbitMQChannelPool = nil

	msgCh, cancel, err := ConsumeExclusiveQueueAsync(
		context.Background(),
		"test-exchange",
		"test-queue",
		[]string{"routing.key.#"},
	)

	if err == nil {
		t.Fatal("expected error when channel pool is nil, got nil")
	}
	if msgCh != nil {
		t.Fatal("expected nil message channel when pool is nil, got non-nil")
	}
	if cancel != nil {
		t.Fatal("expected nil cancel function when pool is nil, got non-nil")
	}
	if !strings.Contains(err.Error(), "channel pool not initialized") {
		t.Errorf("expected error to mention 'channel pool not initialized', got: %s", err.Error())
	}
}

// TestDeclareExclusiveQueueWithTopicBindings_EmptyParams verifies that
// the nil-pool check takes precedence even when exchange/queue names are empty.
func TestDeclareExclusiveQueueWithTopicBindings_EmptyParams(t *testing.T) {
	origPool := RabbitMQChannelPool
	defer func() { RabbitMQChannelPool = origPool }()
	RabbitMQChannelPool = nil

	tests := []struct {
		name         string
		exchangeName string
		queueName    string
		routingKeys  []string
	}{
		{"empty exchange", "", "test-queue", []string{"#"}},
		{"empty queue", "test-exchange", "", []string{"#"}},
		{"both empty", "", "", []string{}},
		{"empty routing keys", "test-exchange", "test-queue", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := DeclareExclusiveQueueWithTopicBindings(
				context.Background(),
				tt.exchangeName,
				tt.queueName,
				tt.routingKeys,
			)
			if err == nil {
				t.Fatal("expected error when channel pool is nil, got nil")
			}
			if ch != nil {
				t.Fatal("expected nil channel when pool is nil, got non-nil")
			}
			if !strings.Contains(err.Error(), "channel pool not initialized") {
				t.Errorf("expected pool-nil error to take precedence, got: %s", err.Error())
			}
		})
	}
}

// TestConsumeExclusiveQueueAsync_EmptyParams verifies that the nil-pool
// error takes precedence over any empty parameter issues.
func TestConsumeExclusiveQueueAsync_EmptyParams(t *testing.T) {
	origPool := RabbitMQChannelPool
	defer func() { RabbitMQChannelPool = origPool }()
	RabbitMQChannelPool = nil

	tests := []struct {
		name         string
		exchangeName string
		queueName    string
		routingKeys  []string
	}{
		{"empty exchange", "", "test-queue", []string{"#"}},
		{"empty queue", "test-exchange", "", []string{"#"}},
		{"both empty", "", "", []string{}},
		{"empty routing keys", "test-exchange", "test-queue", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgCh, cancel, err := ConsumeExclusiveQueueAsync(
				context.Background(),
				tt.exchangeName,
				tt.queueName,
				tt.routingKeys,
			)
			if err == nil {
				t.Fatal("expected error when channel pool is nil, got nil")
			}
			if msgCh != nil {
				t.Fatal("expected nil message channel when pool is nil, got non-nil")
			}
			if cancel != nil {
				t.Fatal("expected nil cancel function when pool is nil, got non-nil")
			}
			if !strings.Contains(err.Error(), "channel pool not initialized") {
				t.Errorf("expected pool-nil error to take precedence, got: %s", err.Error())
			}
		})
	}
}
