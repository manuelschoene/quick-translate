package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"quick-translate/internal/store/db"
)

// Wraps queries in a transaction. Commits when fn succeeds, rolls back otherwise, also on a panic. Use the argument of fn directly, not the instance provided to this function.
func transaction(ctx context.Context, conn *sql.DB, queries *db.Queries, fn func(*db.Queries) error) error {
	return transactionRaw(ctx, conn, func(tx *sql.Tx) error {
		return fn(queries.WithTx(tx))
	})
}

// Wraps queries in a transaction. Commits when fn succeeds, rolls back otherwise, also on a panic. Use the *sql.Tx directly and not the *sql.DB.
func transactionRaw(ctx context.Context, conn *sql.DB, fn func(*sql.Tx) error) (err error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// After Commit this is a no-op returning ErrTxDone.
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
	}()

	// Not wrapped: fn knows best what it was doing.
	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
