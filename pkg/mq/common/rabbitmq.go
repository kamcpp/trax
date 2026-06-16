package mqcommon

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	agoracommon "github.com/xshyft/trax/pkg/common"
)

var (
	RabbitMQURL         string = ""
	RabbitMQConnection  *amqp.Connection
	RabbitMQChannelPool *ChannelPool // Thread-safe channel pool (replaces single shared channel)
	channelMutex        sync.RWMutex
	reconnectListeners  []chan struct{}
	reconnectMutex      sync.Mutex
)

func init() {
	reconnectListeners = make([]chan struct{}, 0)
}

// NotifyReconnect sends a reconnection notification to all listeners
func NotifyReconnect() {
	reconnectMutex.Lock()
	defer reconnectMutex.Unlock()

	agoracommon.L.Info("[RabbitMQ] Broadcasting reconnection to listeners",
		zap.Int("listenerCount", len(reconnectListeners)))

	// Send to all registered listeners
	for i, ch := range reconnectListeners {
		select {
		case ch <- struct{}{}:
			agoracommon.L.Info("[RabbitMQ] Sent reconnection notification to listener",
				zap.Int("listenerIndex", i))
		default:
			agoracommon.L.Warn("[RabbitMQ] Listener channel full, notification dropped",
				zap.Int("listenerIndex", i))
		}
	}

	agoracommon.L.Info("[RabbitMQ] Reconnection broadcast complete")
}

// RegisterReconnectListener creates and registers a new channel for reconnection notifications.
// Returns the notification channel and a cleanup function that MUST be called when the listener
// is no longer needed (e.g., when the consumer goroutine exits) to prevent listener leaks.
func RegisterReconnectListener() (<-chan struct{}, func()) {
	reconnectMutex.Lock()
	defer reconnectMutex.Unlock()

	ch := make(chan struct{}, 10) // Buffered to avoid blocking
	reconnectListeners = append(reconnectListeners, ch)

	listenerNum := len(reconnectListeners)
	agoracommon.L.Info("[RabbitMQ] Registered reconnection listener",
		zap.Int("listenerNum", listenerNum),
		zap.Int("total", listenerNum))

	// Return cleanup function to unregister the listener
	cleanup := func() {
		reconnectMutex.Lock()
		defer reconnectMutex.Unlock()

		// Find and remove this listener from the slice
		for i, listener := range reconnectListeners {
			if listener == ch {
				// Remove by swapping with last element and truncating
				reconnectListeners[i] = reconnectListeners[len(reconnectListeners)-1]
				reconnectListeners = reconnectListeners[:len(reconnectListeners)-1]
				close(ch)
				agoracommon.L.Info("[RabbitMQ] Unregistered reconnection listener",
					zap.Int("remaining", len(reconnectListeners)))
				return
			}
		}
	}

	return ch, cleanup
}
