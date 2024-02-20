package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metric struct {
	Type           MetricType
	CounterValue   *int64
	GaugeValue     *float64 // Подскажите, а зачем они нужны?
	fullValueGauge string   //float64 обрезает нули
}

// интерфейс для работы с объектом Metric
type Metrics interface {
	GetValueForDisplay() (string, error)
	GetValue() (string, error)
	SetValue()
}

func (m Metric) GetValueForDisplay() (string, error) {

	if m.Type == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.CounterValue), nil
	}

	if m.Type == MetricTypeGauge {
		return m.fullValueGauge, nil
	}

	return "", errors.New("the metric type is incorrect")
}

func (m Metric) GetValue() (string, error) {

	if m.Type == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.CounterValue), nil
	}

	if m.Type == MetricTypeGauge {
		return m.fullValueGauge, nil
	}

	return "", errors.New("the metric type is incorrect")
}

// Структура для хранения метрик
type MemStorage struct {
	metrics map[string]Metric
}

func (m MemStorage) Get(name string) (Metric, error) {
	value, ok := m.metrics[name]

	if ok {
		return value, nil
	}

	return Metric{}, errors.New("the metric was not found")
}

func (m MemStorage) Set(name string, metric Metric) error {
	m.metrics[name] = metric
	return nil
}

func (m MemStorage) GetAll() (map[string]Metric, error) {
	return m.metrics, nil
}

// Интерфейс для работы с хранилищем
type Storage interface {
	Get(name string) (Metric, error)
	Set(name string, metric Metric) error
	GetAll() (map[string]Metric, error)
}

// Хранилище метрик
var storage *MemStorage

func AddMetric(c echo.Context) error {
	typeMetric := c.Param("type_metric")
	nameMetric := c.Param("name_metric")
	valueMetric := c.Param("value_metric")

	metric, _ := storage.Get(nameMetric)

	//был вариант добавить метод SetValue(29 строка) для логики сохранения в одном месте,
	//но не понятно, как в него передать все нужные параметры
	//видимо для этого и необходим  context.Context?
	if typeMetric == string(MetricTypeGauge) {
		value, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.Type = MetricTypeGauge
		metric.GaugeValue = &value
		metric.fullValueGauge = valueMetric
	} else if typeMetric == string(MetricTypeCounter) {
		value, err := strconv.ParseInt(valueMetric, 10, 64)

		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.Type = MetricTypeCounter
		metric.CounterValue = &value
	} else {
		return c.String(http.StatusBadRequest, "")
	}

	err := storage.Set(nameMetric, metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, "")
}

func getMetric(c echo.Context) error {
	name := c.Param("name_metric")
	metric, err := storage.Get(name)

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	value, err := metric.GetValueForDisplay()

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, value)
}

func getAllMetrics(c echo.Context) error {
	metrics, err := storage.GetAll()

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	buf.WriteString("<html><body><ul>")

	for name, metric := range metrics {
		value, err := metric.GetValue()

		if err != nil {
			value = "Not found"
		}

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
		metrics: make(map[string]Metric),
	}

	fmt.Println("Running server on", conf.flagRunAddr)
	err := e.Start(conf.flagRunAddr)
	if err != nil {
		panic(err)
	}
}
