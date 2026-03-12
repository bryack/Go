package grpcserver

import (
	"context"
	"myproject/application"
	"myproject/domain"
	"myproject/infrastructure/testhelpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTask(t *testing.T) {
	taskService := &testhelpers.SpyTaskService{
		ResultTask: domain.Task{ID: 42},
	}
	store := &testhelpers.StubTaskStore{}
	authService := &testhelpers.SpyAuthService{}
	server := NewTaskManageServer(store, authService, taskService)

	testUserID := 99
	ctx := context.WithValue(context.Background(), application.UserIDKey, testUserID)

	request := &CreateTaskRequest{Description: "Buy milk"}

	reply, err := server.CreateTask(ctx, request)
	require.NoError(t, err)

	assert.Equal(t, "Buy milk", taskService.LastDescription)
	assert.Equal(t, testUserID, taskService.LastUserID)
	assert.Equal(t, int32(42), reply.TaskId)
}
