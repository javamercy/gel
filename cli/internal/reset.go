package internal

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"Gel/internal/tree"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResetMode selects how much repository state reset updates.
type ResetMode int

const (
	// ResetModeSoft moves HEAD only.
	ResetModeSoft ResetMode = iota
	// ResetModeMixed moves HEAD and resets the index to the target tree.
	ResetModeMixed
	// ResetModeHard moves HEAD, resets the index, and rewrites the working tree.
	ResetModeHard
)

// ResetOptions configures reset execution mode.
type ResetOptions struct {
	Mode ResetMode
}

// ResetResult reports the commit hash reset resolved and applied.
type ResetResult struct {
	TargetHash domain.Hash
}

// ResetService moves repository refs and snapshots to a target revision.
type ResetService struct {
	refService      *core.RefService
	objectService   *core.ObjectService
	readTreeService *tree.ReadTreeService
	treeResolver    *core.TreeResolver
	commitResolver  *core.CommitResolver
	workspace       *domain.Workspace
}

// NewResetService creates a reset service with the dependencies needed to move refs and snapshots.
func NewResetService(
	refService *core.RefService,
	objectService *core.ObjectService,
	readTreeService *tree.ReadTreeService,
	treeResolver *core.TreeResolver,
	commitResolver *core.CommitResolver,
	workspace *domain.Workspace,
) *ResetService {
	return &ResetService{
		refService:      refService,
		objectService:   objectService,
		readTreeService: readTreeService,
		treeResolver:    treeResolver,
		commitResolver:  commitResolver,
		workspace:       workspace,
	}
}

// Reset applies the requested reset mode to the target revision and returns the resolved commit hash.
func (r *ResetService) Reset(target string, options ResetOptions) (*ResetResult, error) {
	if err := validateMode(options.Mode); err != nil {
		return nil, fmt.Errorf("reset: %w", err)
	}

	if strings.TrimSpace(target) == "" {
		target = domain.HeadFileName
	}

	targetHash, err := r.commitResolver.Resolve(target)
	if err != nil {
		return nil, fmt.Errorf("reset: %w", err)
	}

	targetCommit, err := r.objectService.ReadCommit(targetHash)
	if err != nil {
		return nil, fmt.Errorf("reset: %w", err)
	}
	if options.Mode == ResetModeMixed || options.Mode == ResetModeHard {
		if err := r.readTreeService.ReadTree(targetCommit.TreeHash()); err != nil {
			return nil, fmt.Errorf("reset: %w", err)
		}
	}
	if options.Mode == ResetModeHard {
		if err := r.checkoutWorkingTree(targetHash); err != nil {
			return nil, fmt.Errorf("reset: %w", err)
		}
	}
	if err := r.moveHEADPointer(targetHash); err != nil {
		return nil, fmt.Errorf("reset: %w", err)
	}
	return &ResetResult{
		TargetHash: targetHash,
	}, nil
}

// moveHEADPointer advances the current symbolic branch ref to the resolved target hash.
func (r *ResetService) moveHEADPointer(hash domain.Hash) error {
	ref, err := r.refService.ReadSymbolic(domain.HeadFileName)
	if err != nil {
		return err
	}
	return r.refService.Write(ref, hash)
}

// checkoutWorkingTree makes the working tree match the target commit during hard reset.
func (r *ResetService) checkoutWorkingTree(targetHash domain.Hash) error {
	headPathHashes, err := r.treeResolver.ResolveHEAD()
	if err != nil {
		return err
	}

	targetPathHashes, err := r.treeResolver.ResolveCommit(targetHash)
	if err != nil {
		return err
	}

	workingTreePathHashes, err := r.treeResolver.ResolveWorkingTree()
	if err != nil {
		return err
	}

	deletePaths := collectDeletePaths(headPathHashes, targetPathHashes)
	writePaths := collectWritePaths(targetPathHashes, workingTreePathHashes)

	for _, path := range deletePaths {
		absPath, err := path.ToAbsolutePath(r.workspace.RepoRoot())
		if err != nil {
			return err
		}
		if err := os.Remove(absPath.String()); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to delete '%s': %w", absPath, err)
		}
		if err := pruneEmptyParentDirs(absPath.String(), r.workspace.RepoRoot().String()); err != nil {
			return fmt.Errorf("failed to prune empty parents for '%s': %w", absPath, err)
		}
	}
	for _, path := range writePaths {
		blobHash := targetPathHashes[path]
		blob, err := r.objectService.ReadBlob(blobHash)
		if err != nil {
			return err
		}

		absPath, err := path.ToAbsolutePath(r.workspace.RepoRoot())
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(absPath.String()), domain.DefaultDirPermission); err != nil {
			return fmt.Errorf("failed to create parent dir for '%s': %w", absPath.String(), err)
		}
		if err := os.WriteFile(absPath.String(), blob.Content(), domain.DefaultFilePermission); err != nil {
			return fmt.Errorf("failed to write '%s': %w", absPath, err)
		}
	}
	return nil
}

// pruneEmptyParentDirs removes empty ancestor directories up to repoRoot.
// This keeps hard reset working when deleting the last file from a directory
// and then recreating that path as a regular file.
func pruneEmptyParentDirs(filePath, repoRoot string) error {
	dir := filepath.Dir(filePath)
	repoRoot = filepath.Clean(repoRoot)

	for dir != repoRoot {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if len(entries) != 0 {
			return nil
		}

		if err := os.Remove(dir); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		dir = filepath.Dir(dir)
	}

	return nil
}

// collectDeletePaths returns tracked paths present in the old snapshot but absent from the new one.
func collectDeletePaths(oldPathHashes, newPathHashes core.PathHashes) (deletePaths []domain.NormalizedPath) {
	for path := range oldPathHashes {
		if _, inNew := newPathHashes[path]; !inNew {
			deletePaths = append(deletePaths, path)
		}
	}
	return
}

// collectWritePaths returns paths whose target content differs from the current working tree.
func collectWritePaths(targetPathHashes, workingTreePathHashes core.PathHashes) (writePaths []domain.NormalizedPath) {
	for path, targetHash := range targetPathHashes {
		if workingHash, inWorking := workingTreePathHashes[path]; !inWorking || workingHash != targetHash {
			writePaths = append(writePaths, path)
		}
	}
	return
}

// validateMode rejects unknown reset modes before any repository state is changed.
func validateMode(mode ResetMode) error {
	switch mode {
	case ResetModeSoft, ResetModeMixed, ResetModeHard:
		return nil
	default:
		return fmt.Errorf("invalid reset mode: %d", mode)
	}
}
