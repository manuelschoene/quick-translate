package history

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"quick-translate/internal/models"
	"sync"
)

const dbPath = "history.db"

// Returned by every function that needs the database while the history is disabled.
var ErrDisabled = errors.New("The history is disabled. Set 'history.max_entries' to a value greater than zero in the configuration file.")

type History struct {
	limit    int
	database *db
	mutex    sync.Mutex
}

// Creates a new history for the given limit of entries. A limit of zero or lower disables the history, so every function besides Enabled() and Close() returns ErrDisabled. The database connection is not opened yet, it is established with the first call that needs it.
func NewHistory(limit int) *History {
	return &History{
		limit:    limit,
		database: nil,
	}
}

// Checks if the history is enabled. A disabled history neither stores nor returns translations.
func (h *History) Enabled() bool {
	return h.limit > 0
}

// Appends the translation to the history and returns the ID it was stored under. The history is a log, so entries are never changed or removed and the ID stays valid for the lifetime of the history. Returns an error if the history is disabled, the database can not be reached or the translation can not be stored.
func (h *History) Append(translation *models.Translation) (int, error) {
	db, err := h.connection()
	if err != nil {
		return 0, err
	}

	id, err := db.saveTranslation(translation)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Returns the translation that was stored before the one with the given ID. Only the newest entries up to the limit are reachable, older ones stay in the log but are not returned. An ID lower than one is treated as the position behind the newest entry, so the newest translation is returned. Returns nil if the given translation is the oldest reachable one. Returns an error if the history is disabled, the database can not be reached or the translation can not be read.
func (h *History) Previous(id int) (*models.Translation, error) {
	db, err := h.connection()
	if err != nil {
		return nil, err
	}

	return db.previousTranslation(id, h.limit)
}

// Returns the translation that was stored after the one with the given ID. An ID lower than one is treated as the position behind the newest entry, so nothing is returned. Returns nil if the given translation is the newest one. Returns an error if the history is disabled, the database can not be reached or the translation can not be read.
func (h *History) Next(id int) (*models.Translation, error) {
	db, err := h.connection()
	if err != nil {
		return nil, err
	}

	return db.nextTranslation(id)
}

// Checks if a translation older than the one with the given ID is reachable, which means it is one of the newest entries up to the limit. An ID lower than one is treated as the position behind the newest entry, so every stored translation counts as a previous one. Returns an error if the history is disabled or the database can not be reached.
func (h *History) HasPrevious(id int) (bool, error) {
	db, err := h.connection()
	if err != nil {
		return false, err
	}

	return db.hasPreviousTranslation(id, h.limit)
}

// Checks if a translation newer than the one with the given ID exists. Returns an error if the history is disabled or the database can not be reached.
func (h *History) HasNext(id int) (bool, error) {
	db, err := h.connection()
	if err != nil {
		return false, err
	}

	return db.hasNextTranslation(id)
}

// Closes the database connection if one is established. The connection is reopened automatically on the next call to the history. Safe to call on a disabled history.
func (h *History) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.database == nil {
		return
	}

	err := h.database.close()
	if err != nil {
		fmt.Printf("Error closing database connection: %v\n", err)
	}

	h.database = nil
}

// Returns the open database connection and establishes it if the history is not connected yet. Returns ErrDisabled if the history is disabled.
func (h *History) connection() (*db, error) {
	if !h.Enabled() {
		return nil, ErrDisabled
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.database != nil {
		return h.database, nil
	}

	if err := h.initialize(); err != nil {
		return nil, err
	}

	return h.database, nil
}

// Opens the database connection and stores it for further calls. Does not check if a connection is already established.
func (h *History) initialize() error {
	path, err := filePath()
	if err != nil {
		return fmt.Errorf("Could not initialize the database: %w", err)
	}

	instance, err := newDb(path)
	if err != nil {
		return fmt.Errorf("Could not initialize the database: %w", err)
	}

	h.database = instance
	return nil
}

// Returns the path to the database file inside the user's config directory and creates the directory if it does not exist.
func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir = dir + "/quick-translate"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Path to the history directory doesn't exists. Creating directories: %s\n", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return filepath.Join(dir, dbPath), nil
}
