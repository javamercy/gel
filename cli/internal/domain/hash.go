package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	// HashByteLength is the length of a SHA-256 hash in raw bytes.
	HashByteLength = sha256.Size

	// HashHexLength is the length of a SHA-256 hash encoded as hexadecimal.
	HashHexLength = HashByteLength * 2
)

// ErrInvalidHash indicates an invalid SHA-256 hash representation.
var ErrInvalidHash = errors.New("invalid hash")

// Hash represents a SHA-256 object identifier.
type Hash [HashByteLength]byte

// ParseHash constructs a Hash from its 64-character lowercase
// hexadecimal representation.
func ParseHash(hexHash string) (Hash, error) {
	if len(hexHash) != HashHexLength {
		return Hash{}, fmt.Errorf(
			"%w: got %d characters, want %d",
			ErrInvalidHash,
			len(hexHash),
			HashHexLength,
		)
	}

	for _, c := range hexHash {
		if !isLowerHex(c) {
			return Hash{}, fmt.Errorf(
				"%w: expected lowercase hexadecimal characters",
				ErrInvalidHash,
			)
		}
	}

	decoded, err := hex.DecodeString(hexHash)
	if err != nil {
		return Hash{}, fmt.Errorf(
			"%w: decode hexadecimal representation: %v",
			ErrInvalidHash,
			err,
		)
	}

	return NewHash(decoded)
}

// NewHash constructs a Hash from its 32-byte binary representation.
func NewHash(data []byte) (Hash, error) {
	if len(data) != HashByteLength {
		return Hash{}, fmt.Errorf(
			"%w: got %d bytes, want %d",
			ErrInvalidHash,
			len(data),
			HashByteLength,
		)
	}

	var hash Hash
	copy(hash[:], data)
	return hash, nil
}

// Hex returns the lowercase hexadecimal representation of h.
func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

// IsZero reports whether h is the zero-value Hash.
func (h Hash) IsZero() bool {
	return h == Hash{}
}

// String returns the lowercase hexadecimal representation of h.
func (h Hash) String() string {
	return h.Hex()
}

func isLowerHex(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f'
}
