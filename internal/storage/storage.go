package storage

import "github.com/dip96/metrics/internal/model/metric"

var Storage StorageInterface

// TODO имплеменить в файл storage/files
type StorageInterface interface {
	Get(name string) (metric.Metric, error)
	Set(metric metric.Metric) error
	GetAll() (map[string]metric.Metric, error)
	SetAll(map[string]metric.Metric) error
	Clear() error
}
