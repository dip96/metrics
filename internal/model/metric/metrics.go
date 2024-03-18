package metric

type MetricService interface {
	GetValueForDisplay() (string, error)
	GetValue() (string, error)
}
