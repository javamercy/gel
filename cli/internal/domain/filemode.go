package domain

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"
)

var (
	// ErrInvalidFileMode indicates a value that cannot be represented as a
	// supported Gel file mode.
	ErrInvalidFileMode = errors.New("invalid file mode")
)

// FileMode identifies the mode stored for an entry in a tree object.
type FileMode uint32

const (
	// FileModeRegular identifies a non-executable regular file.
	FileModeRegular FileMode = 0o100644

	// FileModeExecutable identifies an executable regular file.
	FileModeExecutable FileMode = 0o100755

	// FileModeDirectory identifies a directory whose object is a tree.
	FileModeDirectory FileMode = 0o040000
)

const (
	treeModeRegular    = "100644"
	treeModeExecutable = "100755"
	treeModeDirectory  = "40000"
)

// NewFileMode validates mode and returns the corresponding FileMode.
func NewFileMode(mode uint32) (FileMode, error) {
	fileMode := FileMode(mode)
	if !fileMode.IsValid() {
		return 0, fmt.Errorf(
			"%w: %#o",
			ErrInvalidFileMode,
			mode,
		)
	}
	return fileMode, nil
}

// ParseFileMode parses a file mode stored in a tree entry.
func ParseFileMode(mode string) (FileMode, error) {
	switch mode {
	case treeModeRegular:
		return FileModeRegular, nil
	case treeModeExecutable:
		return FileModeExecutable, nil
	case treeModeDirectory:
		return FileModeDirectory, nil
	default:
		return 0, fmt.Errorf(
			"%w: tree mode %q",
			ErrInvalidFileMode,
			mode,
		)
	}
}

// FileModeFromFS converts a portable filesystem mode to a supported Gel FileMode.
//
// Regular files are executable when at least one executable permission bit is
// available and set. Filesystems that do not expose executable permissions
// therefore produce FileModeRegular.
func FileModeFromFS(mode fs.FileMode) (FileMode, error) {
	switch {
	case mode.IsDir():
		return FileModeDirectory, nil

	case mode.IsRegular():
		if mode.Perm()&0o111 != 0 {
			return FileModeExecutable, nil
		}
		return FileModeRegular, nil

	default:
		return 0, fmt.Errorf(
			"%w: unsupported filesystem mode %q",
			ErrInvalidFileMode,
			mode.String(),
		)
	}
}

// String returns the octal tree-entry representation of m.
func (m FileMode) String() string {
	return strconv.FormatUint(uint64(m), 8)
}

// IsValid reports whether m is a supported Gel file mode.
func (m FileMode) IsValid() bool {
	switch m {
	case FileModeRegular, FileModeExecutable, FileModeDirectory:
		return true
	default:
		return false
	}
}

// IsDirectory reports whether m identifies a directory.
func (m FileMode) IsDirectory() bool {
	return m == FileModeDirectory
}

// ObjectType returns the object type referenced by an entry with mode m.
func (m FileMode) ObjectType() (ObjectType, error) {
	switch m {
	case FileModeRegular, FileModeExecutable:
		return ObjectTypeBlob, nil
	case FileModeDirectory:
		return ObjectTypeTree, nil
	default:
		return "", fmt.Errorf(
			"%w: %#o",
			ErrInvalidFileMode,
			uint32(m),
		)
	}
}
