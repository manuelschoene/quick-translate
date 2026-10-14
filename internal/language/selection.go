package language

import (
	"errors"
	"fmt"
)

type Selection struct {
	Detection bool   `json:"detection"` // The provider detects the source language
	Source    string `json:"source"`    // Empty while detecting or when none is selected
	Target    string `json:"target"`
	Detected  string `json:"detected"` // The language detected last; only set while detecting
}

// A source or target language chosen by a fallback, with the rank of the step that chose it.
type pick struct {
	tag    string
	detect bool
	rank   int // 3: carried over, 2: detection or preference, 1: preference or locale, 0: none
}

// Carries the previous selection over to a new catalog as far as it allows, falling back to detection, the
// preferences and the locale. Source and target never end up equal.
func initial(c *catalog, previous Selection, locale string) Selection {
	source := sourceFallback(previous, c.preferredSource, c.sourceMatcher, c.detection)
	target := targetFallback(previous.Target, c.preferredTarget, c.targetMatcher, locale)

	if !source.detect && len(source.tag) != 0 && source.tag == target.tag {
		switch {
		case source.rank > target.rank:
			target = rematchTarget(target.tag, previous.Target, c.preferredTarget, c.targetMatcher, locale)
		case source.rank < target.rank:
			source = rematchSource(source.tag, previous, c.preferredSource, c.sourceMatcher, c.detection)
		case c.detection:
			source = pick{detect: true}
		default:
			target = rematchTarget(target.tag, previous.Target, c.preferredTarget, c.targetMatcher, locale)
		}
	}

	selection := Selection{Detection: source.detect, Source: source.tag, Target: target.tag}
	if selection.Detection {
		selection.Detected = c.matchDetected(previous.Detected)
	}

	return selection
}

// Picks the source language: the previous one, detection, the preference, or none, in that order. Detection that
// was selected before wins over everything.
func sourceFallback(previous Selection, preferred string, m matcher, detection bool) pick {
	if previous.Detection && detection {
		return pick{detect: true, rank: 2}
	}

	if !previous.Detection {
		if tag, ok := m.match(previous.Source); ok {
			return pick{tag: tag, rank: 3}
		}
	}

	if detection {
		return pick{detect: true, rank: 2}
	}

	if len(preferred) != 0 {
		return pick{tag: preferred, rank: 1}
	}

	return pick{}
}

// Picks the target language: the previous one, the preference, the locale, or none, in that order.
func targetFallback(previous string, preferred string, m matcher, locale string) pick {
	if tag, ok := m.match(previous); ok {
		return pick{tag: tag, rank: 3}
	}

	if len(preferred) != 0 {
		return pick{tag: preferred, rank: 2}
	}

	if tag, ok := m.match(locale); ok {
		return pick{tag: tag, rank: 1}
	}

	return pick{}
}

// Picks the source language again without the colliding tag.
func rematchSource(collided string, previous Selection, preferred string, m matcher, detection bool) pick {
	if collided == preferred {
		preferred = ""
	}

	return sourceFallback(previous, preferred, m.without(collided), detection)
}

// Picks the target language again without the colliding tag.
func rematchTarget(collided string, previous string, preferred string, m matcher, locale string) pick {
	if collided == preferred {
		preferred = ""
	}

	return targetFallback(previous, preferred, m.without(collided), locale)
}

// Sets the source language. A colliding target is swapped with the previous source, or dropped.
func setSource(c *catalog, s Selection, tag string) (Selection, error) {
	source, ok := c.matchSource(tag)
	if !ok {
		return s, fmt.Errorf("language: %s not available as a source language", tag)
	}

	next := Selection{Source: source, Target: s.Target}

	if target, ok := c.matchSource(s.Target); ok && target == source {
		next.Target = ""

		if !s.Detection {
			if previous, ok := c.matchTarget(s.Source); ok {
				next.Target = previous
			}
		}
	}

	return next, nil
}

// Switches to detection and keeps the detected language as far as it is a source language.
func detect(c *catalog, s Selection, detected string) (Selection, error) {
	if !c.detection {
		return s, fmt.Errorf("language: %s cannot detect the source language", c.provider)
	}

	return Selection{Detection: true, Target: s.Target, Detected: c.matchDetected(detected)}, nil
}

// Sets the target language. A colliding source is swapped with the previous target, replaced by detection, or
// dropped.
func setTarget(c *catalog, s Selection, tag string) (Selection, error) {
	target, ok := c.matchTarget(tag)
	if !ok {
		return s, fmt.Errorf("language: %s not available as a target language", tag)
	}

	if s.Detection || len(s.Source) == 0 {
		return Selection{Detection: s.Detection, Target: target}, nil
	}

	next := Selection{Source: s.Source, Target: target}

	if source, ok := c.matchTarget(s.Source); ok && source == target {
		switch previous, ok := c.matchSource(s.Target); {
		case ok:
			next.Source = previous
		case c.detection:
			next = Selection{Detection: true, Target: target}
		default:
			next.Source = ""
		}
	}

	return next, nil
}

// Swaps source and target. While detecting, the detected language becomes the target; a missing language stays
// missing on the other side.
func switchLanguages(c *catalog, s Selection) (Selection, error) {
	if s.Detection {
		return switchDetection(c, s)
	}

	if len(s.Source) == 0 && len(s.Target) == 0 {
		return s, nil
	}

	var next Selection

	if len(s.Source) != 0 {
		target, ok := c.matchTarget(s.Source)
		if !ok {
			return s, fmt.Errorf("language: %s not available as a target language", s.Source)
		}

		next.Target = target
	}

	if len(s.Target) != 0 {
		source, ok := c.matchSource(s.Target)
		if !ok {
			return s, fmt.Errorf("language: %s not available as a source language", s.Target)
		}

		next.Source = source
	}

	return next, nil
}

// Swaps the target with the detected language and ends detection. Without a usable detection the target stays
// empty.
func switchDetection(c *catalog, s Selection) (Selection, error) {
	if len(s.Target) == 0 {
		return s, errors.New("language: nothing to switch while detecting without a target language")
	}

	source, ok := c.matchSource(s.Target)
	if !ok {
		return s, fmt.Errorf("language: %s not available as a source language", s.Target)
	}

	next := Selection{Source: source}

	if target, ok := c.matchTarget(s.Detected); ok && s.Detected != source {
		next.Target = target
	}

	return next, nil
}
