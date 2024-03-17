package postgres

import (
	"context"
	"github.com/dip96/metrics/internal/config"
	metricModel "github.com/dip96/metrics/internal/model/metric"
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

func (d *DB) Get(name string) (metricModel.Metric, error) {
	err := d.Ping()

	if err != nil {
		panic(err)
	}

	sql := "SELECT name_metric, type, delta, value FROM metrics " +
		"WHERE name_metric = $1"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := d.Pool.QueryRow(ctx, sql, name)

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

func (d *DB) Set(metric metricModel.Metric) {
	err := d.Ping()

	if err != nil {
		panic(err)
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
		panic(err)
	}
}

func (d *DB) GetAll() (map[string]metricModel.Metric, error) {
	err := d.Ping()

	if err != nil {
		panic(err)
	}

	sql := "SELECT name_metric, type, delta, value FROM metrics"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, sql)

	if err != nil {
		return nil, err
	}

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

func (d *DB) CreateTable() error {
	err := d.Ping()

	if err != nil {
		panic(err)
	}

	sql := "CREATE TABLE IF NOT EXISTS metrics (" +
		"id smallserial PRIMARY KEY, " +
		"name_metric CHARACTER VARYING(100) UNIQUE, " +
		"type CHARACTER VARYING(30) NOT NULL, " +
		"delta integer, " +
		"value double precision " +
		")"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = d.Pool.Exec(ctx, sql)

	if err != nil {
		panic(err)
	}

	indexNameMetricSQL := "CREATE INDEX IF NOT EXISTS idx_metrics_name_metric ON metrics (name_metric);"
	_, err = d.Pool.Exec(ctx, indexNameMetricSQL)

	if err != nil {
		panic(err)
	}

	indexTypeSQL := "CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics (type);"
	_, err = d.Pool.Exec(ctx, indexTypeSQL)

	if err != nil {
		panic(err)
	}

	return nil
}

func (d *DB) Ping() error {
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := d.Pool.Ping(pingCtx)
	if err != nil {
		return err
	}

	return nil
}
