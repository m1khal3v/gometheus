package storage

type Storage interface {
	Save(metric *Metric) error
	Get(name string) *Metric
}
