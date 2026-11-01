package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"slices"
	"strconv"
	"strings"
)

type migration struct {
	version  int
	filename string
}

// Migrates the database to the latest version by applying all pending migrations in order. Returns an error if any migration fails or the database belongs to a newer version of the application. The migrations are read from the provided filesystem and executed on the provided database connection.
func migrate(ctx context.Context, fsys fs.FS, conn *sql.DB, logger *slog.Logger) error {
	logger = logger.With("step", "migration")

	fsys, err := fs.Sub(fsys, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	migrations, err := parseMigrations(fsys)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// ver is 1-based, migrations slice is 0-based
	// Zero value assignment is fine as it will allow the first transaction to check whether a migration is needed
	var ver int

	for ver < len(migrations) {
		err := transactionRaw(ctx, conn, func(tx *sql.Tx) error {
			if err := userVersion(ctx, tx, len(migrations), &ver); err != nil {
				return err
			}

			if ver == len(migrations) {
				logger.Info("no new migrations found", slog.Int("version", ver))
				return nil
			}

			m := migrations[ver]
			if err := applyMigration(ctx, tx, m, fsys, logger); err != nil {
				return fmt.Errorf("apply %s: %w", m.filename, err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("migrate: %w", err)
		}

		ver++ // Increment the version after a successful migration to prevent the unnecessary check in the next iteration
	}

	return nil
}

// Parses the migration files inside the filesystem and returns the migrations sorted by version. Returns an error if the migration versions are not continuous or if any migration file has an invalid filename format.
func parseMigrations(fsys fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no migration files found")
	}

	migrations := make([]migration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			return nil, fmt.Errorf("%s: not a .sql file", entry.Name())
		}

		ver, err := verFromFilename(entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration{version: ver, filename: entry.Name()})
	}

	slices.SortFunc(migrations, func(a, b migration) int { return a.version - b.version })

	if err := isContinuous(migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

// Extracts the version number from a migration filename. The filename must be in the format "X_description.sql" where X is a positive integer. Returns an error if the filename does not match this format or if the version number is not a positive integer.
func verFromFilename(filename string) (int, error) {
	base := path.Base(filename)

	before, _, ok := strings.Cut(base, "_")
	if !ok {
		return 0, fmt.Errorf("%s: missing version prefix", base)
	}

	ver, err := strconv.Atoi(before)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid version number", base)
	}

	if ver < 1 {
		return 0, fmt.Errorf("%s: version must be greater than 0", base)
	}

	return ver, nil
}

// Checks if the migration versions are continuous, i.e., there are no gaps or duplicates in the version numbers. Assumes that the migrations are sorted by version number.
func isContinuous(migrations []migration) error {
	current := 1

	for _, m := range migrations {
		if m.version < current {
			return fmt.Errorf("duplicate migration version %d", m.version)
		}

		if m.version > current {
			return fmt.Errorf("missing migration version %d", current)
		}

		current++
	}

	return nil
}

// Queries the user_version pragma and checks if it matches the expected migration versions. Returns an error if the database version is higher than the latest migration version, indicating that the application is outdated or the database is corrupted. The queried version is stored in the provided ver pointer.
func userVersion(ctx context.Context, tx *sql.Tx, mlen int, ver *int) error {
	err := tx.QueryRowContext(ctx, "PRAGMA user_version;").Scan(ver)
	if err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	if *ver > mlen {
		return fmt.Errorf("database version %d is newer than the latest migration %d: use a newer version of the application or purge your data", *ver, mlen)
	}

	return nil
}

// Applies a single migration by executing the SQL statements in the migration file and updating the user_version pragma. Returns an error if any of the SQL statements fail or if the user_version pragma cannot be updated.
func applyMigration(ctx context.Context, tx *sql.Tx, m migration, fsys fs.FS, logger *slog.Logger) error {
	logger.Info("applying migration", slog.Int("version", m.version))

	b, err := fs.ReadFile(fsys, m.filename)
	if err != nil {
		return err
	}

	// The driver's error already names the failing statement.
	if _, err := tx.ExecContext(ctx, string(b)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d;", m.version)); err != nil {
		return fmt.Errorf("set user_version: %w", err)
	}

	logger.Info("migration applied successfully", slog.Int("version", m.version))
	return nil
}
