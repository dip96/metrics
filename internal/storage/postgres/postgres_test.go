package postgres

import (
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewDB(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, err := NewDB()
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.Pool)
	})
}

func TestGet(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	t.Run("success", func(t *testing.T) {
		metric := metricModel.Metric{
			ID:    "test_metric",
			MType: metricModel.MetricTypeGauge,
			Value: Float64Ptr(42),
		}
		err = db.Set(metric)
		require.NoError(t, err)

		result, err := db.Get("test_metric")
		require.NoError(t, err)
		assert.Equal(t, metric, result)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := db.Get("non_existing_metric")
		assert.Error(t, err)
	})
}

func TestSet(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	t.Run("success", func(t *testing.T) {
		metric := metricModel.Metric{
			ID:    "test_metric",
			MType: metricModel.MetricTypeGauge,
			Value: Float64Ptr(42),
		}

		err = db.Set(metric)
		require.NoError(t, err)

		result, err := db.Get("test_metric")
		require.NoError(t, err)
		assert.Equal(t, metric, result)
	})

	t.Run("update existing", func(t *testing.T) {
		metric := metricModel.Metric{
			ID:    "test_metric",
			MType: metricModel.MetricTypeGauge,
			Value: Float64Ptr(43),
		}

		err = db.Set(metric)
		require.NoError(t, err)

		result, err := db.Get("test_metric")
		require.NoError(t, err)
		assert.Equal(t, metric, result)
	})
}

func TestSetAll(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	db.Clear()

	t.Run("success", func(t *testing.T) {
		metrics := map[string]metricModel.Metric{
			"metric1": {
				ID:    "metric1",
				MType: metricModel.MetricTypeGauge,
				Value: Float64Ptr(42),
			},
			"metric2": {
				ID:    "metric2",
				MType: metricModel.MetricTypeCounter,
				Delta: Int64Ptr(1),
			},
		}

		err = db.SetAll(metrics)
		require.NoError(t, err)

		result, err := db.GetAll()
		require.NoError(t, err)
		assert.Equal(t, metrics, result)
	})
}

func TestGetAll(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	t.Run("success", func(t *testing.T) {
		metrics := map[string]metricModel.Metric{
			"metric1": {
				ID:    "metric1",
				MType: metricModel.MetricTypeGauge,
				Value: Float64Ptr(0),
			},
			"metric2": {
				ID:    "metric2",
				MType: metricModel.MetricTypeCounter,
				Delta: Int64Ptr(1),
			},
		}

		err = db.SetAll(metrics)
		require.NoError(t, err)

		result, err := db.GetAll()
		require.NoError(t, err)
		assert.Equal(t, metrics, result)
	})

	t.Run("empty", func(t *testing.T) {
		db.Clear()
		result, err := db.GetAll()
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestPing(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	t.Run("success", func(t *testing.T) {
		err = db.Ping()
		assert.NoError(t, err)
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
