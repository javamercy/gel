package storage

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	refDirPermission  fs.FileMode = 0o755
	refFilePermission fs.FileMode = 0o644
)

var (
	ErrRefNotExist = errors.New("ref does not exist")
	ErrRefExists   = errors.New("ref already exists")
)

// RefStore persists Gel references beneath a repository's .gel directory.
type RefStore struct {
	gelDir domain.AbsolutePath
}

// NewRefStore returns a reference store rooted at gelDir.
func NewRefStore(gelDir domain.AbsolutePath) *RefStore {
	return &RefStore{
		gelDir: gelDir,
	}
}

// Read reads and decodes the reference identified by name.
//
// It returns an error when the reference cannot be read or is malformed.
func (s *RefStore) Read(name domain.RefName) (domain.Ref, error) {
	refPath, err := s.gelDir.Join(name.String())
	if err != nil {
		return domain.Ref{}, fmt.Errorf(
			"join ref %q: %w",
			name, err,
		)
	}

	data, err := os.ReadFile(refPath.String())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.Ref{}, fmt.Errorf(
				"read ref %q: %w: %w",
				name,
				ErrRefNotExist,
				err,
			)
		}
		return domain.Ref{}, fmt.Errorf(
			"read ref %q: %w",
			name, err,
		)
	}

	ref, err := domain.DecodeRef(data)
	if err != nil {
		return domain.Ref{}, fmt.Errorf(
			"decode ref %q: %w",
			name, err,
		)
	}
	return ref, nil
}

// Write encodes ref and atomically replaces the reference identified by name.
//
// It creates parent reference directories as needed.
func (s *RefStore) Write(name domain.RefName, ref domain.Ref) error {
	encoded, err := domain.EncodeRef(ref)
	if err != nil {
		return fmt.Errorf(
			"encode ref %q: %w",
			name, err,
		)
	}

	refPath, err := s.gelDir.Join(name.String())
	if err != nil {
		return fmt.Errorf(
			"join ref %q: %w",
			name, err,
		)
	}

	refDir := filepath.Dir(refPath.String())
	if err := os.MkdirAll(refDir, refDirPermission); err != nil {
		return fmt.Errorf(
			"create ref directory %q: %w",
			refDir, err,
		)
	}
	if err := replaceFileAtomically(
		refPath.String(),
		encoded,
		refFilePermission,
	); err != nil {
		return fmt.Errorf(
			"write ref %q: %w",
			name, err,
		)
	}
	return nil
}

func (s *RefStore) List(prefix domain.RefName) ([]domain.RefName, error) {
	refsPath, err := s.gelDir.Join(prefix.String())
	if err != nil {
		return nil, fmt.Errorf(
			"join ref %q: %w",
			prefix, err,
		)
	}

	var refs []domain.RefName
	err = filepath.WalkDir(
		refsPath.String(), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf(
					"ref path %q is not a regular file",
					path,
				)
			}

			relative, err := filepath.Rel(s.gelDir.String(), path)
			if err != nil {
				return fmt.Errorf(
					"make ref path %q relative: %w",
					path,
					err,
				)
			}

			refName, err := domain.NewRefName(filepath.ToSlash(relative))
			if err != nil {
				return fmt.Errorf(
					"parse ref name %q: %w",
					path, err,
				)
			}

			refs = append(refs, refName)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"walk refs under %q: %w",
			refsPath, err,
		)
	}
	return refs, nil
}

func (s *RefStore) Create(name domain.RefName, ref domain.Ref) error {
	encoded, err := domain.EncodeRef(ref)
	if err != nil {
		return fmt.Errorf(
			"encode ref %q: %w",
			name, err,
		)
	}

	refPath, err := s.gelDir.Join(name.String())
	if err != nil {
		return fmt.Errorf(
			"join ref %q: %w",
			name, err,
		)
	}

	_, err = os.Stat(refPath.String())
	if err == nil {
		return fmt.Errorf(
			"create ref %q: %w",
			name,
			ErrRefExists,
		)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf(
			"inspect ref %q: %w",
			name, err,
		)
	}

	refDir := filepath.Dir(refPath.String())
	if err := os.MkdirAll(refDir, refDirPermission); err != nil {
		return fmt.Errorf(
			"create ref directory %q: %w",
			refDir, err,
		)
	}

	if err := os.WriteFile(
		refPath.String(),
		encoded,
		refFilePermission,
	); err != nil {
		return fmt.Errorf(
			"write ref %q: %w",
			name, err,
		)
	}
	return nil
}

func (s *RefStore) Delete(name domain.RefName) error {
	refPath, err := s.gelDir.Join(name.String())
	if err != nil {
		return fmt.Errorf(
			"join ref %q: %w",
			name, err,
		)
	}

	info, err := os.Lstat(refPath.String())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf(
				"delete ref %q: %w: %w",
				name,
				ErrRefNotExist,
				err,
			)
		}
		return fmt.Errorf("delete ref %q: %w", name, err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf(
			"ref path %q is not a regular file",
			refPath,
		)
	}

	if err := os.Remove(refPath.String()); err != nil {
		return fmt.Errorf("delete ref %q: %w", name, err)
	}
	return nil
}
