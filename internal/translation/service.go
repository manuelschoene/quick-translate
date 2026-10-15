package translation

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"quick-translate/internal/language"
	"quick-translate/internal/models"
	"quick-translate/internal/provider"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	EventStarted = "translation:started" // Emitted without data when a translation begins
	EventChanged = "translation:changed" // Emitted with the translation that is shown now
	EventFailed  = "translation:failed"  // Emitted with the message when a started translation fails
)

func init() {
	application.RegisterEvent[application.Void](EventStarted)
	application.RegisterEvent[models.Translation](EventChanged)
	application.RegisterEvent[string](EventFailed)
}

var (
	ErrNothingSelected  = errors.New("translation: nothing selected")
	ErrNoTranslation    = errors.New("translation: nothing translated yet")
	ErrNoTargetLanguage = errors.New("translation: no target language")
	ErrNoSourceLanguage = errors.New("translation: no source language")
)

// The part of the provider service the service uses.
type providers interface {
	Translator() (*provider.Translator, error)
}

// The part of the language service the service uses.
type languages interface {
	Snapshot() (language.Snapshot, error)
	SetDetected(ID models.ProviderID, tag string)
}

// The part of the clipboard service the service uses.
type clipboard interface {
	Read(ctx context.Context) (string, error)
	Write(ctx context.Context, text string) error
}

type Service struct {
	app       *application.App
	providers providers
	languages languages
	clipboard clipboard
	logger    *slog.Logger

	work sync.Mutex // Serializes translations, from started to changed or failed. Taken before mutex.

	mutex   sync.Mutex // Guards the fields below, never a request.
	current *models.Translation
	hooks   []func(models.Translation)
}

// Creates the service.
func New(app *application.App, providers providers, languages languages, clipboard clipboard, logger *slog.Logger) *Service {
	return &Service{
		app:       app,
		providers: providers,
		languages: languages,
		clipboard: clipboard,
		logger:    logger.With("category", "translation"),
	}
}

// Returns the translation that is shown, nil before the first one.
func (s *Service) Current() *models.Translation {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return nil
	}

	current := *s.current

	return &current
}

// Translates the current text again with the current provider and languages.
func (s *Service) Retranslate(ctx context.Context) error {
	s.work.Lock()
	defer s.work.Unlock()

	current := s.Current()
	if current == nil {
		return ErrNoTranslation
	}

	return s.run(ctx, func(context.Context) (string, error) { return current.Text, nil })
}

// Writes the current translation to the regular clipboard.
func (s *Service) Copy(ctx context.Context) error {
	current := s.Current()
	if current == nil {
		return ErrNoTranslation
	}

	return s.clipboard.Write(ctx, current.Translation)
}

// Translates the selected text. Failures only reach the frontend as an event.
//
//wails:ignore
func (s *Service) TranslateSelection() {
	s.work.Lock()
	defer s.work.Unlock()

	_ = s.run(s.app.Context(), s.selection)
}

// Shows a stored translation without calling the hooks.
//
//wails:ignore
func (s *Service) Show(entry models.Translation) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.show(entry)
}

// Adds a hook every new translation is reported to, in the order they were added. A hook must not call back
// into the service except through Show.
//
//wails:ignore
func (s *Service) OnTranslated(hook func(models.Translation)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.hooks = append(s.hooks, hook)
}

// Emits started, translates the text and ends with exactly one changed or failed. Requires work to be held.
func (s *Service) run(ctx context.Context, text func(context.Context) (string, error)) error {
	s.app.Event.Emit(EventStarted)

	t, fresh, err := s.translate(ctx, text)
	if err != nil {
		s.logger.Warn("translate", slog.Any("error", err))
		s.app.Event.Emit(EventFailed, err.Error())

		return err
	}

	s.mutex.Lock()
	s.show(t)
	hooks := slices.Clone(s.hooks)
	s.mutex.Unlock()

	// After changed, so the history reports after the translation.
	if fresh {
		for _, hook := range hooks {
			hook(t)
		}
	}

	return nil
}

// Translates the text; false if the current translation already answers it and no request was made. Requires
// work to be held.
func (s *Service) translate(ctx context.Context, text func(context.Context) (string, error)) (models.Translation, bool, error) {
	input, err := text(ctx)
	if err != nil {
		return models.Translation{}, false, err
	}

	translator, snapshot, err := s.state()
	if err != nil {
		return models.Translation{}, false, err
	}

	request, err := requestOf(input, snapshot.Selection)
	if err != nil {
		return models.Translation{}, false, err
	}

	ID := translator.Info().ID

	if current := s.Current(); current != nil && answers(*current, ID, request, snapshot.Selection) {
		return *current, false, nil
	}

	result, err := translator.Translate(ctx, request)
	if err != nil {
		return models.Translation{}, false, err
	}

	t := models.Translation{
		Text:        request.Text,
		Translation: result.Text,
		CreatedAt:   time.Now().UTC(),
		ProviderID:  ID,
		TargetTag:   request.Target,
		Detection:   snapshot.Detection,
	}

	if snapshot.Detection {
		t.SourceTag = s.detected(ID, result.Detected)
	} else {
		t.SourceTag = &request.Source
	}

	return t, true, nil
}

// Returns the translator and the language snapshot of the same provider. Reads both again once if a switch ran
// in between; the second read waits for it to finish.
func (s *Service) state() (*provider.Translator, language.Snapshot, error) {
	for range 2 {
		translator, err := s.providers.Translator()
		if err != nil {
			return nil, language.Snapshot{}, err
		}

		snapshot, err := s.languages.Snapshot()
		if err != nil {
			return nil, language.Snapshot{}, err
		}

		if snapshot.Provider == translator.Info().ID {
			return translator, snapshot, nil
		}
	}

	return nil, language.Snapshot{}, errors.New("translation: provider switched during the translation")
}

// Records the detected language and returns it as the language service resolved it; nil if it is unknown or the
// selection changed meanwhile.
func (s *Service) detected(ID models.ProviderID, tag string) *string {
	s.languages.SetDetected(ID, tag)

	snapshot, err := s.languages.Snapshot()
	if err != nil || snapshot.Provider != ID || !snapshot.Detection || len(snapshot.Detected) == 0 {
		return nil
	}

	return &snapshot.Detected
}

// Reads the selected text, trimmed.
func (s *Service) selection(ctx context.Context) (string, error) {
	text, err := s.clipboard.Read(ctx)
	if err != nil {
		return "", err
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return "", ErrNothingSelected
	}

	return text, nil
}

// Sets the translation as the current one and emits it. Requires the lock to be held.
func (s *Service) show(t models.Translation) {
	s.current = &t
	s.app.Event.Emit(EventChanged, t)
}

// Builds the request for the selection. Fails without a target, or without a source while not detecting.
func requestOf(text string, selection language.Selection) (provider.Request, error) {
	if len(selection.Target) == 0 {
		return provider.Request{}, ErrNoTargetLanguage
	}

	if !selection.Detection && len(selection.Source) == 0 {
		return provider.Request{}, ErrNoSourceLanguage
	}

	request := provider.Request{Text: text, Source: selection.Source, Target: selection.Target}
	if selection.Detection {
		request.Source = "" // Lets the provider detect it
	}

	return request, nil
}

// Reports whether the translation already answers the request, so the same text is not paid for twice.
func answers(t models.Translation, ID models.ProviderID, request provider.Request, selection language.Selection) bool {
	if t.ProviderID != ID || t.Text != request.Text || t.TargetTag != request.Target || t.Detection != selection.Detection {
		return false
	}

	return selection.Detection || (t.SourceTag != nil && *t.SourceTag == request.Source)
}
