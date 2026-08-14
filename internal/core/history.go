package core

import (
	"fmt"

	"quick-translate/internal/models"
)

// Checks if translations are kept in the history. A disabled history has nothing to step through, so stepping should not be offered.
func (c *Core) HistoryEnabled() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.history.Enabled()
}

// Returns whether an older and a newer translation exist for the current one. Meant for deciding whether stepping through the history can be offered after a translation. A disabled history reports no translations instead of an error.
func (c *Core) Navigation() (hasPrevious bool, hasNext bool, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.history.Enabled() {
		return false, false, nil
	}

	return c.flags()
}

// Returns the translation that was stored before the current one, and whether an older and a newer translation exist for it. The provider and the languages of the translation are taken over as far as they are still available. The current translation is returned unchanged when it is the oldest one that is still kept. Returns an error if the history is disabled or can not be read.
func (c *Core) PreviousTranslation() (translation *models.Translation, hasPrevious bool, hasNext bool, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, err := c.history.Previous(c.currentID())

	return c.navigate(entry, err)
}

// Returns the translation that was stored after the current one, and whether an older and a newer translation exist for it. The provider and the languages of the translation are taken over as far as they are still available. The current translation is returned unchanged when it is the newest one. Returns an error if the history is disabled or can not be read.
func (c *Core) NextTranslation() (translation *models.Translation, hasPrevious bool, hasNext bool, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, err := c.history.Next(c.currentID())

	return c.navigate(entry, err)
}

// Shows the translation the history has answered with and takes its provider and languages over. Returns the current translation when the history has nothing more to offer in that direction. Requires the lock to be held.
func (c *Core) navigate(entry *models.Translation, err error) (*models.Translation, bool, bool, error) {
	if err != nil {
		return nil, false, false, fmt.Errorf("Could not step through the history: %w", err)
	}

	if entry == nil && c.translation == nil {
		return nil, false, false, ErrNoTranslation
	}

	if entry != nil {
		c.adopt(entry)
		c.translation = entry
	}

	hasPrevious, hasNext, err := c.flags()

	return c.current(), hasPrevious, hasNext, err
}

// Takes the provider and the languages of the translation over, as far as they are still available. Every step is done on a best effort basis and only reported when it fails, because a translation made with a provider that has been removed from the configuration should still be shown. Requires the lock to be held.
func (c *Core) adopt(entry *models.Translation) {
	if entry.Provider != c.providerSlug {
		if err := c.setProvider(entry.Provider); err != nil {
			fmt.Printf("Could not take the provider of the translation over: %v\n", err)
		}
	}

	// The source language is set first, because setting the target language may correct the source language afterwards.
	if len(entry.Source) > 0 {
		if err := c.langs.SetSource(entry.Source, entry.DetectedSource); err != nil {
			fmt.Printf("Could not take the source language of the translation over: %v\n", err)
		}
	}

	if len(entry.Target) > 0 {
		if err := c.langs.SetTarget(entry.Target); err != nil {
			fmt.Printf("Could not take the target language of the translation over: %v\n", err)
		}
	}
}

// Returns whether an older and a newer translation exist for the current one. Requires the lock to be held.
func (c *Core) flags() (bool, bool, error) {
	id := c.currentID()

	hasPrevious, err := c.history.HasPrevious(id)
	if err != nil {
		return false, false, fmt.Errorf("Could not look for an older translation: %w", err)
	}

	hasNext, err := c.history.HasNext(id)
	if err != nil {
		return false, false, fmt.Errorf("Could not look for a newer translation: %w", err)
	}

	return hasPrevious, hasNext, nil
}

// Returns the ID the history is stepped through from. An ID below one stands for the position behind the newest translation, which is where stepping starts when nothing has been translated yet. Requires the lock to be held.
func (c *Core) currentID() int {
	if c.translation == nil {
		return 0
	}

	return c.translation.ID
}
