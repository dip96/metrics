package metric

import (
	"errors"
	"fmt"
)

func (m Metric) GetValueForDisplay() (string, error) {
	if m.MType == MetricTypeCounter {
		return fmt.Sprintf("%d", *m.Delta), nil
	}

	if m.MType == MetricTypeGauge {
		if m.FullValueGauge == "" {
			return fmt.Sprintf("%f", *m.Value), nil
		}
		return m.FullValueGauge, nil
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
