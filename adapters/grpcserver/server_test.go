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

func TestGetTasks(t *testing.T) {
	taskService := &testhelpers.SpyTaskService{
		TasksTable: []domain.Task{
			{ID: 1, Description: "task 1", Done: false},
			{ID: 2, Description: "task 2", Done: true},
		},
	}
	store := &testhelpers.StubTaskStore{}
	authService := &testhelpers.SpyAuthService{}
	server := NewTaskManageServer(store, authService, taskService)

	testUserID := 99
	ctx := context.WithValue(context.Background(), application.UserIDKey, testUserID)

	request := &GetTasksRequest{}

	reply, err := server.GetTasks(ctx, request)
	require.NoError(t, err)

	assert.Len(t, reply.Tasks, 2)
	assert.Equal(t, "task 1", reply.Tasks[0].Description)
	assert.Equal(t, "task 2", reply.Tasks[1].Description)
	assert.Equal(t, int32(1), reply.Tasks[0].Id)
	assert.False(t, reply.Tasks[0].Done)
	assert.Equal(t, int32(2), reply.Tasks[1].Id)
	assert.True(t, reply.Tasks[1].Done)

	assert.Equal(t, testUserID, taskService.LastUserID)
}
