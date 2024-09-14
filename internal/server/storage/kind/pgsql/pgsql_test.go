package pgsql

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/m1khal3v/gometheus/internal/common/metric"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/counter"
	"github.com/m1khal3v/gometheus/internal/common/metric/kind/gauge"
	"github.com/m1khal3v/gometheus/pkg/slice"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseDSN string
var connection *pgx.Conn

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		dsn, cleanup := prepareDSN(t)
		t.Cleanup(cleanup)

		New(dsn)
	})
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		preset     []metric.Metric
		metricName string
		want       metric.Metric
	}{
		{
			name: "one metric",
			preset: []metric.Metric{
				counter.New("m1", 123),
			},
			metricName: "m1",
			want:       counter.New("m1", 123),
		},
		{
			name: "multiple metrics",
			preset: []metric.Metric{
				counter.New("m1", 123),
				counter.New("m2", 321),
				gauge.New("m3", 123.321),
			},
			metricName: "m2",
			want:       counter.New("m2", 321),
		},
		{
			name:       "no metrics",
			preset:     []metric.Metric{},
			metricName: "m1",
			want:       nil,
		},
		{
			name: "metric mismatch",
			preset: []metric.Metric{
				counter.New("m1", 123),
			},
			metricName: "m2",
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := createStorage(t, ctx, tt.preset)
			got, err := storage.Get(ctx, tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		preset []metric.Metric
	}{
		{
			name: "one metric",
			preset: []metric.Metric{
				counter.New("m1", 123),
			},
		},
		{
			name: "multiple metrics",
			preset: []metric.Metric{
				counter.New("m1", 123),
				counter.New("m2", 321),
				gauge.New("m3", 123.321),
			},
		},
		{
			name:   "no metrics",
			preset: []metric.Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := createStorage(t, ctx, tt.preset)
			got, err := storage.GetAll(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.preset, slice.FromChannel(got))
		})
	}
}

func TestStorage_Reset(t *testing.T) {
	ctx := context.Background()
	storage := createStorage(t, ctx, []metric.Metric{
		counter.New("m1", 123),
		gauge.New("m3", 123.321),
		counter.New("m2", 321),
		gauge.New("m4", 321.123),
	})
	storage.Reset(ctx)
	got, err := storage.GetAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, []metric.Metric{}, slice.FromChannel(got))
}

func createStorage(t *testing.T, ctx context.Context, preset []metric.Metric) *Storage {
	t.Helper()
	dsn, cleanup := prepareDSN(t)
	t.Cleanup(cleanup)

	storage := New(dsn)
	t.Cleanup(func() {
		err := storage.Reset(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	switch len(preset) {
	case 0:
		break
	case 1:
		require.NoError(t, storage.Save(ctx, preset[0]))
	default:
		require.NoError(t, storage.SaveBatch(ctx, preset))
	}

	return storage
}

func prepareDSN(t *testing.T) (string, func()) {
	t.Helper()
	name := pgx.Identifier{fmt.Sprintf("%d%s", rand.UintN(1000), t.Name())}.Sanitize()
	if _, err := connection.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", name)); err != nil {
		t.Fatalf("Could not create schema: %s", err)
	}

	return fmt.Sprintf("%s&search_path=%s", baseDSN, name), func() {
		if _, err := connection.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE;", name)); err != nil {
			t.Fatalf("Could not drop schema: %s", err)
		}
	}
}

func TestMain(m *testing.M) {
	cleanup, ok := tryUseExistingPostgres()
	if !ok {
		cleanup = createPostgresContainer()
	}
	defer cleanup()

	m.Run()
}

func tryUseExistingPostgres() (func(), bool) {
	baseDSN = os.Getenv("TEST_DATABASE_DSN")
	if baseDSN == "" {
		baseDSN = os.Getenv("DATABASE_DSN")
	}

	var err error
	connection, err = pgx.Connect(context.Background(), baseDSN)

	return func() {
		if err := connection.Close(context.Background()); err != nil {
			log.Fatalf("Could not close connection: %s", err)
		}
	}, err == nil
}

func createPostgresContainer() func() {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	port, err := getFreePort()
	if err != nil {
		log.Fatalf("Could not get a free port: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{
				HostPort: fmt.Sprintf("%d", port),
			}},
		},
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=test",
			"POSTGRES_DB=test",
			"listen_addresses = '0.0.0.0'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := resource.Expire(180); err != nil {
		log.Fatalf("Could not set expire: %s", err)
	}

	baseDSN = fmt.Sprintf("postgres://test:test@%s/test?sslmode=disable", resource.GetHostPort("5432/tcp"))
	err = pool.Retry(func() error {
		connection, err = pgx.Connect(context.Background(), baseDSN)
		return err
	})
	if err != nil {
		log.Fatalf("Unable to connect to Postgres: %s", err)
	}

	return func() {
		if err := connection.Close(context.Background()); err != nil {
			log.Fatalf("Could not close connection: %s", err)
		}
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
