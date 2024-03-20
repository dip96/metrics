package postgres

import (
	"context"
	"github.com/dip96/metrics/internal/retriable"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type PoolWrapper struct {
	pool *pgxpool.Pool
}

func NewPoolWrapper(pool *pgxpool.Pool) *PoolWrapper {
	return &PoolWrapper{pool: pool}
}

func (pw *PoolWrapper) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	//TODO вынести в отдельный метод логику повторным запросам
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		rows, err := pw.pool.Query(ctx, sql, args...)
		if err == nil {
			return rows, err
		}
		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
		retriable.CheckError(err)

		if !retriable.IsConnectionException(err) {
			return nil, err
		}

		time.Sleep(delay)
	}

	return nil, nil
}

func (pw *PoolWrapper) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	//TODO вынести в отдельный метод логику повторным запросам
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		tag, err := pw.pool.Exec(ctx, sql, arguments...)
		if err == nil {
			return tag, err
		}
		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
		retriable.CheckError(err)

		if !retriable.IsConnectionException(err) {
			return tag, err
		}

		time.Sleep(delay)
	}

	return pgconn.CommandTag{}, nil
}

func (pw *PoolWrapper) Ping(ctx context.Context) error {
	//TODO вынести в отдельный метод логику повторным запросам
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		err := pw.pool.Ping(ctx)
		if err == nil {
			return err
		}
		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
		retriable.CheckError(err)

		if !retriable.IsConnectionException(err) {
			return err
		}

		time.Sleep(delay)
	}

	return nil
}

func (pw *PoolWrapper) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return pw.pool.QueryRow(ctx, sql, args)
}

func (pw *PoolWrapper) Begin(ctx context.Context) (pgx.Tx, error) {
	//TODO вынести в отдельный метод логику повторным запросам
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		pgx, err := pw.pool.BeginTx(ctx, pgx.TxOptions{})
		if err == nil {
			return pgx, err
		}
		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
		retriable.CheckError(err)

		if !retriable.IsConnectionException(err) {
			return nil, err
		}

		time.Sleep(delay)
	}

	return nil, nil
}
