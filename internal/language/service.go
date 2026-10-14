package language

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"quick-translate/internal/config"
	"quick-translate/internal/models"
	"quick-translate/internal/provider"
	"quick-translate/internal/store"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	EventOptions   = "languages:changed" // Emitted with Options when the catalog was rebuilt
	EventSelection = "selection:changed" // Emitted with Selection after every change of it
	languageTTL    = 7 * 24 * time.Hour  // How long stored languages are used before they are refreshed.
)

func init() {
	application.RegisterEvent[Options](EventOptions)
	application.RegisterEvent[Selection](EventSelection)
}

type Options struct {
	Detection       bool              `json:"detection"`       // Depends on the provider
	PreferredSource string            `json:"preferredSource"` // Empty if none is configured or available
	PreferredTarget string            `json:"preferredTarget"`
	SourceLanguages []models.Language `json:"sourceLanguages"` // Preferred first, the rest sorted by tag
	TargetLanguages []models.Language `json:"targetLanguages"`
}

type Snapshot struct {
	Provider models.ProviderID
	Selection
}

// The part of the configuration the service reads.
type configuration interface {
	Preferences() config.LanguagePreferences
	Providers() ([]models.ProviderID, error)
}

// The part of the provider service the service uses.
type providers interface {
	Current() (provider.Info, error)
	Fetcher(ID models.ProviderID) (*provider.Fetcher, error)
	OnChange(hook func(*provider.Fetcher) error)
}

// The part of the store the service uses.
type storage interface {
	NewLanguage() *store.LanguageRepository
}

// The part of the environment the service uses.
type environment interface {
	Locale() (string, error)
}

type Service struct {
	app       *application.App
	config    configuration
	store     storage
	providers providers
	logger    *slog.Logger
	locale    string // Empty if the system has none

	mutex     sync.Mutex
	catalog   *catalog // Nil until a provider is selected
	selection Selection
	refresher *refresher // Nil before ServiceStartup
}

// Creates the service and registers its hook with the provider service, which calls it in its own startup.
func New(app *application.App, config configuration, store storage, providers providers, env environment, logger *slog.Logger) *Service {
	logger = logger.With("category", "language")

	locale, err := env.Locale()
	if err != nil {
		logger.Warn("no locale, the target language falls back without it", slog.Any("error", err))
	}

	s := &Service{
		app:       app,
		config:    config,
		store:     store,
		providers: providers,
		logger:    logger,
		locale:    locale,
	}

	providers.OnChange(s.change)

	return s
}

// Starts the background refresh. Never fails; without a store nothing is refreshed.
func (s *Service) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	repo := s.store.NewLanguage()
	if repo == nil {
		s.logger.Error("store not started, languages are not refreshed")
		return nil
	}

	r := newRefresher(languageTTL, repo, s.config.Providers, s.fetch, s.refreshed, s.logger.With("step", "refresh"))

	s.mutex.Lock()
	s.refresher = r
	s.mutex.Unlock()

	r.Start(ctx)

	return nil
}

// Waits for the refresh to end before the store closes. Wails has cancelled the application context already,
// which cancels the running requests.
func (s *Service) ServiceShutdown() error {
	s.mutex.Lock()
	r := s.refresher
	s.mutex.Unlock()

	if r != nil {
		r.Wait()
	}

	return nil
}

// Returns what the current provider offers.
func (s *Service) Options() (Options, error) {
	c, _ := s.state()
	if c == nil {
		return Options{}, s.notReady()
	}

	return optionsOf(c), nil
}

// Returns the current selection.
func (s *Service) Selection() (Selection, error) {
	c, selection := s.state()
	if c == nil {
		return Selection{}, s.notReady()
	}

	return selection, nil
}

// Sets the source language. A colliding target is swapped or dropped.
func (s *Service) SetSource(tag string) error {
	return s.update(func(c *catalog, selection Selection) (Selection, error) {
		return setSource(c, selection, tag)
	})
}

// Lets the provider detect the source language. Fails if it cannot.
func (s *Service) Detect() error {
	return s.update(func(c *catalog, selection Selection) (Selection, error) {
		return detect(c, selection, selection.Detected)
	})
}

// Sets the target language. A colliding source is swapped, replaced by detection or dropped.
func (s *Service) SetTarget(tag string) error {
	return s.update(func(c *catalog, selection Selection) (Selection, error) {
		return setTarget(c, selection, tag)
	})
}

// Swaps source and target; while detecting, the detected language becomes the target.
func (s *Service) Switch() error {
	return s.update(switchLanguages)
}

// Returns the selection with the provider it belongs to.
//
//wails:ignore
func (s *Service) Snapshot() (Snapshot, error) {
	c, selection := s.state()
	if c == nil {
		return Snapshot{}, s.notReady()
	}

	return Snapshot{Provider: c.provider, Selection: selection}, nil
}

// Takes the languages of a stored translation over, as far as the provider still offers them. Fails only if the
// translation belongs to another provider. Select the provider first.
//
//wails:ignore
func (s *Service) Adopt(entry models.Translation) error {
	return s.update(func(c *catalog, selection Selection) (Selection, error) {
		if entry.ProviderID != c.provider {
			return selection, fmt.Errorf("language: translation made with %s, not %s", entry.ProviderID, c.provider)
		}

		next := selection
		var err error

		// The source comes first, because setting the target may correct it.
		switch {
		case entry.Detection:
			var detected string
			if entry.SourceTag != nil {
				detected = *entry.SourceTag
			}

			next, err = detect(c, next, detected)
		case entry.SourceTag != nil:
			next, err = setSource(c, next, *entry.SourceTag)
		}

		if err != nil {
			s.logger.Warn("adopt source language", slog.Any("error", err))
		}

		if len(entry.TargetTag) != 0 {
			if next, err = setTarget(c, next, entry.TargetTag); err != nil {
				s.logger.Warn("adopt target language", slog.Any("error", err))
			}
		}

		return next, nil
	})
}

// Records the language the provider detected. Ignored if the provider changed meanwhile or detection is off; a
// tag that is no source language is dropped.
//
//wails:ignore
func (s *Service) SetDetected(ID models.ProviderID, tag string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.catalog == nil || s.catalog.provider != ID || !s.selection.Detection {
		return
	}

	next := s.selection
	next.Detected = s.catalog.matchDetected(tag)

	s.choose(next)
}

// The provider hook: builds the catalog of the new provider and carries the selection over. Fetches only if
// nothing is stored yet, so a switch to a known provider works offline.
func (s *Service) change(fetcher *provider.Fetcher) error {
	languages, err := s.load(fetcher)
	if err != nil {
		return err
	}

	c, err := newCatalog(fetcher.Info(), languages, s.config.Preferences(), s.logger)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.commit(c)

	// Before startup there is no refresher; its first pass covers this provider.
	if s.refresher != nil {
		s.refresher.Trigger()
	}

	return nil
}

// Reads the stored languages of the provider, or fetches and stores them if there are none.
func (s *Service) load(fetcher *provider.Fetcher) ([]models.Language, error) {
	repo := s.store.NewLanguage()
	if repo == nil {
		return nil, errors.New("language: store not started")
	}

	ctx := s.app.Context() // Exists before startup, which the hook runs in
	ID := fetcher.Info().ID

	languages, err := repo.Get(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("language: %w", err)
	}

	if len(languages) != 0 {
		return languages, nil
	}

	if languages, err = fetcher.Languages(ctx); err != nil {
		return nil, err
	}

	// Not fatal, the next refresh tries again.
	if err := repo.Put(ctx, ID, languages); err != nil {
		s.logger.Warn("store fetched languages", slog.Any("provider", ID), slog.Any("error", err))
	}

	return languages, nil
}

// Fetches the languages of any configured provider for the refresher.
func (s *Service) fetch(ctx context.Context, ID models.ProviderID) ([]models.Language, error) {
	fetcher, err := s.providers.Fetcher(ID)
	if err != nil {
		return nil, err
	}

	return fetcher.Languages(ctx)
}

// Rebuilds the catalog when the refresher stored new languages of the current provider.
func (s *Service) refreshed(ID models.ProviderID, languages []models.Language) {
	current, _ := s.state()
	if current == nil || current.provider != ID {
		return
	}

	c, err := newCatalog(provider.Info{ID: ID, Detection: current.detection}, languages, s.config.Preferences(), s.logger)
	if err != nil {
		s.logger.Warn("rebuild refreshed languages", slog.Any("error", err))
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// A switch may have happened while the catalog was built.
	if s.catalog == nil || s.catalog.provider != ID {
		return
	}

	s.commit(c)
}

// Swaps the catalog, carries the selection over and emits both events. Requires the lock to be held.
func (s *Service) commit(c *catalog) {
	s.catalog = c
	s.selection = initial(c, s.selection, s.locale)

	s.app.Event.Emit(EventOptions, optionsOf(c))
	s.app.Event.Emit(EventSelection, s.selection)
}

// Sets the selection and emits it if it changed. Requires the lock to be held.
func (s *Service) choose(next Selection) {
	if next == s.selection {
		return
	}

	s.selection = next
	s.app.Event.Emit(EventSelection, next)
}

// Applies a rule to the selection. Nothing changes if the rule fails.
func (s *Service) update(rule func(*catalog, Selection) (Selection, error)) error {
	ready, err := s.apply(rule)
	if !ready {
		return s.notReady()
	}

	return err
}

// Applies the rule under the lock; false if there is no catalog yet.
func (s *Service) apply(rule func(*catalog, Selection) (Selection, error)) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.catalog == nil {
		return false, nil
	}

	next, err := rule(s.catalog, s.selection)
	if err != nil {
		return true, err
	}

	s.choose(next)

	return true, nil
}

// Returns the catalog and the selection. The catalog is immutable, so both can be used without the lock.
func (s *Service) state() (*catalog, Selection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.catalog, s.selection
}

// Explains why there is no catalog yet. Must be called without the lock, as it takes the provider lock.
func (s *Service) notReady() error {
	if _, err := s.providers.Current(); err != nil {
		return err
	}

	return errors.New("language: no languages loaded")
}

// Returns the options of the catalog with copies of its lists.
func optionsOf(c *catalog) Options {
	return Options{
		Detection:       c.detection,
		PreferredSource: c.preferredSource,
		PreferredTarget: c.preferredTarget,
		SourceLanguages: slices.Clone(c.sources),
		TargetLanguages: slices.Clone(c.targets),
	}
}
