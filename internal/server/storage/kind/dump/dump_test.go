package dump

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/internal/server/storage"
	"github.com/m1khal3v/gometheus/internal/server/storage/kind/memory"
	"github.com/m1khal3v/gometheus/pkg/slice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "valid 1",
			args: args{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 0,
				restore:       false,
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
		},
		{
			name: "valid 3",
			args: args{
				storage:       storage,
				filepath:      "/tmp/dump.json",
				storeInterval: 3000,
				restore:       true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.wantPanic != "" {
				assert.PanicsWithValue(t, tt.wantPanic, func() {
					New(ctx, tt.args.storage, tt.args.filepath, tt.args.storeInterval, tt.args.restore)
				})
				return
			}
			_, err := New(ctx, tt.args.storage, tt.args.filepath, tt.args.storeInterval, tt.args.restore)
			require.NoError(t, err)
		})
	}
}

func TestStorage_dump(t *testing.T) {
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
			ctx := context.Background()
			storage := memory.New()
			file, err := os.CreateTemp("", "dump_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			defer os.Remove(file.Name())

			decorator, err := New(ctx, storage, file.Name(), 9999, false)
			require.NoError(t, err)
			for _, item := range tt.items {
				decorator.Save(ctx, item)
			}
			decorator.dump(ctx)

			require.FileExists(t, file.Name())
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
			require.Len(t, items, len(tt.wantItems))

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
			ctx := context.Background()
			storage := memory.New()
			decorator, err := New(ctx, storage, "/tmp/test", 9999, false)
			require.NoError(t, err)
			for _, item := range tt.items {
				decorator.Save(ctx, item)
			}
			metric, err := decorator.Get(ctx, "m1")
			require.NoError(t, err)
			assert.Equal(t, tt.want, metric)
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		items     []metric.Metric
		wantItems []metric.Metric
	}{
		{
			name:      "empty storage",
			items:     []metric.Metric{},
			wantItems: []metric.Metric{},
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
			wantItems: []metric.Metric{
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
			ctx := context.Background()
			storage := memory.New()
			decorator, err := New(ctx, storage, "/tmp/test", 9999, false)
			require.NoError(t, err)
			for _, item := range tt.items {
				require.NoError(t, decorator.Save(ctx, item))
			}

			all, err := decorator.GetAll(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantItems, slice.FromChannel(all))
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
			ctx := context.Background()
			storage := memory.New()
			decorator, err := New(ctx, storage, "/tmp/test", 9999, false)
			require.NoError(t, err)
			decorator.Save(ctx, tt.metric)
			metric, err := decorator.Get(ctx, tt.metric.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.metric, metric)
		})
	}
}

func Test_restoreFromFile(t *testing.T) {
	tests := []struct {
		name      string
		items     []metric.Metric
		wantItems []metric.Metric
	}{
		{
			name:      "empty storage",
			items:     []metric.Metric{},
			wantItems: []metric.Metric{},
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
			wantItems: []metric.Metric{
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
			ctx := context.Background()
			storage := memory.New()
			file, err := os.CreateTemp("", "restore_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			defer os.Remove(file.Name())

			decorator, err := New(ctx, storage, file.Name(), 9999, false)
			require.NoError(t, err)
			for _, item := range tt.items {
				decorator.Save(ctx, item)
			}
			require.NoError(t, decorator.dump(ctx))

			decorator.restoreFromFile(ctx)
			all, err := decorator.GetAll(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantItems, slice.FromChannel(all))
		})
	}
}
