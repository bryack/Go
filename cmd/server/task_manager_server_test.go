package main_test

import (
	"myproject/adapters/webserver"
	"myproject/infrastructure/testhelpers"
	"myproject/specifications"
	"net/http"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
)

func TestTaskManagerServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}
	port := nat.Port("8080/tcp")
	baseURL := testhelpers.StartDockerServer(t, port, "./Dockerfile")
	driver := webserver.Driver{BaseURL: baseURL, Client: &client}

	t.Run("happy path", func(t *testing.T) {
		specifications.TaskManagerSpecification(t, driver)
	})
	t.Run("isolation", func(t *testing.T) {
		specifications.TaskManagerSpecification_Isolation(t, driver)
	})
}
