package language

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"quick-translate/internal/models"
	"quick-translate/internal/system"
)

type cache struct {
	Ttl          *time.Time
	files        *system.FileService
	ProviderSlug string
	Languages    []*models.Language
}

// Creates a new cache instance for the given provider. The cache is not loaded automatically. Use exists() and read() to check the cache status.
func newCache(providerSlug string, files *system.FileService) *cache {
	return &cache{
		files:        files,
		ProviderSlug: providerSlug,
		Ttl:          nil,
		Languages:    nil,
	}
}

// Checks if the cache already exists for the given provider.
func (c *cache) exists() bool {
	return c.files.Exists(system.Languages, c.fileName())
}

// Checks if the cache is expired based on the TTL (Time To Live) value. Requires cache to be read first. If the TTL is nil, it is considered expired.
func (c *cache) isExpired() bool {
	if c.Ttl == nil {
		return true
	}

	return time.Now().After(*c.Ttl)
}

// Invalidates the cache if it exists.
func (c *cache) invalidate() error {
	_, err := c.files.Remove(system.Languages, c.fileName())

	return err
}

// Reads the cache and decodes it into this cache struct. Returns an error if reading or decoding fails. Requires the cache to exist.
func (c *cache) read() error {
	content, err := c.files.Read(system.Languages, c.fileName())
	if err != nil {
		return fmt.Errorf("Could not open cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	if err := gob.NewDecoder(bytes.NewReader(content)).Decode(c); err != nil {
		return fmt.Errorf("Could not decode cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	return nil
}

// The time the languages of a provider are kept before they are fetched again.
const cacheTtl = 7 * 24 * time.Hour

// Write to the cache by encoding this cache struct. Sets the TTL to the full lifetime, because the languages that are written are the ones that were just fetched. Returns an error if writing fails.
func (c *cache) write() error {
	if c.Languages == nil {
		fmt.Println("Did not find any languages to write to the cache. Skipping cache write.")
		return nil
	}

	ttl := time.Now().Add(cacheTtl)
	c.Ttl = &ttl

	content := new(bytes.Buffer)
	if err := gob.NewEncoder(content).Encode(c); err != nil {
		return fmt.Errorf("Could not encode cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	if err := c.files.Write(system.Languages, content.Bytes(), c.fileName()); err != nil {
		return fmt.Errorf("Could not write cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	return nil
}

// Invalidates caches for other providers. This function checks if the cache exists and is expired for each provider slug, and invalidates it if necessary. Takes a list of all provider slugs to invalidate.
func (c *cache) invalidateOtherCaches(providerSlugs []string) {
	for _, slug := range providerSlugs {
		if slug == c.ProviderSlug {
			continue
		}

		other := newCache(slug, c.files)

		if other.exists() {
			fmt.Printf("Invalidating cache for provider '%s'...\n", slug)

			err := other.read()
			if err != nil {
				fmt.Println(err)
			}

			if other.isExpired() {
				err := other.invalidate()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

// Returns the name the cache file of this provider carries.
func (c *cache) fileName() string {
	return c.ProviderSlug + ".gob"
}
