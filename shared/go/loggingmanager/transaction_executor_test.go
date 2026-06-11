package loggingmanager

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"lite-nas/shared/loggingmanager/query"
)

type fakeSQLTx struct {
	execErr   error
	commitErr error
}

type fakeSQLResult struct{}

func (fakeSQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (fakeSQLResult) RowsAffected() (int64, error) {
	return 1, nil
}

func (tx *fakeSQLTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	if tx.execErr != nil {
		return nil, tx.execErr
	}
	return fakeSQLResult{}, nil
}

func (tx *fakeSQLTx) Commit() error {
	return tx.commitErr
}

func (tx *fakeSQLTx) Rollback() error {
	return nil
}

func TestExecuteWithTransactionReturnsExecError(t *testing.T) {
	t.Parallel()

	tx := &fakeSQLTx{execErr: errors.New("exec failed")}
	err := executeWithTransaction(context.Background(), tx, query.TransactionSQL{
		Queries: []query.Query{{SQL: "INSERT INTO x VALUES(?)", Args: []any{1}}},
	})
	if err == nil {
		t.Fatal("expected executeWithTransaction() error")
	}
}

func TestExecuteWithTransactionReturnsCommitError(t *testing.T) {
	t.Parallel()

	tx := &fakeSQLTx{commitErr: errors.New("commit failed")}
	err := executeWithTransaction(context.Background(), tx, query.TransactionSQL{
		Queries: []query.Query{{SQL: "INSERT INTO x VALUES(?)", Args: []any{1}}},
	})
	if err == nil {
		t.Fatal("expected executeWithTransaction() error")
	}
}
