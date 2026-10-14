package language

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"quick-translate/internal/config"
	"quick-translate/internal/models"
	"quick-translate/internal/provider"

	"golang.org/x/text/language"
)

// Matches tags against a fixed list. Built once, because building is the expensive part of matching.
type matcher struct {
	tags    []string
	matcher language.Matcher // Nil for an empty list
}

// Builds a matcher over the tags. Earlier tags win a tie, so the preferred one comes first.
func newMatcher(tags []string) matcher {
	if len(tags) == 0 {
		return matcher{}
	}

	parsed := make([]language.Tag, len(tags))
	for i, tag := range tags {
		parsed[i] = language.Make(tag)
	}

	return matcher{tags: tags, matcher: language.NewMatcher(parsed)}
}

// Returns the best match for the tag; false for an empty tag or no match.
func (m matcher) match(tag string) (string, bool) {
	if len(tag) == 0 || m.matcher == nil {
		return "", false
	}

	_, i, confidence := m.matcher.Match(language.Make(tag))
	if confidence == language.No {
		return "", false
	}

	return m.tags[i], true
}

// Builds a new matcher without the tag.
func (m matcher) without(tag string) matcher {
	return newMatcher(slices.DeleteFunc(slices.Clone(m.tags), func(t string) bool { return t == tag }))
}

// The languages of one provider with the resolved preferences. Immutable.
type catalog struct {
	provider        models.ProviderID
	detection       bool
	preferredSource string // Empty if none is configured or available
	preferredTarget string
	sources         []models.Language // Preferred first, the rest sorted by tag
	targets         []models.Language
	sourceMatcher   matcher
	targetMatcher   matcher
	logger          *slog.Logger
}

// Builds the catalog of a provider. Drops invalid and duplicate tags; fails if no target language is left.
func newCatalog(info provider.Info, langs []models.Language, prefs config.LanguagePreferences, logger *slog.Logger) (*catalog, error) {
	logger = logger.With(slog.Any("provider", info.ID))

	valid := make([]models.Language, 0, len(langs))
	seen := make(map[string]struct{}, len(langs))
	for _, lang := range langs {
		if _, err := language.Parse(lang.Tag); err != nil {
			logger.Warn("invalid language tag dropped", slog.String("tag", lang.Tag))
			continue
		}

		if _, ok := seen[lang.Tag]; ok {
			logger.Warn("duplicate language tag dropped", slog.String("tag", lang.Tag))
			continue
		}

		seen[lang.Tag] = struct{}{}
		valid = append(valid, lang)
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("language: %s offers no valid language", info.ID)
	}

	slices.SortFunc(valid, func(a, b models.Language) int { return cmp.Compare(a.Tag, b.Tag) })

	var sources, targets []models.Language
	for _, lang := range valid {
		if lang.Source {
			sources = append(sources, lang)
		}

		if lang.Target {
			targets = append(targets, lang)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("language: %s offers no target language", info.ID)
	}

	preferredSource := preference(prefs.Source, sources, logger)
	preferredTarget := preference(prefs.Target, targets, logger)
	if preferredSource == preferredTarget {
		preferredTarget = ""
	}

	sources = preferredFirst(preferredSource, sources)
	targets = preferredFirst(preferredTarget, targets)

	return &catalog{
		provider:        info.ID,
		detection:       info.Detection,
		preferredSource: preferredSource,
		preferredTarget: preferredTarget,
		sources:         sources,
		targets:         targets,
		sourceMatcher:   newMatcher(tagsOf(sources)),
		targetMatcher:   newMatcher(tagsOf(targets)),
		logger:          logger,
	}, nil
}

// Resolves a preferred tag against the languages; empty if it is unset, invalid or unavailable.
func preference(tag string, langs []models.Language, logger *slog.Logger) string {
	if len(tag) == 0 {
		return ""
	}

	if _, err := language.Parse(tag); err != nil {
		logger.Warn("invalid preferred language ignored", slog.String("tag", tag))
		return ""
	}

	match, _ := newMatcher(tagsOf(langs)).match(tag)

	return match
}

// Returns a copy of the languages with the one carrying the tag moved to the front.
func preferredFirst(tag string, langs []models.Language) []models.Language {
	i := slices.IndexFunc(langs, func(l models.Language) bool { return l.Tag == tag })
	if i <= 0 {
		return langs
	}

	return slices.Concat(langs[i:i+1], langs[:i], langs[i+1:])
}

// Returns the tags of the languages in their order.
func tagsOf(langs []models.Language) []string {
	tags := make([]string, len(langs))
	for i, lang := range langs {
		tags[i] = lang.Tag
	}

	return tags
}

// Matches the tag against the source languages. False is returned for an empty tag or no match.
func (c *catalog) matchSource(tag string) (string, bool) {
	return c.sourceMatcher.match(tag)
}

// Matches the tag against the target languages. False is returned for an empty tag or no match.
func (c *catalog) matchTarget(tag string) (string, bool) {
	return c.targetMatcher.match(tag)
}

// Matches a detected tag against the source languages; empty if it cannot be resolved, which is not an error.
func (c *catalog) matchDetected(tag string) string {
	if len(tag) == 0 {
		return ""
	}

	if _, err := language.Parse(tag); err != nil {
		c.logger.Warn("invalid detected language ignored", slog.String("tag", tag))
		return ""
	}

	match, ok := c.matchSource(tag)
	if !ok {
		c.logger.Warn("detected language not available as a source language", slog.String("tag", tag))
		return ""
	}

	return match
}
