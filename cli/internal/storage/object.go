package storage

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ObjectStorage provides low-level object database persistence under .gel/objects.
type ObjectStorage struct {
	workspace *domain.Workspace
}

// NewObjectStorage creates object storage bound to the repository workspace.
func NewObjectStorage(workspace *domain.Workspace) *ObjectStorage {
	return &ObjectStorage{
		workspace: workspace,
	}
}

// Write stores compressed object data at its hash-derived object path.
func (o *ObjectStorage) Write(hash domain.Hash, data []byte) error {
	objectPath, err := o.hashToObjectPath(hash)
	if err != nil {
		return err
	}
	dir := filepath.Dir(objectPath.String())
	if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}
	if err := os.WriteFile(objectPath.String(), data, domain.DefaultFilePermission); err != nil {
		return fmt.Errorf("failed to write object '%s': %w", hash, err)
	}
	return nil
}

// Read loads compressed object data by hash.
func (o *ObjectStorage) Read(hash domain.Hash) ([]byte, error) {
	objectPath, err := o.hashToObjectPath(hash)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(objectPath.String())
	if err != nil {
		return nil, fmt.Errorf("failed to read object '%s': %w", hash, err)
	}
	return data, nil
}

// Exists reports whether an object exists for the given hash.
func (o *ObjectStorage) Exists(hash domain.Hash) (bool, error) {
	objectPath, err := o.hashToObjectPath(hash)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(objectPath.String())
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check object '%s' existence: %w", hash, err)
}

// hashToObjectPath converts a hash to .gel/objects/<2-char-prefix>/<remaining> path.
func (o *ObjectStorage) hashToObjectPath(hash domain.Hash) (domain.AbsolutePath, error) {
	hexHash := hash.Hex()
	dir := hexHash[:2]
	file := hexHash[2:]
	joined := filepath.Join(o.workspace.ObjectsDir().String(), dir, file)
	absPath, err := domain.NewAbsolutePath(joined)
	if err != nil {
		return domain.AbsolutePath{}, err
	}
	return absPath, nil
}
