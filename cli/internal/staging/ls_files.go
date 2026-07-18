package staging

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	globPatterns string = "*?[]"
)

// LsFilesOptions controls ls-files output mode.
type LsFilesOptions struct {
	// Cached lists tracked paths from the index.
	Cached bool
	// Stage prints staged entry metadata (mode/hash/stage/path).
	Stage bool
	// Modified lists tracked paths modified in the working tree.
	Modified bool
	// Deleted lists tracked paths missing from the working tree.
	Deleted bool
}

// LsFilesService implements ls-files queries over index and working tree state.
type LsFilesService struct {
	indexService   *core.IndexService
	changeDetector *core.ChangeDetector
	workspace      *domain.Workspace
}

// NewLsFilesService creates an ls-files service with required dependencies.
func NewLsFilesService(
	indexService *core.IndexService,
	changeDetector *core.ChangeDetector,
	workspace *domain.Workspace,
) *LsFilesService {
	return &LsFilesService{
		indexService:   indexService,
		changeDetector: changeDetector,
		workspace:      workspace,
	}
}

// LsFiles returns tracked files based on mode flags and optional pathspec.
//
// Pathspec supports prefix matching by default and glob matching when wildcard
// characters are present.
func (l *LsFilesService) LsFiles(pathspec string, options LsFilesOptions) ([]string, error) {
	if !options.Stage && !options.Cached && !options.Modified && !options.Deleted {
		return nil, fmt.Errorf("ls-files: must specify --stage, --cached, --modified, or --deleted")
	}

	index, err := l.indexService.Read()
	if err != nil {
		return nil, fmt.Errorf("ls-files: %w", err)
	}

	var entries []domain.IndexEntry
	if pathspec != "" {
		if strings.ContainsAny(pathspec, globPatterns) {
			//entries = index.FindEntriesByPathPattern(pathspec)
		} else {
			//entries = index.FindEntriesByPathPrefix(pathspec)
		}
	} else {
		entries = index.Entries()
	}

	switch {
	case options.Stage:
		return l.lsFilesWithStage(entries), nil
	case options.Cached:
		return l.lsFilesWithCached(entries), nil
	case options.Modified:
		return l.lsFilesWithModified(entries)
	case options.Deleted:
		return l.lsFilesWithDeleted(entries)
	}
	return nil, nil
}

// lsFilesWithStage formats entries as mode/hash/stage/path.
func (l *LsFilesService) lsFilesWithStage(entries []domain.IndexEntry) []string {
	files := make([]string, len(entries))
	for i, entry := range entries {
		files[i] = fmt.Sprintf(
			"%s %s %d\t%s",
			entry.Mode(),
			entry.Hash(),
			entry.Stage(),
			entry.Path(),
		)
	}
	return files
}

// lsFilesWithCached returns entry paths exactly as stored in index.
func (l *LsFilesService) lsFilesWithCached(entries []domain.IndexEntry) []string {
	files := make([]string, len(entries))
	for i, entry := range entries {
		files[i] = entry.Path().String()
	}
	return files
}

// lsFilesWithModified returns tracked paths classified as modified.
func (l *LsFilesService) lsFilesWithModified(entries []domain.IndexEntry) ([]string, error) {
	files := make([]string, 0)
	for _, entry := range entries {
		changeResult, err := l.changeDetector.DetectFileChange(entry)
		if err != nil {
			return nil, err
		}
		if changeResult.FileState == core.FileStateModified {
			files = append(files, entry.Path().String())
		}
	}
	return files, nil
}

// lsFilesWithDeleted returns tracked paths that no longer exist on disk.
func (l *LsFilesService) lsFilesWithDeleted(entries []domain.IndexEntry) ([]string, error) {
	files := make([]string, 0)
	for _, entry := range entries {
		absolutePath, err := entry.Path().ToAbsolutePath(l.workspace.RepoRoot())
		if err != nil {
			return nil, fmt.Errorf("ls-files: %w", err)
		}
		_, err = os.Stat(absolutePath.String())
		switch {
		case errors.Is(err, os.ErrNotExist):
			files = append(files, entry.Path().String())
		case err != nil:
			return nil, fmt.Errorf("ls-files: failed to stat file '%s': %w", entry.Path, err)
		}
	}
	return files, nil
}
