package core

import (
	"Gel/internal/domain"
	"errors"
	"os"
)

// FileState describes how a working tree path compares to its index entry.
type FileState int

const (
	// FileStateUnchanged means metadata matches index entry stat fields.
	FileStateUnchanged FileState = iota
	// FileStateModified means file exists but differs from index metadata.
	FileStateModified
	// FileStateDeleted means the path no longer exists in the working tree.
	FileStateDeleted
)

// ChangeResult is the normalized output of ChangeDetector.
type ChangeResult struct {
	// FileState indicates whether file is unchanged, modified, or deleted.
	FileState FileState
	// NewHash is populated for modified files and zero-value otherwise.
	NewHash domain.Hash
}

// ChangeDetector compares index entries to working tree files.
type ChangeDetector struct {
	objectService *ObjectService
	repoDir       domain.AbsolutePath
}

// NewChangeDetector creates a detector rooted at repoDir.
func NewChangeDetector(objectService *ObjectService, repoDir domain.AbsolutePath) *ChangeDetector {
	return &ChangeDetector{
		objectService: objectService,
		repoDir:       repoDir,
	}
}

// DetectFileChange compares a tracked index entry to the current working tree.
//
// It resolves the entry path, classifies missing files as FileStateDeleted,
// returns FileStateUnchanged when stat metadata matches, and otherwise computes
// a fresh blob hash and returns FileStateModified.
func (c *ChangeDetector) DetectFileChange(entry *domain.IndexEntry) (ChangeResult, error) {
	absPath, err := entry.Path.ToAbsolutePath(c.repoDir)
	if err != nil {
		return ChangeResult{}, err
	}

	stat, err := domain.ReadFileStat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ChangeResult{FileState: FileStateDeleted}, nil
		}
		return ChangeResult{}, err
	}
	if entry.MatchesStat(stat) {
		return ChangeResult{FileState: FileStateUnchanged}, nil
	}

	hash, _, err := c.objectService.ComputeObjectHash(absPath)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{FileState: FileStateModified, NewHash: hash}, nil
}
