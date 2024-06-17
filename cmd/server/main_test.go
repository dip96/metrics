package main

import (
	"bytes"
	"encoding/json"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	postgresStorage "github.com/dip96/metrics/internal/storage/postgres"
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
	metric := metricModel.Metric{
		ID:             "test_metric",
		MType:          metricModel.MetricTypeGauge,
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
	metric1 := metricModel.Metric{
		ID:    "test_metric_1",
		MType: metricModel.MetricTypeGauge,
		Value: Float64Ptr(42.0),
		Delta: nil,
	}
	metric2 := metricModel.Metric{
		ID:    "test_metric_2",
		MType: metricModel.MetricTypeCounter,
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

func TestAddMetricV2(t *testing.T) {
	// Создаем новый экземпляр Echo
	e := echo.New()

	// Очищаем хранилище перед каждым тестом
	storage.Storage.Clear()

	t.Run("add gauge metric", func(t *testing.T) {
		gaugeMetric := metricModel.Metric{
			ID:    "GaugeMetric",
			MType: metricModel.MetricTypeGauge,
			Value: Float64Ptr(42.0),
			Delta: nil,
		}

		body, err := json.Marshal(gaugeMetric)
		require.NoError(t, err)

		e.POST("/update/", AddMetricV2)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Вызываем хендлер
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		resp := metricModel.Metric{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.NoError(t, err)

		assert.Equal(t, gaugeMetric.ID, resp.ID)
		assert.Equal(t, gaugeMetric.MType, resp.MType)
		assert.Equal(t, *gaugeMetric.Value, *resp.Value)
		assert.Nil(t, resp.Delta)
	})

	t.Run("add counter metric", func(t *testing.T) {
		counterMetric := metricModel.Metric{
			ID:    "CounterMetric",
			MType: metricModel.MetricTypeCounter,
			Value: nil,
			Delta: Int64Ptr(10),
		}

		body, err := json.Marshal(counterMetric)
		require.NoError(t, err)

		e.POST("/update/", AddMetricV2)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Вызываем хендлер
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		resp := metricModel.Metric{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, counterMetric, resp)
	})
}

// Mock DB object
type mockDB struct{}

func (m *mockDB) Ping() error {
	// Simulate successful ping
	return nil
}

func TestPing(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	db, err := postgresStorage.NewDB()
	assert.NoError(t, err)

	err = ping(e.NewContext(req, rec), db)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetOrDefault(t *testing.T) {
	tests := []struct {
		value        string
		defaultValue string
		expected     string
	}{
		{"value", "default", "value"},
		{"", "default", "default"},
		{"value", "", "value"},
		{"", "", ""},
	}

	for _, test := range tests {
		result := test.value
		if result != test.expected {
			t.Errorf("getOrDefault(%q, %q) = %q; want %q", test.value, test.defaultValue, result, test.expected)
		}
	}
}

// Вспомогательная функция для создания указателя на float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// Вспомогательная функция для создания указателя на int64
func Int64Ptr(i int64) *int64 {
	return &i
}
