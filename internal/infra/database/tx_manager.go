package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/shared/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type pgxTxManager struct {
	pool *pgxpool.Pool
}

var _ database.TxManager = (*pgxTxManager)(nil)

func NewTxManager(pool *pgxpool.Pool) database.TxManager {
	return &pgxTxManager{pool: pool}
}

type auditContextPayload struct {
	UserID   string `json:"user_id"`
	ClinicID string `json:"clinic_id"`
}

func (m *pgxTxManager) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx_manager: failed to begin transaction: %w", err)
	}

	// 1. SAFETY NET (DEFER)
	// Only responsible for Rollback in case of Panic or Early Exit.
	defer func() {
		if p := recover(); p != nil {
			// Panic Recovery
			_ = tx.Rollback(ctx)
			log.Error().Msgf("panic recovered in transaction: %v", p)
			panic(p)
		} else {
			// Blind Rollback (Safe to call even if committed)
			// We discard the error because if it was already committed,
			// Rollback() just returns pgx.ErrTxClosed (which is fine).
			_ = tx.Rollback(ctx)
		}
	}()

	// 2. AUDIT CONTEXT INJECTION
	// Strict check: Only proceed if system user, or if injection works.
	if payload, authErr := middleware.GetAuthPayload(ctx); authErr == nil {
		auditJSON, err := json.Marshal(auditContextPayload{
			UserID:   payload.UserID.String(),
			ClinicID: payload.ClinicID.String(),
		})
		if err != nil {
			return fmt.Errorf("tx_manager: failed to marshal audit context: %w", err)
		}

		// STRICT SECURITY: Do not swallow error here.
		if _, err := tx.Exec(ctx, "SET LOCAL app.audit_context = $1", string(auditJSON)); err != nil {
			return fmt.Errorf("tx_manager: failed to set audit context (audit log integrity risk): %w", err)
		}
	} else {
		// Log warning: We are running without an audit user (System background job?)
		log.Trace().Msg("tx_manager: executing transaction without user context")
	}

	// 3. EXECUTE BUSINESS LOGIC
	if err := fn(tx); err != nil {
		return err // Returns original error, Defer triggers Rollback
	}

	// 4. EXPLICIT COMMIT
	// We handle the happy path explicitly for readability
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx_manager: failed to commit transaction: %w", err)
	}

	return nil
}
