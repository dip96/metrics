package metric_test

import (
	"testing"

	"github.com/dip96/metrics/internal/model/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetric_GetValueForDisplay(t *testing.T) {
	t.Run("counter metric", func(t *testing.T) {
		m := metric.Metric{
			MType: metric.MetricTypeCounter,
			Delta: Int64Ptr(10),
		}

		value, err := m.GetValueForDisplay()
		require.NoError(t, err)
		assert.Equal(t, "10", value)
	})

	t.Run("gauge metric", func(t *testing.T) {
		m := metric.Metric{
			MType:          metric.MetricTypeGauge,
			Value:          Float64Ptr(3.14),
			FullValueGauge: "3.140000",
		}

		value, err := m.GetValueForDisplay()
		require.NoError(t, err)
		assert.Equal(t, "3.140000", value)
	})

	t.Run("invalid metric type", func(t *testing.T) {
		m := metric.Metric{
			MType: "invalid",
		}

		_, err := m.GetValueForDisplay()
		assert.Error(t, err)
	})
}

func TestMetric_GetValue(t *testing.T) {
	t.Run("counter metric", func(t *testing.T) {
		m := metric.Metric{
			MType: metric.MetricTypeCounter,
			Delta: Int64Ptr(10),
		}

		value, err := m.GetValue()
		require.NoError(t, err)
		assert.Equal(t, "10", value)
	})

	t.Run("gauge metric", func(t *testing.T) {
		m := metric.Metric{
			MType: metric.MetricTypeGauge,
			Value: Float64Ptr(3.14),
		}

		value, err := m.GetValue()
		require.NoError(t, err)
		assert.Equal(t, "3.140000", value)
	})

	t.Run("invalid metric type", func(t *testing.T) {
		m := metric.Metric{
			MType: "invalid",
		}

		_, err := m.GetValue()
		assert.Error(t, err)
	})
}

// Вспомогательная функция для создания указателя на float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// Вспомогательная функция для создания указателя на int64
func Int64Ptr(i int64) *int64 {
	return &i
}
