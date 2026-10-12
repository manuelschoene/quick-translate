package store

import (
	"context"
	"fmt"
	"quick-translate/internal/store/db"
	"quick-translate/internal/models"
	"time"
)

type ProviderRepository struct {
	queries *db.Queries
}

// Retrieves the provider with the given ID from the database. Returns ErrNotFound if the provider does not exist, i.e. its languages were never fetched.
func (r *ProviderRepository) Get(ctx context.Context, ID models.ProviderID) (models.Provider, error) {
	p, err := r.queries.GetProvider(ctx, ID)
	if err != nil {
		return models.Provider{}, fmt.Errorf("get provider: %w", notFound(err))
	}

	return models.Provider{
		ID:                 p.ID,
		LanguagesUpdatedAt: time.Unix(p.LanguagesUpdatedAt, 0),
	}, nil
}

// Inserts or updates the provider in the database. If the provider already exists, its LanguagesUpdatedAt field will be updated.
func (r *ProviderRepository) Upsert(ctx context.Context, provider models.Provider) error {
	err := r.queries.UpsertProvider(ctx, db.UpsertProviderParams{
		ID:                 provider.ID,
		LanguagesUpdatedAt: provider.LanguagesUpdatedAt.Unix(),
	})
	if err != nil {
		return fmt.Errorf("upsert provider: %w", err)
	}

	return nil
}
