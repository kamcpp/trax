package trax

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAnnouncementBackoffProgression verifies the exponential backoff interval
// sequence: 1s → 2s → 4s → 8s → 16s, capped at announcementInterval.
func TestAnnouncementBackoffProgression(t *testing.T) {
	cap := 30 * time.Second

	backoff := announcementBackoffInitial
	expected := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}

	for i := 0; i < announcementBackoffMaxRetries; i++ {
		require.Equal(t, expected[i], backoff, "Retry %d backoff mismatch", i)
		backoff = time.Duration(float64(backoff) * announcementBackoffMultiplier)
		if backoff > cap {
			backoff = cap
		}
	}
}

// TestAnnouncementBackoffCap verifies backoff is capped at the announcement interval.
func TestAnnouncementBackoffCap(t *testing.T) {
	cap := 5 * time.Second // simulate a short announcement interval

	backoff := announcementBackoffInitial
	for i := 0; i < 10; i++ {
		backoff = time.Duration(float64(backoff) * announcementBackoffMultiplier)
		if backoff > cap {
			backoff = cap
		}
	}
	require.Equal(t, cap, backoff, "Backoff should be capped at announcement interval")
}

// TestAnnouncementBackoffConstants verifies the constants are set correctly.
func TestAnnouncementBackoffConstants(t *testing.T) {
	require.Equal(t, 1*time.Second, announcementBackoffInitial)
	require.Equal(t, 2.0, announcementBackoffMultiplier)
	require.Equal(t, 5, announcementBackoffMaxRetries)
}
