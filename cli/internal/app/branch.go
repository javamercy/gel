package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"strings"
)

const headsRefPrefix = "refs/heads"

// BranchListItem describes one local branch.
type BranchListItem struct {
	// Name is the branch name without the refs/heads/ prefix.
	Name string
	// IsCurrent reports whether this is the branch targeted by HEAD.
	IsCurrent bool
}

// BranchListResult contains the local branches discovered by Branch.List.
type BranchListResult struct {
	// Branches contains one item for each local branch.
	Branches []BranchListItem
}

// Branch manages local branch references.
type Branch struct {
	refStore    *storage.RefStore
	objectStore *storage.ObjectStore
}

// NewBranch returns a Branch backed by refStore and objectStore.
//
// Both stores must not be nil when the returned Branch is used.
func NewBranch(
	refStore *storage.RefStore,
	objectStore *storage.ObjectStore,
) *Branch {
	return &Branch{
		refStore:    refStore,
		objectStore: objectStore,
	}
}

// List returns the local branches and marks the branch targeted by HEAD.
//
// It returns an error when HEAD cannot be read or does not target a local
// branch, or when branch references cannot be listed. Branch names in the
// result omit the refs/heads/ prefix.
func (b *Branch) List() (BranchListResult, error) {
	var result BranchListResult

	currentBranchName, err := b.currentBranchName()
	if err != nil {
		return result, fmt.Errorf(
			"read current branch: %w",
			err,
		)
	}

	headsName, err := domain.NewRefName(headsRefPrefix)
	if err != nil {
		return result, fmt.Errorf(
			"create heads ref name: %w",
			err,
		)
	}

	branchNames, err := b.refStore.List(headsName)
	if err != nil {
		return result, fmt.Errorf(
			"list branch ref names: %w",
			err,
		)
	}

	for _, name := range branchNames {
		result.Branches = append(
			result.Branches,
			BranchListItem{
				Name:      strings.TrimPrefix(name.String(), headsRefPrefix+"/"),
				IsCurrent: name == currentBranchName,
			},
		)
	}
	return result, nil
}

// Create creates a local branch named shortName.
//
// An empty startPoint uses the current branch tip and fails when HEAD is
// unborn. A non-empty startPoint may be a local branch name or the hash of an
// existing commit. The new branch points directly to the resolved commit and
// returns an error matching [storage.ErrRefExists] when it already exists.
func (b *Branch) Create(shortName, startPoint string) error {
	newBranchName, err := branchRefName(shortName)
	if err != nil {
		return err
	}

	startHash, err := b.resolveStartPoint(startPoint)
	if err != nil {
		return fmt.Errorf(
			"create branch ref: %w",
			err,
		)
	}

	branchRef, err := domain.NewDirectRef(startHash)
	if err != nil {
		return fmt.Errorf(
			"create branch ref: %w",
			err,
		)
	}

	if err := b.refStore.Create(newBranchName, branchRef); err != nil {
		return fmt.Errorf(
			"create branch %q: %w",
			shortName,
			err,
		)
	}
	return nil
}

// Delete removes the local branch named shortName.
//
// Delete never removes the current branch. When force is false, the branch
// must be fully merged into the current branch; when force is true, that
// ancestry check is skipped. It returns an error matching
// [storage.ErrRefNotExist] when the branch does not exist.
func (b *Branch) Delete(shortName string, force bool) error {
	currentBranchName, err := b.currentBranchName()
	if err != nil {
		return err
	}

	branchName, err := branchRefName(shortName)
	if err != nil {
		return err
	}

	if branchName == currentBranchName {
		return fmt.Errorf(
			"cannot delete the current branch %q",
			branchName,
		)
	}

	if !force {
		branchHash, err := b.branchCommitHash(branchName)
		if err != nil {
			return err
		}

		currentHash, err := b.branchCommitHash(currentBranchName)
		if err != nil {
			return err
		}

		isMerged, err := b.isCommitAncestor(branchHash, currentHash)
		if err != nil {
			return fmt.Errorf(
				"check whether branch %q is merged: %w",
				shortName,
				err,
			)
		}
		if !isMerged {
			return fmt.Errorf(
				"branch %q is not fully merged",
				shortName,
			)
		}
	}

	if err := b.refStore.Delete(branchName); err != nil {
		return fmt.Errorf(
			"delete branch %q: %w",
			branchName,
			err,
		)
	}
	return nil
}

func (b *Branch) resolveStartPoint(startPoint string) (domain.Hash, error) {
	if startPoint == "" {
		currentBranchName, err := b.currentBranchName()
		if err != nil {
			if errors.Is(err, storage.ErrRefNotExist) {
				return domain.Hash{}, errors.New(
					"cannot create branch from unborn HEAD",
				)
			}
			return domain.Hash{}, err
		}
		return b.branchCommitHash(currentBranchName)
	}

	if hash, err := domain.ParseHash(startPoint); err == nil {
		_, err := b.objectStore.ReadAs[*domain.Commit](hash)
		if err != nil {
			return domain.Hash{}, fmt.Errorf(
				"read start point %q: %w",
				startPoint,
				err,
			)
		}
		return hash, nil
	}

	branchName, err := branchRefName(startPoint)
	if err != nil {
		return domain.Hash{}, err
	}
	return b.branchCommitHash(branchName)
}

func (b *Branch) currentBranchName() (domain.RefName, error) {
	headName, err := domain.NewRefName(domain.HeadFileName)
	if err != nil {
		return domain.RefName{}, fmt.Errorf(
			"create HEAD ref name: %w",
			err,
		)
	}

	headRef, err := b.refStore.Read(headName)
	if err != nil {
		return domain.RefName{}, fmt.Errorf(
			"read HEAD: %w",
			err,
		)
	}

	branchName, ok := headRef.SymbolicTarget()
	if !ok {
		return domain.RefName{}, errors.New(
			"detached HEAD is unsupported",
		)
	}

	if !strings.HasPrefix(branchName.String(), headsRefPrefix+"/") {
		return domain.RefName{}, fmt.Errorf(
			"HEAD targets non-branch ref %q",
			branchName.String(),
		)
	}
	return branchName, nil
}

func (b *Branch) branchCommitHash(branchName domain.RefName) (domain.Hash, error) {
	ref, err := b.refStore.Read(branchName)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"read branch ref %q: %w",
			branchName.String(),
			err,
		)
	}

	hash, ok := ref.DirectHash()
	if !ok {
		return domain.Hash{}, fmt.Errorf(
			"branch ref %q is not direct",
			branchName.String(),
		)
	}

	if _, err := b.objectStore.ReadAs[*domain.Commit](hash); err != nil {
		return domain.Hash{}, fmt.Errorf(
			"read branch commit %q: %w",
			branchName.String(),
			err,
		)
	}

	return hash, nil
}

func (b *Branch) isCommitAncestor(
	ancestor domain.Hash,
	descendant domain.Hash,
) (bool, error) {
	pending := []domain.Hash{descendant}
	seen := make(map[domain.Hash]struct{})

	for len(pending) > 0 {
		last := len(pending) - 1
		hash := pending[last]
		pending = pending[:last]

		if hash == ancestor {
			return true, nil
		}
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}

		commit, err := b.objectStore.ReadAs[*domain.Commit](hash)
		if err != nil {
			return false, fmt.Errorf(
				"read commit %q: %w",
				hash,
				err,
			)
		}

		pending = append(
			pending,
			commit.ParentHashes()...,
		)
	}
	return false, nil
}

func branchRefName(shortName string) (domain.RefName, error) {
	if shortName == "" {
		return domain.RefName{}, errors.New("branch name is empty")
	}

	refName, err := domain.NewRefName(headsRefPrefix + "/" + shortName)
	if err != nil {
		return domain.RefName{}, fmt.Errorf(
			"parse branch name %q: %w",
			shortName,
			err,
		)
	}
	return refName, nil
}
