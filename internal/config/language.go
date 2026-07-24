package config

import (
	"fmt"
	"quick-translate/internal/models"
)

type preferenceConfig struct {
	LanguagePreferences struct{
		Source string `yaml:"source"`
		Target string `yaml:"target"`
	} `yaml:"language_preferences"`
}

// Reads the preferred languages from the configuration file. The language tags are not validated besides being strings. Returns an error if the reading the configuration file fails.
func PreferredLanguages() (*models.LanguagePreferences, error) {
	if err := initFile(); err != nil {
		return nil, err
	}

	config := &preferenceConfig{}
	if err := readStruct(config); err != nil {
		return nil, fmt.Errorf("Could not read language preferences: %w", err)
	}

	return &models.LanguagePreferences{
		Source: config.LanguagePreferences.Source,
		Target: config.LanguagePreferences.Target,
	}, nil
}