package main_test

import (
	"myproject/adapters/grpcserver"
	"myproject/infrastructure/testhelpers"
	"myproject/specifications"
	"testing"

	"github.com/docker/go-connections/nat"
)

func TestTaskManagerServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	port := nat.Port("50051/tcp")
	baseURL := testhelpers.StartDockerGRPCServer(t, port, "./cmd/grpcserver/Dockerfile")
	driver := grpcserver.Driver{BaseURL: baseURL}

	t.Run("happy path", func(t *testing.T) {
		specifications.TaskManagerSpecification(t, driver)
	})
}
