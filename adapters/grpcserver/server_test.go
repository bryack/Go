package grpcserver

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	authService := &testhelpers.SpyAuthService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewTaskManageServer(authService, taskService, logger)

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
	authService := &testhelpers.SpyAuthService{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewTaskManageServer(authService, taskService, logger)

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
	tests := []struct {
		name         string
		serviceErr   error
		expectedCode codes.Code
		call         func(ctx context.Context, s *TaskManageServer) (any, error)
	}{
		{
			name:         "CreateTask empty description",
			serviceErr:   domain.ErrDescriptionRequired,
			expectedCode: codes.InvalidArgument,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.CreateTask(ctx, &CreateTaskRequest{Description: ""})
			},
		},
		{
			name:         "CreateTask description more than 200 symbols",
			serviceErr:   domain.ErrDescriptionTooLong,
			expectedCode: codes.InvalidArgument,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.CreateTask(ctx, &CreateTaskRequest{Description: fmt.Sprintf("task %s", strings.Repeat("1", 200))})
			},
		},
		{
			name:         "GetTasks storage failure",
			serviceErr:   domain.ErrStorageFailure,
			expectedCode: codes.Internal,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.GetTasks(ctx, &GetTasksRequest{})
			},
		},
		{
			name:         "Register email already exists",
			serviceErr:   domain.ErrEmailAlreadyExists,
			expectedCode: codes.AlreadyExists,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.Register(ctx, &RegisterRequest{
					Email:    "testRegister@email.com",
					Password: "Register123",
				})
			},
		},
		{
			name:         "Register invalid password",
			serviceErr:   domain.ErrInvalidCredentials,
			expectedCode: codes.Unauthenticated,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.Register(ctx, &RegisterRequest{
					Email:    "testRegister@email.com",
					Password: "Register123",
				})
			},
		},
		{
			name:         "Login invalid emain",
			serviceErr:   domain.ErrInvalidEmail,
			expectedCode: codes.InvalidArgument,
			call: func(ctx context.Context, s *TaskManageServer) (any, error) {
				return s.Login(ctx, &LoginRequest{
					Email:    "testLogin@email.com",
					Password: "Login123",
				})
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &testhelpers.SpyAuthService{
				ResultErr: tt.serviceErr,
			}
			taskService := &testhelpers.SpyTaskService{
				ResultErr:     tt.serviceErr,
				GetTasksError: tt.serviceErr,
			}
			server := NewTaskManageServer(authService, taskService, logger)

			testUserID := 99
			ctx := context.WithValue(context.Background(), application.UserIDKey, testUserID)

			_, err := tt.call(ctx, server)
			require.Error(t, err)

			status, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.expectedCode, status.Code())
		})
	}
}
