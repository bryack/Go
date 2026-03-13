package grpcserver

import (
	"context"
	"fmt"
	"myproject/application"
	"myproject/domain"
	"myproject/infrastructure/testhelpers"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	assert.Equal(t, codes.OK, status.Code(err))

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
	assert.Equal(t, codes.OK, status.Code(err))

	assert.Len(t, reply.Tasks, 2)
	assert.Equal(t, "task 1", reply.Tasks[0].Description)
	assert.Equal(t, "task 2", reply.Tasks[1].Description)
	assert.Equal(t, int32(1), reply.Tasks[0].Id)
	assert.False(t, reply.Tasks[0].Done)
	assert.Equal(t, int32(2), reply.Tasks[1].Id)
	assert.True(t, reply.Tasks[1].Done)

	assert.Equal(t, testUserID, taskService.LastUserID)
}

func TestErrorsMapping(t *testing.T) {
	t.Run("CreateTask empty description", func(t *testing.T) {
		taskService := &testhelpers.SpyTaskService{
			ResultErr: domain.ErrDescriptionRequired,
		}
		store := &testhelpers.StubTaskStore{}
		authService := &testhelpers.SpyAuthService{}
		server := NewTaskManageServer(store, authService, taskService)

		testUserID := 99
		ctx := context.WithValue(context.Background(), application.UserIDKey, testUserID)

		request := &CreateTaskRequest{Description: ""}
		_, err := server.CreateTask(ctx, request)
		require.Error(t, err)
		status, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, status.Code())
		assert.Contains(t, status.Message(), domain.ErrDescriptionRequired.Error())
	})

	t.Run("CreateTask description more than 200 symbols", func(t *testing.T) {
		taskService := &testhelpers.SpyTaskService{
			ResultErr: domain.ErrDescriptionTooLong,
		}
		store := &testhelpers.StubTaskStore{}
		authService := &testhelpers.SpyAuthService{}
		server := NewTaskManageServer(store, authService, taskService)

		testUserID := 99
		ctx := context.WithValue(context.Background(), application.UserIDKey, testUserID)

		desc := fmt.Sprintf("task %s", strings.Repeat("1", 200))
		request := &CreateTaskRequest{Description: desc}
		_, err := server.CreateTask(ctx, request)
		require.Error(t, err)
		status, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, status.Code())
		assert.Contains(t, status.Message(), domain.ErrDescriptionTooLong.Error())
	})
}
