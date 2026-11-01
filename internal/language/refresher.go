package language

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"quick-translate/internal/models"
)

// How many providers are fetched at the same time.
const refreshLimit = 3

// The part of the language repository the refresher uses.
type refreshRepository interface {
	FetchedAt(ctx context.Context, IDs []models.ProviderID) (map[models.ProviderID]time.Time, error)
	Put(ctx context.Context, ID models.ProviderID, languages []models.Language) error
}

// Refreshes the stored languages of every configured provider once they are older than the TTL. Runs in its own
// goroutine, one pass at start and one per trigger.
type refresher struct {
	ttl         time.Duration
	repo        refreshRepository
	providers   func() ([]models.ProviderID, error)
	fetch       func(context.Context, models.ProviderID) ([]models.Language, error)
	onRefreshed func(models.ProviderID, []models.Language) // Called after a list was stored
	logger      *slog.Logger

	trigger chan struct{} // Holds at most one pending pass
	done    chan struct{} // Closed when the goroutine ended; nil before Start
}

// Creates a refresher. Nothing runs before Start.
func newRefresher(
	ttl time.Duration,
	repo refreshRepository,
	providers func() ([]models.ProviderID, error),
	fetch func(context.Context, models.ProviderID) ([]models.Language, error),
	onRefreshed func(models.ProviderID, []models.Language),
	logger *slog.Logger,
) *refresher {
	return &refresher{
		ttl:         ttl,
		repo:        repo,
		providers:   providers,
		fetch:       fetch,
		onRefreshed: onRefreshed,
		logger:      logger,
		trigger:     make(chan struct{}, 1),
	}
}

// Starts the goroutine with a first pass. It runs until ctx is done, which is its only way to stop, so pass the
// application context. Must be called once.
func (r *refresher) Start(ctx context.Context) {
	r.done = make(chan struct{})

	go func() {
		defer close(r.done)

		for {
			r.pass(ctx)

			select {
			case <-ctx.Done():
				return
			case <-r.trigger:
			}
		}
	}()
}

// Asks for another pass without waiting. Triggers coalesce while a pass is pending.
func (r *refresher) Trigger() {
	select {
	case r.trigger <- struct{}{}:
	default:
	}
}

// Waits for the goroutine to end after its context is done. Returns at once if it was never started.
func (r *refresher) Wait() {
	if r.done == nil {
		return
	}

	<-r.done
}

// Fetches every stale provider, at most refreshLimit at a time, and waits for all of them.
func (r *refresher) pass(ctx context.Context) {
	stale, err := r.stale(ctx)
	if err != nil {
		r.logger.Error("find stale languages", slog.Any("error", err))
		return
	}

	var wg sync.WaitGroup
	slots := make(chan struct{}, refreshLimit)

	for _, ID := range stale {
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			wg.Wait()
			return
		}

		wg.Go(func() {
			defer func() { <-slots }()
			r.refresh(ctx, ID)
		})
	}

	wg.Wait()
}

// Returns the configured providers that were never fetched or not within the TTL.
func (r *refresher) stale(ctx context.Context) ([]models.ProviderID, error) {
	IDs, err := r.providers()
	if err != nil {
		return nil, err
	}

	fetched, err := r.repo.FetchedAt(ctx, IDs)
	if err != nil {
		return nil, err
	}

	var stale []models.ProviderID
	for _, ID := range IDs {
		if at, ok := fetched[ID]; !ok || time.Since(at) >= r.ttl {
			stale = append(stale, ID)
		}
	}

	return stale, nil
}

// Fetches and stores the languages of one provider. A failure keeps the stored list.
func (r *refresher) refresh(ctx context.Context, ID models.ProviderID) {
	logger := r.logger.With(slog.Any("provider", ID))

	languages, err := r.fetch(ctx, ID)
	if err != nil {
		logger.Warn("refresh languages", slog.Any("error", err))
		return
	}

	if len(languages) == 0 {
		logger.Warn("refresh languages: provider returned none, keeping the stored ones")
		return
	}

	if err := r.repo.Put(ctx, ID, languages); err != nil {
		logger.Error("store refreshed languages", slog.Any("error", err))
		return
	}

	logger.Info("languages refreshed", slog.Int("count", len(languages)))

	if r.onRefreshed != nil {
		r.onRefreshed(ID, languages)
	}
}
