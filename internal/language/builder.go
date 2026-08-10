package language

import (
	"fmt"
	"quick-translate/internal/models"
)

type CollectionBuilder struct {
	languageDetection bool
	preferences       *models.LanguagePreferences
	source, target    string
}

// Creates a new CollectionBuilder instance for building a Collection.
func NewBuilder() *CollectionBuilder {
	return &CollectionBuilder{}
}

// Sets the source language tag for the Collection.
func (b *CollectionBuilder) SetSource(source string) *CollectionBuilder {
	b.source = source
	return b
}

// Sets the target language tag for the Collection.
func (b *CollectionBuilder) SetTarget(target string) *CollectionBuilder {
	b.target = target
	return b
}

// Sets whether language detection of the provider is enabled for the Collection.
func (b *CollectionBuilder) SetLanguageDetection(languageDetection bool) *CollectionBuilder {
	b.languageDetection = languageDetection
	return b
}

// Sets the language preferences for the Collection.
func (b *CollectionBuilder) SetPreferences(preferences *models.LanguagePreferences) *CollectionBuilder {
	b.preferences = preferences
	return b
}

// Builds and returns a new Collection instance based on the provided parameters. Requires the current provider and a list of provider slugs for cache invalidation.
func (b *CollectionBuilder) Build(providerSlug string, provider models.Provider, providerSlugs []string) (*Collection, error) {
	if providerSlug == "" || provider == nil {
		return nil, fmt.Errorf("Provider slug and provider must be set")
	}

	langs, err := retrieveLanguages(providerSlug, provider, providerSlugs)
	if err != nil {
		return nil, err
	}

	langs, err = validateAndSortLanguages(langs)
	if err != nil {
		return nil, err
	}

	pref := parsePreferences(b.preferences, langs)
	sourceLangs, targetLangs := prioritySplit(pref, langs)
	source, target := resolveInitialLanguages(b.source, b.target, b.languageDetection, pref, sourceLangs, targetLangs)

	return &Collection{
		languageDetection: b.languageDetection,
		provider:          provider,
		preferences:       pref,
		providerSlug:      providerSlug,
		source:            source,
		target:            target,
		sourceLangs:       sourceLangs,
		targetLangs:       targetLangs,
	}, nil
}

// Retrieves languages for the current provider. Tries to read from the cache first. If the cache does not exist or is expired, it fetches languages from the provider and updates the cache. This function also invalidates caches for other providers in a separate goroutine.
func retrieveLanguages(providerSlug string, provider models.Provider, providerSlugs []string) ([]*models.Language, error) {
	cache := newCache(providerSlug)
	go cache.invalidateOtherCaches(providerSlugs)

	if cache.exists() {
		err := cache.read()
		if err != nil {
			return nil, err
		}

		if cache.isExpired() {
			go fetchProviderLanguages(cache, provider)
		}

		return cache.Languages, nil
	}

	langs, err := fetchProviderLanguages(cache, provider)
	if err != nil {
		return nil, err
	}

	return langs, nil
}

// Fetches languages from the provider and writes it to the cache. Returns the languages and any fetching errors.
func fetchProviderLanguages(cache *cache, provider models.Provider) ([]*models.Language, error) {
	langs, err := provider.Languages()
	if err != nil {
		return nil, err
	}

	cache.Languages = langs

	err = cache.write()
	if err != nil {
		return nil, err
	}

	return langs, nil
}
