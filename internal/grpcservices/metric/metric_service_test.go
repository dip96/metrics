package metric

import (
	"context"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/dip96/metrics/internal/storage"
	postgresStorage "github.com/dip96/metrics/internal/storage/postgres"
	pbBase "github.com/dip96/metrics/protobuf/protos/metric/base"
	pbV1 "github.com/dip96/metrics/protobuf/protos/metric/v1"
	pbV2 "github.com/dip96/metrics/protobuf/protos/metric/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"testing"
)

func InitTestDB() (*postgresStorage.DB, error) {
	db, err := postgresStorage.NewDB()
	if err != nil {
		return nil, err
	}

	storage.Storage = db
	return db, nil
}

func TestMetricService_AddMetric(t *testing.T) {
	db, err := InitTestDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	service := NewMetricService(db)

	tests := []struct {
		name string
		req  *pbV1.AddMetricRequest
		want *pbV1.AddMetricResponse
	}{
		{
			name: "Add Gauge Metric",
			req: &pbV1.AddMetricRequest{
				Type:  pbBase.MetricType_GAUGE,
				Name:  "test_gauge",
				Value: "10.5",
			},
			want: &pbV1.AddMetricResponse{Success: true},
		},
		{
			name: "Add Counter Metric",
			req: &pbV1.AddMetricRequest{
				Type:  pbBase.MetricType_COUNTER,
				Name:  "test_counter",
				Value: "5",
			},
			want: &pbV1.AddMetricResponse{Success: true},
		},
		{
			name: "Invalid Gauge Value",
			req: &pbV1.AddMetricRequest{
				Type:  pbBase.MetricType_GAUGE,
				Name:  "invalid_gauge",
				Value: "not_a_number",
			},
			want: &pbV1.AddMetricResponse{Success: false},
		},
		{
			name: "Invalid Counter Value",
			req: &pbV1.AddMetricRequest{
				Type:  pbBase.MetricType_COUNTER,
				Name:  "invalid_counter",
				Value: "not_a_number",
			},
			want: &pbV1.AddMetricResponse{Success: false},
		},
		{
			name: "Invalid Metric Type",
			req: &pbV1.AddMetricRequest{
				Type:  pbBase.MetricType(999),
				Name:  "invalid_type",
				Value: "10",
			},
			want: &pbV1.AddMetricResponse{Success: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.AddMetric(context.Background(), tt.req)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			//проверяем сохраненые метрики
			if tt.want.Success {
				metric, err := db.Get(tt.req.Name)
				assert.NoError(t, err)
				assert.Equal(t, tt.req.Name, metric.ID)

				switch tt.req.Type {
				case pbBase.MetricType_GAUGE:
					val, err := strconv.ParseFloat(tt.req.Value, 64)
					assert.NoError(t, err)
					assert.Equal(t, val, *metric.Value)
				case pbBase.MetricType_COUNTER:
					delta, _ := strconv.ParseInt(tt.req.Value, 10, 64)
					assert.Equal(t, delta, *metric.Delta)
				}
			}
		})
	}
	db.Clear()
}

func TestMetricService_AddMetricV2(t *testing.T) {
	db, err := InitTestDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	service := NewMetricService(db)

	tests := []struct {
		name    string
		req     *pbV2.AddMetricV2Request
		want    *pbV2.AddMetricV2Response
		wantErr bool
	}{
		{
			name: "Add Gauge Metric",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id:    "test_gauge",
					Type:  pbBase.MetricType_GAUGE,
					Value: 10.5,
				},
			},
			want: &pbV2.AddMetricV2Response{
				Metric: &pbBase.Metric{
					Id:    "test_gauge",
					Type:  pbBase.MetricType_GAUGE,
					Value: 10.5,
				},
			},
			wantErr: false,
		},
		{
			name: "Add Counter Metric",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id:    "test_counter",
					Type:  pbBase.MetricType_COUNTER,
					Delta: 5,
				},
			},
			want: &pbV2.AddMetricV2Response{
				Metric: &pbBase.Metric{
					Id:    "test_counter",
					Type:  pbBase.MetricType_COUNTER,
					Delta: 5,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Metric Type",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id:   "invalid_type",
					Type: pbBase.MetricType(999),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.AddMetricV2(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// Дополнительная проверка: получим метрику из базы данных и проверим её
			metric, err := db.Get(tt.req.Metric.Id)
			assert.NoError(t, err)
			assert.Equal(t, tt.req.Metric.Id, metric.ID)

			switch tt.req.Metric.Type {
			case pbBase.MetricType_GAUGE:
				assert.Equal(t, tt.req.Metric.Value, *metric.Value)
				assert.Equal(t, metricModel.MetricTypeGauge, metric.MType)
			case pbBase.MetricType_COUNTER:
				assert.Equal(t, tt.req.Metric.Delta, *metric.Delta)
				assert.Equal(t, metricModel.MetricTypeCounter, metric.MType)
			}
		})
	}
	db.Clear()
}

func TestMetricService_GetMetricV2(t *testing.T) {
	db, err := InitTestDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	service := NewMetricService(db)

	gaugeValue := 10.5
	counterValue := int64(5)
	testMetrics := []metricModel.Metric{
		{
			ID:    "test_gauge",
			MType: metricModel.MetricTypeGauge,
			Value: &gaugeValue,
		},
		{
			ID:    "test_counter",
			MType: metricModel.MetricTypeCounter,
			Delta: &counterValue,
		},
	}

	// Добавление тестовых метрик в базу данных
	for _, m := range testMetrics {
		err := db.Set(m)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		req     *pbV2.AddMetricV2Request
		want    *pbV2.AddMetricV2Response
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "Get Gauge Metric",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id: "test_gauge",
				},
			},
			want: &pbV2.AddMetricV2Response{
				Metric: &pbBase.Metric{
					Id:    "test_gauge",
					Type:  pbBase.MetricType_GAUGE,
					Value: 10.5,
				},
			},
			wantErr: false,
		},
		{
			name: "Get Counter Metric",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id: "test_counter",
				},
			},
			want: &pbV2.AddMetricV2Response{
				Metric: &pbBase.Metric{
					Id:    "test_counter",
					Type:  pbBase.MetricType_COUNTER,
					Delta: 5,
				},
			},
			wantErr: false,
		},
		{
			name: "Get Non-existent Metric",
			req: &pbV2.AddMetricV2Request{
				Metric: &pbBase.Metric{
					Id: "non_existent",
				},
			},
			want:    nil,
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetMetricV2(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != codes.OK {
					statusErr, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.errCode, statusErr.Code())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	db.Clear()
}

func TestMetricService_GetMetric(t *testing.T) {
	db, err := InitTestDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	service := NewMetricService(db)

	// Подготовка тестовых данных
	gaugeValue := 10.5
	counterValue := int64(5)
	testMetrics := []metricModel.Metric{
		{
			ID:             "test_gauge",
			MType:          metricModel.MetricTypeGauge,
			Value:          &gaugeValue,
			FullValueGauge: strconv.FormatFloat(gaugeValue, 'f', -1, 64),
		},
		{
			ID:    "test_counter",
			MType: metricModel.MetricTypeCounter,
			Delta: &counterValue,
		},
	}

	// Добавление тестовых метрик в базу данных
	for _, m := range testMetrics {
		err := db.Set(m)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		req     *pbV1.GetMetricRequest
		want    *pbV1.GetMetricResponse
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "Get Gauge Metric",
			req: &pbV1.GetMetricRequest{
				Name: "test_gauge",
				Type: pbBase.MetricType_GAUGE,
			},
			want: &pbV1.GetMetricResponse{
				Value: "10.500000",
			},
			wantErr: false,
		},
		{
			name: "Get Counter Metric",
			req: &pbV1.GetMetricRequest{
				Name: "test_counter",
				Type: pbBase.MetricType_COUNTER,
			},
			want: &pbV1.GetMetricResponse{
				Value: "5",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetMetric(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != codes.OK {
					statusErr, ok := status.FromError(err)
					assert.True(t, ok)
					assert.Equal(t, tt.errCode, statusErr.Code())
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	db.Clear()
}

func TestMetricService_GetAllMetricsHTML(t *testing.T) {
	db, err := InitTestDB()
	require.NoError(t, err)
	defer db.Pool.Close()

	service := NewMetricService(db)

	// Подготовка тестовых данных
	gaugeValue := 10.5
	counterValue := int64(5)
	testMetrics := []metricModel.Metric{
		{
			ID:    "test_gauge",
			MType: metricModel.MetricTypeGauge,
			Value: &gaugeValue,
		},
		{
			ID:    "test_counter",
			MType: metricModel.MetricTypeCounter,
			Delta: &counterValue,
		},
	}

	// Добавление тестовых метрик в базу данных
	for _, m := range testMetrics {
		err := db.Set(m)
		require.NoError(t, err)
	}

	t.Run("Get All Metrics HTML", func(t *testing.T) {
		got, err := service.GetAllMetricsHTML(context.Background(), &pbV1.GetAllMetricsHTMLRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, got)

		// Проверяем, что HTML содержит ожидаемые элементы
		assert.True(t, strings.Contains(got.HtmlContent, "<html><body><ul>"))
		assert.True(t, strings.Contains(got.HtmlContent, "</ul></body></html>"))
		assert.True(t, strings.Contains(got.HtmlContent, "<li>test_gauge: 10.500000</li>"))
		assert.True(t, strings.Contains(got.HtmlContent, "<li>test_counter: 5</li>"))
	})

	t.Run("Get All Metrics HTML - Empty Database", func(t *testing.T) {
		// Очищаем базу данных
		err := db.Clear()
		require.NoError(t, err)

		got, err := service.GetAllMetricsHTML(context.Background(), &pbV1.GetAllMetricsHTMLRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, got)

		// Проверяем, что HTML содержит только базовую структуру
		assert.Equal(t, "<html><body><ul></ul></body></html>", got.HtmlContent)
	})
	db.Clear()
}
