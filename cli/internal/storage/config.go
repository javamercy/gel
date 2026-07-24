package storage

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

const configFilePermission fs.FileMode = 0o644

// ConfigStore persists the repository configuration.
type ConfigStore struct {
	configPath domain.AbsolutePath
}

// NewConfigStore creates a store for the config file at configPath.
func NewConfigStore(configPath domain.AbsolutePath) *ConfigStore {
	return &ConfigStore{
		configPath: configPath,
	}
}

// Load reads and decodes the configuration file.
func (s *ConfigStore) Load() (*domain.Config, error) {
	file, err := os.Open(s.configPath.String())
	if err != nil {
		return nil, fmt.Errorf(
			"open config %q: %w",
			s.configPath,
			err,
		)
	}

	defer func() {
		_ = file.Close()
	}()

	sections := make(map[string]domain.ConfigSection)
	if _, err := toml.NewDecoder(file).Decode(&sections); err != nil {
		return nil, fmt.Errorf(
			"decode config %q: %w",
			s.configPath,
			err,
		)
	}

	config, err := domain.NewConfigFromSections(sections)
	if err != nil {
		return nil, fmt.Errorf(
			"validate config %q: %w",
			s.configPath,
			err,
		)
	}
	return config, nil
}

// Save encodes and atomically replaces the configuration file.
func (s *ConfigStore) Save(config *domain.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	encoded, err := toml.Marshal(config.Sections())
	if err != nil {
		return fmt.Errorf(
			"encode config %q: %w",
			s.configPath,
			err,
		)
	}
	if err := replaceFileAtomically(
		s.configPath.String(),
		encoded,
		configFilePermission,
	); err != nil {
		return fmt.Errorf(
			"write config %q: %w",
			s.configPath,
			err,
		)
	}
	return nil
}
