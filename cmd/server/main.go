package main

//TODO изменить наименования на корретные - https://go.dev/blog/package-names

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/database/migrator"
	"github.com/dip96/metrics/internal/hash"
	"github.com/dip96/metrics/internal/middleware"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	"github.com/dip96/metrics/internal/storage/files"
	memStorage "github.com/dip96/metrics/internal/storage/mem"
	postgresStorage "github.com/dip96/metrics/internal/storage/postgres"
	"github.com/dip96/metrics/internal/utils"
	echopprof "github.com/hiko1129/echo-pprof"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// AddMetric - Ендпоинт для добавления метрики.
// Принимает тип метрики (gauge или counter), имя метрики и значение.
// Возвращает статус-код 200 в случае успешного добавления, иначе - 400.
func AddMetric(c echo.Context) error {
	typeMetric := c.Param("type_metric")
	nameMetric := c.Param("name_metric")
	valueMetric := c.Param("value_metric")

	metric, _ := storage.Storage.Get(nameMetric)
	metric.ID = nameMetric

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

	err := storage.Storage.Set(metric)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, "")
}

// getMetric - Эндпоинт для получения значения метрики по ее имени.
// Принимает имя метрики.
// Возвращает значение метрики в виде строки и статус-код 200,
// или сообщение об ошибке и статус-код 404 в случае, если метрика не найдена.
func getMetric(c echo.Context) error {
	name := c.Param("name_metric")
	metric, err := storage.Storage.Get(name)

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	value, err := metric.GetValueForDisplay()

	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, value)
}

// getAllMetrics - Эндпоинт для получения списка всех метрик в HTML-формате.
// Возвращает HTML-страницу со списком метрик и их значений и статус-код 200.
func getAllMetrics(c echo.Context) error {
	metrics, err := storage.Storage.GetAll()

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

// AddMetricV2 - Эндпоинт для добавления метрики в формате JSON.
// Принимает структуру Metric в теле запроса.
// Возвращает добавленную метрику в формате JSON и статус-код 200 в случае успеха,
// иначе - сообщение об ошибке и статус-код 400.
func AddMetricV2(c echo.Context) error {
	body := new(metricModel.Metric)

	if err := c.Bind(body); err != nil {
		return err
	}

	typeMetric := body.MType
	nameMetric := body.ID
	metric, _ := storage.Storage.Get(nameMetric)
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

	err := storage.Storage.Set(metric)

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

		//возможно стоит высчитывать хеш до gzip
		hashServer := hash.CalculateHashServer(b)
		if hashServer != "" {
			c.Response().Header().Set("HashSHA256", hashServer)
		}

		return c.JSONBlob(http.StatusOK, b)
	}

	//возможно стоит высчитывать хеш до gzip
	hashServer := hash.CalculateHashServer(jsonData)
	if hashServer != "" {
		c.Response().Header().Set("HashSHA256", hashServer)
	}
	return c.JSON(http.StatusOK, metric)
}

// GetMetricV2 - Эндпоинт для получения метрики по ее имени в формате JSON.
// Принимает структуру Metric с заполненным полем ID в теле запроса.
// Возвращает метрику в формате JSON и статус-код 200 в случае успеха,
// иначе - сообщение об ошибке и статус-код 404 или 400.
func GetMetricV2(c echo.Context) error {
	body := new(metricModel.Metric)

	if err := c.Bind(body); err != nil {
		return err
	}

	nameMetric := body.ID
	metric, err := storage.Storage.Get(nameMetric)

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
		//возможно стоит высчитывать хеш до gzip
		hashServer := hash.CalculateHashServer(b)
		if hashServer != "" {
			c.Response().Header().Set("HashSHA256", hashServer)
		}
		return c.JSONBlob(http.StatusOK, b)
	}

	//возможно стоит высчитывать хеш до gzip
	hashServer := hash.CalculateHashServer(jsonData)
	if hashServer != "" {
		c.Response().Header().Set("HashSHA256", hashServer)
	}
	return c.JSON(http.StatusOK, jsonData)
}

// ping - Функция для проверки соединения с базой данных PostgreSQL.
// Принимает контекст Echo и экземпляр подключения к базе данных.
// Возвращает статус-код 200 в случае успешного соединения,
// иначе - статус-код 500.
func ping(c echo.Context, db *postgresStorage.DB) error {
	if err := db.Ping(); err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusOK, "")
}

// AddMetrics - Эндпоинт для добавления нескольких метрик в формате JSON.
// Принимает срез структур Metric в теле запроса.
// Возвращает добавленные метрики в формате JSON и статус-код 200 в случае успеха,
// иначе - сообщение об ошибке и статус-код 400.
func AddMetrics(c echo.Context) error {
	var metrics []metricModel.Metric

	if err := c.Bind(&metrics); err != nil {
		return err
	}

	metricsSave := make(map[string]metricModel.Metric)
	//TODO придумать что-нибудь получше
	var newMetric bool
	for _, metricValue := range metrics {
		if metricValue.MType == metricModel.MetricTypeCounter {
			metric, _ := storage.Storage.Get(metricValue.ID)

			if _, ok := metricsSave[metricValue.ID]; !ok {
				metricsSave[metricValue.ID] = metricValue
				newMetric = true
			}

			if metric.Delta != nil {
				valueMetric := metric.Delta
				*metricsSave[metricValue.ID].Delta += *valueMetric
			} else if !newMetric {
				*metricsSave[metricValue.ID].Delta += *metricValue.Delta
			}

			newMetric = false
			continue
		}

		metricsSave[metricValue.ID] = metricValue
	}

	err := storage.Storage.SetAll(metricsSave)

	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	jsonData, err := json.Marshal(metrics)

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

		//не получилось перезаписать данные в body используя middleware
		//возможно стоит высчитывать хеш до gzip
		hashServer := hash.CalculateHashServer(b)
		if hashServer != "" {
			c.Response().Header().Set("HashSHA256", hashServer)
		}

		return c.JSONBlob(http.StatusOK, b)
	}

	//не получилось перезаписать данные в body используя middleware
	hashServer := hash.CalculateHashServer(jsonData)
	if hashServer != "" {
		c.Response().Header().Set("HashSHA256", hashServer)
	}
	return c.JSON(http.StatusOK, jsonData)
}

func main() {
	printBuildInfo()
	cfg, err := config.LoadServer()

	if err != nil {
		fmt.Printf("Failed to prepare server config: %v\n", err)
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger)
	e.Use(middleware.CheckHash)
	e.Use(middleware.UnzipMiddleware)
	e.Use(middleware.DecodeMiddleware)

	e.POST("/update/:type_metric/:name_metric/:value_metric", AddMetric)
	e.GET("/value/:type_metric/:name_metric", getMetric)
	e.GET("/", getAllMetrics)

	e.POST("/update/", AddMetricV2)
	e.POST("/value/", GetMetricV2)

	e.POST("/updates/", AddMetrics)

	if cfg.DatabaseDsn != "" {
		db, err := postgresStorage.NewDB()
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			panic(err)
		}
		defer db.Pool.Close()

		e.GET("/ping", func(c echo.Context) error {
			return ping(c, db)
		})

		storage.Storage = db
		m, err := migrator.NewMigrator()

		if err != nil {
			log.Fatal(err.Error())
		}

		if err := m.Up(); err != nil {
			log.Fatal(err.Error())
		}
	} else {
		storage.Storage = memStorage.NewStorage()
	}

	// Канал для сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	//TODO вынести логику в отдельный файл
	err = files.InitMetrics()

	if err != nil {
		log.Fatal(err.Error())
	}

	go files.UpdateMetrics()

	fmt.Println("Running server on", cfg.FlagRunAddr)
	echopprof.Wrap(e)

	go func() {
		if err := e.Start(cfg.FlagRunAddr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Ожидаем сигнал завершения
	<-stop

	// Запускаем graceful shutdown
	if err := gracefulShutdown(e, storage.Storage); err != nil {
		log.Fatalf("Error during graceful shutdown: %v", err)
	}
}

func gracefulShutdown(e *echo.Echo, store storage.StorageInterface) error {
	// Устанавливаем таймаут для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	// Ожидаем завершения всех текущих запросов
	// Это время также используется для завершения сохранения метрик
	<-ctx.Done()

	// Закрываем соединение с базой данных
	store.Close()
	log.Println("Graceful shutdown completed")
	return nil
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
