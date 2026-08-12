package history

import (
	"database/sql"
	"fmt"
	"quick-translate/internal/models"

	_ "github.com/glebarez/go-sqlite"
)

// The columns of the history table in the order they are scanned into a translation.
const columns = `id, source, target, text, translation, detected_source, provider, created_at`

type db struct {
	conn *sql.DB
}

// Opens the database at the given path and creates the history table if it does not exist yet. Returns an error if the database can not be opened or migrated.
func newDb(path string) (*db, error) {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("Could not open the database: %w", err)
	}

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("Could not ping the database: %w", err)
	}

	db := &db{conn: conn}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *db) close() error {
	return db.conn.Close()
}

func (db *db) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source VARCHAR(255) NOT NULL,
		target VARCHAR(255) NOT NULL,
		text TEXT NOT NULL,
		translation TEXT NOT NULL,
		detected_source VARCHAR(255),
		provider VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("Could not create history table: %w", err)
	}

	return nil
}

// Appends the translation to the history and returns the ID it was stored with. Entries are never updated or removed, so the ID identifies the translation for the lifetime of the history.
func (db *db) saveTranslation(translation *models.Translation) (int64, error) {
	query := `
	INSERT INTO history (source, target, text, translation, detected_source, provider)
	VALUES (?, ?, ?, ?, ?, ?);
	`

	result, err := db.conn.Exec(query, translation.Source, translation.Target, translation.Text, translation.Translation, translation.DetectedSource, translation.Provider)
	if err != nil {
		return 0, fmt.Errorf("Could not save translation to history: %w", err)
	}

	return result.LastInsertId()
}

// Returns the newest translation that is older than the given ID. Only the newest entries inside the given window are reachable, older ones stay stored but are not returned. An ID lower than one is treated as the position behind the newest entry, so the newest translation is returned. Returns nil if there is no older translation inside the window.
func (db *db) previousTranslation(id int, window int) (*models.Translation, error) {
	query := fmt.Sprintf(`SELECT %s FROM history WHERE id < ? AND id > (SELECT MAX(id) FROM history) - ? ORDER BY id DESC LIMIT 1;`, columns)
	args := []any{id, window}

	if id < 1 {
		query = fmt.Sprintf(`SELECT %s FROM history ORDER BY id DESC LIMIT 1;`, columns)
		args = nil
	}

	translation, err := scanTranslation(db.conn.QueryRow(query, args...))
	if err != nil {
		return nil, fmt.Errorf("Could not get the previous translation: %w", err)
	}

	return translation, nil
}

// Returns the oldest translation that is newer than the given ID. An ID lower than one is treated as the position behind the newest entry, so no translation is returned. Returns nil if there is no newer translation.
func (db *db) nextTranslation(id int) (*models.Translation, error) {
	if id < 1 {
		return nil, nil
	}

	query := fmt.Sprintf(`SELECT %s FROM history WHERE id > ? ORDER BY id ASC LIMIT 1;`, columns)

	translation, err := scanTranslation(db.conn.QueryRow(query, id))
	if err != nil {
		return nil, fmt.Errorf("Could not get the next translation: %w", err)
	}

	return translation, nil
}

// Checks if a translation older than the given ID exists inside the given window. An ID lower than one is treated as the position behind the newest entry, so any stored translation counts as a previous one.
func (db *db) hasPreviousTranslation(id int, window int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM history WHERE id < ? AND id > (SELECT MAX(id) FROM history) - ?);`
	args := []any{id, window}

	if id < 1 {
		query = `SELECT EXISTS(SELECT 1 FROM history);`
		args = nil
	}

	var exists bool
	if err := db.conn.QueryRow(query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("Could not check for a previous translation: %w", err)
	}

	return exists, nil
}

// Checks if a translation newer than the given ID exists. An ID lower than one is treated as the position behind the newest entry, so there is never a next translation.
func (db *db) hasNextTranslation(id int) (bool, error) {
	if id < 1 {
		return false, nil
	}

	query := `SELECT EXISTS(SELECT 1 FROM history WHERE id > ?);`

	var exists bool
	if err := db.conn.QueryRow(query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("Could not check for a next translation: %w", err)
	}

	return exists, nil
}

// Scans a single row into a translation. Returns nil without an error if the row is empty.
func scanTranslation(row *sql.Row) (*models.Translation, error) {
	var (
		translation    models.Translation
		detectedSource sql.NullString
		createdAt      sql.NullTime
	)

	err := row.Scan(
		&translation.ID,
		&translation.Source,
		&translation.Target,
		&translation.Text,
		&translation.Translation,
		&detectedSource,
		&translation.Provider,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No translation found
		}
		return nil, err
	}

	translation.DetectedSource = detectedSource.String
	translation.CreatedAt = createdAt.Time

	return &translation, nil
}
