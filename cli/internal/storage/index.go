package storage

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const indexFilePermission fs.FileMode = 0o644

// IndexStore persists the repository index on the filesystem.
type IndexStore struct {
	path domain.AbsolutePath
}

// NewIndexStore creates an index store for path.
func NewIndexStore(path domain.AbsolutePath) *IndexStore {
	return &IndexStore{
		path: path,
	}
}

// Load reads and decodes the index.
func (s *IndexStore) Load() (*domain.Index, error) {
	data, err := os.ReadFile(s.path.String())
	if err != nil {
		return nil, fmt.Errorf(
			"read index %q: %w",
			s.path.String(),
			err,
		)
	}

	index, err := domain.DecodeIndex(data)
	if err != nil {
		return nil, fmt.Errorf(
			"decode index %q: %w",
			s.path.String(),
			err,
		)
	}
	return index, nil
}

// Save encodes and writes the index.
func (s *IndexStore) Save(index *domain.Index) error {
	data, err := domain.EncodeIndex(index)
	if err != nil {
		return fmt.Errorf(
			"encode index: %w",
			err,
		)
	}

	pathText := s.path.String()
	if err := writeFileAtomically(
		pathText,
		data,
		indexFilePermission,
	); err != nil {
		return fmt.Errorf("write index %q: %w", pathText, err)
	}
	return nil
}

func writeFileAtomically(
	path string,
	data []byte,
	perm fs.FileMode,
) (returnErr error) {
	dir := filepath.Dir(path)
	pattern := "." + filepath.Base(path) + ".temp-*"

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fmt.Errorf("create temporary file in %q: %w", dir, err)
	}

	tempPath := file.Name()
	closed := false
	committed := false

	defer func() {
		if !closed {
			if err := file.Close(); err != nil {
				returnErr = errors.Join(
					returnErr, fmt.Errorf(
						"close temporary file %q: %w",
						tempPath, err,
					),
				)
			}
		}
		if !committed {
			if err := os.Remove(tempPath); err != nil &&
				!errors.Is(err, os.ErrNotExist) {
				returnErr = errors.Join(
					returnErr, fmt.Errorf(
						"remove temporary file %q: %w",
						tempPath, err,
					),
				)
			}
		}
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf(
			"write temporary file %q: %w",
			tempPath,
			err,
		)
	}
	if err := file.Chmod(perm); err != nil {
		return fmt.Errorf(
			"set permissions on temporary file %q: %w",
			tempPath,
			err,
		)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync temporary file %q: %w", tempPath, err)
	}

	err = file.Close()
	closed = true
	if err != nil {
		return fmt.Errorf("close temporary file %q: %w", tempPath, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %q: %w", path, err)
	}

	committed = true
	return nil
}
