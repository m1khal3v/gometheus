package storage

type Storage interface {
	Save(metric *Metric) error
}
