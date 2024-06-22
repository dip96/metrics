package postgres

import (
	"context"
	"github.com/dip96/metrics/internal/config"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DB struct {
	Pool *PoolWrapper
}

// NewDB создает новое подключение к базе данных PostgreSQL.
func NewDB() (*DB, error) {
	cnf := config.LoadServer()
	pool, err := pgxpool.New(context.Background(), cnf.DatabaseDsn)
	if err != nil {
		return nil, err
	}
	wrappedPool := NewPoolWrapper(pool)
	return &DB{Pool: wrappedPool}, nil
}

func (d *DB) Get(name string) (metricModel.Metric, error) {
	err := d.Ping()
	if err != nil {
		return metricModel.Metric{}, err
	}

	sql := "SELECT name_metric, type, delta, value FROM metrics " +
		"WHERE name_metric = $1"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := d.Pool.pool.QueryRow(ctx, sql, name)

	var metrics metricModel.Metric

	err = row.Scan(
		&metrics.ID,
		&metrics.MType,
		&metrics.Delta,
		&metrics.Value,
	)

	if err != nil {
		return metricModel.Metric{}, err
	}

	return metrics, nil
}

func (d *DB) Set(metric metricModel.Metric) error {
	err := d.Ping()
	if err != nil {
		return err
	}

	//TODO использовать именованные параметры в запросе
	sql := "INSERT INTO metrics (name_metric, type, delta, value)" +
		"VALUES ($1,$2,$3,$4)" +
		"ON CONFLICT (name_metric)" +
		"DO UPDATE SET delta = excluded.delta, value = excluded.value"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = d.Pool.Exec(ctx, sql,
		metric.ID,
		metric.MType,
		metric.Delta,
		metric.Value,
	)

	if err != nil {
		return err
	}

	return nil
}

func (d *DB) SetAll(metrics map[string]metricModel.Metric) error {
	err := d.Ping()
	if err != nil {
		return err
	}

	ctx := context.Background()
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := "INSERT INTO metrics (name_metric, type, delta, value)" +
		"VALUES ($1,$2,$3,$4)" +
		"ON CONFLICT (name_metric)" +
		"DO UPDATE SET delta = excluded.delta, value = excluded.value"

	for _, metricValue := range metrics {
		_, err = tx.Exec(context.Background(), sql,
			metricValue.ID,
			metricValue.MType,
			metricValue.Delta,
			metricValue.Value,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (d *DB) GetAll() (map[string]metricModel.Metric, error) {
	err := d.Ping()
	if err != nil {
		return nil, err
	}

	sql := "SELECT name_metric, type, delta, value FROM metrics"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]metricModel.Metric)

	for rows.Next() {
		metric := metricModel.Metric{}
		err = rows.Scan(
			&metric.ID,
			&metric.MType,
			&metric.Delta,
			&metric.Value)
		if err != nil {
			return nil, err
		}

		metrics[metric.ID] = metric
	}

	return metrics, nil
}

func (d *DB) Clear() error {
	err := d.Ping()
	if err != nil {
		return err
	}

	sql := "TRUNCATE TABLE metrics"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = d.Pool.Exec(ctx, sql)
	if err != nil {
		return err
	}

	return nil
}

// Ping проверяет соединение с базой данных PostgreSQL.
func (d *DB) Ping() error {
	pingCtx := context.Background()

	err := d.Pool.Ping(pingCtx)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) Close() {
	d.Pool.Close()
}
