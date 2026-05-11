package loggingmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	errNilDBForExecutor     = errors.New("loggingmanager transaction executor database is required")
	errNilTransactionFromDB = errors.New("loggingmanager writer database returned nil transaction")
)

// TransactionExecutor executes prepared transaction SQL against persistent storage.
type TransactionExecutor interface {
	// Execute runs one transaction unit.
	Execute(ctx context.Context, txSQL TransactionSQL) error
}

type sqlTransaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type sqlTransactionDB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	SetMaxOpenConns(n int)
}

// SQLTransactionExecutor executes transactions using a *sql.DB-backed adapter.
type SQLTransactionExecutor struct {
	db sqlTransactionDB
}

// NewSQLTransactionExecutor builds an SQL-backed transaction executor.
func NewSQLTransactionExecutor(db *sql.DB) (*SQLTransactionExecutor, error) {
	if db == nil {
		return nil, errNilDBForExecutor
	}

	db.SetMaxOpenConns(1)
	return &SQLTransactionExecutor{db: db}, nil
}

// Execute persists one transaction batch in a single SQL transaction.
func (executor *SQLTransactionExecutor) Execute(ctx context.Context, txSQL TransactionSQL) error {
	tx, err := executor.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if tx == nil {
		return errNilTransactionFromDB
	}

	return executeWithTransaction(ctx, tx, txSQL)
}

func executeWithTransaction(ctx context.Context, tx sqlTransaction, txSQL TransactionSQL) error {
	rollbackNeeded := true
	defer func() {
		if rollbackNeeded {
			_ = tx.Rollback()
		}
	}()

	for _, query := range txSQL.Queries {
		if _, execErr := tx.ExecContext(ctx, query.SQL, query.Args...); execErr != nil {
			return fmt.Errorf("exec query %q: %w", query.SQL, execErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	rollbackNeeded = false
	return nil
}
