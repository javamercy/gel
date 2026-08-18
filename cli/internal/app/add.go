package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"os"
)

// AddInput specifies the pathspecs to stage and whether to avoid persistence.
type AddInput struct {
	// Pathspecs identifies files or directories relative to the resolver's
	// working directory.
	Pathspecs []string
	// DryRun reports whether Add should calculate changes without writing blobs
	// or saving the index.
	DryRun bool
}

// AddResult reports the paths identified for staging or removal.
type AddResult struct {
	// Staged contains changed existing paths selected for staging.
	Staged []domain.NormalizedPath
	// Removed is reserved for tracked paths selected for removal. A missing path
	// currently causes Run to return an error and a zero result.
	Removed []domain.NormalizedPath
}

// Add stages working-tree files in the repository index.
type Add struct {
	indexStore       *storage.IndexStore
	objectStore      *storage.ObjectStore
	pathspecResolver *PathspecResolver
}

// NewAdd returns an Add backed by the supplied stores and pathspec resolver.
//
// The dependencies must not be nil when Run is called.
func NewAdd(
	indexStore *storage.IndexStore,
	objectStore *storage.ObjectStore,
	pathspecResolver *PathspecResolver,
) *Add {
	return &Add{
		indexStore:       indexStore,
		objectStore:      objectStore,
		pathspecResolver: pathspecResolver,
	}
}

// Run resolves input pathspecs and stages changed working-tree files.
//
// In normal mode, Run writes changed files as blob objects and saves the
// updated index. In dry-run mode, it reports selected changes without changing
// object storage or the index. An existing index is loaded when available; a
// missing index is treated as an empty index. A missing path is an error, and
// any error returns a zero AddResult.
func (a *Add) Run(input AddInput) (AddResult, error) {
	var result AddResult

	resolvedPaths, err := a.pathspecResolver.Resolve(input.Pathspecs)
	if err != nil {
		return AddResult{}, fmt.Errorf(
			"resolve pathspecs: %w",
			err,
		)
	}

	index, err := a.indexStore.Load()
	if err != nil {
		if !errors.Is(err, storage.ErrIndexNotExist) {
			return AddResult{}, fmt.Errorf(
				"load index: %w",
				err,
			)
		}
		index, err = domain.NewIndexFromEntries([]domain.IndexEntry{})
		if err != nil {
			return AddResult{}, fmt.Errorf(
				"create new index: %w",
				err,
			)
		}
	}

	var changed bool
	for _, resolvedPath := range resolvedPaths {
		entry, inIndex := index.FindEntry(
			resolvedPath.Normalized,
			domain.IndexStageNormal,
		)

		if resolvedPath.Missing {
			if inIndex {
				if !input.DryRun {
					removed := index.RemoveEntry(
						resolvedPath.Normalized,
						domain.IndexStageNormal,
					)
					if removed {
						changed = true
					}
				}

				result.Removed = append(
					result.Removed,
					resolvedPath.Normalized,
				)
			}
			return AddResult{}, fmt.Errorf(
				"path %q does not exist",
				resolvedPath.Absolute.String(),
			)
		}

		stat, err := domain.ReadFileStat(resolvedPath.Absolute)
		if err != nil {
			return AddResult{}, fmt.Errorf(
				"read file stat: %w",
				err,
			)
		}

		if inIndex && entry.MatchesStat(stat) {
			continue
		}

		if !input.DryRun {
			data, err := os.ReadFile(resolvedPath.Absolute.String())
			if err != nil {
				return AddResult{}, fmt.Errorf(
					"read file: %w",
					err,
				)
			}

			blob := domain.NewBlob(data)
			hash, err := a.objectStore.Write(blob)
			if err != nil {
				return AddResult{}, fmt.Errorf(
					"write blob: %w",
					err,
				)
			}

			entryToStage, err := domain.NewIndexEntryFromFileStat(
				resolvedPath.Normalized,
				hash,
				stat,
			)
			if err != nil {
				return AddResult{}, fmt.Errorf(
					"create index entry: %w",
					err,
				)
			}

			if err := index.SetEntry(entryToStage); err != nil {
				return AddResult{}, fmt.Errorf(
					"set index entry: %w",
					err,
				)
			}

			changed = true
		}

		result.Staged = append(
			result.Staged,
			resolvedPath.Normalized,
		)
	}

	if !input.DryRun && changed {
		if err := a.indexStore.Save(index); err != nil {
			return AddResult{}, fmt.Errorf(
				"save index: %w",
				err,
			)
		}
	}
	return result, nil
}
