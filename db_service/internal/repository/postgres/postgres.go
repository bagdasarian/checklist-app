package postgres

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
	"github.com/bagdasarian/checklist-app/db_service/config"
)

type Postgres struct {
    Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*Postgres, error) {
    dbURL := cfg.GetDBURL()
    
    poolConfig, err := pgxpool.ParseConfig(dbURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse db config: %w", err)
    }

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Printf("Successfully connected to PostgreSQL at %s:%s", cfg.DB.Host, cfg.DB.Port)
    return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
    if p.Pool != nil {
        p.Pool.Close()
    }
}