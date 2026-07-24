package config

import (
	"fmt"
	"quick-translate/internal/models"
	"slices"

	"go.yaml.in/yaml/v4"
)

type defaultProvider struct {
	DefaultProvider string `yaml:"default_provider"`
}

type provider struct {
	Provider map[string]yaml.Node `yaml:"provider"`
}

// Reads the default provider slug from the configuration file. Returns an error if reading the configuration file fails or if the default provider is not set. The slug is not verified for existence in the list of providers.
func DefaultProvider() (string, error) {
	if err := initFile(); err != nil {
		return "", err
	}

	def := &defaultProvider{}
	if err := readStruct(def); err != nil {
		return "", fmt.Errorf("Could not read default provider: %w", err)
	}

	val := def.DefaultProvider
	if len(val) == 0 {
		return "", fmt.Errorf("Default provider is not set. Please set the default provider in the configuration file.")
	}

	return val, nil
}

// Reads the list of configured provider slugs from the configuration file. Returns an error if reading the configuration file fails, if any of the slugs in the list are not in the list of allowed providers or no provider is found in the configuration file.
func ListProviders(slugs []string) ([]string, error) {
	if err := initFile(); err != nil {
		return nil, err
	}

	prov := &provider{}
	if err := readStruct(prov); err != nil {
		return nil, fmt.Errorf("Could not read providers: %w", err)
	}

	var list []string
	for k := range prov.Provider {
		if !slices.Contains(slugs, k) {
			return nil, fmt.Errorf("Provider '%s' is not in the list of allowed providers: %v", k, slugs)
		}

		list = append(list, k)
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("No providers found in the configuration file. Please add at least one provider.")
	}

	return list, nil
}

// Reads the configuration for a specific provider from the configuration file and decodes it into the provided config struct. Returns an error if reading the configuration file fails, if the provider is not configured in the configuration file, or if decoding the configuration fails.
func LoadProvider[T models.Provider](slug string, config *T) error {
	if err := initFile(); err != nil {
		return err
	}

	prov := &provider{}
	if err := readStruct(prov); err != nil {
		return fmt.Errorf("Could not read providers: %w", err)
	}

	node, ok := prov.Provider[slug]
	if !ok {
		return fmt.Errorf("Provider '%s' is not configured in the configuration file", slug)
	}

	if err := node.Decode(config); err != nil {
		return fmt.Errorf("Could not decode provider '%s' configuration: %w", slug, err)
	}

	return nil
}
