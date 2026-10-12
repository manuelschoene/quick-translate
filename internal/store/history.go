package store

import (
	"context"
	"fmt"
	"quick-translate/internal/store/db"
	"quick-translate/internal/models"
	"time"
)

type HistoryRepository struct {
	queries *db.Queries
}

// Appends a new translation to the history. The translation ID and creation timestamp will be set on the provided Translation struct.
func (r *HistoryRepository) Append(ctx context.Context, t *models.Translation) error {
	row, err := r.queries.AppendTranslation(ctx, db.AppendTranslationParams{
		Text:        t.Text,
		Translation: t.Translation,
		ProviderID:  t.ProviderID,
		SourceTag:   t.SourceTag,
		TargetTag:   t.TargetTag,
		Detection:   t.Detection,
	})
	if err != nil {
		return fmt.Errorf("append translation: %w", err)
	}

	t.ID = row.ID
	t.CreatedAt = time.Unix(row.CreatedAt, 0)

	return nil
}

// Retrieves the previous translation relative to the given translation ID, within the specified window size. Returns ErrNotFound if no previous translation exists.
func (r *HistoryRepository) Previous(ctx context.Context, tID int64, windowSize int64) (models.Translation, error) {
	row, err := r.queries.PreviousTranslation(ctx, db.PreviousTranslationParams{
		ID:         tID,
		WindowSize: windowSize,
	})
	if err != nil {
		return models.Translation{}, fmt.Errorf("previous translation: %w", notFound(err))
	}

	return translationFromRow(row), nil
}

// Retrieves the next translation relative to the given translation ID. Returns ErrNotFound if no next translation exists.
func (r *HistoryRepository) Next(ctx context.Context, tID int64) (models.Translation, error) {
	row, err := r.queries.NextTranslation(ctx, tID)
	if err != nil {
		return models.Translation{}, fmt.Errorf("next translation: %w", notFound(err))
	}

	return translationFromRow(row), nil
}

// Checks if there is a previous translation relative to the given translation ID, within the specified window size. If the translation with the given ID does not exist, it will be treated as if there is no previous translation.
func (r *HistoryRepository) HasPrevious(ctx context.Context, tID int64, windowSize int64) (bool, error) {
	hasPrevious, err := r.queries.HasPreviousTranslation(ctx, db.HasPreviousTranslationParams{
		ID:         tID,
		WindowSize: windowSize,
	})
	if err != nil {
		return false, fmt.Errorf("has previous translation: %w", err)
	}

	return hasPrevious, nil
}

// Checks if there is a next translation relative to the given translation ID. If the translation with the given ID does not exist, it will be treated as if there is no next translation.
func (r *HistoryRepository) HasNext(ctx context.Context, tID int64) (bool, error) {
	hasNext, err := r.queries.HasNextTranslation(ctx, tID)
	if err != nil {
		return false, fmt.Errorf("has next translation: %w", err)
	}

	return hasNext, nil
}

// Retrieves the latest translation from the history. Returns ErrNotFound if the history is empty.
func (r *HistoryRepository) Latest(ctx context.Context) (models.Translation, error) {
	row, err := r.queries.LatestTranslation(ctx)
	if err != nil {
		return models.Translation{}, fmt.Errorf("latest translation: %w", notFound(err))
	}

	return translationFromRow(row), nil
}

// Maps a generated row onto the model.
func translationFromRow(row db.Translation) models.Translation {
	return models.Translation{
		ID:          row.ID,
		Text:        row.Text,
		Translation: row.Translation,
		CreatedAt:   time.Unix(row.CreatedAt, 0),
		ProviderID:  row.ProviderID,
		SourceTag:   row.SourceTag,
		TargetTag:   row.TargetTag,
		Detection:   row.Detection,
	}
}
