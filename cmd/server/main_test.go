package main

import (
	"github.com/dip96/metrics/internal/model/metric"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddMetric(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем хендлер AddMetric
	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/42.0", nil)

	rec := httptest.NewRecorder()

	// Вызываем хендлер
	e.ServeHTTP(rec, req)

	// Проверяем статус-код ответа
	assert.Equal(t, http.StatusOK, rec.Code)

	// Получаем добавленную метрику
	metric, err := storage.Storage.Get("test_metric")
	require.NoError(t, err)

	// Проверяем тип и значение метрики
	assert.Equal(t, metricModel.MetricTypeGauge, metric.MType)
	assert.Equal(t, 42.0, *metric.Value)
}

func TestGetMetric(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем хендлер getMetric
	e.GET("/value/:name_metric", getMetric)

	// Добавляем тестовую метрику
	metric := metric.Metric{
		ID:             "test_metric",
		MType:          metric.MetricTypeGauge,
		Value:          Float64Ptr(42.0),
		FullValueGauge: "42",
		Delta:          nil,
	}
	err := storage.Storage.Set(metric)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/value/test_metric", nil)
	rec := httptest.NewRecorder()

	// Вызываем хендлер
	e.ServeHTTP(rec, req)

	// Проверяем статус-код ответа
	assert.Equal(t, http.StatusOK, rec.Code)

	// Проверяем значение метрики в ответе
	assert.Equal(t, "42", rec.Body.String())
}

func TestGetAllMetrics(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем хендлер getAllMetrics
	e.GET("/", getAllMetrics)

	// Добавляем тестовые метрики
	metric1 := metric.Metric{
		ID:    "test_metric_1",
		MType: metric.MetricTypeGauge,
		Value: Float64Ptr(42.0),
		Delta: nil,
	}
	metric2 := metric.Metric{
		ID:    "test_metric_2",
		MType: metric.MetricTypeCounter,
		Value: nil,
		Delta: Int64Ptr(100),
	}
	err := storage.Storage.Set(metric1)
	require.NoError(t, err)
	err = storage.Storage.Set(metric2)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Вызываем хендлер
	e.ServeHTTP(rec, req)

	// Проверяем статус-код ответа
	assert.Equal(t, http.StatusOK, rec.Code)

	// Проверяем, что в ответе содержатся добавленные метрики
	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "test_metric_1: 42")
	assert.Contains(t, string(body), "test_metric_2: 100")
}

// Вспомогательная функция для создания указателя на float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// Вспомогательная функция для создания указателя на int64
func Int64Ptr(i int64) *int64 {
	return &i
}
