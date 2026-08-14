package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"quick-translate/internal/language"
	"quick-translate/internal/models"
)

// Returned when a translation is requested while nothing is selected or copied.
var ErrNothingSelected = errors.New("Nothing is selected. Please select the text you want to translate.")

// Returned when the current translation is needed but nothing has been translated yet.
var ErrNoTranslation = errors.New("There is no translation yet. Please translate a text first.")

// Returned when a translation is requested while no target language is set.
var ErrNoTargetLanguage = errors.New("No target language is set. Please choose the language you want to translate into.")

// Returned when a translation is requested while no source language is set and the provider can not detect it.
var ErrNoSourceLanguage = errors.New("No source language is set. Please choose the language you want to translate from.")

// Reads the selected text and translates it with the current provider and languages. The translation becomes the current one and is appended to the history. Returns ErrNothingSelected if nothing is selected and an error if the text can not be translated.
func (c *Core) TranslateFromClipboard() (*models.Translation, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	text, err := c.clipboard.Read()
	if err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil, ErrNothingSelected
	}

	return c.translate(text)
}

// Translates the text of the current translation again, which is meant for a translation whose languages or provider were changed afterwards. Returns ErrNoTranslation if nothing has been translated yet and an error if the text can not be translated.
func (c *Core) TranslateFromHistory() (*models.Translation, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.translation == nil {
		return nil, ErrNoTranslation
	}

	return c.translate(c.translation.Text)
}

// Writes the current translation to the clipboard, so it can be pasted. Returns ErrNoTranslation if nothing has been translated yet and an error if the clipboard can not be written.
func (c *Core) CopyTranslation() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.translation == nil {
		return ErrNoTranslation
	}

	return c.clipboard.Write(c.translation.Translation)
}

// Returns the translation that is currently shown, or nil if nothing has been translated yet.
func (c *Core) CurrentTranslation() *models.Translation {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.current()
}

// Translates the text with the current provider and languages, keeps it as the current translation and appends it to the history. The provider is not asked again when the text, the languages and the provider are unchanged, so pressing the shortcut twice does not pay for the same translation twice. Requires the lock to be held.
func (c *Core) translate(text string) (*models.Translation, error) {
	source, target := c.langs.Source(), c.langs.Target()

	if len(target) == 0 {
		return nil, ErrNoTargetLanguage
	}

	// A missing source language means the provider has to detect it, which not every provider can do.
	if source == language.LanguageDetectionTag || len(source) == 0 {
		if !c.detection[c.providerSlug] {
			return nil, ErrNoSourceLanguage
		}

		source = ""
	}

	if c.unchanged(text, source, target) {
		return c.current(), nil
	}

	translated, detected, err := c.provider.Translate(source, target, text)
	if err != nil {
		return nil, err
	}

	if len(source) == 0 {
		if err := c.langs.SetSource(language.LanguageDetectionTag, detected); err != nil {
			fmt.Printf("Could not keep the detected source language: %v\n", err)
		}
	}

	c.store(&models.Translation{
		CreatedAt:      time.Now().UTC(),
		Source:         c.langs.Source(),
		Target:         target,
		Text:           text,
		Translation:    translated,
		DetectedSource: c.langs.DetectedSource(),
		Provider:       c.providerSlug,
	})

	return c.current(), nil
}

// Checks if a translation of the text would return the same result as the current translation. The languages are compared in the form the collection has resolved them to, because a requested tag is matched against the languages of the provider and can differ from the stored one. Requires the lock to be held.
func (c *Core) unchanged(text string, source string, target string) bool {
	if c.translation == nil {
		return false
	}

	if len(source) == 0 {
		source = language.LanguageDetectionTag
	}

	return c.translation.Text == text &&
		c.translation.Source == source &&
		c.translation.Target == target &&
		c.translation.Provider == c.providerSlug
}

// Keeps the translation as the current one and appends it to the history. A translation that can not be stored is still kept, because it was already paid for. Requires the lock to be held.
func (c *Core) store(translation *models.Translation) {
	c.translation = translation

	if !c.history.Enabled() {
		return
	}

	id, err := c.history.Append(translation)
	if err != nil {
		fmt.Printf("Could not store the translation in the history: %v\n", err)
		return
	}

	translation.ID = id
}

// Returns a copy of the current translation, or nil if nothing has been translated yet. The copy keeps callers from changing the state of the core after it was handed out. Requires the lock to be held.
func (c *Core) current() *models.Translation {
	if c.translation == nil {
		return nil
	}

	copied := *c.translation

	return &copied
}
