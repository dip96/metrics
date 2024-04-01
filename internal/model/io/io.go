package io

import metricModel "github.com/dip96/metrics/internal/model/metric"

type ProducerInterface interface {
	WriteEvent(metricModel.Metric) error
	Close() error
}

type ConsumerInterface interface {
	ReadEvent() (*metricModel.Metric, error)
	Close() error
}
