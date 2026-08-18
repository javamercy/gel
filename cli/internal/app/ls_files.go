package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// LsFilesInput selects which index entries to return.
type LsFilesInput struct {
	// Pathspecs limits results to matching paths. An empty slice selects all
	// entries.
	Pathspecs []string
	// Unmerged reports whether to return only non-normal-stage entries.
	Unmerged bool
}

// LsFilesResult contains the index entries reported by LsFiles.Run.
type LsFilesResult struct {
	// Entries are the index entries selected by the input pathspecs, in index
	// order.
	Entries []domain.IndexEntry
}

// LsFiles reports entries from the repository index.
type LsFiles struct {
	indexStore *storage.IndexStore
	repoRoot   domain.AbsolutePath
	workingDir domain.AbsolutePath
}

// NewLsFiles returns an LsFiles operating within repoRoot and workingDir.
//
// It returns an error when repoRoot or workingDir is zero or when workingDir
// resolves outside repoRoot. indexStore must not be nil when Run is called.
func NewLsFiles(
	indexStore *storage.IndexStore,
	repoRoot domain.AbsolutePath,
	workingDir domain.AbsolutePath,
) (*LsFiles, error) {
	if repoRoot.IsZero() {
		return nil, fmt.Errorf("repository root is zero")
	}
	if workingDir.IsZero() {
		return nil, fmt.Errorf("working directory is zero")
	}
	if _, err := workingDir.ToNormalizedPath(repoRoot); err != nil {
		return nil, fmt.Errorf(
			"working directory is outside repository: %w",
			err,
		)
	}
	return &LsFiles{
		indexStore: indexStore,
		repoRoot:   repoRoot,
		workingDir: workingDir,
	}, nil
}

// Run returns index entries selected by input.
//
// An empty input.Pathspecs matches every entry. When input.Unmerged is true,
// only entries with a non-normal stage are returned. When input.Unmerged is
// false, only normal-stage entries are returned. A pathspec matches an entry
// when it exactly matches the entry path, when it is the repository root, or
// when the entry path is a descendant of the pathspec. An absent index produces
// an empty result with no error.
func (l *LsFiles) Run(input LsFilesInput) (LsFilesResult, error) {
	var result LsFilesResult

	paths, err := l.normalizePathspecs(input.Pathspecs)
	if err != nil {
		return result, fmt.Errorf(
			"normalize paths: %w",
			err,
		)
	}

	index, err := l.indexStore.Load()
	if err != nil {
		if errors.Is(err, storage.ErrIndexNotExist) {
			return result, nil
		}
		return result, fmt.Errorf(
			"load index: %w",
			err,
		)
	}

	for _, entry := range index.Entries() {
		if input.Unmerged && entry.Stage() == domain.IndexStageNormal {
			continue
		}

		matched := len(paths) == 0
		for _, selector := range paths {
			selectorText := selector.String()
			entryText := entry.Path().String()
			if selector.IsRoot() ||
				entryText == selectorText ||
				strings.HasPrefix(entryText, selectorText+"/") {
				matched = true
				break
			}
		}

		if matched {
			result.Entries = append(result.Entries, entry)
		}
	}
	return result, nil
}

func (l *LsFiles) normalizePathspecs(pathspecs []string) ([]domain.NormalizedPath, error) {
	normalizedPaths := make([]domain.NormalizedPath, 0, len(pathspecs))
	for _, pathspec := range pathspecs {
		normalized, err := l.normalizePathspec(pathspec)
		if err != nil {
			return nil, err
		}
		normalizedPaths = append(normalizedPaths, normalized)
	}
	return normalizedPaths, nil
}

func (l *LsFiles) normalizePathspec(pathspec string) (domain.NormalizedPath, error) {
	if pathspec == "" {
		return domain.NormalizedPath{}, fmt.Errorf(
			"path is empty",
		)
	}

	path := pathspec
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.workingDir.String(), path)
	}

	absolute, err := domain.NewAbsolutePath(path)
	if err != nil {
		return domain.NormalizedPath{}, fmt.Errorf(
			"create absolute path: %w",
			err,
		)
	}

	normalized, err := absolute.ToNormalizedPath(l.repoRoot)
	if err != nil {
		return domain.NormalizedPath{}, fmt.Errorf(
			"path %q is outside repository: %w",
			pathspec,
			err,
		)
	}
	return normalized, nil
}
