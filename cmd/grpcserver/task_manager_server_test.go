package main_test

import (
	"myproject/adapters/grpcserver"
	"myproject/infrastructure/testhelpers"
	"myproject/specifications"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTaskManageServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	port := nat.Port("50051/tcp")
	baseURL := testhelpers.StartDockerServer(t, port, "grpcserver", wait.ForListeningPort(port), false)
	driver := grpcserver.Driver{Addr: baseURL}

	t.Run("happy path", func(t *testing.T) {
		specifications.TaskManagerSpecification(t, &driver)
	})
}
