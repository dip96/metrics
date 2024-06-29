package files

import (
	"bufio"
	"encoding/json"
	metricModel "github.com/dip96/metrics/internal/model/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestProducer(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "test.txt")
	file, err := os.Create(tempFilePath)
	require.NoError(t, err)
	defer os.Remove(tempFilePath)

	producer := &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}
	defer producer.Close()

	t.Run("WriteEvent", func(t *testing.T) {
		metric := metricModel.Metric{
			ID:    "test_metric",
			MType: metricModel.MetricTypeGauge,
			Value: Float64Ptr(42),
		}

		err := producer.WriteEvent(metric)
		require.NoError(t, err)

		// Verify the content of the file
		file.Seek(0, 0)
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		line := scanner.Bytes()

		var result metricModel.Metric
		err = json.Unmarshal(line, &result)
		require.NoError(t, err)
		assert.Equal(t, metric, result)
	})
}

func TestConsumer(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "test.txt")
	file, err := os.Create(tempFilePath)
	require.NoError(t, err)
	defer os.Remove(tempFilePath)

	metric := metricModel.Metric{
		ID:    "test_metric",
		MType: metricModel.MetricTypeGauge,
		Value: Float64Ptr(42),
	}
	data, err := json.Marshal(&metric)
	require.NoError(t, err)

	_, err = file.Write(append(data, '\n'))
	require.NoError(t, err)

	file.Close()

	consumer, err := NewConsumer(tempFilePath)
	require.NoError(t, err)
	defer consumer.Close()

	t.Run("ReadEvent", func(t *testing.T) {
		result, err := consumer.ReadEvent()
		require.NoError(t, err)
		assert.Equal(t, &metric, result)
	})
}

func TestNewConsumer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFilePath := filepath.Join(tempDir, "test.txt")
		file, err := os.Create(tempFilePath)
		require.NoError(t, err)
		file.Close()

		consumer, err := NewConsumer(tempFilePath)
		require.NoError(t, err)
		defer consumer.Close()
		assert.NotNil(t, consumer)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewConsumer("/path/to/non/existing/file")
		assert.Error(t, err)
	})
}

// Mock implementation for testing
type MockProducer struct {
	writtenMetrics []metricModel.Metric
}

func (m *MockProducer) WriteEvent(metric metricModel.Metric) error {
	m.writtenMetrics = append(m.writtenMetrics, metric)
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

// MockConfig is a mock for the configuration
type MockConfig struct {
	DirStorageTmpPath string
}

// LoadServer mock implementation
func LoadServer() *MockConfig {
	return &MockConfig{
		DirStorageTmpPath: os.TempDir(),
	}
}

func TestInitTmpProducer(t *testing.T) {
	producer, _ := initTmpProducer()

	// Check if the file was created in the correct directory
	cfg := LoadServer()
	assert.Contains(t, producer.file.Name(), cfg.DirStorageTmpPath)

	// Check if the file can be written to
	testData := []byte("test data")
	_, err := producer.writer.Write(testData)
	assert.NoError(t, err)

	// Check if the buffer can be flushed to the file
	err = producer.writer.Flush()
	assert.NoError(t, err)
}

// Вспомогательная функция для создания указателя на float64
func Float64Ptr(f float64) *float64 {
	return &f
}
