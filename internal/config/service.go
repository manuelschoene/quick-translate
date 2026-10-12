package config

import (
	"context"
	"fmt"
	"log/slog"
	"quick-translate/internal/system"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct {
	config *config
	files  *system.FileService
	logger *slog.Logger
}

// Creates a new config service instance. The config file is read on application startup.
func New(files *system.FileService, logger *slog.Logger) *Service {
	return &Service{
		files:  files,
		logger: logger.With("category", "config"),
	}
}

// Creates the config file from the template, or reads and validates the existing one. Shuts down the application if the template was just written or the file is broken. Called by Wails on application startup.
func (s *Service) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	path := s.files.Path(system.Settings)

	created, err := s.files.EnsureFile(system.Settings, []byte(fileTemplate))
	if err != nil {
		return fmt.Errorf("config: create file: %w", err)
	}

	if created {
		return fmt.Errorf("config: created file at %s: set a default provider and its settings, then start again", path)
	}

	s.logger.Info("using existing config file", slog.String("path", path))

	config, err := s.load()
	if err != nil {
		return err
	}

	s.config = config

	return nil
}
