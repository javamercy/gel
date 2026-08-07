package app

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const (
	defaultDirPermission  os.FileMode = 0o755
	defaultFilePermission fs.FileMode = 0o644

	defaultBranchRef string = "refs/heads/main"
)

// InitResult describes the repository metadata prepared by Init.
type InitResult struct {
	// GelDir is the absolute path to the repository metadata directory.
	GelDir domain.AbsolutePath
	// Reinitialized reports whether the metadata directory existed before Init.
	Reinitialized bool
}

// Init creates the Gel repository layout for workspace.
//
// It creates the repository root, metadata, object, and branch-head
// directories when needed. It also creates HEAD pointing to main and an empty
// configuration file without replacing existing regular files. workspace must
// not be nil.
func Init(workspace *domain.Workspace) (InitResult, error) {
	if workspace == nil {
		return InitResult{}, errors.New("workspace is nil")
	}
	if err := ensureDirectory(workspace.RepoRoot()); err != nil {
		return InitResult{}, fmt.Errorf(
			"init repository root: %w",
			err,
		)
	}

	reinitialized, err := directoryExists(workspace.GelDir())
	if err != nil {
		return InitResult{}, fmt.Errorf(
			"check repository: %w",
			err,
		)
	}

	for _, path := range []domain.AbsolutePath{
		workspace.GelDir(),
		workspace.ObjectsDir(),
		workspace.HeadsDir(),
	} {
		if err := ensureDirectory(path); err != nil {
			return InitResult{}, fmt.Errorf(
				"prepare directory %q: %w",
				path,
				err,
			)
		}
	}

	defaultBranch, err := domain.NewRefName(defaultBranchRef)
	if err != nil {
		return InitResult{}, fmt.Errorf("create default branch ref: %w", err)
	}

	head, err := domain.NewSymbolicRef(defaultBranch)
	if err != nil {
		return InitResult{}, fmt.Errorf("create HEAD ref: %w", err)
	}

	headContent, err := domain.EncodeRef(head)
	if err != nil {
		return InitResult{}, fmt.Errorf("encode HEAD: %w", err)
	}
	if err := ensureRegularFile(workspace.HeadPath(), headContent); err != nil {
		return InitResult{}, fmt.Errorf("prepare HEAD: %w", err)
	}
	if err := ensureRegularFile(workspace.ConfigPath(), nil); err != nil {
		return InitResult{}, fmt.Errorf("prepare config: %w", err)
	}
	return InitResult{
		GelDir:        workspace.GelDir(),
		Reinitialized: reinitialized,
	}, nil
}

func ensureDirectory(path domain.AbsolutePath) error {
	info, err := os.Stat(path.String())
	if err == nil {
		if !info.IsDir() {
			return errors.New("path is not a directory")
		}
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(path.String(), defaultDirPermission); err != nil {
			return fmt.Errorf("create directory %q: %w", path, err)
		}
		return nil
	}
	return fmt.Errorf("stat %q: %w", path, err)
}

func ensureRegularFile(path domain.AbsolutePath, body []byte) error {
	// BUG: check-then-create race
	info, err := os.Stat(path.String())
	if err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("path is not a regular file")
		}
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(path.String(), body, defaultFilePermission); err != nil {
			return fmt.Errorf("create file %q: %w", path, err)
		}
		return nil
	}
	return fmt.Errorf("stat %q: %w", path, err)
}

func directoryExists(path domain.AbsolutePath) (bool, error) {
	info, err := os.Stat(path.String())
	if err == nil {
		return info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %q: %w", path, err)
}
