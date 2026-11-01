package provider

import (
	"context"
	"errors"
	"fmt"
)

// Translates with one provider. Immutable, so it is used without a lock and keeps working after a switch.
type Translator struct {
	info Info
	impl implementation
}

// Returns the provider this translator was built for.
func (t *Translator) Info() Info {
	return t.info
}

// Translates the text. An empty source detects the language, which fails with ErrDetectionUnsupported for a
// provider that cannot.
func (t *Translator) Translate(ctx context.Context, request Request) (Result, error) {
	if len(request.Text) == 0 {
		return Result{}, errors.New("provider: nothing to translate")
	}

	if len(request.Target) == 0 {
		return Result{}, errors.New("provider: no target language")
	}

	if len(request.Source) == 0 && !t.info.Detection {
		return Result{}, ErrDetectionUnsupported
	}

	result, err := t.impl.translate(ctx, request)
	if err != nil {
		return Result{}, fmt.Errorf("provider: %s: %w", t.info.ID, err)
	}

	if len(request.Source) != 0 {
		result.Detected = ""
	}

	return result, nil
}
