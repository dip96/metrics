package postgres

import (
	"context"
	"github.com/dip96/metrics/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB() (*DB, error) {
	cnf := config.LoadServer()
	pool, err := pgxpool.New(context.Background(), cnf.DatabaseDsn)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := d.Pool.Ping(pingCtx)
	if err != nil {
		return err
	}

	return nil
}
