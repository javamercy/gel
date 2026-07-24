package storage

import (
	"Gel/internal/domain"
	"fmt"
	"io/fs"
	"os"
)

const indexFilePermission fs.FileMode = 0o644

// IndexStore persists the repository index on the filesystem.
type IndexStore struct {
	indexPath domain.AbsolutePath
}

// NewIndexStore creates an index store for path.
func NewIndexStore(indexPath domain.AbsolutePath) *IndexStore {
	return &IndexStore{
		indexPath: indexPath,
	}
}

// Load reads and decodes the index.
func (s *IndexStore) Load() (*domain.Index, error) {
	data, err := os.ReadFile(s.indexPath.String())
	if err != nil {
		return nil, fmt.Errorf(
			"read index %q: %w",
			s.indexPath.String(),
			err,
		)
	}

	index, err := domain.DecodeIndex(data)
	if err != nil {
		return nil, fmt.Errorf(
			"decode index %q: %w",
			s.indexPath.String(),
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

	pathText := s.indexPath.String()
	if err := replaceFileAtomically(
		pathText,
		data,
		indexFilePermission,
	); err != nil {
		return fmt.Errorf("write index %q: %w", pathText, err)
	}
	return nil
}
