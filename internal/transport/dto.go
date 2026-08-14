package transport

import (
	"fmt"

	"quick-translate/internal/models"
)

type TranslationDto struct {
	Source, Target, DetectedSource string
	Text, Translation              string
	HasPrevious, HasNext           bool
}

type LanguageDto struct {
	Source, Target                   string
	Detection                        bool
	SourceLanguages, TargetLanguages []*models.Language
}

type ProviderDto struct {
	Current   string
	Providers []string
}

type FullDto struct {
	Provider    string
	Languages   *LanguageDto
	Translation *TranslationDto
}

// Bundles everything the translation view shows. The languages are taken from the core and not from the translation, because they are what the user has chosen and can differ from the translation while it is being made. Requires the lock to be held.
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

	dto.Text, dto.Translation = current.Text, current.Translation

	return dto
}

// Bundles the languages of the current provider together with the ones that are chosen. Requires the lock to be held.
func (a *Adapter) languages() *LanguageDto {
	return &LanguageDto{
		Source:          a.core.Source(),
		Target:          a.core.Target(),
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

// Bundles the whole state of the application, which is needed whenever the provider changes, because the languages that can be chosen change with it. Requires the lock to be held.
func (a *Adapter) full() *FullDto {
	return &FullDto{
		Provider:    a.core.CurrentProvider(),
		Languages:   a.languages(),
		Translation: a.translation(),
	}
}
