package core

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrRepositoryNotFound indicates that repository discovery reached the
	// filesystem root without finding a .gel directory.
	ErrRepositoryNotFound = errors.New("repository not found")
)

// DiscoverWorkspace searches startDir and its ancestors for a Gel repository.
//
// startDir may be relative or absolute, but it must identify an existing
// directory. Discovery returns an error matching ErrRepositoryNotFound when
// no .gel directory exists in the directory hierarchy.
func DiscoverWorkspace(startDir string) (*domain.Workspace, error) {
	if startDir == "" {
		return nil, errors.New("workspace start directory is empty")
	}
	if isRootedButNotAbsolute(startDir) {
		return nil, fmt.Errorf(
			"workspace start path %q is rooted but not fully absolute",
			startDir,
		)
	}

	absolute, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve workspace start directory %q: %w",
			startDir,
			err,
		)
	}

	startPath, err := domain.NewAbsolutePath(absolute)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve workspace start directory %q: %w",
			startDir,
			err,
		)
	}

	info, err := os.Stat(startPath.String())
	if err != nil {
		return nil, fmt.Errorf(
			"inspect workspace start directory %q: %w",
			startPath,
			err,
		)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf(
			"workspace start path %q is not a directory",
			startPath,
		)
	}

	repoRootPath, err := findRepositoryRoot(startPath.String())
	if err != nil {
		return nil, err
	}

	repoRoot, err := domain.NewAbsolutePath(repoRootPath)
	if err != nil {
		return nil, fmt.Errorf(
			"construct repository root %q: %w",
			repoRootPath,
			err,
		)
	}
	return domain.NewWorkspace(repoRoot)
}

func findRepositoryRoot(startDir string) (string, error) {
	currentDir := startDir

	for {
		gelDir := filepath.Join(currentDir, domain.GelDirName)
		info, err := os.Stat(gelDir)
		switch {
		case err == nil:
			if !info.IsDir() {
				return "", fmt.Errorf(
					"repository metadata path %q is not a directory",
					gelDir,
				)
			}
			return currentDir, nil

		case !errors.Is(err, os.ErrNotExist):
			return "", fmt.Errorf(
				"inspect repository metadata path %q: %w",
				gelDir,
				err,
			)
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", ErrRepositoryNotFound
		}
		currentDir = parentDir
	}
}
