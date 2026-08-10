package language

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"quick-translate/internal/models"
	"time"
)

type cache struct {
	Ttl          *time.Time
	ProviderSlug string
	Languages    []*models.Language
}

// Creates a new cache instance for the given provider. The cache is not loaded automatically. Use exists() and read() to check the cache status.
func newCache(providerSlug string) *cache {
	return &cache{
		ProviderSlug: providerSlug,
		Ttl:          nil,
		Languages:    nil,
	}
}

// Checks if the cache already exists for the given provider.
func (c *cache) exists() bool {
	path, err := filePath(c.ProviderSlug)
	if err != nil {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
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
	path, err := filePath(c.ProviderSlug)
	if err != nil {
		return fmt.Errorf("Could not invalidate cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	if !c.exists() {
		return nil
	}

	return os.Remove(path)
}

// Reads the cache and decodes it into this cache struct. Returns an error if reading or decoding fails. Requires the cache to exist.
func (c *cache) read() error {
	path, err := filePath(c.ProviderSlug)
	if err != nil {
		return fmt.Errorf("Could not read cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Could not open cache file for provider '%s': %w", c.ProviderSlug, err)
	}
	defer file.Close()

	err = gob.NewDecoder(file).Decode(c)
	if err != nil {
		return fmt.Errorf("Could not decode cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	return nil
}

// Write to the cache by encoding this cache struct. Adds or updates the TTL. Returns an error if writing fails.
func (c *cache) write() error {
	if c.Languages == nil {
		fmt.Println("Did not find any languages to write to the cache. Skipping cache write.")
		return nil
	}

	path, err := filePath(c.ProviderSlug)
	if err != nil {
		return fmt.Errorf("Could not write cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Path to cache directory doesn't exists. Creating directories: %s\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("Could not create cache directory for provider '%s': %w", c.ProviderSlug, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Could not create cache file for provider '%s': %w", c.ProviderSlug, err)
	}
	defer file.Close()

	if c.Ttl == nil {
		c.Ttl = new(time.Time)
	}
	if c.Ttl.Before(time.Now()) {
		*c.Ttl = time.Now()
	}
	c.Ttl.Add(time.Hour * 24 * 7)

	err = gob.NewEncoder(file).Encode(c)
	if err != nil {
		return fmt.Errorf("Could not encode cache file for provider '%s': %w", c.ProviderSlug, err)
	}

	return nil
}

// Invalidates caches for other providers. This function checks if the cache exists and is expired for each provider slug, and invalidates it if necessary. Takes a list of all provider slugs to invalidate.
func (c *cache) invalidateOtherCaches(providerSlugs []string) {
	for _, slug := range providerSlugs {
		if slug == c.ProviderSlug {
			continue
		}

		c := newCache(slug)

		if c.exists() {
			fmt.Printf("Invalidating cache for provider '%s'...\n", slug)

			err := c.read()
			if err != nil {
				fmt.Println(err)
			}

			if c.isExpired() {
				err := c.invalidate()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

// Returns the path to the cache file based on the user's cache directory and the provider slug.
func filePath(providerSlug string) (string, error) {
	path, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path + "/quick-translate/translations/" + providerSlug + ".gob", nil
}
