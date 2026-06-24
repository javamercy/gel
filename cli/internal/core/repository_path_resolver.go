package core

import (
	"Gel/internal/domain"
	"fmt"
	"os"
	"path/filepath"
)

// ResolvedRepositoryPath contains both filesystem and repository representations of a path.
type ResolvedRepositoryPath struct {
	Absolute   domain.AbsolutePath
	Normalized domain.NormalizedPath
}

// RepositoryPathResolver resolves raw filesystem paths relative to an explicit base directory and repository root.
type RepositoryPathResolver struct {
	baseDir  domain.AbsolutePath
	repoRoot domain.AbsolutePath
}

// NewRepositoryPathResolver creates a resolver for paths interpreted relative
// to baseDir and contained within repoRoot.
func NewRepositoryPathResolver(baseDir, repoRoot domain.AbsolutePath) (*RepositoryPathResolver, error) {
	if _, err := baseDir.ToNormalizedPath(repoRoot); err != nil {
		return nil, fmt.Errorf(
			"base directory must be inside repository root: %w",
			err,
		)
	}
	return &RepositoryPathResolver{
		baseDir:  baseDir,
		repoRoot: repoRoot,
	}, nil
}

// Resolve converts input into absolute and repository-normalized forms.
func (r *RepositoryPathResolver) Resolve(input string) (ResolvedRepositoryPath, error) {
	absolute, err := r.resolveAbsolute(input)
	if err != nil {
		return ResolvedRepositoryPath{}, err
	}

	normalized, err := absolute.ToNormalizedPath(r.repoRoot)
	if err != nil {
		return ResolvedRepositoryPath{}, err
	}
	return ResolvedRepositoryPath{
		Absolute:   absolute,
		Normalized: normalized,
	}, nil
}

func (r *RepositoryPathResolver) resolveAbsolute(input string) (domain.AbsolutePath, error) {
	if input == "" {
		return domain.AbsolutePath{}, fmt.Errorf("input path is empty")
	}
	if filepath.IsAbs(input) {
		return domain.NewAbsolutePath(input)
	}
	if isRootedButNotAbsolute(input) {
		return domain.AbsolutePath{}, fmt.Errorf(
			"path %q is rooted but not fully absolute",
			input,
		)
	}

	joined := filepath.Join(r.baseDir.String(), input)
	return domain.NewAbsolutePath(joined)
}

func isRootedButNotAbsolute(path string) bool {
	if filepath.VolumeName(path) != "" {
		return true
	}
	return len(path) > 0 && os.IsPathSeparator(path[0])
}
