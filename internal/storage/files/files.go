package files

import (
	"bufio"
	"encoding/json"
	"github.com/dip96/metrics/internal/config"
	ioModel "github.com/dip96/metrics/internal/model/io"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer
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

func (p *Producer) Close() error {
	cfg, err := config.LoadServer()

	if err != nil {
		return err
	}

	filename := cfg.FileStoragePath
	if err := p.file.Close(); err != nil {
		log.Errorln("Error closing the tmp file:", err.Error())
	}

	if err := os.Remove(filename); err != nil {
		log.Errorln("Error removing the old file:", err.Error())
	}

	if err := os.Rename(p.file.Name(), filename); err != nil {
		log.Errorln("Error renaming the file:", err.Error())
	}

	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
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
	c.file.Close()
	return c.file.Close()
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func SaveMetrics(producer ioModel.ProducerInterface) {
	metrics, _ := storage.Storage.GetAll()
	for metric := range metrics {
		if err := producer.WriteEvent(metrics[metric]); err != nil {
			log.Errorln(err)
		}
	}
}

// InitMetrics инициализирует метрики из файлового хранилища.
func InitMetrics() error {
	cfg, err := config.LoadServer()

	if err != nil {
		return err
	}
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

		err = storage.Storage.Set(*metric)

		if err != nil {
			log.Errorln(err)
			continue
		}
	}

	return nil
}

// UpdateMetrics обновляет метрики в файловом хранилище в бесконечном цикле.
func UpdateMetrics() error {
	cfg, err := config.LoadServer()

	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	if cfg.Restore {
		for range ticker.C {
			producer, err := initTmpProducer()
			if err != nil {
				return err
			}
			SaveMetrics(producer)
			producer.Close()
		}
	}
	return nil
}

func initTmpProducer() (*Producer, error) {
	cfg, err := config.LoadServer()

	if err != nil {
		log.Errorln("Failed to prepare server config:", err.Error())
		return nil, err
	}
	tmpFile, err := os.CreateTemp(cfg.DirStorageTmpPath, "*.tmp")
	if err != nil {
		log.Errorln("Error creating the tmp file:", err.Error())
		return nil, err
	}

	file, err := os.OpenFile(tmpFile.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorln("Error opening the tmp file:", err.Error())
		return nil, err
	}

	producer := &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}

	return producer, nil
}
