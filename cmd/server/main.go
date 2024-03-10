package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/middleware"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage/files"
	memStorage "github.com/dip96/metrics/internal/storage/mem"
	"github.com/dip96/metrics/internal/utils"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
)

// TODO вынести в отдельную директорию api
func AddMetric(c echo.Context) error {
	typeMetric := c.Param("type_metric")
	nameMetric := c.Param("name_metric")
	valueMetric := c.Param("value_metric")

	metric, _ := memStorage.MemStorage.Get(nameMetric)

	if typeMetric == string(metricModel.MetricTypeGauge) {
		value, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.MType = metricModel.MetricTypeGauge
		metric.Value = &value
		metric.FullValueGauge = valueMetric
	} else if typeMetric == string(metricModel.MetricTypeCounter) {
		value, err := strconv.ParseInt(valueMetric, 10, 64)

		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}

		metric.MType = metricModel.MetricTypeCounter
		if metric.Delta == nil {
			metric.Delta = &value
		} else {
			*metric.Delta += value
		}
	} else {
		return c.String(http.StatusBadRequest, "")
	}

	memStorage.MemStorage.Set(nameMetric, metric)

	return c.String(http.StatusOK, "")
}

func getMetric(c echo.Context) error {
	name := c.Param("name_metric")
	metric, err := memStorage.MemStorage.Get(name)

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
	metrics, err := memStorage.MemStorage.GetAll()

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
	body := new(metricModel.Metric)

	if err := c.Bind(body); err != nil {
		return err
	}

	typeMetric := body.MType
	nameMetric := body.ID
	metric, _ := memStorage.MemStorage.Get(nameMetric)
	metric.ID = body.ID

	if typeMetric == metricModel.MetricTypeGauge {
		valueMetric := body.Value
		metric.MType = metricModel.MetricTypeGauge
		metric.Value = valueMetric
		metric.FullValueGauge = fmt.Sprintf("%f", *valueMetric)
	} else if typeMetric == metricModel.MetricTypeCounter {
		valueMetric := body.Delta
		if metric.Delta == nil {
			metric.MType = metricModel.MetricTypeCounter
			metric.Delta = valueMetric
		} else {
			*metric.Delta += *valueMetric
		}
	} else {
		return c.String(http.StatusBadRequest, "")
	}

	memStorage.MemStorage.Set(nameMetric, metric)

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
	body := new(metricModel.Metric)

	if err := c.Bind(body); err != nil {
		return err
	}

	nameMetric := body.ID
	metric, err := memStorage.MemStorage.Get(nameMetric)

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	jsonData, err := json.Marshal(metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	acceptEncoding := c.Request().Header.Get("Accept-Encoding")
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
	cfg := config.LoadServer()

	e := echo.New()
	e.Use(middleware.Logger)
	e.Use(middleware.UnzipMiddleware)

	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)
	e.GET("/value/:type_metric/:name_metric", getMetric)
	e.GET("/", getAllMetrics)

	e.POST("/update/", AddMetricV2)
	e.POST("/value/", GetMetricV2)

	if memStorage.MemStorage == nil {
		memStorage.MemStorage = memStorage.NewStorage()
	}

	//TODO вынести логику в отдельный файл
	files.InitMetrics()
	go files.UpdateMetrics()

	fmt.Println("Running server on", cfg.FlagRunAddr)
	err := e.Start(cfg.FlagRunAddr)
	if err != nil {
		panic(err)
	}
}
