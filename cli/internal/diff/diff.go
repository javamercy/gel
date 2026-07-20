package diff

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"fmt"
	"os"
	"slices"
	"strings"
)

type DiffMode int

const (
	// DiffModeIndexVsWorkingTree compares index snapshot against working tree.
	DiffModeIndexVsWorkingTree DiffMode = iota
	// DiffModeHEADVsIndex compares HEAD tree against index snapshot.
	DiffModeHEADVsIndex
	// DiffModeHeadVsWorkingTree compares HEAD tree against working tree.
	DiffModeHeadVsWorkingTree
	// DiffModeCommitVsWorkingTree compares a specific commit tree against working tree.
	DiffModeCommitVsWorkingTree
	// DiffModeCommitVsCommit compares two commit trees.
	DiffModeCommitVsCommit
)

// DiffOptions configures which snapshots should be compared.
type DiffOptions struct {
	Mode             DiffMode
	BaseCommitHash   domain.Hash
	TargetCommitHash domain.Hash
}

// ContentLoaderFunc loads textual content for a path/hash pair in a snapshot.
type ContentLoaderFunc func(path domain.NormalizedPath, hash domain.Hash) (string, error)

// Snapshot bundles path hashes with a matching content loader.
type Snapshot struct {
	PathHashes    core.PathHashes
	ContentLoader ContentLoaderFunc
}

// DiffStatus classifies a file-level diff result.
type DiffStatus int

const (
	// DiffStatusModified indicates a path exists in both snapshots with different content.
	DiffStatusModified DiffStatus = iota
	// DiffStatusAdded indicates a path exists only in the new snapshot.
	DiffStatusAdded
	// DiffStatusDeleted indicates a path exists only in the old snapshot.
	DiffStatusDeleted
)

// DiffResult represents the patch hunks and metadata for one changed path.
type DiffResult struct {
	Hunks   []*Hunk
	Status  DiffStatus
	OldPath domain.NormalizedPath
	NewPath domain.NormalizedPath
	OldHash domain.Hash
	NewHash domain.Hash
}

// DiffService resolves snapshots and computes unified-style line diffs.
type DiffService struct {
	objectService *core.ObjectService
	treeResolver  *core.TreeResolver
	diffAlgorithm *MyersDiffAlgorithm
	workspace     *domain.Workspace
}

// NewDiffService creates a diff service.
func NewDiffService(
	objectService *core.ObjectService,
	treeResolver *core.TreeResolver,
	diffAlgorithm *MyersDiffAlgorithm,
	workspace *domain.Workspace,
) *DiffService {
	return &DiffService{
		objectService: objectService,
		treeResolver:  treeResolver,
		diffAlgorithm: diffAlgorithm,
		workspace:     workspace,
	}
}

// Diff computes diff results for the requested mode.
func (d *DiffService) Diff(options DiffOptions) ([]*DiffResult, error) {
	oldSnapshot, newSnapshot, err := d.LoadSnapshots(options)
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}
	return d.computeDiffResults(oldSnapshot, newSnapshot)
}

// computeDiffResults compares old and new snapshots and builds per-path patch hunks.
func (d *DiffService) computeDiffResults(oldSnapshot *Snapshot, newSnapshot *Snapshot) ([]*DiffResult, error) {
	var results []*DiffResult
	for newPath, newHash := range newSnapshot.PathHashes {
		newContent, err := newSnapshot.ContentLoader(newPath, newHash)
		if err != nil {
			return nil, err
		}

		newLines := strings.Split(strings.TrimSuffix(newContent, "\n"), "\n")
		oldHash, ok := oldSnapshot.PathHashes[newPath]
		if !ok {
			lineDiffs := d.diffAlgorithm.ComputeLineDiffs(nil, newLines)
			regions := FindRegions(lineDiffs)
			hunks := BuildHunks(lineDiffs, regions)
			results = append(
				results, &DiffResult{
					hunks, DiffStatusAdded, newPath,
					newPath, domain.Hash{}, newHash,
				},
			)
		} else if newHash != oldHash {
			oldContent, err := oldSnapshot.ContentLoader(newPath, oldHash)
			if err != nil {
				return nil, err
			}

			oldLines := strings.Split(strings.TrimSuffix(oldContent, "\n"), "\n")
			lineDiffs := d.diffAlgorithm.ComputeLineDiffs(oldLines, newLines)
			regions := FindRegions(lineDiffs)
			hunks := BuildHunks(lineDiffs, regions)
			results = append(
				results, &DiffResult{
					hunks, DiffStatusModified, newPath,
					newPath, oldHash, newHash,
				},
			)
		}
	}

	for oldPath, oldHash := range oldSnapshot.PathHashes {
		if _, ok := newSnapshot.PathHashes[oldPath]; !ok {
			oldContent, err := oldSnapshot.ContentLoader(oldPath, oldHash)
			if err != nil {
				return nil, err
			}

			oldLines := strings.Split(strings.TrimSuffix(oldContent, "\n"), "\n")
			lineDiffs := d.diffAlgorithm.ComputeLineDiffs(oldLines, nil)
			regions := FindRegions(lineDiffs)
			hunks := BuildHunks(lineDiffs, regions)
			results = append(
				results, &DiffResult{
					hunks, DiffStatusDeleted, oldPath,
					domain.NormalizedPath{}, oldHash, domain.Hash{},
				},
			)
		}
	}
	return sortDiffResults(results), nil
}

// LoadSnapshots resolves old/new snapshots and matching content loaders for a diff mode.
func (d *DiffService) LoadSnapshots(options DiffOptions) (*Snapshot, *Snapshot, error) {
	switch options.Mode {
	case DiffModeHeadVsWorkingTree:
		headPathHashes, err := d.treeResolver.ResolveHEAD()
		if err != nil {
			return nil, nil, err
		}

		workingTreePathHashes, err := d.treeResolver.ResolveWorkingTree()
		if err != nil {
			return nil, nil, err
		}
		return &Snapshot{headPathHashes, d.LoadBlobContent},
			&Snapshot{workingTreePathHashes, d.LoadFileContent},
			nil
	case DiffModeHEADVsIndex:
		headPathHashes, err := d.treeResolver.ResolveHEAD()
		if err != nil {
			return nil, nil, err
		}

		indexPathHashes, err := d.treeResolver.ResolveIndex()
		if err != nil {
			return nil, nil, err
		}
		return &Snapshot{headPathHashes, d.LoadBlobContent},
			&Snapshot{indexPathHashes, d.LoadBlobContent},
			nil

	case DiffModeCommitVsCommit:
		baseCommitPathHashes := make(core.PathHashes)
		if !options.BaseCommitHash.IsZero() {
			var err error
			baseCommitPathHashes, err = d.treeResolver.ResolveCommit(options.BaseCommitHash)
			if err != nil {
				return nil, nil, err
			}
		}

		targetCommitPathHashes, err := d.treeResolver.ResolveCommit(options.TargetCommitHash)
		if err != nil {
			return nil, nil, err
		}
		return &Snapshot{baseCommitPathHashes, d.LoadBlobContent},
			&Snapshot{targetCommitPathHashes, d.LoadBlobContent},
			nil

	case DiffModeCommitVsWorkingTree:
		commitPathHashes, err := d.treeResolver.ResolveCommit(options.BaseCommitHash)
		if err != nil {
			return nil, nil, err
		}

		workingTreePathHashes, err := d.treeResolver.ResolveWorkingTree()
		if err != nil {
			return nil, nil, err
		}
		return &Snapshot{commitPathHashes, d.LoadBlobContent},
			&Snapshot{workingTreePathHashes, d.LoadFileContent},
			nil

	case DiffModeIndexVsWorkingTree:
		indexPathHashes, err := d.treeResolver.ResolveIndex()
		if err != nil {
			return nil, nil, err
		}

		workingTreePathHashes, err := d.treeResolver.ResolveWorkingTree()
		if err != nil {
			return nil, nil, err
		}
		return &Snapshot{indexPathHashes, d.LoadBlobContent},
			&Snapshot{workingTreePathHashes, d.LoadFileContent},
			nil
	default:
		return nil, nil, fmt.Errorf("%d': %w", options.Mode, ErrUnsupportedDiffMode)
	}
}

// LoadBlobContent loads blob text from object storage.
func (d *DiffService) LoadBlobContent(_ domain.NormalizedPath, hash domain.Hash) (string, error) {
	blob, err := d.objectService.ReadBlob(hash)
	if err != nil {
		return "", err
	}
	return string(blob.Content()), nil
}

// LoadFileContent loads file text from the working tree using repository-root resolution.
func (d *DiffService) LoadFileContent(path domain.NormalizedPath, _ domain.Hash) (string, error) {
	absolutePath, err := path.ToAbsolutePath(d.workspace.RepoRoot())
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absolutePath.String())
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func sortDiffResults(results []*DiffResult) []*DiffResult {
	slices.SortFunc(
		results, func(a, b *DiffResult) int {
			return strings.Compare(a.OldPath.String(), b.OldPath.String())
		},
	)
	return results
}
