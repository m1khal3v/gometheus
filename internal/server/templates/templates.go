package templates

import (
	"embed"
	"fmt"
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
	template, err := storage.getTemplate("get_all_metrics")
	if err != nil {
		return err
	}

	return template.Execute(writer, metrics)
}

func (storage *Storage) getTemplate(name string) (*template.Template, error) {
	value, ok := storage.templateMap.Load(name)
	if !ok {
		content, err := embedTemplates.ReadFile(fmt.Sprintf("%s.gohtml", name))
		if err != nil {
			return nil, err
		}

		value, err = template.New(name).Parse(string(content))
		if err != nil {
			return nil, err
		}

		storage.templateMap.Store(name, value)
	}

	return value.(*template.Template), nil
}
