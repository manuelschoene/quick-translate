package core

import (
	"fmt"
	"maps"
	"slices"
	"sync"

	"quick-translate/internal/clipboard"
	"quick-translate/internal/config"
	"quick-translate/internal/history"
	"quick-translate/internal/language"
	"quick-translate/internal/models"
	"quick-translate/internal/provider"
)

type clip interface {
	Read() (string, error)
	Write(text string) error
}

type Core struct {
	clipboard    clip
	factory      func(slug string) (models.Provider, error)
	history      *history.History
	langs        *language.Collection
	preferences  *models.LanguagePreferences
	detection    map[string]bool
	providerList []string
	providerSlug string
	provider     models.Provider
	translation  *models.Translation
	mutex        sync.Mutex
}

// Creates a new core and initializes every service the application is built on. The providers the application knows about are matched against the ones configured in the configuration file, and the configuration is then used to set up the clipboard, the history and the languages of the default provider. Returns an error for every problem that stops the application from running, so a core that is returned is ready to translate.
func NewCore() (*Core, error) {
	detection := provider.All()
	known := slices.Sorted(maps.Keys(detection))

	slug, err := config.DefaultProvider()
	if err != nil {
		return nil, err
	}

	list, err := config.ListProviders(known)
	if err != nil {
		return nil, err
	}
	slices.Sort(list)

	if !slices.Contains(list, slug) {
		return nil, fmt.Errorf("The default provider '%s' is not configured. Please add it to the 'provider' section of the configuration file or choose one of: %v", slug, list)
	}

	preferences, err := config.PreferredLanguages()
	if err != nil {
		return nil, err
	}

	limit, err := config.HistoryLimit()
	if err != nil {
		return nil, err
	}

	board, err := clipboard.NewClipboard()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Using the clipboard backend: %s\n", board.Backend())

	past := history.NewHistory(limit)
	if past.Enabled() {
		// Opens the database right away, so a broken history is reported on startup and not with the first translation.
		if _, err := past.HasPrevious(0); err != nil {
			return nil, err
		}
	}

	core := &Core{
		clipboard:    board,
		factory:      buildProvider,
		history:      past,
		preferences:  preferences,
		detection:    detection,
		providerList: list,
	}

	prov, err := buildProvider(slug)
	if err != nil {
		past.Close()
		return nil, err
	}

	langs, err := core.buildCollection(slug, prov, "", "", "")
	if err != nil {
		past.Close()
		return nil, err
	}

	core.providerSlug, core.provider, core.langs = slug, prov, langs

	return core, nil
}

// Releases the resources the core holds, which closes the connection to the history database. Safe to call more than once.
func (c *Core) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.history.Close()
}

// Returns every provider slug the application knows about, no matter if it is configured. Requires the lock to be held.
func (c *Core) knownSlugs() []string {
	return slices.Sorted(maps.Keys(c.detection))
}
