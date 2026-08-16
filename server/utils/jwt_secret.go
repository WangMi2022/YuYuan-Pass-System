package utils

import "strings"

// IsWeakJWTSigningKey reports whether a JWT signing key is a known placeholder or too short for production use.
func IsWeakJWTSigningKey(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "qmplus", "changeme", "change-me", "default":
		return true
	default:
		return len([]byte(value)) < 32
	}
}
