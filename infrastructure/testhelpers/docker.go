package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartDockerServer(t testing.TB, port nat.Port, dockerfilePath string) string {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../.",
			Dockerfile: dockerfilePath,
		},
		ExposedPorts: []string{string(port)},
		WaitingFor:   wait.ForHTTP("/health").WithPort(port),
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

	mappedPort, err := container.MappedPort(ctx, port)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	return baseURL
}

func StartDockerGRPCServer(t testing.TB, port nat.Port, dockerfilePath string) string {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../.",
			Dockerfile: dockerfilePath,
		},
		ExposedPorts: []string{string(port)},
		WaitingFor:   wait.ForLog("gRPC server listening at"),
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

	mappedPort, err := container.MappedPort(ctx, port)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	return baseURL
}
