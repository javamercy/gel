package domain

import (
	"errors"
	"fmt"
)

// ConfigSection maps configuration keys to string values within one section.
type ConfigSection map[string]string

// ConfigEntry represents one configuration Value identified by Section and Key.
type ConfigEntry struct {
	Section string
	Key     string
	Value   string
}

// Config stores string-valued configuration grouped by section.
// The zero value is a valid empty configuration.
type Config struct {
	sections map[string]ConfigSection
}

// NewConfigFromSections validates sections and constructs a Config containing a defensive copy.
func NewConfigFromSections(sections map[string]ConfigSection) (*Config, error) {
	sectionsCopy := make(map[string]ConfigSection, len(sections))
	for sectionName, section := range sections {
		if sectionName == "" {
			return nil, errors.New("section is empty")
		}

		sectionCopy := make(ConfigSection, len(section))
		for key, value := range section {
			if key == "" {
				return nil, fmt.Errorf(
					"section %q contains an empty key",
					sectionName,
				)
			}
			sectionCopy[key] = value
		}
		sectionsCopy[sectionName] = sectionCopy
	}

	return &Config{
		sections: sectionsCopy,
	}, nil
}

// Get returns the value at section and key and reports whether it exists.
func (c *Config) Get(section, key string) (string, bool) {
	configSection, ok := c.sections[section]
	if !ok {
		return "", false
	}

	value, ok := configSection[key]
	return value, ok
}

// Set stores value at section and key, creating the section when necessary.
func (c *Config) Set(section, key, value string) error {
	if section == "" {
		return errors.New("section is empty")
	}
	if key == "" {
		return errors.New("key is empty")
	}
	if c.sections == nil {
		c.sections = make(map[string]ConfigSection)
	}

	configSection, ok := c.sections[section]
	if !ok {
		configSection = make(ConfigSection)
		c.sections[section] = configSection
	}

	configSection[key] = value
	return nil
}
