package config

import (
	"fmt"
)

// The number of entries kept in the history if the option is not set in the configuration file.
const defaultMaxEntries = 100

type historyConfig struct {
	History struct {
		MaxEntries *int `yaml:"max_entries"`
	} `yaml:"history"`
}

// Reads the maximum number of history entries from the configuration file. A limit of zero disables the history. If the option is missing, the default limit is returned. Returns an error if reading the configuration file fails or if the limit is negative.
func HistoryLimit() (int, error) {
	if err := initFile(); err != nil {
		return 0, err
	}

	config := &historyConfig{}
	if err := readStruct(config); err != nil {
		return 0, fmt.Errorf("Could not read history limit: %w", err)
	}

	val := config.History.MaxEntries
	if val == nil {
		return defaultMaxEntries, nil
	}

	if *val < 0 {
		return 0, fmt.Errorf("History limit must not be negative, got %d. Please set 'history.max_entries' to zero or higher in the configuration file.", *val)
	}

	return *val, nil
}
