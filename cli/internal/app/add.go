package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type AddInput struct {
	Pathspecs []string
	DryRun    bool
}

type AddResult struct {
	Staged  []domain.NormalizedPath
	Removed []domain.NormalizedPath
}

type Add struct {
	indexStore       *storage.IndexStore
	objectStore      *storage.ObjectStore
	pathspecResolver *PathspecResolver
}

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
		if !errors.Is(err, fs.ErrNotExist) {
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
