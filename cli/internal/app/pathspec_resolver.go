package app

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ResolvedPath struct {
	Normalized domain.NormalizedPath
	Absolute   domain.AbsolutePath
	Missing    bool
}

type PathspecResolver struct {
	repoRoot domain.AbsolutePath
	startDir domain.AbsolutePath
}

func NewPathspecResolver(
	repoRoot domain.AbsolutePath,
	startDir domain.AbsolutePath,
) (*PathspecResolver, error) {
	if repoRoot.IsZero() {
		return nil, errors.New("pathspec repository root is zero")
	}
	if startDir.IsZero() {
		return nil, errors.New("pathspec start directory is zero")
	}
	if _, err := startDir.ToNormalizedPath(repoRoot); err != nil {
		return nil, fmt.Errorf(
			"pathspec start directory is outside repository: %w",
			err,
		)
	}
	return &PathspecResolver{
		repoRoot: repoRoot,
		startDir: startDir,
	}, nil
}

func (r *PathspecResolver) Resolve(pathspecs []string) ([]ResolvedPath, error) {
	if len(pathspecs) == 0 {
		return nil, errors.New("pathspecs are empty")
	}

	resolved := make([]ResolvedPath, 0)
	seen := make(map[string]struct{})

	add := func(
		absolute domain.AbsolutePath,
		normalized domain.NormalizedPath,
		missing bool,
	) {
		normText := normalized.String()
		if _, found := seen[normText]; found {
			return
		}

		seen[normText] = struct{}{}
		resolved = append(
			resolved, ResolvedPath{
				Normalized: normalized,
				Absolute:   absolute,
				Missing:    missing,
			},
		)
	}

	for _, pathspec := range pathspecs {
		absolute, normalized, err := r.resolvePathspec(pathspec)
		if err != nil {
			return nil, err
		}

		info, err := os.Lstat(absolute.String())
		switch {
		case errors.Is(err, fs.ErrNotExist):
			add(absolute, normalized, true)

		case err != nil:
			return nil, fmt.Errorf(
				"inspect pathspec %q: %w",
				pathspec,
				err,
			)

		case info.Mode().IsRegular():
			add(absolute, normalized, false)

		case info.Mode().IsDir():
			found, err := r.resolveDirectory(absolute, add)
			if err != nil {
				return nil, fmt.Errorf(
					"resolve directory pathspec %q: %w",
					pathspec,
					err,
				)
			}
			if !found {
				return nil, fmt.Errorf(
					"pathspec %q contains no stageable files",
					pathspec,
				)
			}

		default:
			return nil, fmt.Errorf(
				"pathspec %q has unsupported file type %q",
				pathspec,
				info.Mode().String(),
			)
		}
	}
	return resolved, nil
}

func (r *PathspecResolver) resolvePathspec(pathspec string) (
	domain.AbsolutePath,
	domain.NormalizedPath,
	error,
) {
	if pathspec == "" {
		return domain.AbsolutePath{},
			domain.NormalizedPath{},
			errors.New("pathspec is empty")
	}

	path := pathspec
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.startDir.String(), path)
	}

	absolute, err := domain.NewAbsolutePath(path)
	if err != nil {
		return domain.AbsolutePath{},
			domain.NormalizedPath{},
			fmt.Errorf(
				"create absolute path: %w",
				err,
			)
	}

	normalized, err := absolute.ToNormalizedPath(r.repoRoot)
	if err != nil {
		return domain.AbsolutePath{}, domain.NormalizedPath{}, fmt.Errorf(
			"pathspec %q is outside repository: %w",
			pathspec,
			err,
		)
	}
	if containsGelDirectory(normalized) {
		return domain.AbsolutePath{}, domain.NormalizedPath{}, fmt.Errorf(
			"pathspec %q contains %q directory",
			pathspec,
			domain.GelDirName,
		)
	}
	if err := r.rejectSymlinkComponents(normalized); err != nil {
		return domain.AbsolutePath{}, domain.NormalizedPath{}, fmt.Errorf(
			"pathspec %q: %w",
			pathspec,
			err,
		)
	}
	return absolute, normalized, nil
}

func (r *PathspecResolver) resolveDirectory(
	dir domain.AbsolutePath,
	add func(
		domain.AbsolutePath,
		domain.NormalizedPath,
		bool,
	),
) (bool, error) {
	found := false

	err := filepath.WalkDir(
		dir.String(),
		func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if strings.EqualFold(entry.Name(), domain.GelDirName) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if entry.IsDir() {
				return nil
			}

			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("inspect path %q: %w", path, err)
			}
			if info.Mode()&fs.ModeSymlink != 0 {
				return fmt.Errorf("path %q is a symlink", path)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf(
					"path %q has unsupported file type %q",
					path,
					info.Mode().String(),
				)
			}

			absolute, err := domain.NewAbsolutePath(path)
			if err != nil {
				return fmt.Errorf(
					"create absolute path %q: %w",
					path,
					err,
				)
			}

			normalized, err := absolute.ToNormalizedPath(r.repoRoot)
			if err != nil {
				return fmt.Errorf(
					"normalize path %q: %w",
					path,
					err,
				)
			}

			add(absolute, normalized, false)
			found = true
			return nil
		},
	)
	return found, err
}

func (r *PathspecResolver) rejectSymlinkComponents(path domain.NormalizedPath) error {
	current := r.repoRoot

	for component := range strings.SplitSeq(path.String(), "/") {
		if component == "" {
			continue
		}

		var err error
		current, err = current.Join(component)
		if err != nil {
			return err
		}

		info, err := os.Lstat(current.String())
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil

		case err != nil:
			return fmt.Errorf(
				"inspect path %q: %w",
				current,
				err,
			)

		case info.Mode()&fs.ModeSymlink != 0:
			return fmt.Errorf("path %q is a symlink", current)
		}
	}
	return nil
}

func containsGelDirectory(path domain.NormalizedPath) bool {
	for component := range strings.SplitSeq(path.String(), "/") {
		if strings.EqualFold(component, domain.GelDirName) {
			return true
		}
	}
	return false
}
