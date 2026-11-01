package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"quick-translate/internal/models"
	"quick-translate/internal/store/db"
	"time"
)

type LanguageRepository struct {
	conn    *sql.DB
	logger  *slog.Logger
	queries *db.Queries
}

// Retrieves the list of languages for the given provider ID. Returns an empty list without an error if none are stored, including when the provider does not exist or its languages expired.
func (r *LanguageRepository) Get(ctx context.Context, ID models.ProviderID) ([]models.Language, error) {
	rows, err := r.queries.ListProviderLanguages(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("get provider languages: %w", err)
	}

	languages := make([]models.Language, 0, len(rows))
	for _, row := range rows {
		languages = append(languages, models.Language{
			Tag:    row.Tag,
			Name:   row.Name,
			Source: row.Source,
			Target: row.Target,
			Stable: row.Stable,
		})
	}

	return languages, nil
}

// Replaces the languages of the given provider in one transaction, records now as their fetch time and deletes the languages nothing references any more. Creates the provider row if it does not exist yet.
func (r *LanguageRepository) Put(ctx context.Context, ID models.ProviderID, languages []models.Language) error {
	var orphans int64

	err := transaction(ctx, r.conn, r.queries, func(query *db.Queries) error {
		if err := query.UpsertProvider(ctx, db.UpsertProviderParams{
			ID:                 ID,
			LanguagesUpdatedAt: time.Now().Unix(),
		}); err != nil {
			return fmt.Errorf("upsert provider: %w", err)
		}

		if err := query.DeleteProviderLanguages(ctx, ID); err != nil {
			return fmt.Errorf("delete provider languages: %w", err)
		}

		for _, lang := range languages {
			if err := query.InsertLanguage(ctx, lang.Tag); err != nil {
				return fmt.Errorf("insert language %s: %w", lang.Tag, err)
			}

			if err := query.InsertProviderLanguage(ctx, db.InsertProviderLanguageParams{
				ProviderID: ID,
				Tag:        lang.Tag,
				Name:       lang.Name,
				Source:     lang.Source,
				Target:     lang.Target,
				Stable:     lang.Stable,
			}); err != nil {
				return fmt.Errorf("insert provider language %s: %w", lang.Tag, err)
			}
		}

		count, err := query.DeleteOrphanLanguages(ctx)
		if err != nil {
			return fmt.Errorf("delete orphan languages: %w", err)
		}

		orphans = count

		return nil
	})

	if err != nil {
		return fmt.Errorf("put languages: %w", err)
	}

	r.logger.Info("deleted orphan languages", slog.Int64("count", orphans))

	return nil
}

// Returns when the languages of the given providers were fetched. A provider missing from the map has never been fetched.
func (r *LanguageRepository) FetchedAt(ctx context.Context, IDs []models.ProviderID) (map[models.ProviderID]time.Time, error) {
	rows, err := r.queries.ListLanguagesUpdatedAt(ctx, IDs)
	if err != nil {
		return nil, fmt.Errorf("get languages fetch time: %w", err)
	}

	fetched := make(map[models.ProviderID]time.Time, len(rows))
	for _, row := range rows {
		fetched[row.ID] = time.Unix(row.LanguagesUpdatedAt, 0)
	}

	return fetched, nil
}

// Deletes the languages of the given provider, then the languages nothing references any more. The provider row and its fetch time are kept, because the history references it.
func (r *LanguageRepository) Delete(ctx context.Context, ID models.ProviderID) error {
	var orphans int64

	err := transaction(ctx, r.conn, r.queries, func(query *db.Queries) error {
		if err := query.DeleteProviderLanguages(ctx, ID); err != nil {
			return fmt.Errorf("delete provider languages: %w", err)
		}

		count, err := query.DeleteOrphanLanguages(ctx)
		if err != nil {
			return fmt.Errorf("delete orphan languages: %w", err)
		}

		orphans = count

		return nil
	})

	if err != nil {
		return fmt.Errorf("delete languages: %w", err)
	}

	r.logger.Info("deleted orphan languages", slog.Int64("count", orphans))

	return nil
}
