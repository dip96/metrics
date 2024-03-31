package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/utils"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func main() {
	cfg := config.LoadAgent()
	updateInterval := time.Duration(cfg.FlagRuntime) * time.Second
	sendInterval := time.Duration(cfg.FlagReportInterval) * time.Second

	lastUpdateTime := time.Now()
	lastSendTime := time.Now()

	PollCount := int64(1)

	var metrics []metricModel.Metric
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

func collectMetrics(PollCount int64) []metricModel.Metric {
	var metrics []metricModel.Metric

	// метрики gauge
	metrics = collectRuntimeGauges()

	// счетчик PollCount
	metrics = append(metrics, collectPollCount(PollCount)...)

	return metrics
}

func collectRuntimeGauges() []metricModel.Metric {
	var gauges []metricModel.Metric

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	gauges = append(gauges, createMetricFromUint64("Alloc", metricModel.MetricTypeGauge, memStats.Alloc))
	gauges = append(gauges, createMetricFromUint64("BuckHashSys", metricModel.MetricTypeGauge, memStats.BuckHashSys))
	gauges = append(gauges, createMetricFromUint64("Frees", metricModel.MetricTypeGauge, memStats.Frees))
	gauges = append(gauges, createMetricFromFloat64("GCCPUFraction", metricModel.MetricTypeGauge, memStats.GCCPUFraction))
	gauges = append(gauges, createMetricFromUint64("GCSys", metricModel.MetricTypeGauge, memStats.GCSys))
	gauges = append(gauges, createMetricFromUint64("HeapAlloc", metricModel.MetricTypeGauge, memStats.HeapAlloc))
	gauges = append(gauges, createMetricFromUint64("HeapIdle", metricModel.MetricTypeGauge, memStats.HeapIdle))
	gauges = append(gauges, createMetricFromUint64("HeapInuse", metricModel.MetricTypeGauge, memStats.HeapInuse))
	gauges = append(gauges, createMetricFromUint64("HeapObjects", metricModel.MetricTypeGauge, memStats.HeapObjects))
	gauges = append(gauges, createMetricFromUint64("HeapReleased", metricModel.MetricTypeGauge, memStats.HeapReleased))
	gauges = append(gauges, createMetricFromUint64("HeapSys", metricModel.MetricTypeGauge, memStats.HeapSys))
	gauges = append(gauges, createMetricFromUint64("LastGC", metricModel.MetricTypeGauge, memStats.LastGC))
	gauges = append(gauges, createMetricFromUint64("Lookups", metricModel.MetricTypeGauge, memStats.Lookups))
	gauges = append(gauges, createMetricFromUint64("MCacheInuse", metricModel.MetricTypeGauge, memStats.MCacheInuse))
	gauges = append(gauges, createMetricFromUint64("Lookups", metricModel.MetricTypeGauge, memStats.Lookups))
	gauges = append(gauges, createMetricFromUint64("MCacheSys", metricModel.MetricTypeGauge, memStats.MCacheSys))
	gauges = append(gauges, createMetricFromUint64("Mallocs", metricModel.MetricTypeGauge, memStats.Mallocs))
	gauges = append(gauges, createMetricFromUint64("NextGC", metricModel.MetricTypeGauge, memStats.NextGC))
	gauges = append(gauges, createMetricFromUint32("NumForcedGC", metricModel.MetricTypeGauge, memStats.NumForcedGC))
	gauges = append(gauges, createMetricFromUint32("NumGC", metricModel.MetricTypeGauge, memStats.NumGC))
	gauges = append(gauges, createMetricFromUint64("OtherSys", metricModel.MetricTypeGauge, memStats.OtherSys))
	gauges = append(gauges, createMetricFromUint64("PauseTotalNs", metricModel.MetricTypeGauge, memStats.PauseTotalNs))
	gauges = append(gauges, createMetricFromUint64("StackInuse", metricModel.MetricTypeGauge, memStats.StackInuse))
	gauges = append(gauges, createMetricFromUint64("StackSys", metricModel.MetricTypeGauge, memStats.StackSys))
	gauges = append(gauges, createMetricFromUint64("Sys", metricModel.MetricTypeGauge, memStats.Sys))
	gauges = append(gauges, createMetricFromUint64("TotalAlloc", metricModel.MetricTypeGauge, memStats.TotalAlloc))
	gauges = append(gauges, createMetricFromUint64("StackInuse", metricModel.MetricTypeGauge, memStats.StackInuse))
	gauges = append(gauges, createMetricFromUint64("MSpanInuse", metricModel.MetricTypeGauge, memStats.MSpanInuse))
	gauges = append(gauges, createMetricFromUint64("MSpanSys", metricModel.MetricTypeGauge, memStats.MSpanSys))
	gauges = append(gauges, createMetricFromFloat64("RandomValue", metricModel.MetricTypeGauge, collectRandomValue()))

	return gauges
}

func collectPollCount(PollCount int64) []metricModel.Metric {
	var counter []metricModel.Metric
	counter = append(counter, createMetricFromInt64("PollCount", metricModel.MetricTypeCounter, PollCount))
	return counter
}

func collectRandomValue() float64 {
	return rand.Float64()
}

func sendMetrics(metrics []metricModel.Metric) {
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
			return
		} else {
			err = resp.Body.Close()
			if err != nil {
				log.Println("Error closing the connection:", err)
			}
		}
	}
}

func createMetricFromFloat64(name string, typeMetric metricModel.MetricType, value float64) metricModel.Metric {
	var metric metricModel.Metric
	metric.ID = name
	metric.MType = typeMetric
	metric.Value = &value
	return metric
}

func createMetricFromUint64(name string, typeMetric metricModel.MetricType, value uint64) metricModel.Metric {
	var metric metricModel.Metric
	metric.ID = name
	metric.MType = typeMetric
	floatValue := float64(value)
	metric.Value = &floatValue
	return metric
}

func createMetricFromInt64(name string, typeMetric metricModel.MetricType, value int64) metricModel.Metric {
	var metric metricModel.Metric
	metric.ID = name
	metric.MType = typeMetric
	metric.Delta = &value
	return metric
}

func createMetricFromUint32(name string, typeMetric metricModel.MetricType, value uint32) metricModel.Metric {
	var metric metricModel.Metric
	metric.ID = name
	metric.MType = typeMetric
	floatValue := float64(value)
	metric.Value = &floatValue
	return metric
}
