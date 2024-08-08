package transaction

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/mikhailsoldatkin/platform_common/pkg/db"
	"github.com/mikhailsoldatkin/platform_common/pkg/db/pg"
)

type manager struct {
	db db.Transactor
}

// NewTransactionManager creates a new transaction manager that satisfies the db.TxManager interface.
func NewTransactionManager(db db.Transactor) db.TxManager {
	return &manager{db: db}
}

// transaction performs the user-provided handler within a transaction.
func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn db.Handler) (err error) {
	// If this is a nested transaction, skip starting a new transaction and execute the handler.
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	// Start a new transaction.
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	// Place the transaction in the context.
	ctx = pg.MakeContextTx(ctx, tx)

	// Set up a deferred function to handle commit or rollback of the transaction.
	defer func() {
		// Recover from panic if it occurs
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		// Rollback the transaction if there was an error
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "rollback error: %v", errRollback)
			}
			return
		}

		// If no error, commit the transaction
		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "transaction commit failed")
			}
		}
	}()

	// Execute the code inside the transaction.
	// If the function fails, it returns an error, and the deferred function performs a rollback,
	// otherwise the transaction is committed.
	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

// ReadCommitted performs the handler within a transaction with ReadCommitted isolation level.
func (m *manager) ReadCommitted(ctx context.Context, f db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}
