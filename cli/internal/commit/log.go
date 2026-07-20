package commit

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type LogEntry struct {
	Hash    domain.Hash
	Message string
	Date    string
}

// LogOptions controls commit history query and formatting data selection.
type LogOptions struct {
	Limit   int
	Oneline bool
	Since   string
	Until   string
}

// LogService resolves a starting revision and walks commit history.
type LogService struct {
	refService    *core.RefService
	objectService *core.ObjectService
}

// NewLogService creates a log service.
func NewLogService(
	refService *core.RefService,
	objectService *core.ObjectService,
) *LogService {
	return &LogService{
		refService:    refService,
		objectService: objectService,
	}
}

// Log returns commit history entries starting from name.
// name may be HEAD, a branch name, a full refs/* path, or a commit hash.
// The current traversal follows first-parent history only.
func (l *LogService) Log(name string, options LogOptions) ([]*LogEntry, error) {
	// TODO: handle --since and --until

	hash, err := l.resolveStartHash(name)
	if err != nil {
		return nil, fmt.Errorf("log: %w", err)
	}

	var entries []*LogEntry
	count := 0
	for {
		if options.Limit > 0 && count >= options.Limit {
			break
		}
		count++

		commit, err := l.objectService.ReadCommit(hash)
		if err != nil {
			return nil, fmt.Errorf("log: %w", err)
		}
		if options.Oneline {
			firstLine := strings.Split(commit.Message(), "\n")[0]
			entries = append(
				entries, &LogEntry{
					Hash:    hash,
					Message: firstLine,
				},
			)
		} else {
			entries = append(
				entries, &LogEntry{
					Hash:    hash,
					Message: commit.Message(),
					Date:    "",
				},
			)

		}
		if len(commit.ParentHashes()) == 0 {
			break
		}
		// TODO: use a priority queue/heap to interleave commits from all parents by date
		hash = commit.ParentHashes()[0]
	}
	return entries, nil
}

// resolveStartHash resolves the starting revision for log traversal.
func (l *LogService) resolveStartHash(name string) (domain.Hash, error) {
	if name == domain.HeadFileName {
		hash, err := l.refService.Resolve(name)
		if errors.Is(err, core.ErrRefNotFound) {
			return domain.Hash{}, ErrNoCommitsYet
		}
		return hash, err
	}

	if strings.HasPrefix(name, domain.RefsDirName+"/") {
		return l.refService.Read(name)
	}

	branchRef := filepath.Join(domain.RefsDirName, domain.HeadsDirName, name)
	if hash, err := l.refService.Read(branchRef); err == nil {
		return hash, nil
	} else if !errors.Is(err, core.ErrRefNotFound) {
		return domain.Hash{}, err
	}

	hash, err := domain.NewHashFromHex(name)
	if err != nil {
		return domain.Hash{}, fmt.Errorf("'%s': %w", name, core.ErrRefNotFound)
	}
	if _, err := l.objectService.ReadCommit(hash); err != nil {
		return domain.Hash{}, err
	}
	return hash, nil
}
