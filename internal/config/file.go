package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

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

// The permissions the configuration file and the directory around it are kept at. The file holds the
// authentication keys of the providers and the directory also holds the history database, which holds every
// text that was ever translated, so neither belongs to the other users of the machine.
const (
	fileMode = 0600
	dirMode  = 0700
)

// Builds the path to the config file inside the user's config directory.
func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return dir + "/quick-translate/config.yml", nil
}

// Checks if the config file exists.
func fileExists() (bool, error) {
	path, err := filePath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

// Creates all necessary directories and the config file with default values. Does not check if the file already exists.
func createFile() error {
	path, err := filePath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), dirMode)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.WriteString(defaultYml)

	err = os.WriteFile(path, buf.Bytes(), fileMode)
	if err != nil {
		return err
	}

	return nil
}

// Initializes the config file if it does not exist. If the file exists, it does nothing.
func initFile() error {
	path, err := filePath()
	if err != nil {
		return err
	}

	exists, err := fileExists()
	if err != nil {
		return err
	}

	if exists {
		narrow(filepath.Dir(path), dirMode)
		narrow(path, fileMode)

		return nil
	}

	fmt.Println("Could not find config file. Creating new file with default values at: " + path)

	err = createFile()
	if err != nil {
		return fmt.Errorf("Could not create config file: %w", err)
	}

	return nil
}

// Narrows the permissions of a path that an earlier version has created readable for everyone else on the
// machine. A path that is already narrow enough is left alone and never reported, which is the normal case.
// Failing to narrow only costs the narrowing itself, so it is reported and swallowed rather than keeping the
// application from starting.
func narrow(path string, mode os.FileMode) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.Mode().Perm()&0077 == 0 {
		return
	}

	if err := os.Chmod(path, mode); err != nil {
		fmt.Printf("Could not narrow the permissions of '%s': %v\n", path, err)
		return
	}

	fmt.Printf("Narrowed the permissions of '%s', which is not meant to be readable by other users.\n", path)
}

// Reads a struct from the config file. The struct must have YAML tags from the "go.yaml.in/yaml/v4" package. The function will return an error if the file does not exist or if the struct cannot be decoded from the file.
func readStruct[T any](x *T) error {
	path, err := filePath()
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(x)
	if err != nil {
		return fmt.Errorf("Could not decode YAML from config file: %w", err)
	}

	return nil
}
