package storage

import "github.com/dip96/metrics/internal/model/metric"

var Storage StorageInterface

type StorageInterface interface {
	Get(name string) (metric.Metric, error)
	Set(name string, metric metric.Metric)
	GetAll() (map[string]metric.Metric, error)
}
