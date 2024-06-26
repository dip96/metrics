package main

import (
	"fmt"
	"github.com/dip96/metrics/internal/model/metric"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
	"time"
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
	assert.Contains(t, merged[0].ID, "metric1")
	assert.Contains(t, merged[1].ID, "metric2")
}

// CollectMetrics - реализация метода CollectMetrics для мока
//func CollectMetricsMock(pollCount int64) []metricModel.Metric {
//	metric1 := metricModel.Metric{
//		ID:    "metric1",
//		MType: metricModel.MetricTypeGauge,
//		Value: Float64Ptr(float64(pollCount)),
//		Delta: nil,
//	}
//
//	metric2 := metricModel.Metric{
//		ID:    "metric2",
//		MType: metricModel.MetricTypeCounter,
//		Value: nil,
//		Delta: Int64Ptr(pollCount),
//	}
//
//	return []metricModel.Metric{metric1, metric2}
//}
//
//func TestCollectMetricsRoutine(t *testing.T) {
//	metricsChan := make(chan []metricModel.Metric, 1)
//	stop := make(chan struct{})
//
//	go collectMetricsRoutine(metricsChan, stop)
//
//	// Ждем первый набор метрик
//	metrics := <-metricsChan
//	require.Equal(t, []metricModel.Metric{
//		{ID: "metric1", MType: metricModel.MetricTypeGauge, Value: Float64Ptr(1.0)},
//		{ID: "metric2", MType: metricModel.MetricTypeCounter, Delta: Int64Ptr(1)},
//	}, metrics)
//
//	// Ждем второй набор метрик
//	metrics = <-metricsChan
//	require.Equal(t, []metricModel.Metric{
//		{ID: "metric1", MType: metricModel.MetricTypeGauge, Value: Float64Ptr(2.0)},
//		{ID: "metric2", MType: metricModel.MetricTypeCounter, Delta: Int64Ptr(2)},
//	}, metrics)
//
//	// Останавливаем горутину
//	close(stop)
//}

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

// MetricsSender интерфейс для отправки метрик.
type MetricsSender interface {
	send(metrics []metric.Metric)
}

func TestSendMetricsRoutine(t *testing.T) {
	//TODO доработать
	jobChan := make(chan metric.Metric, 1)
	stop := make(chan struct{})
	//sender := &fakeMetricsSender{}

	go sendMetricsRoutine(jobChan, stop)

	jobChan <- metric.Metric{
		ID:    "test_metric",
		MType: metric.MetricTypeGauge,
		Value: Float64Ptr(10.5),
	}

	time.Sleep(100 * time.Millisecond)

	close(stop)

	//assert.True(t, sender, "sendMetricsButch should have been called")
}
func TestCollectMetricsRoutine(t *testing.T) {
	metricsChan := make(chan []metric.Metric, 1)
	stop := make(chan struct{})

	go collectMetricsRoutine(metricsChan, stop)

	select {
	case metrics := <-metricsChan:
		assert.NotEmpty(t, metrics)
	case <-time.After(2 * time.Second): // Slightly more than the 1-second interval
		t.Fatal("Expected metrics, but got timeout")
	}

	close(stop)
}

func TestCollectGopsutilMetricsRoutine_Stop(t *testing.T) {
	gopsutilMetricsChan := make(chan []metric.Metric, 1)
	stop := make(chan struct{})

	go collectGopsutilMetricsRoutine(gopsutilMetricsChan, stop)
	close(stop)

	time.Sleep(6 * time.Second)

	select {
	case <-gopsutilMetricsChan:
		t.Fatal("Expected no metrics, but received some")
	default:
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Создаем канал stop
	stop := make(chan struct{})

	// Создаем WaitGroup для синхронизации горутин
	var wg sync.WaitGroup
	wg.Add(2) // Теперь ожидаем две горутины

	// Имитируем работающую горутину
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				fmt.Println("Goroutine received stop signal")
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Засекаем время начала
	start := time.Now()

	// Запускаем gracefulShutdown в отдельной горутине
	go func() {
		defer wg.Done()
		gracefulShutdown(stop)
	}()

	// Ожидаем завершения всех горутин
	wg.Wait()

	// Проверяем, что прошло достаточно времени
	elapsed := time.Since(start)
	if elapsed < 5*time.Second {
		t.Errorf("Shutdown took %v, expected at least 5 seconds", elapsed)
	}

	// Проверяем, что канал stop закрыт
	select {
	case _, ok := <-stop:
		if ok {
			t.Error("Stop channel is not closed")
		}
	default:
		t.Error("Stop channel is not closed")
	}
}

func TestPrepareMetricsRoutine(t *testing.T) {
	// Создаем тестовые каналы
	metricsChan := make(chan []metric.Metric)
	gopsutilMetricsChan := make(chan []metric.Metric)
	stop := make(chan struct{})

	// Запускаем тестируемую функцию в отдельной горутине
	go prepareMetricsRoutine(metricsChan, gopsutilMetricsChan, stop)

	// Отправляем тестовые метрики
	go func() {
		metricsChan <- []metric.Metric{{ID: "test1"}}
		gopsutilMetricsChan <- []metric.Metric{{ID: "test2"}}
	}()

	// Ждем некоторое время, чтобы метрики были обработаны
	time.Sleep(2 * time.Second)

	// Отправляем сигнал остановки
	close(stop)
}

// Вспомогательная функция для создания указателя на float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// Вспомогательная функция для создания указателя на int64
func Int64Ptr(i int64) *int64 {
	return &i
}
