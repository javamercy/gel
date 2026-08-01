package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// RefName identifies a validated reference name.
type RefName struct {
	value string
}

// NewRefName parses value as a reference name.
//
// It returns an error when value does not satisfy Gel reference-name rules.
func NewRefName(value string) (RefName, error) {
	if err := validateRefName(value); err != nil {
		return RefName{}, fmt.Errorf(
			"parse ref name: %w",
			err,
		)
	}
	return RefName{value: value}, nil
}

// String returns the reference name.
func (n RefName) String() string {
	return n.value
}

func validateRefName(value string) error {
	if value == "" {
		return errors.New("value is empty")
	}
	for _, r := range value {
		if r == '\\' || unicode.IsControl(r) {
			return errors.New("value contains a backslash or control character")
		}
	}
	if value == "HEAD" {
		return nil
	}
	if !strings.HasPrefix(value, "refs/") {
		return errors.New(`value must be "HEAD" or start with "refs/"`)
	}

	components := strings.SplitSeq(value, "/")
	for component := range components {
		if component == "" || component == "." || component == ".." {
			return errors.New("value must not have empty, '.' or '..' components")
		}
	}
	return nil
}

// RefKind identifies whether a reference stores an object hash or another
// reference name.
type RefKind uint8

const (
	refKindInvalid RefKind = iota
	refKindDirect
	refKindSymbolic
)

// Ref represents either a direct reference to an object hash or a symbolic
// reference to another reference name.
//
// The zero value is invalid and is neither direct nor symbolic.
type Ref struct {
	kind   RefKind
	hash   Hash
	target RefName
}

// NewDirectRef returns a direct reference to hash.
//
// It returns an error when hash is zero.
func NewDirectRef(hash Hash) (Ref, error) {
	if hash.IsZero() {
		return Ref{}, errors.New("hash is zero")
	}
	return Ref{
		kind: refKindDirect,
		hash: hash,
	}, nil
}

// NewSymbolicRef returns a symbolic reference targeting target.
//
// It returns an error when target is not a valid reference name.
func NewSymbolicRef(target RefName) (Ref, error) {
	if err := validateRefName(target.value); err != nil {
		return Ref{}, fmt.Errorf("invalid symbolic target: %w", err)
	}
	return Ref{
		kind:   refKindSymbolic,
		target: target,
	}, nil
}

// DirectHash returns the referenced object hash and true for a direct reference.
//
// It returns a zero hash and false for a symbolic or invalid reference.
func (r Ref) DirectHash() (Hash, bool) {
	return r.hash, r.kind == refKindDirect
}

// SymbolicTarget returns the target name and true for a symbolic reference.
//
// It returns a zero RefName and false for a direct or invalid reference.
func (r Ref) SymbolicTarget() (RefName, bool) {
	return r.target, r.kind == refKindSymbolic
}
