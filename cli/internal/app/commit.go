package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"strings"
	"time"
)

type CommitInput struct {
	Message string
}

type CommitResult struct {
	CommitHash domain.Hash
}

type Commit struct {
	refStore    *storage.RefStore
	indexStore  *storage.IndexStore
	objectStore *storage.ObjectStore
	configStore *storage.ConfigStore
	now         func() time.Time
}

func NewCommit(
	refStore *storage.RefStore,
	indexStore *storage.IndexStore,
	objectStore *storage.ObjectStore,
	configStore *storage.ConfigStore,
) *Commit {
	return &Commit{
		refStore:    refStore,
		indexStore:  indexStore,
		objectStore: objectStore,
		configStore: configStore,
		now:         time.Now,
	}
}

func (c *Commit) Run(input CommitInput) (CommitResult, error) {
	var result CommitResult

	if err := validateCommitInput(input); err != nil {
		return result, err
	}

	headName, err := domain.NewRefName(domain.HeadFileName)
	if err != nil {
		return result, fmt.Errorf(
			"create HEAD ref name: %w",
			err,
		)
	}

	headRef, err := c.refStore.Read(headName)
	if err != nil {
		return result, fmt.Errorf(
			"read HEAD: %w",
			err,
		)
	}

	branchName, ok := headRef.SymbolicTarget()
	if !ok {
		return result, errors.New("detached HEAD is unsupported")
	}

	var parentCommit *domain.Commit
	var parentHashes []domain.Hash
	branchRef, err := c.refStore.Read(branchName)

	switch {
	case errors.Is(err, storage.ErrRefNotExist):
		break

	case err != nil:
		return result, fmt.Errorf(
			"read branch ref %q: %w",
			branchName.String(),
			err,
		)

	default:
		parentHash, ok := branchRef.DirectHash()
		if !ok {
			return result, fmt.Errorf(
				"branch ref %q is not direct",
				branchName.String(),
			)
		}

		parentCommit, err = c.objectStore.ReadAs[*domain.Commit](parentHash)
		if err != nil {
			return result, fmt.Errorf(
				"read parent commit %q: %w",
				parentHash,
				err,
			)
		}

		parentHashes = []domain.Hash{parentHash}
	}

	var entries []domain.IndexEntry
	index, err := c.indexStore.Load()

	switch {
	case errors.Is(err, storage.ErrIndexNotExist):
		break

	case err != nil:
		return result, fmt.Errorf(
			"load index: %w",
			err,
		)

	default:
		entries = index.Entries()
	}

	if parentCommit != nil && len(entries) == 0 {
		return result, errors.New("nothing to commit")
	}

	treeHash, err := writeIndexTree(entries, c.objectStore)
	if err != nil {
		return result, fmt.Errorf(
			"write index tree: %w",
			err,
		)
	}

	if parentCommit != nil && parentCommit.TreeHash() == treeHash {
		return result, errors.New("nothing to commit")
	}

	commitHash, err := writeCommitObject(
		commitObjectInput{
			treeHash:     treeHash,
			parentHashes: parentHashes,
			message:      input.Message,
		},
		c.configStore,
		c.objectStore,
		c.now(),
	)
	if err != nil {
		return result, fmt.Errorf(
			"write commit object: %w",
			err,
		)
	}

	newBranchRef, err := domain.NewDirectRef(commitHash)
	if err != nil {
		return result, fmt.Errorf(
			"create branch ref: %w",
			err,
		)
	}

	if err := c.refStore.Write(branchName, newBranchRef); err != nil {
		return result, fmt.Errorf(
			"write branch ref %q: %w",
			branchName.String(),
			err,
		)
	}

	result.CommitHash = commitHash
	return result, nil
}

func validateCommitInput(input CommitInput) error {
	if strings.TrimSpace(input.Message) == "" {
		return errors.New("commit message is empty")
	}
	return nil
}
