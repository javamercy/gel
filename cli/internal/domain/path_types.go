package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrInvalidAbsolutePath indicates a value that cannot represent an
	// absolute filesystem path.
	ErrInvalidAbsolutePath = errors.New("invalid absolute path")

	// ErrInvalidNormalizedPath indicates a malformed repository-relative path.
	ErrInvalidNormalizedPath = errors.New("invalid normalized path")

	// ErrPathOutsideRepository indicates an absolute path that cannot be
	// represented relative to a repository root.
	ErrPathOutsideRepository = errors.New("path outside repository")
)

// NormalizedPath represents a slash-separated path relative to the repository
// root. The zero value represents the repository root.
type NormalizedPath struct {
	value string
}

// ParseNormalizedPath parses a repository-relative normalized path.
//
// An empty string represents the repository root.
func ParseNormalizedPath(path string) (NormalizedPath, error) {
	if err := validateNormalizedPath(path); err != nil {
		return NormalizedPath{}, err
	}
	return NormalizedPath{value: path}, nil
}

func validateNormalizedPath(path string) error {
	if path == "" {
		return nil
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf(
			"%w: %q is absolute",
			ErrInvalidNormalizedPath,
			path,
		)
	}
	if strings.ContainsRune(path, '\\') {
		return fmt.Errorf(
			"%w: %q contains backslash",
			ErrInvalidNormalizedPath,
			path,
		)
	}
	if strings.IndexByte(path, 0) >= 0 {
		return fmt.Errorf(
			"%w: path contains NUL byte",
			ErrInvalidNormalizedPath,
		)
	}
	for _, segment := range strings.Split(path, "/") {
		switch segment {
		case "":
			return fmt.Errorf(
				"%w: %q contains an empty segment",
				ErrInvalidNormalizedPath,
				path,
			)

		case ".", "..":
			return fmt.Errorf(
				"%w: %q contains traversal segment %q",
				ErrInvalidNormalizedPath,
				path,
				segment,
			)
		}
	}
	return nil
}

// ToAbsolutePath converts p to a native filesystem path rooted at repoRoot.
func (p NormalizedPath) ToAbsolutePath(repoRoot AbsolutePath) (AbsolutePath, error) {
	if repoRoot.value == "" {
		return AbsolutePath{}, fmt.Errorf(
			"%w: repository root is the zero value",
			ErrInvalidAbsolutePath,
		)
	}
	if p.IsRoot() {
		return repoRoot, nil
	}

	localPath, err := filepath.Localize(p.value)
	if err != nil {
		return AbsolutePath{}, fmt.Errorf(
			"normalized path %q is not representable on this platform: %w",
			p.value,
			err,
		)
	}
	return NewAbsolutePath(filepath.Join(repoRoot.value, localPath))
}

// IsRoot reports whether p represents the repository root.
func (p NormalizedPath) IsRoot() bool {
	return p.value == ""
}

// String returns the slash-separated repository representation of p.
func (p NormalizedPath) String() string {
	return p.value
}

// AbsolutePath represents a lexically clean absolute filesystem path.
//
// AbsolutePath uses the current operating system's path syntax. It does not
// guarantee that the path exists or resolve symbolic links.
type AbsolutePath struct {
	value string
}

// NewAbsolutePath validates path and returns its lexically clean absolute
// representation.
func NewAbsolutePath(path string) (AbsolutePath, error) {
	if path == "" {
		return AbsolutePath{}, fmt.Errorf(
			"%w: path is empty",
			ErrInvalidAbsolutePath,
		)
	}
	if strings.IndexByte(path, 0) >= 0 {
		return AbsolutePath{}, fmt.Errorf(
			"%w: path contains NUL byte",
			ErrInvalidAbsolutePath,
		)
	}

	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		return AbsolutePath{}, fmt.Errorf(
			"%w: %q is not absolute",
			ErrInvalidAbsolutePath,
			path,
		)
	}
	return AbsolutePath{value: cleaned}, nil
}

// String returns the native filesystem representation of p.
func (p AbsolutePath) String() string {
	return p.value
}

// ToNormalizedPath converts p to a repository-relative normalized path.
//
// It returns an error matching ErrPathOutsideRepository when p cannot be
// represented relative to repoRoot.
func (p AbsolutePath) ToNormalizedPath(repoRoot AbsolutePath) (NormalizedPath, error) {
	if p.value == "" {
		return NormalizedPath{}, fmt.Errorf(
			"%w: source path is the zero value",
			ErrInvalidAbsolutePath,
		)
	}
	if repoRoot.value == "" {
		return NormalizedPath{}, fmt.Errorf(
			"%w: repository root is the zero value",
			ErrInvalidAbsolutePath,
		)
	}

	relative, err := filepath.Rel(repoRoot.value, p.value)
	if err != nil {
		return NormalizedPath{}, fmt.Errorf(
			"%w: cannot make %q relative to repository root %q: %v",
			ErrPathOutsideRepository,
			p.value,
			repoRoot.value,
			err,
		)
	}
	if relative == "." {
		return NormalizedPath{}, nil
	}
	if startsWithParentTraversal(relative) {
		return NormalizedPath{}, fmt.Errorf(
			"%w: %q is outside repository root %q",
			ErrPathOutsideRepository,
			p.value,
			repoRoot.value,
		)
	}
	return ParseNormalizedPath(filepath.ToSlash(relative))
}

func startsWithParentTraversal(path string) bool {
	return path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator))
}
