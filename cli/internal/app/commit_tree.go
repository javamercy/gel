package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"time"
)

// CommitTreeInput specifies the tree, parents, and message for a new commit.
type CommitTreeInput struct {
	// TreeHash identifies the root tree object and must not be zero.
	TreeHash domain.Hash
	// ParentHashes identifies existing parent commits in semantic order. An
	// empty slice creates a root commit.
	ParentHashes []domain.Hash
	// Message is the commit message.
	Message string
}

// CommitTreeResult contains the identifier of the written commit object.
type CommitTreeResult struct {
	// CommitHash is the SHA-256 identifier of the new commit.
	CommitHash domain.Hash
}

// CommitTree creates commit objects from tree and parent objects.
type CommitTree struct {
	configStore *storage.ConfigStore
	objectStore *storage.ObjectStore
	now         func() time.Time
}

// NewCommitTree returns a CommitTree backed by configStore and objectStore.
//
// Both stores must not be nil when Run is called.
func NewCommitTree(
	configStore *storage.ConfigStore,
	objectStore *storage.ObjectStore,
) *CommitTree {
	return &CommitTree{
		configStore: configStore,
		objectStore: objectStore,
		now:         time.Now,
	}
}

// Run validates the input, creates a commit, and writes it to object storage.
//
// Run verifies that TreeHash identifies a tree and that every parent hash
// identifies a commit. It reads the author name and email from the user.name
// and user.email configuration entries, and uses the current local time for
// both the author and committer identities. Run returns an error when either
// identity is not configured or any input, object, configuration, or write
// operation is invalid. Errors return a zero CommitTreeResult.
func (c *CommitTree) Run(input CommitTreeInput) (CommitTreeResult, error) {
	if err := validateCommitTreeInput(input); err != nil {
		return CommitTreeResult{}, err
	}

	tree, err := c.objectStore.Read(input.TreeHash)
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf(
			"read tree object %q: %w",
			input.TreeHash,
			err,
		)
	}

	if tree.Type() != domain.ObjectTypeTree {
		return CommitTreeResult{}, fmt.Errorf(
			"object %q is %s, not tree",
			input.TreeHash,
			tree.Type(),
		)
	}

	for _, parentHash := range input.ParentHashes {
		parent, err := c.objectStore.Read(parentHash)
		if err != nil {
			return CommitTreeResult{}, fmt.Errorf(
				"read parent commit %q: %w",
				parentHash,
				err,
			)
		}
		if parent.Type() != domain.ObjectTypeCommit {
			return CommitTreeResult{}, fmt.Errorf(
				"object %q is %s, not commit",
				parentHash,
				parent.Type(),
			)
		}
	}

	config, err := c.configStore.Load()
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf(
			"load config: %w",
			err,
		)
	}

	name, found := config.Get("user", "name")
	if !found {
		return CommitTreeResult{}, errors.New("user.name is not configured")
	}

	email, found := config.Get("user", "email")
	if !found {
		return CommitTreeResult{}, errors.New("user.email is not configured")
	}

	now := c.now()
	_, offsetSeconds := now.Zone()

	identity, err := domain.NewCommitIdentity(
		name,
		email,
		now.Unix(),
		offsetSeconds/60,
	)
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf(
			"create commit identity: %w",
			err,
		)
	}

	commit, err := domain.NewCommit(
		input.TreeHash,
		input.ParentHashes,
		identity,
		identity,
		input.Message,
	)
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf("create commit: %w", err)
	}

	commitHash, err := c.objectStore.Write(commit)
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf("write commit object: %w", err)
	}

	return CommitTreeResult{
		CommitHash: commitHash,
	}, nil
}

func validateCommitTreeInput(input CommitTreeInput) error {
	if input.TreeHash.IsZero() {
		return errors.New("tree hash is zero")
	}

	seenParents := make(map[domain.Hash]struct{}, len(input.ParentHashes))
	for i, hash := range input.ParentHashes {
		if hash.IsZero() {
			return fmt.Errorf("parent %d hash is zero", i)
		}
		if _, exists := seenParents[hash]; exists {
			return fmt.Errorf(
				"duplicate parent commit %q",
				hash,
			)
		}
		seenParents[hash] = struct{}{}
	}
	return nil
}
