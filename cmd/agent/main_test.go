package main

import (
	"strings"
	"testing"

	"github.com/dip96/metrics/internal/model/metric"
	"github.com/stretchr/testify/assert"
)

func TestCollectMetrics(t *testing.T) {
	// Проверяем, что функция возвращает срез метрик
	metrics := collectMetrics(1)
	assert.NotEmpty(t, metrics)

	// Проверяем наличие счетчика PollCount
	var pollCountFound bool
	for _, m := range metrics {
		if m.ID == "PollCount" && m.MType == metric.MetricTypeCounter {
			pollCountFound = true
			break
		}
	}
	assert.True(t, pollCountFound)

	// Проверяем наличие метрик gauge
	var gaugesFound bool
	for _, m := range metrics {
		if m.MType == metric.MetricTypeGauge {
			gaugesFound = true
			break
		}
	}
	assert.True(t, gaugesFound)
}

func TestCollectGopsutilMetrics(t *testing.T) {
	metrics := collectGopsutilMetrics()
	assert.NotEmpty(t, metrics)

	// Проверяем наличие метрик памяти
	var memMetricsFound bool
	for _, m := range metrics {
		if m.ID == "TotalMemory" || m.ID == "FreeMemory" {
			memMetricsFound = true
			break
		}
	}
	assert.True(t, memMetricsFound)

	// Проверяем наличие метрик использования CPU
	var cpuMetricsFound bool
	for _, m := range metrics {
		if strings.HasPrefix(m.ID, "CPUutilization") {
			cpuMetricsFound = true
			break
		}
	}
	assert.True(t, cpuMetricsFound)
}

func TestMergeMetrics(t *testing.T) {
	metricsChan := make(chan []metric.Metric)
	gopsutilMetricsChan := make(chan []metric.Metric)

	mergedChan := mergeMetrics(metricsChan, gopsutilMetricsChan)

	// Отправляем данные в каналы
	go func() {
		metricsChan <- []metric.Metric{{ID: "metric1"}}
		gopsutilMetricsChan <- []metric.Metric{{ID: "metric2"}}
		close(metricsChan)
		close(gopsutilMetricsChan)
	}()

	// Получаем данные из объединенного канала
	merged := []metric.Metric{}
	for metrics := range mergedChan {
		merged = append(merged, metrics...)
	}

	// Проверяем, что объединенный канал содержит все метрики
	assert.Equal(t, 2, len(merged))
	assert.Contains(t, merged, metric.Metric{ID: "metric1"})
	assert.Contains(t, merged, metric.Metric{ID: "metric2"})
}

func TestCreateMetricFromFloat64(t *testing.T) {
	name := "float_metric"
	value := 42.0
	m := createMetricFromFloat64(name, metric.MetricTypeGauge, value)
	assert.Equal(t, name, m.ID)
	assert.Equal(t, metric.MetricTypeGauge, m.MType)
	assert.Equal(t, value, *m.Value)
}

func TestCreateMetricFromUint64(t *testing.T) {
	name := "uint64_metric"
	value := uint64(42)
	m := createMetricFromUint64(name, metric.MetricTypeGauge, value)
	assert.Equal(t, name, m.ID)
	assert.Equal(t, metric.MetricTypeGauge, m.MType)
	assert.Equal(t, float64(value), *m.Value)
}

func TestCreateMetricFromInt64(t *testing.T) {
	name := "int64_metric"
	value := int64(42)
	m := createMetricFromInt64(name, metric.MetricTypeCounter, value)
	assert.Equal(t, name, m.ID)
	assert.Equal(t, metric.MetricTypeCounter, m.MType)
	assert.Equal(t, value, *m.Delta)
}

func TestCreateMetricFromUint32(t *testing.T) {
	name := "uint32_metric"
	value := uint32(42)
	m := createMetricFromUint32(name, metric.MetricTypeGauge, value)
	assert.Equal(t, name, m.ID)
	assert.Equal(t, metric.MetricTypeGauge, m.MType)
	assert.Equal(t, float64(value), *m.Value)
}
