package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidRef = errors.New("invalid ref")

const symbolicRefPrefix = "ref: "

// EncodeRef returns the canonical textual representation of ref.
func EncodeRef(ref Ref) ([]byte, error) {
	switch ref.kind {
	case refKindDirect:
		if ref.hash.IsZero() {
			return nil, fmt.Errorf(
				"%w: direct hash is zero",
				ErrInvalidRef,
			)
		}
		return []byte(ref.hash.Hex() + "\n"), nil

	case refKindSymbolic:
		if err := validateRefName(ref.target.value); err != nil {
			return nil, fmt.Errorf(
				"%w: %w",
				ErrInvalidRef,
				err,
			)
		}
		return []byte(
			symbolicRefPrefix + ref.target.value + "\n",
		), nil

	default:
		return nil, fmt.Errorf(
			"%w: unknown kind %d",
			ErrInvalidRef,
			ref.kind,
		)
	}
}

// DecodeRef validates and decodes a canonical textual reference.
func DecodeRef(encoded []byte) (Ref, error) {
	if len(encoded) == 0 {
		return Ref{}, fmt.Errorf(
			"%w: encoding is empty",
			ErrInvalidRef,
		)
	}
	if encoded[len(encoded)-1] != '\n' {
		return Ref{}, fmt.Errorf(
			"%w: encoding must end with a newline",
			ErrInvalidRef,
		)
	}

	content := string(encoded[:len(encoded)-1])
	if content == "" {
		return Ref{}, fmt.Errorf(
			"%w: content is empty",
			ErrInvalidRef,
		)
	}
	if strings.ContainsAny(content, "\r\n") {
		return Ref{}, fmt.Errorf(
			"%w: content contains a carriage return or embedded newline",
			ErrInvalidRef,
		)
	}
	if strings.HasPrefix(content, "ref:") {
		if !strings.HasPrefix(content, symbolicRefPrefix) {
			return Ref{}, fmt.Errorf(
				"%w: symbolic ref must start with %q",
				ErrInvalidRef,
				symbolicRefPrefix,
			)
		}

		target, err := NewRefName(content[len(symbolicRefPrefix):])
		if err != nil {
			return Ref{}, fmt.Errorf(
				"%w: decode symbolic target: %w",
				ErrInvalidRef,
				err,
			)
		}

		ref, err := NewSymbolicRef(target)
		if err != nil {
			return Ref{}, fmt.Errorf(
				"%w: decode symbolic target: %w",
				ErrInvalidRef,
				err,
			)
		}
		return ref, nil
	}

	hash, err := ParseHash(content)
	if err != nil {
		return Ref{}, fmt.Errorf(
			"%w: decode direct hash: %w",
			ErrInvalidRef,
			err,
		)
	}

	ref, err := NewDirectRef(hash)
	if err != nil {
		return Ref{}, fmt.Errorf(
			"%w: decode direct ref: %w",
			ErrInvalidRef,
			err,
		)
	}
	return ref, nil
}
