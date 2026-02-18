package main_test

import (
	"context"
	"fmt"
	"myproject/adapters/webserver"
	"myproject/specifications"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTaskManagerServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../.",
			Dockerfile: "./Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp"),
		Env: map[string]string{
			"TASKMANAGER_JWT_SECRET": "test-only-secret-min32chars-long",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	mappedPort, err := container.MappedPort(ctx, "8080/tcp")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	driver := webserver.Driver{BaseURL: baseURL}
	specifications.TaskManagerSpecification(t, driver)
}
