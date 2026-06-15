package common

import (
	"strings"
	"testing"
)

func TestRandomHexString_Length(t *testing.T) {
	tests := []struct {
		name        string
		numBytes    int
		expectedLen int
	}{
		{"20 bytes", 20, 42},  // 0x + 40 hex chars
		{"32 bytes", 32, 66},  // 0x + 64 hex chars
		{"40 bytes", 40, 82},  // 0x + 80 hex chars
		{"64 bytes", 64, 130}, // 0x + 128 hex chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomHexString(tt.numBytes)

			if len(result) != tt.expectedLen {
				t.Errorf("RandomHexString(%d) should return %d characters, got %d: %s",
					tt.numBytes, tt.expectedLen, len(result), result)
			}

			if !strings.HasPrefix(result, "0x") {
				t.Errorf("RandomHexString should return string with 0x prefix, got: %s", result)
			}
		})
	}
}

func TestRandomHexString_Format(t *testing.T) {
	result := RandomHexString(40)

	// Verify 0x prefix
	if !strings.HasPrefix(result, "0x") {
		t.Errorf("RandomHexString should have 0x prefix, got: %s", result)
	}

	// Verify lowercase hex
	hexPart := result[2:]
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("RandomHexString should return lowercase hex, found invalid char '%c' in: %s", c, result)
		}
	}
}

func TestRandomHexString_Uniqueness(t *testing.T) {
	// Generate multiple random strings and verify they're different
	const numTests = 100
	seen := make(map[string]bool)

	for i := 0; i < numTests; i++ {
		result := RandomHexString(40)
		if seen[result] {
			t.Errorf("RandomHexString generated duplicate value: %s", result)
		}
		seen[result] = true
	}

	// All should be unique
	if len(seen) != numTests {
		t.Errorf("Expected %d unique random strings, got %d", numTests, len(seen))
	}
}

func TestRandomHexString_DifferentCalls(t *testing.T) {
	// Two calls should produce different results
	result1 := RandomHexString(40)
	result2 := RandomHexString(40)

	if result1 == result2 {
		t.Errorf("RandomHexString should produce different values on different calls, both returned: %s", result1)
	}
}

func TestRandomHexString_VariousLengths(t *testing.T) {
	// Test that different lengths work correctly
	lengths := []int{1, 5, 10, 20, 32, 40, 64, 100}

	for _, length := range lengths {
		result := RandomHexString(length)
		expectedLen := 2 + (length * 2) // 0x + 2 hex chars per byte

		if len(result) != expectedLen {
			t.Errorf("RandomHexString(%d) returned wrong length: expected %d, got %d",
				length, expectedLen, len(result))
		}
	}
}
