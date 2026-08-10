package language

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"quick-translate/internal/models"
	"runtime"
	"slices"
	"strings"

	"golang.org/x/text/language"
)

// Validates the integrity of the provided languages and sorts them by their tags in ascending order. Returns an error if no valid languages are found.
func validateAndSortLanguages(langs []*models.Language) ([]*models.Language, error) {
	var validLangs []*models.Language
	for _, lang := range langs {
		_, err := language.Parse(lang.Tag)
		if err != nil {
			fmt.Printf("Provider language has an invalid BCP 47 tag '%s'. Ignoring the language.\n", lang.Tag)
			continue
		}

		validLangs = append(validLangs, lang)
	}

	if len(validLangs) == 0 {
		return nil, fmt.Errorf("No valid languages found for provider")
	}

	slices.SortFunc(validLangs, func(a, b *models.Language) int {
		return cmp.Compare(a.Tag, b.Tag)
	})

	return validLangs, nil
}

// Finds a language by its tag in a sorted list of languages. Returns the language and its index if found, otherwise returns nil and 0.
func findLanguageByTag(tag string, langs []*models.Language) (*models.Language, int) {
	i, found := slices.BinarySearchFunc(langs, tag, func(lang *models.Language, tag string) int {
		return strings.Compare(lang.Tag, tag)
	})

	if found {
		return langs[i], i
	}

	return nil, 0
}

// Parses the user language preferences and resolves them against the available languages. Returns a new LanguagePreferences instance with the resolved source and target language tags. If no match is found for a preference, it will be set to an empty string. Source and target languages are ensured to be different. If no preferences are provided, a new LanguagePreferences instance is created with empty source and target tags.
func parsePreferences(pref *models.LanguagePreferences, langs []*models.Language) *models.LanguagePreferences {
	if pref == nil {
		return new(models.LanguagePreferences)
	}

	_, err := language.Parse(pref.Source)
	if err != nil {
		fmt.Printf("Invalid BCP 47 tag for preferred source language '%s'. Ignoring the language.\n", pref.Source)
		pref.Source = ""
	}

	_, err = language.Parse(pref.Target)
	if err != nil {
		fmt.Printf("Invalid BCP 47 tag for preferred target language '%s'. Ignoring the language.\n", pref.Target)
		pref.Target = ""
	}

	var sourceTag, targetTag string
	var sourceLangs, targetLangs []*models.Language

	for _, lang := range langs {
		if lang.Source {
			sourceLangs = append(sourceLangs, lang)
		}
		if lang.Target {
			targetLangs = append(targetLangs, lang)
		}
	}

	if len(pref.Source) > 0 {
		source := matchLanguage(pref.Source, sourceLangs)
		if source != nil {
			sourceTag = source.Tag
		}
	}

	if len(pref.Target) > 0 {
		target := matchLanguage(pref.Target, targetLangs)
		if target != nil {
			targetTag = target.Tag
		}
	}

	if sourceTag == targetTag && len(sourceTag) > 0 {
		targetTag = ""
	}

	pref.Source = sourceTag
	pref.Target = targetTag

	return pref
}

// Splits the available languages into two separate slices: one for source languages and one for target languages. Orders the preferred tags first, followed by the remaining tags in ascending order.
func prioritySplit(pref *models.LanguagePreferences, langs []*models.Language) ([]*models.Language, []*models.Language) {
	var prefSource, prefTarget *models.Language
	var source []*models.Language
	var target []*models.Language

	for _, lang := range langs {
		if lang.Source {
			if lang.Tag == pref.Source {
				prefSource = lang
			} else {
				source = append(source, lang)
			}
		}
		if lang.Target {
			if lang.Tag == pref.Target {
				prefTarget = lang
			} else {
				target = append(target, lang)
			}
		}
	}

	return append([]*models.Language{prefSource}, source...), append([]*models.Language{prefTarget}, target...)
}

// Resolves the initial source and target languages based on the values of the previous provider and the languages from the new provider. Tries to find the best match for both languages, ensuring that they are different and supported by the new provider. Provide the previous source and target language tags, the new provider's language detection setting, the user's language preferences, and the new provider's available languages. Returns the resolved source and target language tags. Assumes that the languages are sorted ascending by tag.
func resolveInitialLanguages(
	source string,
	target string,
	languageDetection bool,
	pref *models.LanguagePreferences,
	sourceLangs []*models.Language,
	targetLangs []*models.Language,
) (string, string) {
	newSource, sourceCode := sourceFallback(source, pref.Source, sourceLangs, languageDetection)
	newTarget, targetCode := targetFallback(target, pref.Target, targetLangs)

	if (newSource != newTarget) || (newSource == "" && newTarget == "") {
		return newSource, newTarget
	}

	if sourceCode > targetCode {
		newTarget = rematchTarget(newTarget, target, pref, targetLangs)
		return newSource, newTarget
	}

	if sourceCode < targetCode {
		newSource = rematchSource(newSource, source, pref, sourceLangs, languageDetection)
		return newSource, newTarget
	}

	if languageDetection {
		return LanguageDetectionTag, newTarget
	}

	newTarget = rematchTarget(newTarget, target, pref, targetLangs)
	return newSource, newTarget
}

// Rematches the source language by removing the matched tag from the list of available source languages and finding the next best match.
func rematchSource(matchedTag string, initialTag string, pref *models.LanguagePreferences, sourceLangs []*models.Language, languageDetection bool) string {
	var prefSource string
	if matchedTag != pref.Source {
		prefSource = pref.Source
	}

	_, i := findLanguageByTag(matchedTag, sourceLangs)

	newSource, _ := sourceFallback(initialTag, prefSource, append(sourceLangs[:i], sourceLangs[i+1:]...), languageDetection)
	return newSource
}

// Rematches the target language by removing the matched tag from the list of available target languages and finding the next best match.
func rematchTarget(matchedTag string, initialTag string, pref *models.LanguagePreferences, targetLangs []*models.Language) string {
	var prefTarget string
	if matchedTag != pref.Target {
		prefTarget = pref.Target
	}

	_, i := findLanguageByTag(matchedTag, targetLangs)

	newTarget, _ := targetFallback(initialTag, prefTarget, append(targetLangs[:i], targetLangs[i+1:]...))
	return newTarget
}

// Matches the source tag against available source languages and returns the best match. If the source tag is available (code 3), it is returned. If the source tag is not available, but the provider allows language detection, the LanguageDetectionTag is returned (code 2). If the source tag is not available and language detection is not allowed, the preferred source tag is returned (code 1). If no match is found, an empty string is returned (code 0). Assumes that the languages are sorted in order of preference, with the most preferred language first.
func sourceFallback(source string, prefSource string, sourceLangs []*models.Language, languageDetection bool) (string, int) {
	if source == LanguageDetectionTag && languageDetection {
		return source, 2
	}

	if source != LanguageDetectionTag && len(source) > 0 {
		if lang := matchLanguage(source, sourceLangs); lang != nil {
			return lang.Tag, 3
		}
	}

	if languageDetection {
		return LanguageDetectionTag, 2
	}

	if len(prefSource) > 0 {
		return prefSource, 1
	}

	return "", 0
}

// Matches the target tag against available target languages and returns the best match. If the target tag is available (code 3), it is returned. If the target tag is not available, but a preferred target tag is provided, it is returned (code 2). If the target tag is not available and no preferred target tag is provided, the system locale is used to find a match (code 1). If no match is found, an empty string is returned (code 0). Assumes that the languages are sorted in order of preference, with the most preferred language first.
func targetFallback(target string, prefTarget string, targetLangs []*models.Language) (string, int) {
	if len(target) > 0 {
		if lang := matchLanguage(target, targetLangs); lang != nil {
			return lang.Tag, 3
		}
	}

	if len(prefTarget) > 0 {
		return prefTarget, 2
	}

	locale, err := locale()
	if err != nil {
		fmt.Println(fmt.Errorf("Falling back to user selection: %w", err))
		return "", 0
	}

	if lang := matchLanguage(locale, targetLangs); lang != nil {
		return lang.Tag, 1
	}

	return "", 0
}

// Detects the system locale and returns it as a language tag. If the locale cannot be determined, an error is returned.
func locale() (string, error) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "Get-Culture | select -exp Name")

		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("Cannot determine system locale: %w", err)
		}

		return strings.Trim(string(out), "\r\n"), nil
	}

	locale, ok := os.LookupEnv("LANG")
	if !ok {
		return "", fmt.Errorf("Cannot determine system locale: LANG environment variable not set")
	}

	s, _, _ := strings.Cut(locale, ".")

	tag, err := language.Parse(s)
	if err != nil {
		return "", fmt.Errorf("System locale '%s' is not a valid BCP 47 tag", s)
	}

	return tag.String(), nil
}

// Matches the given language against the available languages in the collection and returns the best match. If no match is found, it returns nil. The langs parameter should be sorted in order of preference, with the most preferred language first.
func matchLanguage(tag string, langs []*models.Language) *models.Language {
	tags := make([]language.Tag, len(langs))
	for i, l := range langs {
		tags[i] = language.Make(l.Tag)
	}

	_, j, conf := language.NewMatcher(tags).Match(language.Make(tag))
	if conf == language.No {
		return nil
	}

	return langs[j]
}
