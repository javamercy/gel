package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"time"
)

type CommitTreeInput struct {
	TreeHash     domain.Hash
	ParentHashes []domain.Hash
	Message      string
}

type CommitTreeResult struct {
	CommitHash domain.Hash
}

type CommitTree struct {
	configStore *storage.ConfigStore
	objectStore *storage.ObjectStore
	now         func() time.Time
}

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
