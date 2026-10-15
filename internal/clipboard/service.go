package clipboard

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Returned by the startup when none of the supported programs is installed. Only occurs on Linux.
var ErrNoTool = errors.New("clipboard: no program found, install wl-clipboard (Wayland) or xclip or xsel (X11)")

// Reads and writes the clipboard of the operating system. Calls may run concurrently: the external programs
// are separate processes and the library serializes its calls itself.
type backend interface {
	read(ctx context.Context) (string, error)
	write(ctx context.Context, text string) error
	name() string // For the log
}

type Service struct {
	logger *slog.Logger

	mutex   sync.Mutex
	current backend // Nil before ServiceStartup
}

// Creates the service. The backend is selected in ServiceStartup.
func New(logger *slog.Logger) *Service {
	return &Service{logger: logger.With("category", "clipboard")}
}

// Selects the backend. Fails if the clipboard cannot be reached, on Linux with ErrNoTool if no program is
// installed, which shuts the application down.
func (s *Service) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	b, err := newBackend()
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.current = b
	s.mutex.Unlock()

	s.logger.Info("backend selected", slog.String("backend", b.name()))

	return nil
}

// Returns the selected text: the primary selection on Linux, the clipboard elsewhere. Empty if nothing is
// selected.
//
//wails:ignore
func (s *Service) Read(ctx context.Context) (string, error) {
	b, err := s.backend()
	if err != nil {
		return "", err
	}

	return b.read(ctx)
}

// Writes the text to the regular clipboard, never to the primary selection.
//
//wails:ignore
func (s *Service) Write(ctx context.Context, text string) error {
	b, err := s.backend()
	if err != nil {
		return err
	}

	return b.write(ctx, text)
}

// Returns the backend, which is immutable once selected, so it is used without the lock.
func (s *Service) backend() (backend, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.current == nil {
		return nil, errors.New("clipboard: not started")
	}

	return s.current, nil
}
