package translation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/manuelschoene/quick-translate/internal/models"
	"github.com/manuelschoene/quick-translate/internal/store"
)

var (
	ErrHistoryDisabled = errors.New("translation: history disabled, set history.max_entries above zero")
	ErrNoPrevious      = errors.New("translation: no previous translation")
	ErrNoNext          = errors.New("translation: no next translation")
	ErrHistoryEmpty    = errors.New("translation: history empty")
)

type Navigation struct {
	HasPrevious bool `json:"hasPrevious"`
	HasNext     bool `json:"hasNext"`
}

// The part of the store the service uses.
type storage interface {
	NewHistory() *store.HistoryRepository
}

// The part of the configuration the service reads.
type configuration interface {
	HistoryLimit() int
}

// Reads the history limit and takes the repository. Without a store the history stays disabled.
func (s *Service) startHistory() {
	limit := s.config.HistoryLimit()
	if limit <= 0 {
		s.logger.Info("history disabled")
		return
	}

	repo := s.store.NewHistory()
	if repo == nil {
		s.logger.Error("store not started, the history is disabled")
		return
	}

	s.history.Lock()
	s.repo, s.limit = repo, int64(limit)
	s.history.Unlock()

	s.logger.Info("history enabled", slog.Int("limit", limit))
}

// Returns whether there is a translation to step to. Both are false while the history is disabled.
func (s *Service) Navigation(ctx context.Context) (Navigation, error) {
	s.history.Lock()
	defer s.history.Unlock()

	if s.repo == nil {
		return Navigation{}, nil
	}

	return s.navigation(ctx)
}

// Shows the translation before the shown one, or the newest if the shown one is not stored.
func (s *Service) Previous(ctx context.Context) error {
	return s.step(ctx, func(repo *store.HistoryRepository, cursor int64) (models.Translation, error) {
		if cursor == 0 {
			return latest(ctx, repo)
		}

		entry, err := repo.Previous(ctx, cursor, s.limit)
		return entry, lookup(err, ErrNoPrevious)
	})
}

// Shows the translation after the shown one.
func (s *Service) Next(ctx context.Context) error {
	return s.step(ctx, func(repo *store.HistoryRepository, cursor int64) (models.Translation, error) {
		if cursor == 0 {
			return models.Translation{}, ErrNoNext
		}

		entry, err := repo.Next(ctx, cursor)
		return entry, lookup(err, ErrNoNext)
	})
}

// Shows the newest translation. Fails if the history is disabled or empty.
//
//wails:ignore
func (s *Service) Latest() error {
	ctx := s.app.Context()

	return s.step(ctx, func(repo *store.HistoryRepository, _ int64) (models.Translation, error) {
		return latest(ctx, repo)
	})
}

// Finds an entry and shows it with its provider and languages. Nothing changes if no entry is found.
func (s *Service) step(ctx context.Context, find func(*store.HistoryRepository, int64) (models.Translation, error)) error {
	s.history.Lock()
	defer s.history.Unlock()

	if s.repo == nil {
		return ErrHistoryDisabled
	}

	entry, err := find(s.repo, s.cursor)
	if err != nil {
		return err
	}

	// Not fatal, so an entry of a provider that is no longer configured does not block the navigation.
	if err := s.providers.Select(entry.ProviderID); err != nil {
		s.logger.Warn("select provider of entry", slog.Int64("id", entry.ID), slog.Any("error", err))
	}

	if err := s.languages.Adopt(entry); err != nil {
		s.logger.Warn("adopt languages of entry", slog.Int64("id", entry.ID), slog.Any("error", err))
	}

	s.mutex.Lock()
	s.show(entry)
	s.mutex.Unlock()

	s.move(ctx, entry.ID)

	return nil
}

// Stores a new translation and moves the cursor to it. Requires work to be held.
func (s *Service) record(t models.Translation) {
	s.history.Lock()
	defer s.history.Unlock()

	if s.repo == nil {
		return
	}

	ctx := s.app.Context()

	// Not fatal; the translation is shown anyway, just not stored.
	if err := s.repo.Append(ctx, &t); err != nil {
		s.logger.Warn("record translation", slog.Any("error", err))
		t.ID = 0
	}

	s.move(ctx, t.ID)
}

// Sets the cursor and emits the navigation. Requires history to be held.
func (s *Service) move(ctx context.Context, cursor int64) {
	s.cursor = cursor

	navigation, err := s.navigation(ctx)
	if err != nil {
		s.logger.Warn("read navigation", slog.Any("error", err))
	}

	s.app.Event.Emit(EventHistoryChanged, navigation)
}

// Reads whether there is an entry before and after the cursor. Requires history to be held and the repository
// to be set.
func (s *Service) navigation(ctx context.Context) (Navigation, error) {
	// An unstored translation has nothing after it, and the newest entry before it.
	if s.cursor == 0 {
		_, err := s.repo.Latest(ctx)
		if errors.Is(err, store.ErrNotFound) {
			return Navigation{}, nil
		}

		if err != nil {
			return Navigation{}, fmt.Errorf("translation: history: %w", err)
		}

		return Navigation{HasPrevious: true}, nil
	}

	previous, err := s.repo.HasPrevious(ctx, s.cursor, s.limit)
	if err != nil {
		return Navigation{}, fmt.Errorf("translation: history: %w", err)
	}

	next, err := s.repo.HasNext(ctx, s.cursor)
	if err != nil {
		return Navigation{}, fmt.Errorf("translation: history: %w", err)
	}

	return Navigation{HasPrevious: previous, HasNext: next}, nil
}

// Returns the newest entry; ErrHistoryEmpty if there is none.
func latest(ctx context.Context, repo *store.HistoryRepository) (models.Translation, error) {
	entry, err := repo.Latest(ctx)
	return entry, lookup(err, ErrHistoryEmpty)
}

// Replaces store.ErrNotFound with the sentinel and prefixes any other error.
func lookup(err error, missing error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, store.ErrNotFound):
		return missing
	default:
		return fmt.Errorf("translation: history: %w", err)
	}
}
