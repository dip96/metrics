package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dip96/metrics/internal/middleware"
	"github.com/dip96/metrics/internal/utils"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

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

func (m Metric) GetValueForDisplay() (string, error) {
	if m.MType == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.Delta), nil
	}

	if m.MType == MetricTypeGauge {
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

	//TODO Повторяющийся фрагмент кода, вынести
	acceptEncoding := c.Request().Header.Get("Accept-Encoding")
	if acceptEncoding == "gzip" {
		b, err := utils.GzipCompress(buf.Bytes())

		if err != nil {
			log.Fatal("Error when compress data:", err.Error())
		}

		fmt.Printf("3 %d bytes has been compressed to %d bytes\r\n", len(buf.String()), len(b))
		c.Response().Header().Set("Content-Encoding", "gzip")
		return c.HTMLBlob(http.StatusOK, b)
	}

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

	jsonData, err := json.Marshal(metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	//не получилось перезаписать данные в body используя middleware
	acceptEncoding := c.Request().Header.Get("Accept-Encoding")
	if acceptEncoding == "gzip" {
		b, err := utils.GzipCompress(jsonData)

		if err != nil {
			log.Fatal("Error when compress data:", err.Error())
		}

		fmt.Printf("2 %d bytes has been compressed to %d bytes\r\n", len(jsonData), len(b))
		c.Response().Header().Set("Content-Encoding", "gzip")
		return c.JSONBlob(http.StatusOK, b)
	}

	return c.JSON(http.StatusOK, jsonData)
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

	jsonData, err := json.Marshal(metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	acceptEncoding := c.Request().Header.Get("Accept-Encoding")
	//contentType := c.Request().Header.Get("Content-Type")
	if acceptEncoding == "gzip" {
		b, err := utils.GzipCompress(jsonData)

		if err != nil {
			log.Fatal("Error when compress data:", err.Error())
		}

		fmt.Printf("1 %d bytes has been compressed to %d bytes\r\n", len(jsonData), len(b))
		c.Response().Header().Set("Content-Encoding", "gzip")
		return c.JSONBlob(http.StatusOK, b)
	}

	return c.JSON(http.StatusOK, jsonData)
}

func main() {
	parseFlags()
	e := echo.New()
	e.Use(middleware.Logger)
	e.Use(middleware.UnzipMiddleware)

	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)
	e.GET("/value/:type_metric/:name_metric", getMetric)
	e.GET("/", getAllMetrics)

	e.POST("/update/", AddMetricV2)
	e.POST("/value/", GetMetricV2)

	storage = &MemStorage{
		metrics: make(map[string]Metric),
	}

	//TODO вынести логику в отдельный файл
	initMetrics()
	go saveMetrics()

	fmt.Println("Running server on", conf.flagRunAddr)
	err := e.Start(conf.flagRunAddr)
	if err != nil {
		panic(err)
	}
}

// TODO вынести в отдельный файл
func initMetrics() {
	Consumer, err := NewConsumer(conf.fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer Consumer.Close()

	for {
		metric, err := Consumer.ReadEvent()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error(err)
			continue
		}

		err = storage.Set(metric.ID, *metric)
		if err != nil {
			log.Error(err)
		}
	}
}

// TODO вынести в отдельный файл internal/storage/storageMetrics
func saveMetrics() error {
	ticker := time.NewTicker(time.Duration(conf.storeInterval) * time.Second)

	if conf.restore {
		for range ticker.C {
			Producer, err := NewProducer(conf.fileStoragePath)
			if err != nil {
				log.Fatal(err)
			}
			//defer Producer.Close()
			metrics, _ := storage.GetAll()
			for metric := range metrics {

				if err := Producer.WriteEvent(metrics[metric]); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	return nil
}

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(metric Metric) error {
	data, err := json.Marshal(&metric)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*Metric, error) {
	if !c.scanner.Scan() {
		if c.scanner.Err() == nil {
			return nil, io.EOF
		}
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	metric := Metric{}
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, err
	}

	return &metric, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (p *Producer) Close() error {
	// закрываем файл
	return p.file.Close()
}
