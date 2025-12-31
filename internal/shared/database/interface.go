package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TxManager handles the transaction lifecycle.
type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// Querier is the Common Interface for both *pgxpool.Pool and pgx.Tx.
// This allows repositories to work with or without a transaction seamlessly.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
