// TODO не до конца понимаю, как реализовать условия из "Важно"
// TODO переделать ассоциативный масссив на  map[string]interface{}???

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Num struct {
	val string
}

func (n *Num) Float64() string {
	return n.val
}

func (n *Num) Uint64(num uint64) string {
	n.val = fmt.Sprint(float64(num))
	return n.val
}

func (n *Num) Uint32(num uint32) string {
	n.val = fmt.Sprint(float64(num))
	return n.val
}

func main() {
	parseFlags()

	//не до конца понимаю, как можно связать с http cервером
	//e := echo.New()
	//
	//err := e.Start(flagRunAddr)
	//if err != nil {
	//	panic(err)
	//}
	updateInterval := 2 * time.Second
	sendInterval := 10 * time.Second

	lastUpdateTime := time.Now()
	lastSendTime := time.Now()

	PollCount := int64(1)

	var metrics = make(map[string]map[string]string)
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

func collectMetrics(PollCount int64) map[string]map[string]string {
	var metrics = make(map[string]map[string]string)

	// метрики gauge
	metrics["gauge"] = collectRuntimeGauges()

	// счетчик PollCount
	metrics["counter"] = collectPollCount(PollCount)

	return metrics
}

func collectRuntimeGauges() map[string]string {
	//ассоциативный массив
	var gauges = make(map[string]string)

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	num := &Num{}
	gauges["Alloc"] = num.Uint64(memStats.Alloc)
	gauges["BuckHashSys"] = num.Uint64(memStats.BuckHashSys)
	gauges["Frees"] = num.Uint64(memStats.Frees)
	gauges["GCCPUFraction"] = num.Float64()
	gauges["GCSys"] = num.Uint64(memStats.GCSys)
	gauges["HeapAlloc"] = num.Uint64(memStats.HeapAlloc)
	gauges["HeapIdle"] = num.Uint64(memStats.HeapIdle)
	gauges["HeapInuse"] = num.Uint64(memStats.HeapInuse)
	gauges["HeapObjects"] = num.Uint64(memStats.HeapObjects)
	gauges["HeapReleased"] = num.Uint64(memStats.HeapReleased)
	gauges["HeapSys"] = num.Uint64(memStats.HeapSys)
	gauges["LastGC"] = num.Uint64(memStats.LastGC)
	gauges["Lookups"] = num.Uint64(memStats.Lookups)
	gauges["MCacheInuse"] = num.Uint64(memStats.MCacheInuse)
	gauges["MCacheSys"] = num.Uint64(memStats.MCacheSys)
	gauges["Mallocs"] = num.Uint64(memStats.Mallocs)
	gauges["NextGC"] = num.Uint64(memStats.NextGC)
	gauges["NumForcedGC"] = num.Uint32(memStats.NumForcedGC)
	gauges["NumGC"] = num.Uint32(memStats.NumGC)
	gauges["OtherSys"] = num.Uint64(memStats.OtherSys)
	gauges["PauseTotalNs"] = num.Uint64(memStats.PauseTotalNs)
	gauges["StackInuse"] = num.Uint64(memStats.StackInuse)
	gauges["StackSys"] = num.Uint64(memStats.StackSys)
	gauges["Sys"] = num.Uint64(memStats.Sys)
	gauges["TotalAlloc"] = num.Uint64(memStats.TotalAlloc)

	// случайное значение
	gauges["RandomValue"] = collectRandomValue()

	return gauges
}

func collectPollCount(PollCount int64) map[string]string {
	var counter = make(map[string]string)
	counter["PollCount"] = fmt.Sprint(PollCount)
	return counter
}

func collectRandomValue() string {
	return fmt.Sprint(rand.Float64())
}

func sendMetrics(metrics map[string]map[string]string) {
	//не понимаю, как отправить запрос используя echo, не поднимая сервер
	for key, types := range metrics {
		for name, value := range types {
			url := fmt.Sprintf("http://%s/update/%s/%s/%s", conf.flagRunAddr, key, name, value)
			post, err := http.Post(
				url,
				"text/plain",
				nil)

			if err != nil {
				log.Fatal("Error when sending data:", err)
			}

			err = post.Body.Close()
			if err != nil {
				log.Fatal("Error closing the connection:", err)
			}
		}
	}
}
