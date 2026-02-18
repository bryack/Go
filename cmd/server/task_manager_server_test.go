package main_test

import (
	"myproject/adapters/webserver"
	"myproject/specifications"
	"testing"
)

func TestTaskManagerServer(t *testing.T) {
	driver := webserver.Driver{BaseURL: "http://localhost:8080"}
	specifications.TaskManagerSpecification(t, driver)

}
