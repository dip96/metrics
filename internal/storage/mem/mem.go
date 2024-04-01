package mem

import (
	"errors"
	"github.com/dip96/metrics/internal/model/metric"
)

var MemStorage *Storage

type Storage struct {
	metrics map[string]metric.Metric
}

func (m *Storage) Get(name string) (metric.Metric, error) {
	value, ok := m.metrics[name]

	if ok {
		return value, nil
	}

	return metric.Metric{}, errors.New("the metric was not found")
}

func (m *Storage) Set(name string, metric metric.Metric) {
	m.metrics[name] = metric
}

func (m *Storage) GetAll() (map[string]metric.Metric, error) {
	return m.metrics, nil
}

// NewStorage - конструктор для создания нового экземпляра Storage
func NewStorage() *Storage {
	return &Storage{
		metrics: make(map[string]metric.Metric),
	}
}
