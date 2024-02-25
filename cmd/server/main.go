package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dip96/metrics/internal/middleware"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

//type RequestBody struct {
//	ID    string   `json:"id"`              // имя метрики
//	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
//	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
//	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
//}

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metric struct {
	ID             string     `json:"id"`              // имя метрики
	MType          MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta          *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value          *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
	fullValueGauge string     //float64 обрезает нули
}

//type Metric struct {
//	ID             string
//	Type           MetricType
//	CounterValue   *int64
//	GaugeValue     *float64
//	fullValueGauge string //float64 обрезает нули
//}

func (m Metric) GetValueForDisplay() (string, error) {
	if m.MType == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.Delta), nil
	}

	if m.MType == MetricTypeGauge {
		//return fmt.Sprintf("%f", *m.Value), nil
		return m.fullValueGauge, nil
	}

	return "", errors.New("the metric type is incorrect")
}

func (m Metric) GetValue() (string, error) {
	if m.MType == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.Delta), nil
	}

	if m.MType == MetricTypeGauge {
		return fmt.Sprintf("%f", *m.Value), nil
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

// Хранилище метрик
var storage *MemStorage

func AddMetric(c echo.Context) error {
	typeMetric := c.Param("type_metric")
	nameMetric := c.Param("name_metric")
	valueMetric := c.Param("value_metric")

	metric, _ := storage.Get(nameMetric)

	if typeMetric == string(MetricTypeGauge) {
		value, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.MType = MetricTypeGauge
		metric.Value = &value
		metric.fullValueGauge = valueMetric
	} else if typeMetric == string(MetricTypeCounter) {
		value, err := strconv.ParseInt(valueMetric, 10, 64)

		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.MType = MetricTypeCounter
		if metric.Delta == nil {
			metric.Delta = &value
		} else {
			*metric.Delta += value
		}
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

func AddMetricV2(c echo.Context) error {
	body := new(Metric)

	if err := c.Bind(body); err != nil {
		return err
	}

	typeMetric := body.MType
	nameMetric := body.ID
	metric, _ := storage.Get(nameMetric)
	metric.ID = body.ID

	jsonBytes, _ := json.Marshal(body)
	log.Printf(string(jsonBytes))

	if typeMetric == MetricTypeGauge {
		valueMetric := body.Value
		metric.MType = MetricTypeGauge
		metric.Value = valueMetric
		metric.fullValueGauge = fmt.Sprintf("%f", *valueMetric)
	} else if typeMetric == MetricTypeCounter {
		valueMetric := body.Delta
		if metric.Delta == nil {
			metric.MType = MetricTypeCounter
			metric.Delta = valueMetric
		} else {
			*metric.Delta += *valueMetric
		}
	} else {
		return c.String(http.StatusBadRequest, "")
	}

	err := storage.Set(nameMetric, metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	return c.JSON(http.StatusOK, metric)
}

func GetMetricV2(c echo.Context) error {
	body := new(Metric)

	if err := c.Bind(body); err != nil {
		return err
	}
	nameMetric := body.ID
	metric, err := storage.Get(nameMetric)

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, metric)
}

func main() {
	//conf := NewConfig()
	parseFlags()
	//middleware.InitLogger()
	//defer middleware.CloseLogger()
	e := echo.New()

	e.Use(middleware.Logger)

	//TODO нужно ли удалять два нижних роута?
	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)
	e.GET("/value/:type_metric/:name_metric", getMetric)
	e.GET("/", getAllMetrics)

	e.POST("/update/", AddMetricV2)
	e.POST("/value/", GetMetricV2)

	storage = &MemStorage{
		metrics: make(map[string]Metric),
	}

	fmt.Println("Running server on", conf.flagRunAddr)
	err := e.Start(conf.flagRunAddr)
	if err != nil {
		panic(err)
	}
}
