package mqcommon

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublisherWithConfirms wraps a channel to add publisher confirmation support
type PublisherWithConfirms struct {
	ch             *amqp.Channel
	confirmChan    chan amqp.Confirmation
	nextSeqNo      uint64
	mu             sync.Mutex
	confirmTimeout time.Duration

	// Per-sequence confirmation tracking to avoid race conditions
	pendingConfirms map[uint64]chan amqp.Confirmation
	confirmsMu      sync.RWMutex
	routerDone      chan struct{}
}

// NewPublisherWithConfirms creates a publisher with confirms enabled
func NewPublisherWithConfirms(ch *amqp.Channel, timeout time.Duration) (*PublisherWithConfirms, error) {
	if ch == nil {
		return nil, fmt.Errorf("channel is nil")
	}

	// Enable publisher confirms
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("failed to enable confirms: %w", err)
	}

	// Create confirmation channel (buffered to avoid blocking broker)
	confirmChan := ch.NotifyPublish(make(chan amqp.Confirmation, 100))

	p := &PublisherWithConfirms{
		ch:              ch,
		confirmChan:     confirmChan,
		nextSeqNo:       1,
		confirmTimeout:  timeout,
		pendingConfirms: make(map[uint64]chan amqp.Confirmation),
		routerDone:      make(chan struct{}),
	}

	// Start confirmation router goroutine to prevent race conditions
	// when multiple goroutines publish concurrently
	go p.routeConfirmations()

	return p, nil
}

// routeConfirmations reads from confirmChan and routes confirmations to waiting goroutines
// This prevents the race condition where confirmations are received by the wrong goroutine
func (p *PublisherWithConfirms) routeConfirmations() {
	for {
		select {
		case <-p.routerDone:
			// Publisher is being closed, drain pending confirmations with error
			p.confirmsMu.Lock()
			for seqNo, ch := range p.pendingConfirms {
				close(ch) // Signal timeout/error to waiting goroutine
				delete(p.pendingConfirms, seqNo)
			}
			p.confirmsMu.Unlock()
			return

		case confirm, ok := <-p.confirmChan:
			if !ok {
				// Channel closed, publisher is done
				p.confirmsMu.Lock()
				for seqNo, ch := range p.pendingConfirms {
					close(ch)
					delete(p.pendingConfirms, seqNo)
				}
				p.confirmsMu.Unlock()
				return
			}

			// Route confirmation to the correct waiting goroutine
			p.confirmsMu.Lock()
			if ch, exists := p.pendingConfirms[confirm.DeliveryTag]; exists {
				ch <- confirm
				close(ch)
				delete(p.pendingConfirms, confirm.DeliveryTag)
			}
			// If no pending confirmation found, it's likely a duplicate or stale - ignore it
			p.confirmsMu.Unlock()
		}
	}
}

// Close stops the confirmation router and cleans up resources
func (p *PublisherWithConfirms) Close() {
	close(p.routerDone)
}

// PublishWithConfirm publishes message and waits for broker confirmation
func (p *PublisherWithConfirms) PublishWithConfirm(
	ctx context.Context,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp.Publishing,
) error {
	// Create per-sequence confirmation channel
	responseChan := make(chan amqp.Confirmation, 1)

	// Get sequence number and publish atomically to prevent sequence mismatch
	// when multiple goroutines publish concurrently
	p.mu.Lock()
	seqNo := p.nextSeqNo
	p.nextSeqNo++

	// Register this sequence number BEFORE publishing to avoid race condition
	// where confirmation arrives before we register
	p.confirmsMu.Lock()
	p.pendingConfirms[seqNo] = responseChan
	p.confirmsMu.Unlock()

	// Publish message (must be inside lock to ensure sequence order matches RabbitMQ's delivery tags)
	err := p.ch.PublishWithContext(ctx, exchange, routingKey, mandatory, immediate, msg)
	p.mu.Unlock()

	if err != nil {
		// Remove from pending confirmations on publish error
		p.confirmsMu.Lock()
		delete(p.pendingConfirms, seqNo)
		p.confirmsMu.Unlock()
		return fmt.Errorf("publish failed: %w", err)
	}

	// Wait for confirmation from router goroutine
	select {
	case confirm, ok := <-responseChan:
		if !ok {
			// Channel closed, likely due to publisher shutdown
			return fmt.Errorf("publisher confirmation channel closed (seq %d)", seqNo)
		}
		if confirm.DeliveryTag != seqNo {
			// This should never happen with the new router, but keep the check
			return fmt.Errorf("sequence mismatch: expected %d, got %d", seqNo, confirm.DeliveryTag)
		}
		if !confirm.Ack {
			return ErrPublishNacked
		}
		return nil // ✅ Success - broker confirmed

	case <-time.After(p.confirmTimeout):
		// Remove from pending confirmations on timeout
		p.confirmsMu.Lock()
		delete(p.pendingConfirms, seqNo)
		p.confirmsMu.Unlock()
		return fmt.Errorf("%w: after %v (seq %d)", ErrPublishTimeout, p.confirmTimeout, seqNo)

	case <-ctx.Done():
		// Remove from pending confirmations on context cancellation
		p.confirmsMu.Lock()
		delete(p.pendingConfirms, seqNo)
		p.confirmsMu.Unlock()
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}
