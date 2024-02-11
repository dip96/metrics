package main

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
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

func (m MemStorage) GetAll() map[string]interface{} {
	return m.metrics
}

// Интерфейс для работы с хранилищем
type Storage interface {
	Get(name string) interface{}
	Set(name string, metric interface{})
	GetAll() map[string]interface{}
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

func (g *Gauge) GetName() float64 {
	return g.value
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

func AddMetric(c echo.Context) error {
	typeMetric := c.Param("type_metric")
	nameMetric := c.Param("name_metric")
	valueMetric := c.Param("value_metric")

	var valueMet interface{}
	if typeMetric == "gauge" {
		valueInt, _ := strconv.ParseFloat(valueMetric, 64)
		valueMet = valueInt
	} else if typeMetric == "counter" {
		valueInt, _ := strconv.ParseInt(valueMetric, 10, 64)
		valueMet = valueInt
	} else {
		return c.String(http.StatusBadRequest, "")
	}

	metric := storage.Get(nameMetric)

	switch m := metric.(type) {
	case *Gauge:
		m.Set(valueMet.(float64))
	case *Counter:
		m.Inc(valueMet.(int64))
	default:
		// Создаем метрику, если ее нет
		if typeMetric == "gauge" {
			metric = &Gauge{
				name:  nameMetric,
				value: valueMet.(float64),
			}
		} else if typeMetric == "counter" {
			metric = &Counter{
				name:  nameMetric,
				value: valueMet.(int64),
			}
		} else {
			return c.String(http.StatusBadRequest, "")
		}
		storage.Set(nameMetric, metric)
	}
	return c.String(http.StatusOK, "")

}

func getMetric(c echo.Context) error {
	name := c.Param("name_metric")
	metric := storage.Get(name)

	switch m := metric.(type) {
	case *Gauge:
		return c.String(http.StatusOK, fmt.Sprintf("%f", m.value))

	case *Counter:
		return c.String(http.StatusOK, fmt.Sprintf("%d", m.value))

	default:
		return c.String(http.StatusNotFound, "")
	}

}

func getAllMetrics(c echo.Context) error {
	metrics := storage.GetAll()

	var buf bytes.Buffer

	buf.WriteString("<html><body><ul>")

	for name, value := range metrics {
		buf.WriteString(fmt.Sprintf("<li>%s: %v</li>", name, value))
	}

	buf.WriteString("</ul></body></html>")

	return c.HTML(http.StatusOK, buf.String())
}

func main() {
	parseFlags()

	e := echo.New()

	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)
	e.GET("/value/:type_metric/:name_metric", getMetric)
	e.GET("/", getAllMetrics)

	storage = &MemStorage{
		metrics: make(map[string]interface{}),
	}

	//e.Logger.Fatal()
	fmt.Println("Running server on", flagRunAddr)
	err := e.Start(flagRunAddr)
	if err != nil {
		panic(err)
	}
}
