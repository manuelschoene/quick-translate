package language

import (
	"fmt"
	"quick-translate/internal/models"

	"golang.org/x/text/language"
)

// The tag to be used for language detection. Can be returned by the Source() method and used as a parameter for the SetSource() method. It is not a real language and will not be included in the list of available languages.
const LanguageDetectionTag string = "auto"

type Collection struct {
	languageDetection                            bool
	provider                                     models.Provider
	preferences                                  *models.LanguagePreferences
	providerSlug, source, target, detectedSource string
	sourceLangs, targetLangs                     []*models.Language
}

// Returns the current source language tag. If language detection is used, returns the LanguageDetectionTag. If no source language is set, returns an empty string.
func (c *Collection) Source() string {
	return c.source
}

// Returns the current target language tag. If no target language is set, returns an empty string.
func (c *Collection) Target() string {
	return c.target
}

// Returns the detected source language tag. If no detected source language is set, returns an empty string.
func (c *Collection) DetectedSource() string {
	return c.detectedSource
}

// Returns the list of available languages for the Collection. The list does not include the LanguageDetectionTag, as it is virtually computed and not a real language. The list is sorted in ascending order by tag with the preferred language first.
func (c *Collection) SourceLanguages() []*models.Language {
	return c.sourceLangs
}

// Returns the list of available target languages for the Collection. The list is sorted in ascending order by tag with the preferred language first.
func (c *Collection) TargetLanguages() []*models.Language {
	return c.targetLangs
}

// Switches the source and target languages. When language detection is used, the detected source language will be switched with the target language. If the languages are not available for switching, an error will be returned.
func (c *Collection) SwitchLanguages() error {
	if c.source == "" && c.target == "" {
		return nil
	}

	if c.source == LanguageDetectionTag {
		return c.switchLanguageDetection()
	}

	return c.switchTags()
}

// Sets the source language tag for the Collection. If the source language is set to the LanguageDetectionTag, language detection for the provider will be used. If the target language is the same as the source language, switching the languages will be attempted. If the language is not available as a source language, an error will be returned. If language detection is requested, the detected source language tag can be provided. Otherwise the second parameter will be ignored.
func (c *Collection) SetSource(source string, detectedSource string) error {
	if len(source) == 0 {
		return fmt.Errorf("Source language cannot be empty")
	}

	if source == LanguageDetectionTag {
		return c.setLanguageDetection(detectedSource)
	}

	return c.setTag(source)
}

// Sets the target language tag for the Collection. If the target language is the same as the source language, switching the languages will be attempted. If the language is not available as a target language, an error will be returned. Language detection is not allowed for the target language.
func (c *Collection) SetTarget(target string) error {
	if len(target) == 0 || target == LanguageDetectionTag {
		return fmt.Errorf("Target language cannot be set to '%s' or be empty", LanguageDetectionTag)
	}

	matchTarget := matchLanguage(target, c.targetLangs)
	if matchTarget == nil {
		return fmt.Errorf("Language '%s' is not available as a target language", target)
	}

	if c.source == LanguageDetectionTag || len(c.source) == 0 {
		c.target = matchTarget.Tag
		c.detectedSource = ""
		return nil
	}

	matchSource := matchLanguage(c.source, c.targetLangs)

	if matchSource != nil && matchTarget.Tag == matchSource.Tag {
		matchPrevTarget := matchLanguage(c.target, c.sourceLangs)
		if matchPrevTarget != nil {
			c.source, c.target = matchPrevTarget.Tag, matchTarget.Tag
			return nil
		}

		if c.languageDetection {
			c.source = LanguageDetectionTag
			c.target = matchTarget.Tag
			return nil
		}

		c.source = ""
	}

	c.target = matchTarget.Tag
	return nil
}

// Switches the languages when language detection is currently used. If the target language is set, it will be switched with the detected source language if available. Otherwise, the target language will be empty afterwards. It is required, that the target language is set, otherwise an error will be returned. The detected source language will be cleared after switching.
func (c *Collection) switchLanguageDetection() error {
	if len(c.target) == 0 {
		return fmt.Errorf("Languages cannot be switched when source is set to '%s' and target is not specified.", LanguageDetectionTag)
	}

	matchTarget := matchLanguage(c.target, c.sourceLangs)
	if matchTarget == nil {
		return fmt.Errorf("Target language '%s' is not available as a source language", c.target)
	}

	var matchSource *models.Language
	if len(c.detectedSource) > 0 {
		if matchDetectedSource := matchLanguage(c.detectedSource, c.targetLangs); matchDetectedSource != nil {
			matchSource = matchDetectedSource
		}
	}

	if matchSource == nil || c.detectedSource == matchTarget.Tag {
		c.source, c.target = matchTarget.Tag, ""
		c.detectedSource = ""
		return nil
	}

	c.source, c.target = matchTarget.Tag, matchSource.Tag
	c.detectedSource = ""
	return nil
}

// Switches the languages when no language detection is used. If both languages are set, they will be switched. If only one language is set, the other language will be set empty afterwards. If the languages are not available for switching, an error will be returned.
func (c *Collection) switchTags() error {
	if len(c.source) == 0 {
		matchTarget := matchLanguage(c.target, c.sourceLangs)
		if matchTarget == nil {
			return fmt.Errorf("Target language '%s' is not available as a source language", c.target)
		}

		c.source, c.target = matchTarget.Tag, c.source
		return nil
	}

	matchSource := matchLanguage(c.source, c.targetLangs)
	if matchSource == nil {
		return fmt.Errorf("Source language '%s' is not available as a target language", c.source)
	}

	if len(c.target) == 0 {
		c.source, c.target = c.target, matchSource.Tag
		return nil
	}

	matchTarget := matchLanguage(c.target, c.sourceLangs)
	if matchTarget == nil {
		return fmt.Errorf("Target language '%s' is not available as a source language", c.target)
	}

	c.source, c.target = matchTarget.Tag, matchSource.Tag
	return nil
}

// Sets the source language to language detection. If language detection is not available for the provider, an error will be returned. If the detected source language is provided, it will be validated and set if available. If the detected source language is not available, it will be ignored.
func (c *Collection) setLanguageDetection(detectedSource string) error {
	if !c.languageDetection {
		return fmt.Errorf("Language detection is not available for this provider")
	}

	c.source = LanguageDetectionTag
	c.detectedSource = ""

	if len(detectedSource) == 0 {
		return nil
	}

	_, err := language.Parse(detectedSource)
	if err != nil {
		fmt.Printf("Detected source language '%s' is an invalid BCP 47 tag. Ignoring detection.\n", detectedSource)
		return nil
	}

	matchDetectedSource := matchLanguage(detectedSource, c.sourceLangs)
	if matchDetectedSource == nil {
		fmt.Printf("Detected source language '%s' is not available as a source language. Ignoring detection.\n", detectedSource)
		return nil
	}

	c.detectedSource = matchDetectedSource.Tag
	return nil
}

// Sets the source language for tags. If the source language is the same as the target language, switching the languages will be attempted. If the language is not available as a source language, an error will be returned. The detected source language will be cleared after setting the source language.
func (c *Collection) setTag(source string) error {
	matchSource := matchLanguage(source, c.sourceLangs)
	if matchSource == nil {
		return fmt.Errorf("Language '%s' is not available as a source language", source)
	}

	if len(c.target) == 0 {
		c.source = matchSource.Tag
		c.detectedSource = ""
		return nil
	}

	matchTarget := matchLanguage(c.target, c.sourceLangs)

	if matchTarget != nil && matchSource.Tag == matchTarget.Tag {
		matchPrevSource := matchLanguage(c.source, c.targetLangs)
		if matchPrevSource != nil {
			c.detectedSource = ""
			c.source, c.target = matchSource.Tag, matchPrevSource.Tag
			return nil
		}

		c.target = ""
	}

	c.detectedSource = ""
	c.source = matchSource.Tag
	return nil
}
