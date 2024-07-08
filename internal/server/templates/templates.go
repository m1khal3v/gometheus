package templates

import (
	"embed"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"html/template"
	"io"
	"sync"
)

type Storage struct {
	templateMap sync.Map
}

//go:embed *.gohtml
var embedTemplates embed.FS

func New() *Storage {
	return &Storage{
		templateMap: sync.Map{},
	}
}

func (storage *Storage) ExecuteAllMetricsTemplate(writer io.Writer, metrics <-chan metric.Metric) error {
	content, err := embedTemplates.ReadFile("get_all_metrics.gohtml")
	if err != nil {
		return err
	}

	template, err := storage.getTemplate("get_all_metrics", string(content))
	if err != nil {
		return err
	}

	return template.Execute(writer, metrics)
}

func (storage *Storage) getTemplate(name, content string) (*template.Template, error) {
	value, ok := storage.templateMap.Load(name)
	if !ok {
		var err error
		value, err = template.New(name).Parse(content)
		if err != nil {
			return nil, err
		}

		storage.templateMap.Store(name, value)
	}

	return value.(*template.Template), nil
}
