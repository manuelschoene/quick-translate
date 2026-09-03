package transport

import (
	"fmt"

	"quick-translate/internal/models"
)

// The languages the user has chosen together with the translation that came out of them. The three tags point into the lists of the LanguageDto, so the frontend resolves them there instead of receiving their names twice. An empty tag means the language is not set, and a source tag of "auto" means the provider detects it, in which case DetectedSource carries what it found for the last translation.
type TranslationDto struct {
	Source, Target, DetectedSource string
	Translation                    string
	HasPrevious, HasNext           bool
}

// What the current provider offers. Everything in here changes only with the provider, which is why the chosen languages are not part of it. The two lists come in no promised order, so the frontend sorts them the way it wants to show them. Detection is nil for a provider that does not detect the source language on its own, and the two preferred tags are the ones the user configured, resolved against the lists so they can be pinned.
type LanguageDto struct {
	PreferredSource, PreferredTarget string
	Detection                        *models.Language
	SourceLanguages, TargetLanguages []*models.Language
}

// The providers that can be chosen. Current is the slug of the provider in use, and Providers holds the slugs of every provider that is configured well enough to be offered, the current one included.
type ProviderDto struct {
	Current   string
	Providers []string
}

// The whole state of the application. It is what every call answers with that can change more than the chosen languages, because the languages that can be chosen belong to the provider and change with it.
type FullDto struct {
	Provider    *ProviderDto
	Languages   *LanguageDto
	Translation *TranslationDto
}

// Bundles the chosen languages with the translation that is shown. The languages are taken from the core and not from the translation, because they are what the user has chosen and can differ from the translation while it is being made. Requires the lock to be held.
func (a *Adapter) translation() *TranslationDto {
	hasPrevious, hasNext, err := a.core.Navigation()
	if err != nil {
		// A history that can not be read only costs the stepping, so the translation itself is still shown.
		fmt.Printf("Could not look for stored translations: %v\n", err)
	}

	dto := &TranslationDto{
		Source:         a.core.Source(),
		Target:         a.core.Target(),
		DetectedSource: a.core.DetectedSource(),
		HasPrevious:    hasPrevious,
		HasNext:        hasNext,
	}

	current := a.core.CurrentTranslation()
	if current == nil {
		return dto
	}

	dto.Translation = current.Translation

	return dto
}

// Bundles the languages of the current provider together with the preferences and the language detection they are offered with. Requires the lock to be held.
func (a *Adapter) languages() *LanguageDto {
	return &LanguageDto{
		PreferredSource: a.core.PreferredSource(),
		PreferredTarget: a.core.PreferredTarget(),
		Detection:       a.core.LanguageDetection(),
		SourceLanguages: a.core.SourceLanguages(),
		TargetLanguages: a.core.TargetLanguages(),
	}
}

// Bundles the configured providers together with the one that is in use. Requires the lock to be held.
func (a *Adapter) providers() *ProviderDto {
	return &ProviderDto{
		Current:   a.core.CurrentProvider(),
		Providers: a.core.Providers(),
	}
}

// Bundles the whole state of the application, which is needed for the first render and whenever the provider changes, because the languages that can be chosen change with it. Requires the lock to be held.
func (a *Adapter) full() *FullDto {
	return &FullDto{
		Provider:    a.providers(),
		Languages:   a.languages(),
		Translation: a.translation(),
	}
}
