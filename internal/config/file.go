package config

import (
	"fmt"

	"quick-translate/internal/system"

	"go.yaml.in/yaml/v4"
)

const defaultYml = `# <=== Language Preferences ===>
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

// Creates the config file with the commented default template when it is not there yet, and narrows the permissions of one an earlier version created readable for everyone else on the machine. Returns an error if the file can not be written.
func initFile(files *system.FileService) error {
	created, err := files.EnsureFile(system.Settings, []byte(defaultYml))
	if err != nil {
		return fmt.Errorf("Could not create config file: %w", err)
	}

	if created {
		fmt.Println("Could not find config file. Created a new one with default values at: " + files.Path(system.Settings))
	}

	return nil
}

// Reads a struct from the config file. The struct must have YAML tags from the "go.yaml.in/yaml/v4" package. Every caller runs initFile first, so the file is there; one that is empty leaves the struct at its own defaults. The function will return an error if the file can not be read or if the struct cannot be decoded from it.
func readStruct[T any](files *system.FileService, x *T) error {
	content, err := files.Read(system.Settings)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(content, x); err != nil {
		return fmt.Errorf("Could not decode YAML from config file: %w", err)
	}

	return nil
}
