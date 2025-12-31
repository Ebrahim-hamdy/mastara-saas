package service

import (
	"context"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/shared/database"
	"github.com/jackc/pgx/v5"
)

// BaseService wraps dependencies common to all Domain Services.
// Embedding this in your Service structs gives them transaction powers.
type BaseService struct {
	Tx database.TxManager
}

// RunInTransaction wraps the atomic business operation.
// This allows you to combine multiple Repo calls into one Atomic Unit of Work.
func (s *BaseService) RunInTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return s.Tx.ExecTx(ctx, fn)
}
