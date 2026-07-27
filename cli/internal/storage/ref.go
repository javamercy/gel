package storage

import (
	"Gel/internal/domain"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	refDirPermission  fs.FileMode = 0o755
	refFilePermission fs.FileMode = 0o644
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
		return domain.Ref{}, fmt.Errorf("join ref %q: %w", name, err)
	}

	data, err := os.ReadFile(refPath.String())
	if err != nil {
		return domain.Ref{}, fmt.Errorf("read ref %q: %w", name, err)
	}

	ref, err := domain.DecodeRef(data)
	if err != nil {
		return domain.Ref{}, fmt.Errorf("decode ref %q: %w", name, err)
	}
	return ref, nil
}

// Write encodes ref and atomically replaces the reference identified by name.
//
// It creates parent reference directories as needed.
func (s *RefStore) Write(name domain.RefName, ref domain.Ref) error {
	encoded, err := domain.EncodeRef(ref)
	if err != nil {
		return fmt.Errorf("encode ref %q: %w", name, err)
	}

	refPath, err := s.gelDir.Join(name.String())
	if err != nil {
		return fmt.Errorf("join ref %q: %w", name, err)
	}

	refDir := filepath.Dir(refPath.String())
	if err := os.MkdirAll(refDir, refDirPermission); err != nil {
		return fmt.Errorf("create ref directory %q: %w", refDir, err)
	}
	if err := replaceFileAtomically(
		refPath.String(),
		encoded,
		refFilePermission,
	); err != nil {
		return fmt.Errorf("write ref %q: %w", name, err)
	}
	return nil
}
