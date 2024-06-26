package mem

import (
	"errors"
	"github.com/dip96/metrics/internal/model/metric"
	log "github.com/sirupsen/logrus"
)

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

func (m *Storage) Set(metric metric.Metric) error {
	m.metrics[metric.ID] = metric
	return nil
}

func (m *Storage) GetAll() (map[string]metric.Metric, error) {
	return m.metrics, nil
}

func (m *Storage) SetAll(metrics map[string]metric.Metric) error {
	for _, metricValue := range metrics {
		err := m.Set(metricValue)
		if err != nil {
			log.Printf("Error setting metric %s: %v", metricValue.ID, err)
		}
	}

	return nil
}

func (m *Storage) Clear() error {
	m.metrics = make(map[string]metric.Metric)
	return nil
}

func (m *Storage) Close() {

}

// NewStorage - конструктор для создания нового экземпляра Storage
func NewStorage() *Storage {
	return &Storage{
		metrics: make(map[string]metric.Metric),
	}
}
