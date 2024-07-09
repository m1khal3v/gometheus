package dump

import (
	"fmt"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		storage       storage.Storage
		filepath      string
		storeInterval uint32
		restore       bool
	}
	storage := memory.New()
	tests := []struct {
		name      string
		args      args
		want      *Storage
		wantPanic string
	}{
		{
			name: "nil storage",
			args: args{
				storage: nil,
			},
			wantPanic: "Decorated storage cannot be nil",
		},
		{
			name: "empty filepath",
			args: args{
				storage:  storage,
				filepath: "",
			},
			wantPanic: "Dump file path cannot be empty",
		},
		{
			name: "valid 1",
			args: args{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 0,
				restore:       false,
			},
			want: &Storage{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 0,
			},
		},
		{
			name: "valid 2",
			args: args{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 0,
				restore:       true,
			},
			want: &Storage{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 0,
			},
		},
		{
			name: "valid 3",
			args: args{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 3000,
				restore:       true,
			},
			want: &Storage{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 3000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic != "" {
				assert.PanicsWithValue(t, tt.wantPanic, func() {
					New(tt.args.storage, tt.args.filepath, tt.args.storeInterval, tt.args.restore)
				})
				return
			}
			assert.Equal(t, tt.want, New(tt.args.storage, tt.args.filepath, tt.args.storeInterval, tt.args.restore))
		})
	}
}

func TestStorage_Dump(t *testing.T) {
	tests := []struct {
		name      string
		items     []metric.Metric
		wantItems []metric.Metric
	}{
		{
			name:  "empty storage",
			items: []metric.Metric{},
		},
		{
			name: "storage with items",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.124),
				counter.New("m4", 331),
				counter.New("m5", 545),
			},
		},
		{
			name: "storage with name replace",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m1", 123),
				gauge.New("m1", 123.124),
				counter.New("m1", 331),
				counter.New("m1", 545),
			},
			wantItems: []metric.Metric{
				counter.New("m1", 545),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			file, err := os.CreateTemp("", "dump_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			defer os.Remove(file.Name())

			decorator := New(storage, file.Name(), 9999, false)
			for _, item := range tt.items {
				decorator.Save(nil, item)
			}
			decorator.Dump()

			assert.FileExists(t, file.Name())
			all, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}

			items := strings.Split(strings.TrimRight(string(all), "\n"), "\n")
			if len(items) == 1 && items[0] == "" {
				items = []string{}
			}
			if tt.wantItems == nil {
				tt.wantItems = tt.items
			}
			assert.Len(t, items, len(tt.wantItems))

			for _, metric := range tt.wantItems {
				assert.Contains(t, items, fmt.Sprintf(
					"{\"type\":\"%s\",\"name\":\"%s\",\"value\":\"%s\"}",
					metric.Type(),
					metric.Name(),
					metric.StringValue(),
				))
			}
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name  string
		items []metric.Metric
		want  metric.Metric
	}{
		{
			name:  "empty storage",
			items: []metric.Metric{},
		},
		{
			name: "storage with items",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.124),
				counter.New("m4", 331),
				counter.New("m5", 545),
			},
			want: gauge.New("m1", 123.321),
		},
		{
			name: "storage with name replace",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m1", 123),
				gauge.New("m1", 123.124),
				counter.New("m1", 331),
				counter.New("m1", 545),
			},
			want: counter.New("m1", 545),
		},
		{
			name: "storage without required item",
			items: []metric.Metric{
				counter.New("m2", 123),
				gauge.New("m3", 123.124),
				counter.New("m4", 331),
				counter.New("m5", 545),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			decorator := New(storage, "/tmp/test", 9999, false)
			for _, item := range tt.items {
				decorator.Save(nil, item)
			}
			assert.Equal(t, tt.want, decorator.Get(nil, "m1"))
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		items     []metric.Metric
		wantItems map[string]metric.Metric
	}{
		{
			name:      "empty storage",
			items:     []metric.Metric{},
			wantItems: map[string]metric.Metric{},
		},
		{
			name: "storage with items",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.124),
				counter.New("m4", 331),
				counter.New("m5", 545),
			},
			wantItems: map[string]metric.Metric{
				"m1": gauge.New("m1", 123.321),
				"m2": counter.New("m2", 123),
				"m3": gauge.New("m3", 123.124),
				"m4": counter.New("m4", 331),
				"m5": counter.New("m5", 545),
			},
		},
		{
			name: "storage with name replace",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m1", 123),
				gauge.New("m1", 123.124),
				counter.New("m1", 331),
				counter.New("m1", 545),
			},
			wantItems: map[string]metric.Metric{
				"m1": counter.New("m1", 545),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			decorator := New(storage, "/tmp/test", 9999, false)
			for _, item := range tt.items {
				decorator.Save(nil, item)
			}
			assert.Equal(t, tt.wantItems, decorator.GetAll(nil))
		})
	}
}

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name   string
		metric metric.Metric
	}{
		{
			name:   "gauge",
			metric: gauge.New("m1", 123.321),
		},
		{
			name:   "counter",
			metric: counter.New("m2", 123),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			decorator := New(storage, "/tmp/test", 9999, false)
			decorator.Save(nil, tt.metric)
			assert.Equal(t, tt.metric, decorator.Get(nil, tt.metric.Name()))
		})
	}
}

func Test_restoreFromFile(t *testing.T) {
	tests := []struct {
		name      string
		items     []metric.Metric
		wantItems map[string]metric.Metric
	}{
		{
			name:      "empty storage",
			items:     []metric.Metric{},
			wantItems: map[string]metric.Metric{},
		},
		{
			name: "storage with items",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m2", 123),
				gauge.New("m3", 123.124),
				counter.New("m4", 331),
				counter.New("m5", 545),
			},
			wantItems: map[string]metric.Metric{
				"m1": gauge.New("m1", 123.321),
				"m2": counter.New("m2", 123),
				"m3": gauge.New("m3", 123.124),
				"m4": counter.New("m4", 331),
				"m5": counter.New("m5", 545),
			},
		},
		{
			name: "storage with name replace",
			items: []metric.Metric{
				gauge.New("m1", 123.321),
				counter.New("m1", 123),
				gauge.New("m1", 123.124),
				counter.New("m1", 331),
				counter.New("m1", 545),
			},
			wantItems: map[string]metric.Metric{
				"m1": counter.New("m1", 545),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.New()
			file, err := os.CreateTemp("", "restore_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			defer os.Remove(file.Name())

			decorator := New(storage, file.Name(), 9999, false)
			for _, item := range tt.items {
				decorator.Save(nil, item)
			}
			decorator.Dump()

			restorage := memory.New()
			restoreFromFile(restorage, file.Name())
			assert.Equal(t, tt.wantItems, restorage.GetAll(nil))
		})
	}
}
