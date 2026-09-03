package core

import (
	"slices"

	"quick-translate/internal/language"
	"quick-translate/internal/models"
)

// Returns the languages the current provider can translate from, in no promised order. The list does not contain language detection, because it is not a real language. Use LanguageDetection to find out whether it should be offered on top of this list.
func (c *Core) SourceLanguages() []*models.Language {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return slices.Clone(c.langs.SourceLanguages())
}

// Returns the languages the current provider can translate into, in no promised order.
func (c *Core) TargetLanguages() []*models.Language {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return slices.Clone(c.langs.TargetLanguages())
}

// Returns the virtual language that stands for language detection when the current provider detects the source language on its own, and nil when it does not. It is not part of the source languages, so a caller that offers it has to place it next to them.
func (c *Core) LanguageDetection() *models.Language {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.detection[c.providerSlug] {
		return nil
	}

	return language.DetectionLanguage()
}

// Returns the tag of the source language the user prefers, already resolved against the languages of the current provider. Returns an empty tag when no preference is configured or the preferred language is not available.
func (c *Core) PreferredSource() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.PreferredSource()
}

// Returns the tag of the target language the user prefers, already resolved against the languages of the current provider. Returns an empty tag when no preference is configured or the preferred language is not available.
func (c *Core) PreferredTarget() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.PreferredTarget()
}

// Returns the tag of the current source language. Returns the tag for language detection when the source language is detected by the provider, and an empty tag when no source language is set.
func (c *Core) Source() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.Source()
}

// Returns the tag of the current target language, or an empty tag when no target language is set.
func (c *Core) Target() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.Target()
}

// Returns the tag of the source language the provider detected for the last translation, or an empty tag when the source language was not detected.
func (c *Core) DetectedSource() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.DetectedSource()
}

// Sets the source language to the given tag, or to language detection when the tag for it is given. The target language is switched or dropped when it collides with the new source language. Returns an error if the language is not available as a source language of the current provider.
func (c *Core) SetSource(tag string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.SetSource(tag, c.langs.DetectedSource())
}

// Sets the target language to the given tag. The source language is switched or dropped when it collides with the new target language. Returns an error if the language is not available as a target language of the current provider.
func (c *Core) SetTarget(tag string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.SetTarget(tag)
}

// Switches the source and the target language. When the source language is detected by the provider, the detected language is switched with the target language. Returns an error if the languages can not be switched, which leaves them unchanged.
func (c *Core) SwitchLanguages() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.langs.SwitchLanguages()
}

// Builds the languages of the given provider and carries the given languages over as far as the provider supports them, including the language that was detected for the last translation. The preferences are copied for every build, because they are resolved against the languages of the provider and would otherwise be lost for every provider that follows. Requires the lock to be held.
func (c *Core) buildCollection(slug string, prov models.Provider, source string, target string, detected string) (*language.Collection, error) {
	preferences := *c.preferences

	langs, err := language.NewBuilder().
		SetSource(source, detected).
		SetTarget(target).
		SetLanguageDetection(c.detection[slug]).
		SetPreferences(&preferences).
		Build(slug, prov, c.knownSlugs())

	if err != nil {
		return nil, err
	}

	return langs, nil
}
