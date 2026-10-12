package store

import (
	"database/sql"
	"errors"
)

// Returned in the chain when a lookup finds no row. Check it with errors.Is.
var ErrNotFound = errors.New("not found")

// Replaces sql.ErrNoRows with ErrNotFound, so callers never need database/sql.
func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
