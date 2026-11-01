package provider

import (
	"context"
	"fmt"

	"quick-translate/internal/models"
)

// Fetches the languages of one provider. Immutable, so several can fetch concurrently.
type Fetcher struct {
	info Info
	impl implementation
}

// Returns the provider this fetcher was built for.
func (f *Fetcher) Info() Info {
	return f.info
}

// Fetches the languages the provider offers.
func (f *Fetcher) Languages(ctx context.Context) ([]models.Language, error) {
	languages, err := f.impl.languages(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider: %s: %w", f.info.ID, err)
	}

	return languages, nil
}
