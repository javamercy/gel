package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeSHA256 returns the lowercase hexadecimal SHA-256 digest of data.
func ComputeSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
