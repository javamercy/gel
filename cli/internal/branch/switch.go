package branch

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"Gel/internal/tree"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// SwitchOptions controls branch switch behavior.
type SwitchOptions struct {
	// Create creates the branch before switching when it does not exist.
	Create bool
	// Force bypasses overwrite-conflict checks and proceeds with checkout.
	Force bool
}

// SwitchResult reports the outcome of a switch operation.
type SwitchResult struct {
	// Branch is the target branch name.
	Branch string
	// Created is true when the target branch was created during this switch.
	Created bool
}

// SwitchService coordinates safe branch switching across refs, working tree, and index.
type SwitchService struct {
	refService      *core.RefService
	branchService   *BranchService
	objectService   *core.ObjectService
	readTreeService *tree.ReadTreeService
	treeResolver    *core.TreeResolver
	workspace       *domain.Workspace
}

// NewSwitchService creates a switch service.
func NewSwitchService(
	refService *core.RefService,
	branchService *BranchService,
	objectService *core.ObjectService,
	readTreeService *tree.ReadTreeService,
	treeResolver *core.TreeResolver,
	workspace *domain.Workspace,
) *SwitchService {
	return &SwitchService{
		refService:      refService,
		branchService:   branchService,
		objectService:   objectService,
		readTreeService: readTreeService,
		treeResolver:    treeResolver,
		workspace:       workspace,
	}
}

// Switch changes the current branch and updates working tree/index to match target commit.
// When Force is false, it aborts if local changes would be overwritten.
func (s *SwitchService) Switch(branch string, options SwitchOptions) (*SwitchResult, error) {
	targetRef, created, err := s.resolveTargetRef(branch, options.Create)
	if err != nil {
		return nil, err
	}

	headCommitHash, err := s.refService.Resolve(domain.HeadFileName)
	if err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}

	targetCommitHash, err := s.refService.Read(targetRef)
	if err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}
	if headCommitHash == targetCommitHash {
		if err := s.refService.WriteSymbolic(domain.HeadFileName, targetRef); err != nil {
			return nil, fmt.Errorf("switch: %w", err)
		}
		return &SwitchResult{Branch: branch, Created: created}, nil
	}

	if !options.Force {
		conflicts, err := s.findOverwriteConflicts(headCommitHash, targetCommitHash)
		if err != nil {
			return nil, fmt.Errorf("switch: %w", err)
		}
		if len(conflicts) > 0 {
			return nil, fmt.Errorf("switch: local changes to '%s' would be overwritten", conflicts[0])
		}
	}

	if err := s.checkoutWorkingTree(headCommitHash, targetCommitHash); err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}

	targetCommit, err := s.objectService.ReadCommit(targetCommitHash)
	if err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}
	if err := s.readTreeService.ReadTree(targetCommit.TreeHash()); err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}
	if err := s.refService.WriteSymbolic(domain.HeadFileName, targetRef); err != nil {
		return nil, fmt.Errorf("switch: %w", err)
	}
	return &SwitchResult{Branch: branch, Created: created}, nil
}

// resolveTargetRef resolves refs/heads/<branch> and optionally creates the branch.
func (s *SwitchService) resolveTargetRef(branch string, create bool) (string, bool, error) {
	targetRef := filepath.Join(domain.RefsDirName, domain.HeadsDirName, branch)

	if create {
		exists, err := s.branchService.Exists(branch)
		if err != nil {
			return "", false, fmt.Errorf("switch: %w", err)
		}
		if exists {
			return "", false, fmt.Errorf("switch: '%s': %w", branch, ErrBranchAlreadyExists)
		}
		if err := s.branchService.Create(branch, ""); err != nil {
			return "", false, fmt.Errorf("switch: %w", err)
		}
		return targetRef, true, nil
	}

	exists, err := s.refService.Exists(targetRef)
	if err != nil {
		return "", false, fmt.Errorf("switch: %w", err)
	}
	if !exists {
		return "", false, fmt.Errorf("switch: '%s': %w", branch, ErrBranchNotFound)
	}
	return targetRef, false, nil
}

// findOverwriteConflicts returns changed target paths that would lose local staged/unstaged state.
func (s *SwitchService) findOverwriteConflicts(currentCommitHash, targetCommitHash domain.Hash) (
	[]domain.NormalizedPath, error,
) {
	oldPathHashes, err := s.treeResolver.ResolveCommit(currentCommitHash)
	if err != nil {
		return nil, err
	}

	targetPathHashes, err := s.treeResolver.ResolveCommit(targetCommitHash)
	if err != nil {
		return nil, err
	}

	indexPathHashes, err := s.treeResolver.ResolveIndex()
	if err != nil {
		return nil, err
	}

	workingTreePathHashes, err := s.treeResolver.ResolveWorkingTree()
	if err != nil {
		return nil, err
	}

	paths := changedPaths(oldPathHashes, targetPathHashes)
	conflicts := make([]domain.NormalizedPath, 0)
	for _, path := range paths {
		oldState := newPathStateFromPathHashes(oldPathHashes, path)
		targetState := newPathStateFromPathHashes(targetPathHashes, path)
		indexState := newPathStateFromPathHashes(indexPathHashes, path)
		workingTreeState := newPathStateFromPathHashes(workingTreePathHashes, path)
		if hasOverwriteConflict(oldState, targetState, indexState, workingTreeState) {
			conflicts = append(conflicts, path)
		}
	}
	return conflicts, nil
}

// checkoutWorkingTree applies target commit file contents and removes paths absent in target.
func (s *SwitchService) checkoutWorkingTree(oldCommitHash, targetCommitHash domain.Hash) error {
	oldPathHashes, err := s.treeResolver.ResolveCommit(oldCommitHash)
	if err != nil {
		return err
	}

	targetPathHashes, err := s.treeResolver.ResolveCommit(targetCommitHash)
	if err != nil {
		return err
	}
	for targetPath, targetHash := range targetPathHashes {
		oldHash, ok := oldPathHashes[targetPath]
		if ok && oldHash == targetHash {
			continue
		}

		blob, err := s.objectService.ReadBlob(targetHash)
		if err != nil {
			return err
		}

		absPath, err := targetPath.ToAbsolutePath(s.workspace.RepoRoot())
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(absPath.String()), domain.DefaultDirPermission); err != nil {
			return err
		}
		if err := os.WriteFile(absPath.String(), blob.Content(), domain.DefaultFilePermission); err != nil {
			return err
		}
	}

	for oldPath := range oldPathHashes {
		if _, existsInTarget := targetPathHashes[oldPath]; existsInTarget {
			continue
		}

		absPath, err := oldPath.ToAbsolutePath(s.workspace.RepoRoot())
		if err != nil {
			return err
		}
		if err := os.Remove(absPath.String()); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// pathState models one path's snapshot state, including deletion via exists=false.
type pathState struct {
	exists bool
	hash   domain.Hash
}

// newPathStateFromPathHashes builds a pathState from a path->hash snapshot map.
func newPathStateFromPathHashes(pathHashes core.PathHashes, path domain.NormalizedPath) pathState {
	hash, ok := pathHashes[path]
	return pathState{exists: ok, hash: hash}
}

// isSamePathState reports whether two path states are equivalent.
func isSamePathState(a, b pathState) bool {
	if a.exists != b.exists {
		return false
	}
	if !a.exists {
		return true
	}
	return a.hash == b.hash
}

// isSafeState reports whether state s matches either old or target state.
func isSafeState(s, a, b pathState) bool {
	return isSamePathState(s, a) || isSamePathState(s, b)
}

// hasOverwriteConflict reports whether local index/worktree states fall outside {old,target}.
func hasOverwriteConflict(oldState, targetState, indexState, workingTreeState pathState) bool {
	indexSafe := isSafeState(indexState, oldState, targetState)
	workingTreeSafe := isSafeState(workingTreeState, oldState, targetState)
	return !indexSafe || !workingTreeSafe
}

// changedPaths returns deterministic sorted paths whose old and target states differ.
func changedPaths(oldPathHashes, targetPathHashes core.PathHashes) []domain.NormalizedPath {
	set := make(map[domain.NormalizedPath]bool)
	for oldPath, oldHash := range oldPathHashes {
		if targetHash, ok := targetPathHashes[oldPath]; !ok || targetHash != oldHash {
			set[oldPath] = true
		}
	}
	for targetPath, targetHash := range targetPathHashes {
		if oldHash, ok := oldPathHashes[targetPath]; !ok || oldHash != targetHash {
			set[targetPath] = true
		}
	}

	out := make([]domain.NormalizedPath, 0, len(set))
	for path := range set {
		out = append(out, path)
	}
	slices.SortFunc(
		out, func(a, b domain.NormalizedPath) int {
			return strings.Compare(a.String(), b.String())
		},
	)
	return out
}
