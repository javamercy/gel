package commit

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"Gel/internal/tree"
	"errors"
	"fmt"
)

type CommitService struct {
	writeTreeService  *tree.WriteTreeService
	commitTreeService *CommitTreeService
	refService        *core.RefService
	objectService     *core.ObjectService
}

// NewCommitService creates and initializes a CommitService with the provided dependencies.
func NewCommitService(
	writeTreeService *tree.WriteTreeService,
	commitTreeService *CommitTreeService,
	refService *core.RefService,
	objectService *core.ObjectService,
) *CommitService {
	return &CommitService{
		writeTreeService:  writeTreeService,
		commitTreeService: commitTreeService,
		refService:        refService,
		objectService:     objectService,
	}
}

// Commit writes the current index tree and advances the current branch.
// It refuses no-op commits when the new tree matches the parent tree and
// rejects an empty initial commit when both parent and tree are empty.
func (c *CommitService) Commit(message string) error {
	var parentHashes []domain.Hash
	headRef, err := c.refService.ReadSymbolic(domain.HeadFileName)
	if err != nil {
		return fmt.Errorf("commit: failed to read HEAD: %w", err)
	}

	parentHash, err := c.refService.Read(headRef)
	if errors.Is(err, core.ErrRefNotFound) {
		parentHashes = nil
	} else if err != nil {
		return fmt.Errorf("commit: failed to read parent ref '%s': %w", headRef, err)
	}

	treeHash, err := c.writeTreeService.WriteTree()
	if err != nil {
		return fmt.Errorf("commit: failed to write tree: %w", err)
	}

	if !parentHash.IsZero() {
		parentHashes = append(parentHashes, parentHash)
		parentCommit, err := c.objectService.ReadCommit(parentHash)
		if err != nil {
			return fmt.Errorf("commit: failed to read parent commit '%s': %w", parentHash, err)
		}
		if parentCommit.TreeHash == treeHash {
			return ErrNothingToCommit
		}
	}

	commitHash, err := c.commitTreeService.CommitTree(treeHash, message, parentHashes)
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	if err := c.refService.Write(headRef, commitHash); err != nil {
		return fmt.Errorf("commit: failed to update ref '%s': %w", headRef, err)
	}
	return nil
}
