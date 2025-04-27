package templates

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
)

func generateMetrics() <-chan metric.Metric {
	ch := make(chan metric.Metric, 3)
	ch <- gauge.New("metric1", 10)
	ch <- counter.New("metric2", 20)
	ch <- gauge.New("metric3", 30)
	close(ch)
	return ch
}

func TestStorage_getTemplate(t *testing.T) {
	storage := New()

	templateName := "get_all_metrics"
	tmpl, err := storage.getTemplate(templateName)
	if err != nil {
		t.Fatalf("expected no error, but got one: %v", err)
	}
	if tmpl == nil {
		t.Fatalf("expected template, but got nil")
	}

	_, err = storage.getTemplate("non_existent_template")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestStorage_ExecuteAllMetricsTemplate(t *testing.T) {
	storage := New()

	metrics := generateMetrics()

	var buf bytes.Buffer
	writer := io.Writer(&buf)

	err := storage.ExecuteAllMetricsTemplate(writer, metrics)
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected non-empty output, but got empty")
	}
}

func TestStorage_ExecuteAllMetricsTemplate_Errors(t *testing.T) {
	storage := New()

	mockWriter := &failingWriter{}

	metrics := generateMetrics()
	err := storage.ExecuteAllMetricsTemplate(mockWriter, metrics)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("mock writer error")
}
