package storage

import (
	"Gel/internal/domain"
	"fmt"
	"os"
)

// IndexStorage provides raw index file persistence under .gel/index.
type IndexStorage struct {
	workspace *domain.Workspace
}

// NewIndexStorage creates an index storage bound to the repository workspace.
func NewIndexStorage(workspace *domain.Workspace) *IndexStorage {
	return &IndexStorage{
		workspace: workspace,
	}
}

// Read loads the entire index file as bytes.
func (i *IndexStorage) Read() ([]byte, error) {
	data, err := os.ReadFile(i.workspace.IndexPath().String())
	if err != nil {
		return nil, fmt.Errorf("error reading index file: %w", err)
	}
	return data, nil
}

// Write replaces the index file with data.
func (i *IndexStorage) Write(data []byte) error {
	if err := os.WriteFile(i.workspace.IndexPath().String(), data, domain.DefaultFilePermission); err != nil {
		return fmt.Errorf("error writing index file: %w", err)
	}
	return nil
}
