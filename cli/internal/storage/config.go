package storage

import (
	"Gel/internal/domain"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigStorage struct {
	workspace *domain.Workspace
}

// NewConfigStorage creates storage access for .gel/config.toml.
func NewConfigStorage(workspace *domain.Workspace) *ConfigStorage {
	return &ConfigStorage{
		workspace: workspace,
	}
}

// Read returns raw config bytes from disk.
func (c *ConfigStorage) Read() ([]byte, error) {
	data, err := os.ReadFile(c.workspace.ConfigPath().String())
	if err != nil {
		return nil, fmt.Errorf("config: error reading config file: %w", err)
	}
	return data, nil
}

// Write persists raw config bytes to disk, creating parent directories if needed.
func (c *ConfigStorage) Write(data []byte) error {
	dir := filepath.Dir(c.workspace.ConfigPath().String())
	if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
		return fmt.Errorf("config: failed to create directory '%s': %w", dir, err)
	}
	if err := os.WriteFile(c.workspace.ConfigPath().String(), data, domain.DefaultFilePermission); err != nil {
		return fmt.Errorf("config: error writing config file: %w", err)
	}
	return nil
}
