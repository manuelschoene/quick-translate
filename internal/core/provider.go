package core

import (
	"fmt"
	"slices"

	"quick-translate/internal/config"
	"quick-translate/internal/models"
	"quick-translate/internal/provider"
)

// Returns the slugs of every provider that is configured in the configuration file, sorted by slug.
func (c *Core) Providers() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return slices.Clone(c.providerList)
}

// Returns the slug of the provider that translations are currently made with.
func (c *Core) CurrentProvider() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.providerSlug
}

// Switches to the provider with the given slug and rebuilds the languages for it, carrying the current languages over as far as the new provider supports them. The choice is not written to the configuration file, so the default provider is used again after a restart. Nothing is changed when the provider or its languages can not be loaded. Returns an error if the slug is not configured or the provider can not be built.
func (c *Core) SetProvider(slug string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.setProvider(slug)
}

// Switches to the provider with the given slug. Requires the lock to be held.
func (c *Core) setProvider(slug string) error {
	if slug == c.providerSlug {
		return nil
	}

	if !slices.Contains(c.providerList, slug) {
		return fmt.Errorf("The provider '%s' is not configured. Please choose one of: %v", slug, c.providerList)
	}

	prov, err := c.factory(slug)
	if err != nil {
		return err
	}

	langs, err := c.buildCollection(slug, prov, c.langs.Source(), c.langs.Target(), c.langs.DetectedSource())
	if err != nil {
		return fmt.Errorf("Could not switch to the provider '%s': %w", slug, err)
	}

	// The state is replaced as a whole, so a failed switch leaves the previous provider fully usable.
	c.providerSlug, c.provider, c.langs = slug, prov, langs

	return nil
}

// Creates the provider for the given slug and fills it with its section of the configuration file. Returns an error if the slug is not implemented or the configuration can not be read.
func buildProvider(slug string) (models.Provider, error) {
	switch slug {
	case provider.SlugDeepl:
		return decodeProvider(slug, provider.NewDeepl())
	default:
		return nil, fmt.Errorf("The provider '%s' is known but not implemented yet.", slug)
	}
}

// Reads the configuration of the given provider into the instance and returns it. The instance is returned by value, so a provider that was handed out can not be reconfigured through the core. Returns an error if the provider is not configured in the configuration file.
func decodeProvider[T models.Provider](slug string, instance *T) (models.Provider, error) {
	if err := config.LoadProvider(slug, instance); err != nil {
		return nil, err
	}

	return *instance, nil
}
