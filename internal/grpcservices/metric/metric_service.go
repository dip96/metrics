package metric

import (
	"bytes"
	"context"
	"fmt"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	pbBase "github.com/dip96/metrics/protobuf/protos/metric/base"
	pbV1 "github.com/dip96/metrics/protobuf/protos/metric/v1"
	pbV2 "github.com/dip96/metrics/protobuf/protos/metric/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

type MetricService struct {
	//pbV1.UnimplementedMetricServiceServer
	//pbV2.UnimplementedMetricServiceServer
	storage storage.StorageInterface
}

func NewMetricService(storage storage.StorageInterface) *MetricService {
	return &MetricService{storage: storage}
}

func (s *MetricService) AddMetric(ctx context.Context, req *pbV1.AddMetricRequest) (*pbV1.AddMetricResponse, error) {
	metric, _ := storage.Storage.Get(req.Name)
	metric.ID = req.Name

	switch req.Type {
	case pbBase.MetricType_GAUGE:
		value, err := strconv.ParseFloat(req.Value, 64)
		if err != nil {
			return &pbV1.AddMetricResponse{Success: false}, nil
		}
		metric.MType = metricModel.MetricTypeGauge
		metric.Value = &value
		metric.FullValueGauge = req.Value

	case pbBase.MetricType_COUNTER:
		value, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return &pbV1.AddMetricResponse{Success: false}, nil
		}
		metric.MType = metricModel.MetricTypeCounter
		if metric.Delta == nil {
			metric.Delta = &value
		} else {
			*metric.Delta += value
		}

	default:
		return &pbV1.AddMetricResponse{Success: false}, nil
	}

	err := storage.Storage.Set(metric)
	if err != nil {
		return &pbV1.AddMetricResponse{Success: false}, nil
	}

	return &pbV1.AddMetricResponse{Success: true}, nil
}

func (s *MetricService) AddMetricV2(ctx context.Context, req *pbV2.AddMetricV2Request) (*pbV2.AddMetricV2Response, error) {
	metric := metricModel.Metric{
		ID:    req.Metric.Id,
		MType: protoMetricTypeToModelMetricType(req.Metric.Type),
	}

	if req.Metric.Type == pbBase.MetricType_GAUGE {
		metric.Value = &req.Metric.Value
		metric.FullValueGauge = fmt.Sprintf("%f", req.Metric.Value)
	} else if req.Metric.Type == pbBase.MetricType_COUNTER {
		metric.Delta = &req.Metric.Delta
	} else {
		return nil, fmt.Errorf("invalid metric type")
	}

	err := s.storage.Set(metric)
	if err != nil {
		return nil, err
	}

	respMetric := &pbBase.Metric{
		Id:   metric.ID,
		Type: MetricTypeToProto(metric.MType),
	}

	switch req.Metric.Type {
	case pbBase.MetricType_GAUGE:
		respMetric.Value = *metric.Value
	case pbBase.MetricType_COUNTER:
		respMetric.Delta = *metric.Delta
	}

	return &pbV2.AddMetricV2Response{
		Metric: respMetric,
	}, nil
}

func (s *MetricService) GetMetricV2(ctx context.Context, req *pbV2.AddMetricV2Request) (*pbV2.AddMetricV2Response, error) {
	nameMetric := req.Metric.Id
	metric, err := s.storage.Get(nameMetric)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "метрика не найдена: %v", err)
	}

	// Преобразование метрики в формат protobuf
	pbMetric := &pbBase.Metric{
		Id:   metric.ID,
		Type: MetricTypeToProto(metric.MType),
	}

	switch metric.MType {
	case metricModel.MetricTypeCounter:
		pbMetric.Delta = *metric.Delta
	case metricModel.MetricTypeGauge:
		pbMetric.Value = *metric.Value
	default:
		return nil, status.Errorf(codes.Internal, "неподдерживаемый тип метрики")
	}

	return &pbV2.AddMetricV2Response{
		Metric: pbMetric,
	}, nil
}

func (s *MetricService) GetMetric(ctx context.Context, req *pbV1.GetMetricRequest) (*pbV1.GetMetricResponse, error) {
	metric, err := s.storage.Get(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "метрика не найдена: %v", err)
	}

	value, err := metric.GetValueForDisplay()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка получения значения метрики: %v", err)
	}

	return &pbV1.GetMetricResponse{
		Value: value,
	}, nil
}

func (s *MetricService) GetAllMetricsHTML(ctx context.Context, req *pbV1.GetAllMetricsHTMLRequest) (*pbV1.GetAllMetricsHTMLResponse, error) {
	metrics, err := s.storage.GetAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка при получении метрик: %v", err)
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

	return &pbV1.GetAllMetricsHTMLResponse{
		HtmlContent: buf.String(),
	}, nil
}

func (s *MetricService) SendMetricsBatch(ctx context.Context, req *pbV1.SendMetricsBatchRequest) (*pbV1.SendMetricsBatchResponse, error) {
	for _, pbMetric := range req.Metrics {
		metric := metricModel.Metric{
			ID:    pbMetric.Id,
			MType: protoMetricTypeToModelMetricType(pbMetric.Type),
		}

		switch pbMetric.Type {
		case pbBase.MetricType_GAUGE:
			metric.Value = &pbMetric.Value
			metric.FullValueGauge = fmt.Sprintf("%f", pbMetric.Value)
		case pbBase.MetricType_COUNTER:
			metric.Delta = &pbMetric.Delta
		default:
			return &pbV1.SendMetricsBatchResponse{
				Success: false,
				Message: fmt.Sprintf("Invalid metric type for metric %s", pbMetric.Id),
			}, nil
		}

		err := s.storage.Set(metric)
		if err != nil {
			return &pbV1.SendMetricsBatchResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to set metric %s: %v", pbMetric.Id, err),
			}, nil
		}
	}

	return &pbV1.SendMetricsBatchResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully processed %d metrics", len(req.Metrics)),
	}, nil
}

func MetricTypeToProto(mType metricModel.MetricType) pbBase.MetricType {
	switch mType {
	case metricModel.MetricTypeGauge:
		return pbBase.MetricType_GAUGE
	case metricModel.MetricTypeCounter:
		return pbBase.MetricType_COUNTER
	default:
		log.Printf("Unknown metric type: %v", mType)
		return pbBase.MetricType_GAUGE
	}
}

func protoMetricTypeToModelMetricType(protoType pbBase.MetricType) metricModel.MetricType {
	switch protoType {
	case pbBase.MetricType_GAUGE:
		return metricModel.MetricTypeGauge
	case pbBase.MetricType_COUNTER:
		return metricModel.MetricTypeCounter
	default:
		return ""
	}
}
