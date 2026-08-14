package transport

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Returns the providers that can be chosen together with the one that is in use.
func (a *Adapter) Providers() *ProviderDto {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.providers()
}

// Returns the languages of the current provider together with the ones that are chosen.
func (a *Adapter) Languages() *LanguageDto {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.languages()
}

// Returns the translation that is currently shown. Meant for the first render and for a frontend that was reloaded while a translation was on screen.
func (a *Adapter) Translation() *TranslationDto {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.translation()
}

// Switches to the given provider and translates the current text with it. Returns the whole state, because the languages that can be chosen belong to the provider and change with it. Returns an error if the provider can not be used, which leaves the previous one in place.
func (a *Adapter) ChangeProvider(slug string) (*FullDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.core.SetProvider(slug); err != nil {
		return nil, err
	}

	if err := a.retranslate(); err != nil {
		return nil, err
	}

	return a.full(), nil
}

// Sets the source language and translates the current text again. The target language is switched or dropped when it collides with the new source language, so it is part of the returned state. Returns an error if the language can not be chosen or the translation fails.
func (a *Adapter) ChangeSource(tag string) (*TranslationDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.core.SetSource(tag); err != nil {
		return nil, err
	}

	if err := a.retranslate(); err != nil {
		return nil, err
	}

	return a.translation(), nil
}

// Sets the target language and translates the current text again. The source language is switched or dropped when it collides with the new target language, so it is part of the returned state. Returns an error if the language can not be chosen or the translation fails.
func (a *Adapter) ChangeTarget(tag string) (*TranslationDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.core.SetTarget(tag); err != nil {
		return nil, err
	}

	if err := a.retranslate(); err != nil {
		return nil, err
	}

	return a.translation(), nil
}

// Switches the source and the target language and translates the current text in the other direction. Returns an error if the languages can not be switched, which leaves them unchanged.
func (a *Adapter) SwitchLanguages() (*TranslationDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.core.SwitchLanguages(); err != nil {
		return nil, err
	}

	if err := a.retranslate(); err != nil {
		return nil, err
	}

	return a.translation(), nil
}

// Steps to the translation that was stored before the current one. Nothing is translated again, the stored translation is shown as it was made. Returns the whole state, because the provider and the languages of the stored translation are taken over with it. Returns an error if the history is disabled or can not be read.
func (a *Adapter) PreviousTranslation() (*FullDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, _, _, err := a.core.PreviousTranslation(); err != nil {
		return nil, err
	}

	return a.full(), nil
}

// Steps to the translation that was stored after the current one. Nothing is translated again, the stored translation is shown as it was made. Returns the whole state, because the provider and the languages of the stored translation are taken over with it. Returns an error if the history is disabled or can not be read.
func (a *Adapter) NextTranslation() (*FullDto, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, _, _, err := a.core.NextTranslation(); err != nil {
		return nil, err
	}

	return a.full(), nil
}

// Writes the current translation to the clipboard and hides the window, because the translation is on its way into another application at that point. Returns an error if there is nothing to copy or the clipboard can not be written, which keeps the window open.
func (a *Adapter) CopyTranslation() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.core.CopyTranslation(); err != nil {
		return err
	}

	runtime.WindowHide(a.ctx)

	return nil
}
