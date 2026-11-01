package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"quick-translate/internal/models"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Emitted with State after a switch.
const EventChanged = "provider:changed"

func init() {
	application.RegisterEvent[State](EventChanged)
}

type State struct {
	Current   Info   `json:"current"`
	Providers []Info `json:"providers"` // Only configured providers, sorted by ID
}

// The part of the configuration the service reads.
type configuration interface {
	DefaultProvider() (models.ProviderID, error)
	Providers() ([]models.ProviderID, error)
	Provider(ID models.ProviderID) (decode func(into any) error, err error)
}

type Service struct {
	app    *application.App
	config configuration
	env    environment
	logger *slog.Logger

	mutex      sync.Mutex  // Guards the translator and the switch, never a request.
	translator *Translator // The current provider translations are made with. Nil if none is selected.
	err        error
	hooks      []func(*Fetcher) error
}

// Creates the service. No provider is built before ServiceStartup.
func New(app *application.App, config configuration, logger *slog.Logger) *Service {
	return &Service{
		app:    app,
		config: config,
		env:    newEnvironment(),
		logger: logger.With("category", "provider"),
		err:    errors.New("provider: not started"),
	}
}

// Selects the default provider. A failure is kept and answered by every getter until a provider is selected,
// because an unconfigured provider must not shut the application down.
func (s *Service) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ID, err := s.config.DefaultProvider()
	if err == nil {
		err = s.selectProvider(ID)
	}

	if err != nil {
		s.err = err
		s.logger.Error("select default provider", slog.Any("error", err))
	}

	return nil
}

// Returns every configured provider, sorted by ID.
func (s *Service) Providers() ([]Info, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.providers()
}

// Returns the provider translations are made with.
func (s *Service) Current() (Info, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.translator == nil {
		return Info{}, s.err
	}

	return s.translator.info, nil
}

// Switches to the provider with the given ID. Nothing changes if it cannot be built or the hook fails.
func (s *Service) Select(ID models.ProviderID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.selectProvider(ID)
}

// Returns the translator of the current provider. Waits while a switch is running.
//
//wails:ignore
func (s *Service) Translator() (*Translator, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.translator == nil {
		return nil, s.err
	}

	return s.translator, nil
}

// Builds a fetcher for any configured provider from its settings. Takes no lock, so fetchers can be built and
// used concurrently, also during a switch.
//
//wails:ignore
func (s *Service) Fetcher(ID models.ProviderID) (*Fetcher, error) {
	info, impl, err := s.build(ID)
	if err != nil {
		return nil, err
	}

	return &Fetcher{info: info, impl: impl}, nil
}

// Adds a hook a switch calls with the new provider before it commits. Hooks run in the order they were added,
// and the first error cancels the switch; the hooks before it have already run. A hook must not call back into
// the service.
//
//wails:ignore
func (s *Service) OnChange(hook func(*Fetcher) error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.hooks = append(s.hooks, hook)
}

// Builds the provider and lists the configured ones, then runs the hooks and commits only if all succeed.
// Requires the lock to be held.
func (s *Service) selectProvider(ID models.ProviderID) error {
	if s.translator != nil && s.translator.info.ID == ID {
		return nil
	}

	info, impl, err := s.build(ID)
	if err != nil {
		return err
	}

	providers, err := s.providers()
	if err != nil {
		return err
	}

	for _, hook := range s.hooks {
		if err := hook(&Fetcher{info: info, impl: impl}); err != nil {
			return fmt.Errorf("provider: switch to %s: %w", ID, err)
		}
	}

	s.translator, s.err = &Translator{info: info, impl: impl}, nil
	s.logger.Info("provider selected", slog.Any("provider", ID))

	s.app.Event.Emit(EventChanged, State{Current: info, Providers: providers})

	return nil
}

// Returns every configured provider, sorted by ID. Requires the lock to be held.
func (s *Service) providers() ([]Info, error) {
	IDs, err := s.config.Providers()
	if err != nil {
		return nil, err
	}

	// Sorted here, because the configuration does not promise an order.
	IDs = slices.Sorted(slices.Values(IDs))

	providers := make([]Info, 0, len(IDs))
	for _, ID := range IDs {
		info, ok := describe(ID)
		if !ok {
			return nil, fmt.Errorf("provider: id %d not implemented", ID)
		}

		providers = append(providers, info)
	}

	return providers, nil
}

// Builds the provider from its configured settings. Fails if it is not configured.
func (s *Service) build(ID models.ProviderID) (Info, implementation, error) {
	decode, err := s.config.Provider(ID)
	if err != nil {
		return Info{}, nil, err
	}

	return build(ID, decode, s.env)
}
