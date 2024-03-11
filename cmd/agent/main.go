package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/utils"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func main() {
	cfg := config.LoadAgent()
	updateInterval := time.Duration(cfg.FlagRuntime) * time.Second
	sendInterval := time.Duration(cfg.FlagReportInterval) * time.Second

	lastUpdateTime := time.Now()
	lastSendTime := time.Now()

	PollCount := int64(1)

	var metrics []Metrics
	for {
		// Обновляем метрики каждые 2 секунды
		if time.Since(lastUpdateTime) > updateInterval {
			metrics = collectMetrics(PollCount)
			PollCount++
			lastUpdateTime = time.Now()
		}

		// Отправляем метрики каждые 10 секунд
		if time.Since(lastSendTime) > sendInterval {
			sendMetrics(metrics)
			lastSendTime = time.Now()
		}
	}
}

func collectMetrics(PollCount int64) []Metrics {
	var metrics []Metrics

	// метрики gauge
	metrics = collectRuntimeGauges()

	// счетчик PollCount
	metrics = append(metrics, collectPollCount(PollCount)...)

	return metrics
}

func collectRuntimeGauges() []Metrics {
	var gauges []Metrics

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	gauges = append(gauges, createMetricFromUint64("Alloc", string(MetricTypeGauge), memStats.Alloc))
	gauges = append(gauges, createMetricFromUint64("BuckHashSys", string(MetricTypeGauge), memStats.BuckHashSys))
	gauges = append(gauges, createMetricFromUint64("Frees", string(MetricTypeGauge), memStats.Frees))
	gauges = append(gauges, createMetricFromFloat64("GCCPUFraction", string(MetricTypeGauge), memStats.GCCPUFraction))
	gauges = append(gauges, createMetricFromUint64("GCSys", string(MetricTypeGauge), memStats.GCSys))
	gauges = append(gauges, createMetricFromUint64("HeapAlloc", string(MetricTypeGauge), memStats.HeapAlloc))
	gauges = append(gauges, createMetricFromUint64("HeapIdle", string(MetricTypeGauge), memStats.HeapIdle))
	gauges = append(gauges, createMetricFromUint64("HeapInuse", string(MetricTypeGauge), memStats.HeapInuse))
	gauges = append(gauges, createMetricFromUint64("HeapObjects", string(MetricTypeGauge), memStats.HeapObjects))
	gauges = append(gauges, createMetricFromUint64("HeapReleased", string(MetricTypeGauge), memStats.HeapReleased))
	gauges = append(gauges, createMetricFromUint64("HeapSys", string(MetricTypeGauge), memStats.HeapSys))
	gauges = append(gauges, createMetricFromUint64("LastGC", string(MetricTypeGauge), memStats.LastGC))
	gauges = append(gauges, createMetricFromUint64("Lookups", string(MetricTypeGauge), memStats.Lookups))
	gauges = append(gauges, createMetricFromUint64("MCacheInuse", string(MetricTypeGauge), memStats.MCacheInuse))
	gauges = append(gauges, createMetricFromUint64("Lookups", string(MetricTypeGauge), memStats.Lookups))
	gauges = append(gauges, createMetricFromUint64("MCacheSys", string(MetricTypeGauge), memStats.MCacheSys))
	gauges = append(gauges, createMetricFromUint64("Mallocs", string(MetricTypeGauge), memStats.Mallocs))
	gauges = append(gauges, createMetricFromUint64("NextGC", string(MetricTypeGauge), memStats.NextGC))
	gauges = append(gauges, createMetricFromUint32("NumForcedGC", string(MetricTypeGauge), memStats.NumForcedGC))
	gauges = append(gauges, createMetricFromUint32("NumGC", string(MetricTypeGauge), memStats.NumGC))
	gauges = append(gauges, createMetricFromUint64("OtherSys", string(MetricTypeGauge), memStats.OtherSys))
	gauges = append(gauges, createMetricFromUint64("PauseTotalNs", string(MetricTypeGauge), memStats.PauseTotalNs))
	gauges = append(gauges, createMetricFromUint64("StackInuse", string(MetricTypeGauge), memStats.StackInuse))
	gauges = append(gauges, createMetricFromUint64("StackSys", string(MetricTypeGauge), memStats.StackSys))
	gauges = append(gauges, createMetricFromUint64("Sys", string(MetricTypeGauge), memStats.Sys))
	gauges = append(gauges, createMetricFromUint64("TotalAlloc", string(MetricTypeGauge), memStats.TotalAlloc))
	gauges = append(gauges, createMetricFromUint64("StackInuse", string(MetricTypeGauge), memStats.StackInuse))
	gauges = append(gauges, createMetricFromUint64("MSpanInuse", string(MetricTypeGauge), memStats.MSpanInuse))
	gauges = append(gauges, createMetricFromUint64("MSpanSys", string(MetricTypeGauge), memStats.MSpanSys))
	gauges = append(gauges, createMetricFromFloat64("RandomValue", string(MetricTypeGauge), collectRandomValue()))

	return gauges
}

func collectPollCount(PollCount int64) []Metrics {
	var counter []Metrics
	counter = append(counter, createMetricFromInt64("PollCount", string(MetricTypeCounter), PollCount))
	return counter
}

func collectRandomValue() float64 {
	return rand.Float64()
}

func sendMetrics(metrics []Metrics) {
	cfg := config.LoadAgent()
	for _, metric := range metrics {
		data, err := json.Marshal(metric)

		if err != nil {
			log.Println("Error when serialization object:", err)
		}

		url := fmt.Sprintf("http://%s/update/", cfg.FlagRunAddr)
		b, err := utils.GzipCompress(data)

		if err != nil {
			log.Println("Error when compress data:", err.Error())
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
		if err != nil {
			log.Println("Error when created request data:", err.Error())
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Encoding", "gzip")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error when sending data:", err.Error())
		} else {
			err = resp.Body.Close()
			if err != nil {
				log.Println("Error closing the connection:", err)
			}
		}
	}
}

func createMetricFromFloat64(name string, typeMetric string, value float64) Metrics {
	var metric Metrics
	metric.ID = name
	metric.MType = typeMetric
	metric.Value = &value
	return metric
}

func createMetricFromUint64(name string, typeMetric string, value uint64) Metrics {
	var metric Metrics
	metric.ID = name
	metric.MType = typeMetric
	floatValue := float64(value)
	metric.Value = &floatValue
	return metric
}

func createMetricFromInt64(name string, typeMetric string, value int64) Metrics {
	var metric Metrics
	metric.ID = name
	metric.MType = typeMetric
	metric.Delta = &value
	return metric
}

func createMetricFromUint32(name string, typeMetric string, value uint32) Metrics {
	var metric Metrics
	metric.ID = name
	metric.MType = typeMetric
	floatValue := float64(value)
	metric.Value = &floatValue
	return metric
}
