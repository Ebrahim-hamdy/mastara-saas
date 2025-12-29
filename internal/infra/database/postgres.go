// Package database provides the infrastructure for connecting to and interacting with the PostgreSQL database.
// It encapsulates the pgx connection pool and offers methods for health checks and graceful shutdowns.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Provider holds the active database connection pool.
// It is the central point for all database interactions.
type Provider struct {
	Pool *pgxpool.Pool
}

// NewProvider creates and returns a new database provider.
// It initializes the connection pool based on the provided configuration and performs
// a health check to ensure the database is reachable before returning.
// It will return a non-nil error if the connection cannot be established.
func NewProvider(cfg config.DatabaseConfig) (*Provider, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database on startup: %w", err)
	}

	log.Info().Msg("Database connection pool established successfully.")

	return &Provider{Pool: pool}, nil
}

// HealthCheck performs a simple query to verify the database connection is alive.
func (p *Provider) HealthCheck(ctx context.Context) error {
	if err := p.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Close gracefully terminates the database connection pool.
func (p *Provider) Close() {
	log.Info().Msg("Closing database connection pool.")
	p.Pool.Close()
}
