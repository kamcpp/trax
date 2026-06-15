package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"strings"
)

func SecureRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(62))
		b[i] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"[num.Int64()]
	}
	return string(b)
}

// alphanumericNoSimilarAlphabet is the 55-char alphabet used by
// SecureRandomStringNoSimilar — alphanumerics minus visually
// ambiguous chars (0/O/o, 1/l/I/i). Same alphabet the marketmgr
// participant order id uses.
const alphanumericNoSimilarAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz"

// SecureRandomStringNoSimilar returns an n-char cryptographically
// random string drawn from a 55-char alphanumeric alphabet that
// excludes visually ambiguous characters (0 / O / o / 1 / l / I / i).
// Use for human-copy-paste-friendly secrets and identifiers like API
// keys where one mistyped char makes the value useless.
func SecureRandomStringNoSimilar(n int) string {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alphanumericNoSimilarAlphabet)))
	for i := range b {
		num, _ := rand.Int(rand.Reader, max)
		b[i] = alphanumericNoSimilarAlphabet[num.Int64()]
	}
	return string(b)
}

// SecureRandom3x3DigitId returns a 9-digit decimal id in the canonical
// `XXX-XXX-XXX` shape (e.g. `229-231-393`). Each digit is in [1..9] so
// no zero-only / leading-zero blocks confuse downstream parsers and the
// id reads cleanly. Cryptographically random.
func SecureRandomId() string {
	const digits = "123456789"
	out := make([]byte, 11) // 9 digits + 2 hyphens
	idx := 0
	for i := 0; i < 9; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		out[idx] = digits[num.Int64()]
		idx++
		if i == 2 || i == 5 {
			out[idx] = '-'
			idx++
		}
	}
	return string(out)
}

// GenerateTaggedIid returns a canonical IID in the form
// `{category}_{prefix}_{XXX-XXX-XXX}` (or `{category}_{XXX-XXX-XXX}`
// when [prefix] is empty / has no usable characters). The [prefix]
// segment is the owner's naming tag — preserved as-is (case included)
// after dropping non-alphanumerics. The 9-digit decimal id is
// generated via [SecureRandomId].
//
// Examples:
//
//	GenerateTaggedIid("acc",     "TokeniseCSD") → "acc_TokeniseCSD_482-715-936"
//	GenerateTaggedIid("legstr",  "ACME")        → "legstr_ACME_193-624-857"
//	GenerateTaggedIid("legmech", "ACME")        → "legmech_ACME_715-302-948"
//	GenerateTaggedIid("acc",     "")            → "acc_482-715-936"
func GenerateTaggedIid(category, prefix string) string {
	tag := sanitizeIidPrefix(prefix)
	if tag == "" {
		return fmt.Sprintf("%s_%s", category, SecureRandomId())
	}
	return fmt.Sprintf("%s_%s_%s", category, tag, SecureRandomId())
}

// GenerateAccountIid wraps GenerateTaggedIid("acc", prefix). Owner
// prefix is the participant's naming prefix for participant-owned
// accounts, or the legal structure's prefix for LS-owned ones.
func GenerateAccountIid(prefix string) string {
	return GenerateTaggedIid("acc", prefix)
}

// GenerateLegalStructureIid wraps GenerateTaggedIid("legstr", prefix).
// Prefix is the LS naming prefix supplied by the wizard.
func GenerateLegalStructureIid(prefix string) string {
	return GenerateTaggedIid("legstr", prefix)
}

// GenerateLegalMechanismIid wraps GenerateTaggedIid("legmech", prefix).
// Prefix is inherited from the owning legal structure.
func GenerateLegalMechanismIid(prefix string) string {
	return GenerateTaggedIid("legmech", prefix)
}

// GenerateLegalMechanismDeploymentIid wraps
// GenerateTaggedIid("legmechdep", prefix). Prefix is inherited from
// the owning legal mechanism (which inherits from its LS).
func GenerateLegalMechanismDeploymentIid(prefix string) string {
	return GenerateTaggedIid("legmechdep", prefix)
}

// sanitizeIidPrefix drops everything that isn't [A-Za-z0-9] from
// [prefix] so the resulting IID stays parseable / URL-safe / shell-
// safe. Case is preserved — the prefix is meant to be the operator-
// supplied owner tag verbatim. Empty inputs (or all-junk inputs)
// collapse to "" so callers can decide whether to render a 3-segment
// or 2-segment IID (see [GenerateTaggedIid]).
func sanitizeIidPrefix(prefix string) string {
	var b strings.Builder
	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// RandomHexString generates a cryptographically secure random hex string of specified byte length.
// Returns hex-encoded string with 0x prefix, lowercase (e.g., "0xabc123...")
func RandomHexString(numBytes int) string {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	return "0x" + strings.ToLower(hex.EncodeToString(b))
}

// SeededRandomHexString generates a deterministic random hex string using a seed.
// The seed is hashed with SHA256 to derive a numeric seed for the PRNG.
// Returns hex-encoded string with 0x prefix, lowercase (e.g., "0xabc123...")
func SeededRandomHexString(seed string, numBytes int) string {
	// Hash the seed to get a deterministic numeric value
	hash := sha256.Sum256([]byte(seed))
	// Use first 8 bytes as int64 seed
	numericSeed := int64(0)
	for i := 0; i < 8; i++ {
		numericSeed = (numericSeed << 8) | int64(hash[i])
	}

	// Create seeded random generator
	rng := mathrand.New(mathrand.NewSource(numericSeed))

	// Generate random bytes using seeded generator
	b := make([]byte, numBytes)
	if _, err := rng.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate seeded random bytes: %v", err))
	}

	return "0x" + strings.ToLower(hex.EncodeToString(b))
}
