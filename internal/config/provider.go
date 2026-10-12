package config

import (
	"errors"
	"fmt"
	"maps"
	"quick-translate/internal/models"
	"slices"
)

// Returns the default provider. Fails if it is not set; whether it is configured is not checked.
//
//wails:ignore
func (s *Service) DefaultProvider() (models.ProviderID, error) {
	if s.config.defaultProvider == 0 {
		return 0, errors.New("config: default provider not set")
	}

	return s.config.defaultProvider, nil
}

// Returns every configured provider, sorted by ID. Fails if there is none.
//
//wails:ignore
func (s *Service) Providers() ([]models.ProviderID, error) {
	if len(s.config.providers) == 0 {
		return nil, errors.New("config: no provider configured")
	}

	return slices.Sorted(maps.Keys(s.config.providers)), nil
}

// Returns a decoder for the configuration of the given provider. The error means not configured, the decoder's error means configured but unreadable.
//
//wails:ignore
func (s *Service) Provider(id models.ProviderID) (decode func(into any) error, err error) {
	node, ok := s.config.providers[id]
	if !ok {
		return nil, fmt.Errorf("config: provider %q not configured", id)
	}

	return func(into any) error {
		if err := node.Decode(into); err != nil {
			return fmt.Errorf("config: decode provider %q: %w", id, err)
		}

		return nil
	}, nil
}
