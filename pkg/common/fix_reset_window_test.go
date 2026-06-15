package common

import (
	"testing"
	"time"
)

func TestFIXResetWindowAllows(t *testing.T) {
	// Anchor every test at a deterministic UTC moment.
	at := func(hour, min int) time.Time {
		return time.Date(2026, 5, 8, hour, min, 0, 0, time.UTC)
	}

	type tc struct {
		name string
		spec string
		now  time.Time
		want bool
	}

	cases := []tc{
		{name: "empty spec disables window", spec: "", now: at(12, 0), want: false},
		{name: "garbage spec disables window", spec: "not-a-window", now: at(12, 0), want: false},
		{name: "single-bound spec disables", spec: "12:00", now: at(12, 0), want: false},
		{name: "zero-width window disables", spec: "12:00-12:00", now: at(12, 0), want: false},

		// Same-day window [09:00, 17:00)
		{name: "same-day: before start", spec: "09:00-17:00", now: at(8, 59), want: false},
		{name: "same-day: at start", spec: "09:00-17:00", now: at(9, 0), want: true},
		{name: "same-day: middle", spec: "09:00-17:00", now: at(12, 30), want: true},
		{name: "same-day: at end (exclusive)", spec: "09:00-17:00", now: at(17, 0), want: false},
		{name: "same-day: after end", spec: "09:00-17:00", now: at(17, 1), want: false},

		// Wrap-midnight window [23:30, 00:30)
		{name: "wrap: before pre-midnight start", spec: "23:30-00:30", now: at(23, 29), want: false},
		{name: "wrap: at pre-midnight start", spec: "23:30-00:30", now: at(23, 30), want: true},
		{name: "wrap: just after midnight", spec: "23:30-00:30", now: at(0, 0), want: true},
		{name: "wrap: at post-midnight end (exclusive)", spec: "23:30-00:30", now: at(0, 30), want: false},
		{name: "wrap: well outside", spec: "23:30-00:30", now: at(12, 0), want: false},

		// Whitespace tolerance
		{name: "whitespace tolerance", spec: " 09:00 - 17:00 ", now: at(10, 0), want: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(FIXResetWindowEnv, c.spec)
			got := FIXResetWindowAllows(c.now)
			if got != c.want {
				t.Fatalf("spec=%q now=%s want=%v got=%v", c.spec, c.now.Format(time.RFC3339), c.want, got)
			}
		})
	}
}

func TestFIXResetWindowSpec(t *testing.T) {
	t.Setenv(FIXResetWindowEnv, "  09:00-17:00  ")
	if got := FIXResetWindowSpec(); got != "09:00-17:00" {
		t.Fatalf("expected trimmed spec, got %q", got)
	}
}
