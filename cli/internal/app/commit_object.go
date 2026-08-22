package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"time"
)

type commitObjectInput struct {
	treeHash     domain.Hash
	parentHashes []domain.Hash
	message      string
}

func writeCommitObject(
	input commitObjectInput,
	configStore *storage.ConfigStore,
	objectStore *storage.ObjectStore,
	now time.Time,
) (domain.Hash, error) {
	if err := validateCommitObjectInput(input); err != nil {
		return domain.Hash{}, err
	}

	_, err := objectStore.ReadAs[*domain.Tree](input.treeHash)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"read tree object %q: %w",
			input.treeHash,
			err,
		)
	}
	for _, parentHash := range input.parentHashes {
		if _, err := objectStore.ReadAs[*domain.Commit](parentHash); err != nil {
			return domain.Hash{}, fmt.Errorf(
				"read parent commit %q: %w",
				parentHash,
				err,
			)
		}
	}

	config, err := configStore.Load()
	if err != nil {
		return domain.Hash{}, fmt.Errorf("load config: %w", err)
	}

	name, found := config.Get("user", "name")
	if !found {
		return domain.Hash{}, errors.New("user.name is not configured")
	}

	email, found := config.Get("user", "email")
	if !found {
		return domain.Hash{}, errors.New("user.email is not configured")
	}

	_, offsetSeconds := now.Zone()
	identity, err := domain.NewCommitIdentity(
		name,
		email,
		now.Unix(),
		offsetSeconds/60,
	)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"create commit identity: %w",
			err,
		)
	}

	commit, err := domain.NewCommit(
		input.treeHash,
		input.parentHashes,
		identity,
		identity,
		input.message,
	)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"create commit: %w",
			err,
		)
	}

	hash, err := objectStore.Write(commit)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"write commit object: %w",
			err,
		)
	}
	return hash, nil
}

func validateCommitObjectInput(input commitObjectInput) error {
	if input.treeHash.IsZero() {
		return errors.New("tree hash is zero")
	}

	seenParents := make(
		map[domain.Hash]struct{},
		len(input.parentHashes),
	)
	for i, hash := range input.parentHashes {
		if hash.IsZero() {
			return fmt.Errorf("parent %d hash is zero", i)
		}
		if _, exists := seenParents[hash]; exists {
			return fmt.Errorf(
				"parent %d duplicates hash %q",
				i,
				hash,
			)
		}
		seenParents[hash] = struct{}{}
	}
	return nil
}
