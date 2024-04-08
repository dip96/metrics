package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/hash"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func main() {
	stop := make(chan struct{})
	metricsChan := make(chan []metricModel.Metric)
	gopsutilMetricsChan := make(chan []metricModel.Metric)

	go collectMetricsRoutine(metricsChan, stop)
	go collectGopsutilMetricsRoutine(gopsutilMetricsChan, stop)
	go prepareMetricsRoutine(metricsChan, gopsutilMetricsChan, stop)

	<-stop
}
func collectGopsutilMetricsRoutine(gopsutilMetricsChan chan<- []metricModel.Metric, stop <-chan struct{}) {
	gopsutilInterval := 5 * time.Second

	ticker := time.NewTicker(gopsutilInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			gopsutilMetrics := collectGopsutilMetrics()
			gopsutilMetricsChan <- gopsutilMetrics
		}
	}
}

func collectMetricsRoutine(metricsChan chan<- []metricModel.Metric, stop <-chan struct{}) {
	cfg := config.LoadAgent()
	updateInterval := time.Duration(cfg.FlagRuntime) * time.Second
	lastUpdateTime := time.Now()
	PollCount := int64(1)

	for {
		select {
		case <-stop:
			return
		default:
			if time.Since(lastUpdateTime) > updateInterval {
				metrics := collectMetrics(PollCount)
				PollCount++
				lastUpdateTime = time.Now()
				metricsChan <- metrics
			}
		}
	}
}

func prepareMetricsRoutine(metricsChan <-chan []metricModel.Metric, gopsutilMetricsChan <-chan []metricModel.Metric, stop <-chan struct{}) {
	cfg := config.LoadAgent()
	rateLimit := cfg.RateLimit
	sendInterval := time.Duration(cfg.FlagReportInterval) * time.Second

	mergedMetricsChan := mergeMetrics(metricsChan, gopsutilMetricsChan)

	lastSendTime := time.Now()
	jobChan := make(chan metricModel.Metric, rateLimit)

	//pool worker
	for i := 0; i < rateLimit; i++ {
		go sendMetricsRoutine(jobChan, stop)
	}

	for {
		select {
		case <-stop:
			close(jobChan)
			return
		default:
			if time.Since(lastSendTime) > sendInterval {
				for metrics := range mergedMetricsChan {
					for _, m := range metrics {
						jobChan <- m
					}
				}
				lastSendTime = time.Now()
			}
		}
	}
}

func mergeMetrics(metricsChan <-chan []metricModel.Metric, gopsutilMetricsChan <-chan []metricModel.Metric) <-chan []metricModel.Metric {
	mergedChan := make(chan []metricModel.Metric)

	go func() {
		defer close(mergedChan)

		for {
			select {
			case metrics, ok := <-metricsChan:
				if !ok {
					metricsChan = nil
					continue
				}
				mergedChan <- metrics
			case gopsutilMetrics, ok := <-gopsutilMetricsChan:
				if !ok {
					gopsutilMetricsChan = nil
					continue
				}
				mergedChan <- gopsutilMetrics
			default:
				if metricsChan == nil && gopsutilMetricsChan == nil {
					return
				}
			}
		}
	}()

	return mergedChan
}

func sendMetricsRoutine(jobChan <-chan metricModel.Metric, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case metric := <-jobChan:
			sendMetricsButch([]metricModel.Metric{metric})
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

func collectGopsutilMetrics() []metricModel.Metric {
	var metrics []metricModel.Metric

	virtualMemoryStat, err := mem.VirtualMemory()
	if err == nil {
		metrics = append(metrics, createMetricFromUint64("TotalMemory", metricModel.MetricTypeGauge, virtualMemoryStat.Total))
		metrics = append(metrics, createMetricFromUint64("FreeMemory", metricModel.MetricTypeGauge, virtualMemoryStat.Free))
	}

	cpuPercentages, err := cpu.Percent(time.Second, false)
	if err == nil {
		for i, cpuUsage := range cpuPercentages {
			metricName := fmt.Sprintf("CPUutilization%d", i+1)
			metrics = append(metrics, createMetricFromFloat64(metricName, metricModel.MetricTypeGauge, cpuUsage))
		}
	}

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

func sendMetricsButch(metrics []metricModel.Metric) {
	cfg := config.LoadAgent()
	data, err := json.Marshal(metrics)

	if err != nil {
		log.Println("Error when serialization object:", err)
	}

	url := fmt.Sprintf("http://%s/updates/", cfg.FlagRunAddr)
	b, err := utils.GzipCompress(data)

	if err != nil {
		log.Println("Error when compress data:", err.Error())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		log.Println("Error when created request data:", err.Error())
	}

	hashAgent := hash.CalculateHashAgent(b)
	if hashAgent != "" {
		req.Header.Add("HashSHA256", hashAgent)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	client := &http.Client{}
	//TODO вынести в отдельную функцию
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for attempt, delay := range retryDelays {
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error when sending data (attempt %d/%d): %v", attempt+1, len(retryDelays), err)
			time.Sleep(delay)
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			log.Println("Error closing the connection:", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Println("Data sent successfully.")
			return
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
