package common

import (
	"strings"
	"testing"
)

func TestHashSHA256_KnownInput(t *testing.T) {
	// Test with a known input
	input := "test-slot-123"
	result := HashSHA256(input, 40)

	// Verify format
	if !strings.HasPrefix(result, "0x") {
		t.Errorf("HashSHA256 should return string with 0x prefix, got: %s", result)
	}

	// Verify length: 0x + (40 bytes * 2 hex chars) = 82 characters
	expectedLen := 2 + (40 * 2)
	if len(result) != expectedLen {
		t.Errorf("HashSHA256(40 bytes) should return %d characters, got %d: %s", expectedLen, len(result), result)
	}

	// Verify lowercase hex
	hexPart := result[2:]
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("HashSHA256 should return lowercase hex, found invalid char '%c' in: %s", c, result)
		}
	}

	// Verify deterministic (same input produces same output)
	result2 := HashSHA256(input, 40)
	if result != result2 {
		t.Errorf("HashSHA256 should be deterministic, got different results: %s vs %s", result, result2)
	}
}

func TestHashSHA256_DifferentLengths(t *testing.T) {
	input := "test-input"

	tests := []struct {
		name        string
		outputBytes int
		expectedLen int
	}{
		{"20 bytes", 20, 42},  // 0x + 40 hex chars
		{"32 bytes", 32, 66},  // 0x + 64 hex chars
		{"40 bytes", 40, 82},  // 0x + 80 hex chars
		{"64 bytes", 64, 130}, // 0x + 128 hex chars (exceeds SHA256 but should work)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashSHA256(input, tt.outputBytes)
			if len(result) != tt.expectedLen {
				t.Errorf("HashSHA256(%d bytes) should return %d characters, got %d: %s",
					tt.outputBytes, tt.expectedLen, len(result), result)
			}
		})
	}
}

func TestHashSHA256_EmptyString(t *testing.T) {
	result := HashSHA256("", 40)

	if !strings.HasPrefix(result, "0x") {
		t.Errorf("HashSHA256 of empty string should have 0x prefix, got: %s", result)
	}

	if len(result) != 82 {
		t.Errorf("HashSHA256 of empty string should return 82 characters, got %d: %s", len(result), result)
	}
}

func TestGetDirectOrderHash_Deterministic(t *testing.T) {
	hash1 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-1", "exch_oid_abc")
	hash2 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-1", "exch_oid_abc")

	if hash1 != hash2 {
		t.Errorf("GetDirectOrderHash should be deterministic, got %s vs %s", hash1, hash2)
	}

	if !strings.HasPrefix(string(hash1), "0x") {
		t.Errorf("GetDirectOrderHash should return 0x-prefixed hex, got: %s", hash1)
	}

	// SHA-512/384 = 384 bits = 48 bytes = 96 hex chars + "0x" = 98
	if len(hash1) != 98 {
		t.Errorf("GetDirectOrderHash should return 98 characters (0x + 96 hex), got %d: %s", len(hash1), hash1)
	}
}

func TestGetDirectOrderHash_DifferentInputs(t *testing.T) {
	hash1 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-1", "exch_oid_abc")
	hash2 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "43", "ext-oid-1", "exch_oid_abc")

	if hash1 == hash2 {
		t.Errorf("GetDirectOrderHash should produce different hashes for different orderId")
	}

	hash3 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-2", "exch_oid_abc")
	if hash1 == hash3 {
		t.Errorf("GetDirectOrderHash should produce different hashes for different externalOid")
	}

	hash4 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-1", "exch_oid_xyz")
	if hash1 == hash4 {
		t.Errorf("GetDirectOrderHash should produce different hashes for different exchangeOid")
	}
}

func TestGetDirectOrderHash_CaseInsensitive(t *testing.T) {
	hash1 := GetDirectOrderHash("31337", "Anvil", "0xABC123", "0xDEF456", "42", "Ext-OID-1", "Exch_OID_ABC")
	hash2 := GetDirectOrderHash("31337", "anvil", "0xabc123", "0xdef456", "42", "ext-oid-1", "exch_oid_abc")

	if hash1 != hash2 {
		t.Errorf("GetDirectOrderHash should be case-insensitive, got %s vs %s", hash1, hash2)
	}
}

func TestHashSHA256_DifferentInputs(t *testing.T) {
	// Different inputs should produce different hashes
	hash1 := HashSHA256("input1", 40)
	hash2 := HashSHA256("input2", 40)

	if hash1 == hash2 {
		t.Errorf("HashSHA256 should produce different hashes for different inputs")
	}
}
