package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// IndexStage identifies an entry's role in a resolved or unresolved merge.
type IndexStage uint8

const (
	// IndexStageNormal identifies a resolved stage-zero entry.
	IndexStageNormal IndexStage = 0

	// IndexStageBase identifies the common-ancestor version.
	IndexStageBase IndexStage = 1

	// IndexStageOurs identifies the current-branch version.
	IndexStageOurs IndexStage = 2

	// IndexStageTheirs identifies the merged-branch version.
	IndexStageTheirs IndexStage = 3
)

func (s IndexStage) isValid() bool {
	switch s {
	case IndexStageNormal,
		IndexStageBase,
		IndexStageOurs,
		IndexStageTheirs:
		return true
	default:
		return false
	}
}

// IndexEntry represents one staged object and its filesystem stat cache.
type IndexEntry struct {
	path  NormalizedPath
	stage IndexStage
	hash  Hash
	mode  FileMode

	size       uint64
	deviceID   uint64
	inode      uint64
	userID     uint32
	groupID    uint32
	changeTime time.Time
	modTime    time.Time
}

// NewIndexEntryFromFileStat validates and constructs an IndexEntry.
func NewIndexEntryFromFileStat(
	path NormalizedPath,
	hash Hash,
	stat FileStat,
) (IndexEntry, error) {
	entry := IndexEntry{
		path:       path,
		stage:      IndexStageNormal,
		hash:       hash,
		mode:       stat.Mode,
		size:       stat.Size,
		deviceID:   stat.DeviceID,
		inode:      stat.Inode,
		userID:     stat.UserID,
		groupID:    stat.GroupID,
		changeTime: normalizeIndexTime(stat.ChangeTime),
		modTime:    normalizeIndexTime(stat.ModTime),
	}
	if err := validateIndexEntry(entry); err != nil {
		return IndexEntry{}, fmt.Errorf(
			"create index entry: %w",
			err,
		)
	}
	return entry, nil
}

// Path returns the entry's repository-relative path.
func (e *IndexEntry) Path() NormalizedPath {
	return e.path
}

// Stage returns the entry's merge stage.
func (e *IndexEntry) Stage() IndexStage {
	return e.stage
}

// Hash returns the hash of the staged blob.
func (e *IndexEntry) Hash() Hash {
	return e.hash
}

// Mode returns the staged file mode.
func (e *IndexEntry) Mode() FileMode {
	return e.mode
}

// Size returns the cached working-tree file size.
func (e *IndexEntry) Size() uint64 {
	return e.size
}

// MatchesStat reports whether the entry's stat cache matches stat.
//
// It always returns false for non-zero merge stages. Platform-specific fields
// are compared only when both values are available.
func (e *IndexEntry) MatchesStat(stat FileStat) bool {
	if e.stage != IndexStageNormal {
		return false
	}
	if e.mode != stat.Mode || e.size != stat.Size {
		return false
	}
	if !e.modTime.Equal(stat.ModTime) {
		return false
	}
	if !e.changeTime.IsZero() &&
		!stat.ChangeTime.IsZero() &&
		!e.changeTime.Equal(stat.ChangeTime) {
		return false
	}
	if e.deviceID != 0 &&
		stat.DeviceID != 0 &&
		e.deviceID != stat.DeviceID {
		return false
	}
	if e.inode != 0 &&
		stat.Inode != 0 &&
		e.inode != stat.Inode {
		return false
	}
	if e.userID != 0 &&
		stat.UserID != 0 &&
		e.userID != stat.UserID {
		return false
	}
	if e.groupID != 0 &&
		stat.GroupID != 0 &&
		e.groupID != stat.GroupID {
		return false
	}
	return true
}

func isValidIndexMode(mode FileMode) bool {
	return mode == FileModeRegular || mode == FileModeExecutable
}

func validateIndexEntries(entries []IndexEntry) error {
	for i, entry := range entries {
		if err := validateIndexEntry(entry); err != nil {
			return fmt.Errorf(
				"entry %d: %w",
				i,
				err,
			)
		}
	}
	for i := 1; i < len(entries); i++ {
		if err := validateConsecutiveIndexEntries(entries[i-1], entries[i]); err != nil {
			return fmt.Errorf(
				"entries %d and %d: %w",
				i-1,
				i,
				err,
			)
		}
	}
	return nil
}

func validateIndexEntry(entry IndexEntry) error {
	if !entry.stage.isValid() {
		return fmt.Errorf(
			"index stage %d is outside the supported range 0-3",
			entry.stage,
		)
	}
	if !isValidIndexMode(entry.mode) {
		return fmt.Errorf(
			"mode %#o cannot be stored in an index entry",
			uint32(entry.mode),
		)
	}
	if entry.hash.IsZero() {
		return fmt.Errorf(
			"zero hash cannot identify an index object",
		)
	}
	if err := validateIndexEntryPath(entry.path); err != nil {
		return err
	}
	return nil
}

func validateIndexEntryPath(path NormalizedPath) error {
	if path.IsRoot() {
		return fmt.Errorf(
			"index entry path cannot represent repository root",
		)
	}

	pathText := path.String()
	if !utf8.ValidString(pathText) {
		return fmt.Errorf(
			"index entry path %q is not valid UTF-8",
			pathText,
		)
	}
	if strings.IndexByte(pathText, 0) >= 0 {
		return fmt.Errorf(
			"index entry path %q contains NUL",
			pathText,
		)
	}
	for _, component := range strings.Split(pathText, "/") {
		if component == ".gel" {
			return fmt.Errorf(
				"index entry path %q contains reserved component %q",
				pathText,
				".gel",
			)
		}
	}
	return nil
}

func normalizeIndexTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return t.UTC()
}
