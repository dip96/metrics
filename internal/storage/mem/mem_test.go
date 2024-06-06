package mem_test

import (
	"testing"

	"github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		storage := mem.NewStorage()
		m := metric.Metric{ID: "test", MType: metric.MetricTypeGauge, Value: Float64Ptr(42.0)}
		err := storage.Set(m)
		require.NoError(t, err)

		// Получаем существующую метрику
		got, err := storage.Get("test")
		require.NoError(t, err)
		assert.Equal(t, m, got)

		// Получаем несуществующую метрику
		_, err = storage.Get("nonexistent")
		assert.Error(t, err)
	})

	t.Run("Set", func(t *testing.T) {
		storage := mem.NewStorage()
		m := metric.Metric{ID: "test", MType: metric.MetricTypeGauge, Value: Float64Ptr(42.0)}
		err := storage.Set(m)
		require.NoError(t, err)

		// Проверяем, что метрика была успешно добавлена
		got, err := storage.Get("test")
		require.NoError(t, err)
		assert.Equal(t, m, got)
	})

	t.Run("GetAll", func(t *testing.T) {
		storage := mem.NewStorage()
		metrics := map[string]metric.Metric{
			"test1": {ID: "test1", MType: metric.MetricTypeGauge, Value: Float64Ptr(42.0)},
			"test2": {ID: "test2", MType: metric.MetricTypeCounter, Delta: Int64Ptr(10)},
		}
		for _, m := range metrics {
			err := storage.Set(m)
			require.NoError(t, err)
		}

		got, err := storage.GetAll()
		require.NoError(t, err)
		assert.Equal(t, metrics, got)
	})

	t.Run("SetAll", func(t *testing.T) {
		storage := mem.NewStorage()
		metrics := map[string]metric.Metric{
			"test1": {ID: "test1", MType: metric.MetricTypeGauge, Value: Float64Ptr(42.0)},
			"test2": {ID: "test2", MType: metric.MetricTypeCounter, Delta: Int64Ptr(10)},
		}
		err := storage.SetAll(metrics)
		require.NoError(t, err)

		got, err := storage.GetAll()
		require.NoError(t, err)
		assert.Equal(t, metrics, got)
	})

	t.Run("Clear", func(t *testing.T) {
		storage := mem.NewStorage()
		m := metric.Metric{ID: "test", MType: metric.MetricTypeGauge, Value: Float64Ptr(42.0)}
		err := storage.Set(m)
		require.NoError(t, err)

		err = storage.Clear()
		require.NoError(t, err)

		// Проверяем, что хранилище очищено
		got, err := storage.GetAll()
		require.NoError(t, err)
		assert.Empty(t, got)
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
