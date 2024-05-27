// Package metric содержит типы и структуры для работы с метриками.
package metric

// MetricType представляет собой тип метрики.
type MetricType string

const (
	// MetricTypeGauge - тип метрики для измерения значений на определенный момент времени.
	MetricTypeGauge MetricType = "gauge"
	// MetricTypeCounter - тип метрики для подсчета количества событий или значений в течение определенного периода времени.
	MetricTypeCounter MetricType = "counter"
)

// Metric представляет собой структуру для хранения информации о метрике.
type Metric struct {
	// ID - уникальный идентификатор метрики.
	ID string `json:"id"`
	// MType - тип метрики (gauge или counter).
	MType MetricType `json:"type"`
	// Delta - значение метрики в случае передачи counter.
	// Используется только для MetricTypeCounter.
	Delta *int64 `json:"delta,omitempty"`
	// Value - значение метрики в случае передачи gauge.
	// Используется только для MetricTypeGauge.
	Value *float64 `json:"value,omitempty"`
	// FullValueGauge - строковое представление значения метрики типа gauge с сохранением всех десятичных знаков после запятой.
	FullValueGauge string
}
