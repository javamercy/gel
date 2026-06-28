package inspect

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RestoreMode selects source/target snapshots for restore operations.
type RestoreMode int

const (
	// RestoreModeIndexVsWorkingTree restores working tree paths from index entries.
	RestoreModeIndexVsWorkingTree RestoreMode = iota
	// RestoreModeHEADVsIndex restores index entries from HEAD.
	RestoreModeHEADVsIndex
	// RestoreModeCommitVsWorkingTree restores working tree paths from a source commit.
	RestoreModeCommitVsWorkingTree
	// RestoreModeCommitVsIndex restores index entries from a source commit.
	RestoreModeCommitVsIndex
)

// RestoreOptions configures how restore should resolve and apply source data.
type RestoreOptions struct {
	// Mode controls source/target behavior for the restore operation.
	Mode RestoreMode
	// Source is a revision name/hash used by commit-based restore modes.
	Source string
}

// RestoreService applies restore operations to working tree files and index entries.
type RestoreService struct {
	indexService   *core.IndexService
	objectService  *core.ObjectService
	refService     *core.RefService
	treeResolver   *core.TreeResolver
	changeDetector *core.ChangeDetector
	workspace      *domain.Workspace
}

// NewRestoreService creates a restore service.
func NewRestoreService(
	indexService *core.IndexService,
	objectService *core.ObjectService,
	refService *core.RefService,
	treeResolver *core.TreeResolver,
	changeDetector *core.ChangeDetector,
	workspace *domain.Workspace,
) *RestoreService {
	return &RestoreService{
		indexService:   indexService,
		objectService:  objectService,
		refService:     refService,
		treeResolver:   treeResolver,
		changeDetector: changeDetector,
		workspace:      workspace,
	}
}

// Restore executes a restore operation for the provided absolute paths.
func (r *RestoreService) Restore(paths []domain.AbsolutePath, options RestoreOptions) error {
	var err error
	switch options.Mode {
	case RestoreModeIndexVsWorkingTree:
		err = r.restoreIndexVsWorkingTree(paths)
	case RestoreModeHEADVsIndex:
		err = r.restoreHEADVsIndex(paths)
	case RestoreModeCommitVsWorkingTree:
		commitHash, resolveErr := r.resolveSource(options.Source)
		if resolveErr != nil {
			return fmt.Errorf("restore: %w", resolveErr)
		}
		err = r.restoreCommitVsWorkingTree(commitHash, paths)
	case RestoreModeCommitVsIndex:
		commitHash, resolveErr := r.resolveSource(options.Source)
		if resolveErr != nil {
			return fmt.Errorf("restore: %w", resolveErr)
		}
		err = r.restoreCommitVsIndex(commitHash, paths)
	default:
		err = ErrInvalidRestoreMode
	}
	if err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// restoreIndexVsWorkingTree updates working tree files from index blob snapshots.
func (r *RestoreService) restoreIndexVsWorkingTree(paths []domain.AbsolutePath) error {
	index, err := r.indexService.Read()
	if err != nil {
		return err
	}
	for _, path := range paths {
		normalizedPath, err := path.ToNormalizedPath(r.workspace.RepoRoot())
		if err != nil {
			return err
		}

		entry, _ := index.FindEntry(normalizedPath)
		if entry == nil {
			continue
		}

		changeResult, err := r.changeDetector.DetectFileChange(entry)
		if err != nil {
			return err
		}
		if changeResult.FileState == core.FileStateUnchanged {
			continue
		}

		blob, err := r.objectService.ReadBlob(entry.Hash)
		if err != nil {
			return err
		}

		dir := filepath.Dir(path.String())
		if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
			return err
		}
		if err := os.WriteFile(path.String(), blob.Body(), domain.DefaultFilePermission); err != nil {
			return err
		}
	}
	return nil
}

// restoreHEADVsIndex resets index entries to match HEAD commit contents.
func (r *RestoreService) restoreHEADVsIndex(paths []domain.AbsolutePath) error {
	commitHash, err := r.refService.Resolve(domain.HeadFileName)
	if err != nil {
		return err
	}
	return r.restoreCommitVsIndex(commitHash, paths)
}

// restoreCommitVsWorkingTree updates working tree files to match the source commit.
// Paths missing in the source commit are removed from the working tree.
func (r *RestoreService) restoreCommitVsWorkingTree(commitHash domain.Hash, paths []domain.AbsolutePath) error {
	commitPathHashes, err := r.treeResolver.ResolveCommit(commitHash)
	if err != nil {
		return err
	}

	workingTreePathHashes, err := r.treeResolver.ResolveWorkingTree()
	if err != nil {
		return err
	}

	for _, path := range paths {
		normalizedPath, err := path.ToNormalizedPath(r.workspace.RepoRoot())
		if err != nil {
			return err
		}

		commitHash, inCommit := commitPathHashes[normalizedPath]
		if !inCommit {
			err := os.Remove(path.String())
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}

		workingTreeHash, inWorkingTree := workingTreePathHashes[normalizedPath]
		if !inWorkingTree || commitHash != workingTreeHash {
			blob, err := r.objectService.ReadBlob(commitHash)
			if err != nil {
				return err
			}

			dir := filepath.Dir(path.String())
			if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
				return err
			}
			if err := os.WriteFile(path.String(), blob.Body(), domain.DefaultFilePermission); err != nil {
				return err
			}
		}
	}
	return nil
}

// restoreCommitVsIndex updates index entries to match the source commit.
// Paths missing in the source commit are removed from index.
func (r *RestoreService) restoreCommitVsIndex(commitHash domain.Hash, paths []domain.AbsolutePath) error {
	commit, err := r.objectService.ReadCommit(commitHash)
	if err != nil {
		return err
	}

	index, err := r.indexService.Read()
	if err != nil {
		return err
	}
	for _, path := range paths {
		normalizedPath, err := path.ToNormalizedPath(r.workspace.RepoRoot())
		if err != nil {
			return err
		}

		treeEntry, err := r.treeResolver.LookupPathInTree(commit.TreeHash, normalizedPath)
		if err != nil && !errors.Is(err, core.ErrPathNotFoundInTree) {
			return err
		}

		inCommit := err == nil
		indexEntry, _ := index.FindEntry(normalizedPath)
		inIndex := indexEntry != nil

		switch {
		case !inCommit && !inIndex:
			continue
		case inCommit && inIndex && indexEntry.Hash == treeEntry.Hash():
			continue
		case inCommit:
			newIndexEntry := domain.NewEmptyIndexEntry(normalizedPath, treeEntry.Hash(), uint32(treeEntry.Mode()))
			index.SetEntry(newIndexEntry)
		default:
			index.RemoveEntry(normalizedPath)
		}
	}
	return r.indexService.Write(index)
}

// resolveSource resolves a restore source string to a commit hash.
// Supported inputs are HEAD, main, full refs/* paths, local branch names, and commit hashes.
func (r *RestoreService) resolveSource(source string) (domain.Hash, error) {
	switch source {
	case domain.HeadFileName:
		return r.refService.Resolve(domain.HeadFileName)
	case domain.DefaultBranchName:
		ref := filepath.Join(domain.RefsDirName, domain.HeadsDirName, domain.DefaultBranchName)
		return r.refService.Read(ref)
	}

	if strings.HasPrefix(source, domain.RefsDirName+"/") {
		return r.refService.Read(source)
	}

	branchRef := filepath.Join(domain.RefsDirName, domain.HeadsDirName, source)
	if hash, err := r.refService.Read(branchRef); err == nil {
		return hash, nil
	} else if !errors.Is(err, core.ErrRefNotFound) {
		return domain.Hash{}, err
	}
	return domain.NewHashFromHex(source)
}
