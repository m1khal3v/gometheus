package pgsql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand/v2"
	"net"
	"testing"
)

var baseDSN string
var connection *pgx.Conn

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		New(prepareDSN(t, "new"))
	})
}

func TestMain(m *testing.M) {
	if !tryConnectToExistingPostgres() {
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

		defer func() {
			if err := connection.Close(context.Background()); err != nil {
				log.Fatalf("Could not close connection: %s", err)
			}
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("Could not purge resource: %s", err)
			}
		}()
	}

	m.Run()
}

func tryConnectToExistingPostgres() bool {
	baseDSN = fmt.Sprintf("postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable")
	var err error
	connection, err = pgx.Connect(context.Background(), baseDSN)

	return err == nil
}

func prepareDSN(t *testing.T, name string) string {
	t.Helper()
	name = pgx.Identifier{fmt.Sprintf("test%s%d", name, rand.Uint32())}.Sanitize()
	if _, err := connection.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", name)); err != nil {
		t.Fatalf("Could not create schema: %s", err)
	}

	return fmt.Sprintf("%s&search_path=%s", baseDSN, name)
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
