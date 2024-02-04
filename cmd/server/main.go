package main

import (
	"net/http"
	"strconv"
	"strings"
)

// Структура для хранения метрик
type MemStorage struct {
	metrics map[string]interface{}
}

func (m MemStorage) Get(name string) interface{} {
	return m.metrics[name]
}

func (m MemStorage) Set(name string, metric interface{}) {
	m.metrics[name] = metric
}

// Интерфейс для работы с хранилищем
type Storage interface {
	Get(name string) interface{}
	Set(name string, metric interface{})
}

// Структура метрики Gauge
type Gauge struct {
	name  string
	value float64
}

// Метод Set для Gauge
func (g *Gauge) Set(value float64) {
	g.value = value
}

// Структура метрики Counter
type Counter struct {
	name  string
	value int64
}

// Метод Inc для Counter
func (c *Counter) Inc(delta int64) {
	c.value += delta
}

// Хранилище метрик
var storage Storage

func AddMetric(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	//Нужна ли проверка на заголовок Content-Type?
	//if req.Header.Get("Content-Type") != "text/plain" {
	//
	//}

	parts := strings.Split(req.URL.Path, "/")

	if len(parts) != 5 {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	//Как определить, что в запросе нет имени метрики?

	metricType := parts[2]
	name := parts[3]
	valueMetric := parts[4]

	var value interface{}
	if valueInt, err := strconv.ParseInt(valueMetric, 10, 64); err == nil {
		value = valueInt
	} else if valueFloat, err := strconv.ParseFloat(valueMetric, 64); err == nil {
		value = valueFloat
	} else {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	// Получаем метрику из хранилища
	metric := storage.Get(name)

	// Обновляем значение
	switch m := metric.(type) {
	case *Gauge:
		m.Set(value.(float64))
	case *Counter:
		m.Inc(value.(int64))
	default:
		// Создаем метрику, если ее нет
		if metricType == "gauge" {
			metric = &Gauge{
				name:  name,
				value: value.(float64),
			}
		} else if metricType == "counter" {
			metric = &Counter{
				name:  name,
				value: value.(int64),
			}
		} else {
			http.Error(res, "Bad request", http.StatusBadRequest)
			return
		}
		storage.Set(name, metric)
		m = metric

		res.WriteHeader(http.StatusOK)
	}
}

func main() {
	http.HandleFunc(`/update/`, AddMetric)

	storage = &MemStorage{
		metrics: make(map[string]interface{}),
	}

	err := http.ListenAndServe(`:80`, nil)
	if err != nil {
		panic(err)
	}
}
