package core

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"Gel/internal/validate"
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
)

const (
	// ConfigSectionUser stores author/committer identity defaults.
	ConfigSectionUser = "user"
	// ConfigKeyName is the user name key under [user].
	ConfigKeyName = "name"
	// ConfigKeyEmail is the user email key under [user].
	ConfigKeyEmail = "email"
)

// ConfigService manages repository config stored in .gel/config.toml.
type ConfigService struct {
	configStorage *storage.ConfigStorage
}

// NewConfigService creates a config service.
func NewConfigService(configStorage *storage.ConfigStorage) *ConfigService {
	return &ConfigService{
		configStorage: configStorage,
	}
}

// GetUserInfo returns user.name and user.email used by commit operations.
func (c *ConfigService) GetUserInfo() (string, string, error) {
	config, err := c.Read()
	if err != nil {
		return "", "", err
	}

	name, ok := config.Get(ConfigSectionUser, ConfigKeyName)
	if !ok {
		return "", "", fmt.Errorf("config: user.name is not set")
	}

	email, ok := config.Get(ConfigSectionUser, ConfigKeyEmail)
	if !ok {
		return "", "", fmt.Errorf("config: user.email is not set")
	}
	return name, email, nil
}

// Set writes section.key=value to config, creating missing sections as needed.
func (c *ConfigService) Set(section, key, value string) error {
	config, err := c.Read()
	if err != nil {
		return err
	}
	if err := config.Set(section, key, value); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	return c.Write(config)
}

// Get returns a config value by section and key.
func (c *ConfigService) Get(section, key string) (string, error) {
	if err := validate.StringMustNotBeEmpty(section); err != nil {
		return "", fmt.Errorf("config: %w", err)
	}
	if err := validate.StringMustNotBeEmpty(key); err != nil {
		return "", fmt.Errorf("config: %w", err)
	}

	config, err := c.Read()
	if err != nil {
		return "", err
	}
	value, ok := config.Get(section, key)
	if !ok {
		return "", fmt.Errorf("config: key '%s.%s' not found", section, key)
	}
	return value, nil
}

// List returns all config entries in "section.key=value" format.
// Results are sorted by section then key for deterministic ordering.
func (c *ConfigService) List() ([]string, error) {
	/*config, err := c.Read()
	if err != nil {
		return nil, err
	}

	entries := config.Entries()
	slices.SortFunc(
		entries, func(a, b domain.ConfigEntry) int {
			if a.section != b.section {
				return strings.Compare(a.section, b.section)
			}
			return strings.Compare(a.key, b.key)
		},
	)

	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, fmt.Sprintf("%s.%s=%s", entry.section, entry.key, entry.value))
	}*/
	return nil, nil
}

// Write encodes and persists config data to storage.
func (c *ConfigService) Write(config *domain.Config) error {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if err := encoder.Encode(nil); err != nil {
		return fmt.Errorf("config: failed to encode config: %w", err)
	}
	if err := c.configStorage.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// Read loads and decodes config from storage.
// Empty files are treated as empty config maps.
func (c *ConfigService) Read() (*domain.Config, error) {
	data, err := c.configStorage.Read()
	if err != nil {
		return nil, err
	}

	sections := make(map[string]domain.ConfigSection)
	if _, err := toml.Decode(string(data), sections); err != nil {
		return nil, fmt.Errorf("config: failed to decode config: %w", err)
	}

	config, err := domain.NewConfigFromSections(sections)
	if err != nil {
		return nil, fmt.Errorf("config: invalid config: %w", err)
	}
	return config, nil
}
