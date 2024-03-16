package files

import (
	"bufio"
	"encoding/json"
	"github.com/dip96/metrics/internal/config"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

func NewProducer(filename string) {
	_, err := os.Stat(filename)

	if err == nil {
		tmpFile, err := os.CreateTemp("", "*.tmp")
		if err != nil {
			log.Errorln("Error creating the tmp file:", err.Error())
			return
		}
		defer os.Remove(tmpFile.Name())

		file, err := os.OpenFile(tmpFile.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Errorln("Error opening the tmp file:", err.Error())
			return
		}

		producer := &Producer{
			file:   file,
			writer: bufio.NewWriter(file),
		}

		SaveMetrics(producer)

		if err := producer.Close(); err != nil {
			log.Errorln("Error closing the tmp file:", err.Error())
			return
		}

		if err := os.Remove(filename); err != nil {
			log.Errorln("Error removing the old file:", err.Error())
			return
		}

		if err := os.Rename(tmpFile.Name(), filename); err != nil {
			log.Errorln("Error renaming the file:", err.Error())
			return
		}
		return
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorln("Error opening the file:", err.Error())
		return
	}

	producer := &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}
	SaveMetrics(producer)
}

func (p *Producer) WriteEvent(metric metricModel.Metric) error {
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

type Producer struct {
	file   *os.File
	writer *bufio.Writer
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

func (c *Consumer) ReadEvent() (*metricModel.Metric, error) {
	if !c.scanner.Scan() {
		if c.scanner.Err() == nil {
			return nil, io.EOF
		}
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	metric := metricModel.Metric{}
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

func SaveMetrics(producer *Producer) {
	metrics, _ := storage.Storage.GetAll()
	for metric := range metrics {
		if err := producer.WriteEvent(metrics[metric]); err != nil {
			log.Errorln(err)
		}
	}
}

func InitMetrics() {
	cfg := config.ServerConfig
	Consumer, err := NewConsumer(cfg.FileStoragePath)
	if err != nil {
		log.Errorln(err)
	}
	defer Consumer.Close()

	for {
		metric, err := Consumer.ReadEvent()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorln(err)
			continue
		}

		storage.Storage.Set(metric.ID, *metric)
	}
}

func UpdateMetrics() error {
	cfg := config.ServerConfig
	ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	if cfg.Restore {
		for range ticker.C {
			NewProducer(cfg.FileStoragePath)
		}
	}
	return nil
}
