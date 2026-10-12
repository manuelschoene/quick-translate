package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"quick-translate/internal/store/db"
	"quick-translate/internal/system"

	_ "github.com/glebarez/go-sqlite"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Service struct {
	conn    *sql.DB
	files   *system.FileService
	logger  *slog.Logger
	queries *db.Queries
}

// Creates a new store service instance. The service will not be connected to the database yet. Connection is established on application startup.
func New(f *system.FileService, l *slog.Logger) *Service {
	return &Service{
		files:  f,
		logger: l.With("category", "store"),
	}
}

// Connects to the database and initializes the queries. Ensures the database file exists. Shuts down the application if the connection fails. Called by Wails on application startup.
func (s *Service) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	path, err := s.createOrGetDatabaseFile()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if err := s.connect(ctx, path); err != nil {
		return fmt.Errorf("database: connect: %w", err)
	}

	if err := migrate(ctx, migrations, s.conn, s.logger); err != nil {
		s.conn.Close()
		return fmt.Errorf("database: %w", err)
	}

	return nil
}

// Ensures the database file exists and returns its path. If the file does not exist, it will be created.
func (s *Service) createOrGetDatabaseFile() (path string, err error) {
	created, err := s.files.EnsureFile(system.Database, []byte{})
	if err != nil {
		return
	}

	// Not empty as EnsureFile would have errored already
	path = s.files.Path(system.Database)

	if created {
		s.logger.Info("creating new database file", slog.String("path", path))
		return
	}

	s.logger.Info("using existing database file", slog.String("path", path))
	return
}

// Configures and establishes a database connection. Initializes the queries with the connection. Fails if the connection cannot be established or if the database is unreachable.
func (s *Service) connect(ctx context.Context, path string) error {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	conn.SetMaxOpenConns(1)

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return err
	}

	s.conn = conn
	s.queries = db.New(conn)

	s.logger.Info("connected to database", slog.Int("connections", conn.Stats().OpenConnections))

	return nil
}

// Shuts down the database connection. Called by Wails on application shutdown.
func (s *Service) ServiceShutdown() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return fmt.Errorf("database: close: %w", err)
		}

		s.logger.Info("database connection closed")
	}

	return nil
}

// Creates a new history repository instance. Returns nil if called before service startup and no database connection is established.
//
//wails:ignore
func (s *Service) NewHistory() *HistoryRepository {
	if s.conn == nil {
		return nil
	}

	return &HistoryRepository{
		queries: s.queries,
	}
}

// Creates a new language repository instance. Returns nil if called before service startup and no database connection is established.
//
//wails:ignore
func (s *Service) NewLanguage() *LanguageRepository {
	if s.conn == nil {
		return nil
	}

	return &LanguageRepository{
		conn:    s.conn,
		logger:  s.logger.With("repository", "language"),
		queries: s.queries,
	}
}

// Creates a new provider repository instance. Returns nil if called before service startup and no database connection is established.
//
//wails:ignore
func (s *Service) NewProvider() *ProviderRepository {
	if s.conn == nil {
		return nil
	}

	return &ProviderRepository{
		queries: s.queries,
	}
}
