package metric

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metric struct {
	ID             string     `json:"id"`              // имя метрики
	MType          MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta          *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value          *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
	FullValueGauge string     //float64 обрезает нули
}
