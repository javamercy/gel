package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"cmp"
	"fmt"
	"slices"
)

// Config provides access to repository configuration values.
type Config struct {
	configStore *storage.ConfigStore
}

// NewConfig returns a Config backed by configStore.
//
// configStore must not be nil when the returned Config is used.
func NewConfig(configStore *storage.ConfigStore) (*Config, error) {
	return &Config{
		configStore: configStore,
	}, nil
}

// Get returns the value for section and key and reports whether the key exists.
//
// Get loads the configuration from storage for each call. It returns found as
// false when the section or key does not exist.
func (c *Config) Get(section, key string) (value string, found bool, err error) {
	config, err := c.load()
	if err != nil {
		return "", false, err
	}

	value, found = config.Get(section, key)
	return value, found, nil
}

// Set stores value under section and key, replacing any existing value.
//
// Set loads the current configuration and persists the updated configuration
// to storage. It returns an error when the configuration cannot be loaded, the
// section or key is invalid, or the updated configuration cannot be saved.
func (c *Config) Set(section, key, value string) error {
	config, err := c.load()
	if err != nil {
		return err
	}
	if err := config.Set(section, key, value); err != nil {
		return fmt.Errorf(
			"set %s.%s: %w",
			section, key, err,
		)
	}
	if err := c.configStore.Save(config); err != nil {
		return fmt.Errorf(
			"save config: %w",
			err,
		)
	}
	return nil
}

// List returns all configuration entries sorted by section and then key.
//
// It loads the configuration from storage for each call. An empty
// configuration produces no entries.
func (c *Config) List() ([]domain.ConfigEntry, error) {
	config, err := c.load()
	if err != nil {
		return nil, err
	}

	var entries []domain.ConfigEntry
	for section, values := range config.Sections() {
		for key, value := range values {
			entries = append(
				entries, domain.ConfigEntry{
					Section: section,
					Key:     key,
					Value:   value,
				},
			)
		}
	}

	slices.SortFunc(
		entries, func(a, b domain.ConfigEntry) int {
			if result := cmp.Compare(a.Section, b.Section); result != 0 {
				return result
			}
			return cmp.Compare(a.Key, b.Key)
		},
	)
	return entries, nil
}

func (c *Config) load() (*domain.Config, error) {
	config, err := c.configStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return config, nil
}
