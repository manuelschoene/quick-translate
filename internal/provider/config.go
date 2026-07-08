package provider

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"go.yaml.in/yaml/v4"
)

const defaultYaml = `# <=== Default Provider ===>
#
# The default provider is selected on startup and used for all translations unless changed in the GUI.
# Set this option to the slug of the provider defined in the provider section below (e.g. "deepl").
# It is recommended to set this option to a provider that supports language detection.
#
default: ""


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

type config struct {
	Default  string               `yaml:"default"`
	Provider map[string]yaml.Node `yaml:"provider"`
}

func metas() []*Meta {
	return []*Meta{
		&Meta{
			Slug:              "deepl",
			Name:              "DeepL",
			LanguageDetection: true,
		},
	}
}

func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return dir + "/quick-translate/provider.yaml", nil
}

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

func touch() error {
	path, err := filePath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func readFile(x any) error {
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
		return err
	}

	return nil
}

func initFile() error {
	if exists, err := fileExists(); err != nil {
		return err
	} else if !exists {
		if err := touch(); err != nil {
			return err
		}

		path, err := filePath()
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		buf.WriteString(defaultYaml)

		err = os.WriteFile(path, buf.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeProvider(config config, meta *Meta) (provider, error) {
	node, ok := config.Provider[meta.Slug]
	if !ok {
		return nil, fmt.Errorf("Invalid config: provider '%s' is not configured", meta.Slug)
	}

	p := meta.provider()
	if err := node.Decode(p); err != nil {
		return nil, fmt.Errorf("Invalid config: provider '%s' is not configured correctly: %w", meta.Slug, err)
	}

	return p, nil
}

func loadConfig() ([]*Meta, *Meta, provider, error) {
	var config config
	err := readFile(&config)
	if err != nil {
		return nil, nil, nil, err
	}

	slugs := slices.Collect(maps.Keys(config.Provider))

	if len(slugs) == 0 {
		return nil, nil, nil, fmt.Errorf("Invalid config: no providers are configured")
	}

	def := config.Default
	if def == "" {
		return nil, nil, nil, fmt.Errorf("Invalid config: no default provider is configured")
	}

	var defMeta *Meta
	all := metas()

	for k, m := range all {
		var found bool
		for _, s := range slugs {
			if m.Slug == s {
				found = true

				if m.Slug == def {
					defMeta = m
				}

				break
			}
		}

		if !found {
			all = append(all[:k], all[k+1:]...)
		}
	}

	if len(all) == 0 {
		return nil, nil, nil, fmt.Errorf("Invalid config: no configured provider is supported")
	}

	if defMeta == nil {
		return nil, nil, nil, fmt.Errorf("Invalid config: default provider '%s' is not supported or configured", def)
	}

	p, err := decodeProvider(config, defMeta)
	if err != nil {
		return nil, nil, nil, err
	}

	return all, defMeta, p, nil
}

func loadProvider(meta *Meta) (provider, error) {
	var config config
	if err := readFile(&config); err != nil {
		return nil, err
	}

	p, err := decodeProvider(config, meta)
	if err != nil {
		return nil, err
	}

	return p, nil
}
