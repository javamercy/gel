package staging

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"fmt"
)

// AddOptions controls add command execution mode.
type AddOptions struct {
	// DryRun computes add/remove results without writing index changes.
	DryRun bool
	// Verbose includes added/removed path output after a real update.
	Verbose bool
}

// AddResult returns the outcome of an add invocation.
type AddResult struct {
	Added   []domain.NormalizedPath
	Removed []domain.NormalizedPath
	Error   error
}

// AddService stages working tree paths and reconciles scoped removals.
type AddService struct {
	indexService       *core.IndexService
	updateIndexService *UpdateIndexService
	pathResolver       *core.PathResolver
	workspace          *domain.Workspace
}

// NewAddService creates an add service with required dependencies.
func NewAddService(
	indexService *core.IndexService,
	updateIndexService *UpdateIndexService,
	pathResolver *core.PathResolver,
	workspace *domain.Workspace,
) *AddService {
	return &AddService{
		indexService:       indexService,
		updateIndexService: updateIndexService,
		pathResolver:       pathResolver,
		workspace:          workspace,
	}
}

// Add resolves pathspecs, computes staged additions/removals, and updates index.
// It uses pathspec scope to decide which previously tracked entries should be
// removed when they are no longer present in the resolved set.
func (a *AddService) Add(pathspecs []string, options AddOptions) AddResult {
	index, err := a.indexService.Read()
	if err != nil {
		return AddResult{Error: fmt.Errorf("add: %w", err)}
	}

	resolvedPaths, err := a.pathResolver.Resolve(pathspecs)
	if err != nil {
		return AddResult{Error: fmt.Errorf("add: %w", err)}
	}

	pathsToAdd, pathsToRemove, err := a.collectPaths(index, resolvedPaths)
	if err != nil {
		return AddResult{Error: fmt.Errorf("add: %w", err)}
	}

	if options.DryRun {
		return AddResult{Added: pathsToAdd, Removed: pathsToRemove}
	}

	addedFiles, err := a.updateIndexService.UpdateIndex(
		pathsToAdd, UpdateIndexOptions{
			Add:    true,
			Remove: false,
			Write:  !options.DryRun,
		},
	)
	if err != nil {
		return AddResult{Error: fmt.Errorf("add: %w", err)}
	}

	removedFiles, err := a.updateIndexService.UpdateIndex(
		pathsToRemove, UpdateIndexOptions{
			Add:    false,
			Remove: true,
			Write:  !options.DryRun,
		},
	)
	if err != nil {
		return AddResult{Error: fmt.Errorf("add: %w", err)}
	}
	if options.Verbose {
		return AddResult{Added: addedFiles, Removed: removedFiles}
	}
	return AddResult{}
}

// collectPaths computes deterministic add/remove path lists for update-index.
// Paths are deduplicated and sorted to avoid map-iteration order instability.
func (a *AddService) collectPaths(
	index *domain.Index,
	resolvedPaths []core.ResolvedPath,
) (
	pathsToAdd []domain.NormalizedPath, pathsToRemove []domain.NormalizedPath, err error,
) {
	pathsToAddSet := make(map[domain.NormalizedPath]struct{})
	pathsToRemoveSet := make(map[domain.NormalizedPath]struct{})

	for _, resolved := range resolvedPaths {
		for path := range resolved.NormalizedPaths {
			pathsToAddSet[path] = struct{}{}
		}

		var indexEntries []*domain.IndexEntry
		switch resolved.Type {
		case core.PathspecTypeFile, core.PathspecTypeNonExistent:
			// TODO: fix this to find the entry in the index and add it to indexEntries
			_, err := domain.ParseNormalizedPath(resolved.NormalizedScope)
			if err != nil {
				return nil, nil, fmt.Errorf("add: %w", err)
			}
			/* if entry, _ := index.FindEntry(normalizedScope); entry != nil {
				indexEntries = nil
			} else {
				prefix := resolved.NormalizedScope
				if prefix != "" {
					prefix += "/"
				}
				//indexEntries = index.FindEntriesByPathPrefix(prefix)
			}*/
		case core.PathspecTypeDirectory:
			prefix := resolved.NormalizedScope
			if prefix != "" {
				prefix += "/"
			}
			//indexEntries = index.FindEntriesByPathPrefix(prefix)
		case core.PathspecTypeGlobPattern:
			//indexEntries = index.FindEntriesByPathPattern(resolved.NormalizedScope)
		}

		for _, entry := range indexEntries {
			if !resolved.NormalizedPaths[entry.Path()] {
				pathsToRemoveSet[entry.Path()] = struct{}{}
			}
		}
		if len(resolved.NormalizedPaths) == 0 && len(indexEntries) == 0 {
			return nil, nil, fmt.Errorf("'%s': %w", resolved.NormalizedScope, ErrPathDidNotMatch)
		}

	}
	for path := range pathsToAddSet {
		delete(pathsToRemoveSet, path)
	}
	// return domain.SortedPathSet(pathsToAddSet), domain.SortedPathSet(pathsToRemoveSet), nil
	return nil, nil, nil
}
