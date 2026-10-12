package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"quick-translate/internal/store/db"
	"quick-translate/internal/models"
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

// Replaces the languages of the given provider in one transaction and records now as their fetch time. Creates the provider row if it does not exist yet.
func (r *LanguageRepository) Put(ctx context.Context, ID models.ProviderID, languages []models.Language) error {
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

		return nil
	})

	if err != nil {
		return fmt.Errorf("put languages: %w", err)
	}

	return nil
}

// Deletes the languages of every provider fetched longer than ttl ago, then the languages neither a provider nor the history references any more.
func (r *LanguageRepository) DeleteExpired(ctx context.Context, ttl time.Duration) error {
	var count, countOrphan int64

	err := transaction(ctx, r.conn, r.queries, func(query *db.Queries) error {
		c, err := query.DeleteExpiredLanguages(ctx, time.Now().Add(-ttl).Unix())
		if err != nil {
			return fmt.Errorf("delete expired languages: %w", err)
		}

		co, err := query.DeleteOrphanLanguages(ctx)
		if err != nil {
			return fmt.Errorf("delete orphan languages: %w", err)
		}

		count = c
		countOrphan = co

		return nil
	})

	if err != nil {
		return err
	}

	r.logger.Info("deleted expired languages", slog.Int64("count", count))
	r.logger.Info("deleted orphan languages", slog.Int64("count", countOrphan))

	return nil
}
