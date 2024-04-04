package postgres

import (
	"context"
	"errors"
	"fmt"
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
			return rows, nil
		}

		if !retriable.IsConnectionException(err) {
			return nil, err
		}

		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
		err = errors.Join(err, fmt.Errorf("retry %d: %w", attempt, err))
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

		if !retriable.IsConnectionException(err) {
			return err
		}

		time.Sleep(delay)
	}

	return nil
}

func (pw *PoolWrapper) Begin(ctx context.Context) (pgx.Tx, error) {
	//TODO вынести в отдельный метод логику повторным запросам
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		tx, err := pw.pool.BeginTx(ctx, pgx.TxOptions{})
		if err == nil {
			return tx, err
		}
		log.Printf("Err (attempt %d/%d): %v", attempt+1, len(retryDelays), err)

		if !retriable.IsConnectionException(err) {
			return nil, err
		}

		time.Sleep(delay)
	}

	return nil, nil
}

func (pw *PoolWrapper) Close() {
	pw.pool.Close()
}
