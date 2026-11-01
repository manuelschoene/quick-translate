package config

import (
	"fmt"

	"quick-translate/internal/models"
	"quick-translate/internal/system"

	"go.yaml.in/yaml/v4"
)

const fileTemplate = `# <=== Language Preferences ===>
#
# The language preferences section allows you to specify your preferred source and target languages for translation.
# Use an BCP 47 language tag (e.g. "en" for English, "de" for German) to specify the source and target languages.
# When leaving the options empty and the default provider supports language detection, the source language will be detected automatically.
# The target language will be set to the system locale by default, if not specified. 
#
language_preferences:
  source:
  target:


# <=== History ===>
#
# The history section allows you to specify the maximum number of translations to keep in the history.
# Setting the max_entries to 0 will disable the history feature.
#
history:
  max_entries: 100


# <=== Default Provider ===>
#
# The default provider is selected on startup and used for all translations unless changed in the GUI.
# Set this option to the slug of the provider defined in the provider section below (e.g. "deepl").
# It is recommended to set this option to a provider that supports language detection.
#
default_provider: ""


# <=== Provider Configuration ===>
#
# The provider configuration section contains the configuration for each provider.
# Each provider is identified by its slug (e.g. "deepl") and can have its own configuration options.
# See the documentation for each provider for more information on the available configuration options.
#
# Example:
#
# provider:
# 	deepl:
#   	auth_key: YOUR_DEEPL_API_KEY
#		free_version: true
#       fast_mode: false
#       formality: default
#
provider: {}
`

type file struct {
	LanguagePreferences LanguagePreferences `yaml:"language_preferences"`
	History             struct {
		MaxEntries *int `yaml:"max_entries"`
	} `yaml:"history"`
	DefaultProvider string               `yaml:"default_provider"`
	Provider        map[string]yaml.Node `yaml:"provider"`
}

// The validated content of the config file, with provider slugs resolved to IDs.
type config struct {
	preferences     LanguagePreferences
	historyLimit    int
	defaultProvider models.ProviderID // Zero when not set
	providers       map[models.ProviderID]yaml.Node
}

// Reads, decodes and validates the config file.
func (s *Service) load() (*config, error) {
	content, err := s.files.Read(system.Settings)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	f := new(file)
	if err := yaml.Unmarshal(content, f); err != nil {
		return nil, fmt.Errorf("config: decode file: %w", err)
	}

	return validate(f)
}

// Validates the config file content and returns the validated config. An empty file leaves every option at its default; an unknown provider slug is an error.
func validate(f *file) (*config, error) {
	c := &config{
		preferences:  f.LanguagePreferences,
		historyLimit: defaultMaxEntries,
		providers:    make(map[models.ProviderID]yaml.Node, len(f.Provider)),
	}

	if f.History.MaxEntries != nil {
		if *f.History.MaxEntries < 0 {
			return nil, fmt.Errorf("config: history.max_entries must not be negative, got %d", *f.History.MaxEntries)
		}

		c.historyLimit = *f.History.MaxEntries
	}

	if f.DefaultProvider != "" {
		ID, ok := models.ProviderIDFromSlug(f.DefaultProvider)
		if !ok {
			return nil, fmt.Errorf("config: default_provider: unknown provider %q", f.DefaultProvider)
		}

		c.defaultProvider = ID
	}

	for slug, node := range f.Provider {
		ID, ok := models.ProviderIDFromSlug(slug)
		if !ok {
			return nil, fmt.Errorf("config: provider: unknown provider %q", slug)
		}

		c.providers[ID] = node
	}

	return c, nil
}
