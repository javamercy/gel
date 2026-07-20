package domain

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// Commit represents an immutable Gel commit object.
//
// The zero value is invalid.
type Commit struct {
	treeHash     Hash
	parentHashes []Hash
	author       CommitIdentity
	committer    CommitIdentity
	message      string
}

// NewCommit validates and constructs a commit with a canonical encoded body.
func NewCommit(
	treeHash Hash,
	parentHashes []Hash,
	author CommitIdentity,
	committer CommitIdentity,
	message string,
) (*Commit, error) {
	commit := &Commit{
		treeHash:     treeHash,
		parentHashes: slices.Clone(parentHashes),
		author:       author,
		committer:    committer,
		message:      message,
	}
	if err := validateCommit(commit); err != nil {
		return nil, fmt.Errorf(
			"create commit: %w",
			err,
		)
	}
	return commit, nil
}

// TreeHash returns the root tree object hash.
func (c *Commit) TreeHash() Hash {
	return c.treeHash
}

// ParentHashes returns the parent commit hashes in semantic order.
func (c *Commit) ParentHashes() []Hash {
	return slices.Clone(c.parentHashes)
}

// Author returns the commit author identity.
func (c *Commit) Author() CommitIdentity {
	return c.author
}

// Committer returns the commit committer identity.
func (c *Commit) Committer() CommitIdentity {
	return c.committer
}

// Message returns the commit message.
func (c *Commit) Message() string {
	return c.message
}

// Type returns ObjectTypeCommit.
func (c *Commit) Type() ObjectType {
	return ObjectTypeCommit
}

func (c *Commit) isObject() {}

func validateCommit(commit *Commit) error {
	if commit == nil {
		return fmt.Errorf("commit is nil")
	}
	if commit.treeHash.IsZero() {
		return fmt.Errorf(
			"tree hash cannot be zero",
		)
	}

	seenParents := make(map[Hash]struct{}, len(commit.parentHashes))
	for i, parentHash := range commit.parentHashes {
		if parentHash.IsZero() {
			return fmt.Errorf(
				"parent %d hash cannot be zero",
				i,
			)
		}
		if _, exists := seenParents[parentHash]; exists {
			return fmt.Errorf(
				"parent %d duplicates hash %q",
				i,
				parentHash.Hex(),
			)
		}
		seenParents[parentHash] = struct{}{}
	}
	if err := validateCommitIdentity(commit.author); err != nil {
		return fmt.Errorf(
			"author: %w",
			err,
		)
	}
	if err := validateCommitIdentity(commit.committer); err != nil {
		return fmt.Errorf(
			"committer: %w",
			err,
		)
	}
	if !utf8.ValidString(commit.message) {
		return fmt.Errorf("message is not valid UTF-8")
	}
	if strings.ContainsRune(commit.message, 0) {
		return fmt.Errorf("message contains NUL")
	}
	if strings.ContainsRune(commit.message, '\r') {
		return fmt.Errorf(
			"message contains carriage return",
		)
	}
	return nil
}
