package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"quick-translate/internal/models"
)

// How long a provider gets to answer. Go applies no timeout of its own.
const requestTimeout = 30 * time.Second

// Returned when the source language is left empty for a provider that cannot detect it.
var ErrDetectionUnsupported = errors.New("provider: language detection not supported")

// What the package hands every provider, so none configures its own client or timeout.
type environment struct {
	client *http.Client
}

// Creates the environment shared by every provider of one service.
func newEnvironment() environment {
	return environment{client: &http.Client{Timeout: requestTimeout}}
}

// What a provider implements. It is built once from its settings and never changed afterwards, so it must be
// safe for concurrent use.
type implementation interface {
	translate(ctx context.Context, request Request) (Result, error)
	languages(ctx context.Context) ([]models.Language, error)
}

type Request struct {
	Text   string
	Source string // Empty means "detect it"
	Target string
}

type Result struct {
	Text     string
	Detected string // Only set when the request had no source, otherwise empty
}

type definition struct {
	name      string                                                                     // Display name, not the slug
	detection bool                                                                       // Whether it can detect the source language
	build     func(decode func(into any) error, env environment) (implementation, error) // Builds the provider from its settings. Fails if they are invalid.
}

// Every implemented provider. Adding one means a models.ProviderID, a file implementing implementation and a
// row here.
var registry = map[models.ProviderID]definition{
	models.DeepL: {name: "DeepL", detection: true, build: newDeepl},
}

// What the frontend and the consumers know about a provider.
type Info struct {
	ID        models.ProviderID `json:"id"`
	Name      string            `json:"name"`
	Detection bool              `json:"detection"`
}

// Returns the registry entry of the provider as Info; false if it is not implemented.
func describe(ID models.ProviderID) (Info, bool) {
	def, ok := registry[ID]
	if !ok {
		return Info{}, false
	}

	return Info{ID: ID, Name: def.name, Detection: def.detection}, true
}

// Builds the provider from its settings. Fails if it is not implemented or the settings are invalid.
func build(ID models.ProviderID, decode func(into any) error, env environment) (Info, implementation, error) {
	info, ok := describe(ID)
	if !ok {
		return Info{}, nil, fmt.Errorf("provider: id %d not implemented", ID)
	}

	impl, err := registry[ID].build(decode, env)
	if err != nil {
		return Info{}, nil, fmt.Errorf("provider: %s: %w", ID, err)
	}

	return info, impl, nil
}
