package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"myproject/application"
	"myproject/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskManageServer struct {
	UnimplementedTaskManagerServer
	authService domain.AuthService
	taskService domain.TaskService
	logger      *slog.Logger
}

func NewTaskManageServer(authService domain.AuthService, taskService domain.TaskService, logger *slog.Logger) *TaskManageServer {
	return &TaskManageServer{
		authService: authService,
		taskService: taskService,
		logger:      logger,
	}
}

func (g TaskManageServer) Register(ctx context.Context, request *RegisterRequest) (*RegisterReply, error) {
	token, err := g.authService.Register(ctx, request.Email, request.Password)
	if err != nil {
		return nil, mapError(err, g.logger)
	}
	return &RegisterReply{Token: token}, nil
}

func (g TaskManageServer) Login(ctx context.Context, request *LoginRequest) (*LoginReply, error) {
	token, err := g.authService.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, mapError(err, g.logger)
	}
	return &LoginReply{Token: token}, nil
}

func (g TaskManageServer) CreateTask(ctx context.Context, request *CreateTaskRequest) (*CreateTaskReply, error) {
	userID, err := application.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID from context: %v", err)
	}

	task, err := g.taskService.CreateTask(ctx, request.Description, userID)
	if err != nil {
		return nil, mapError(err, g.logger)
	}

	return &CreateTaskReply{TaskId: int32(task.ID)}, nil
}

func (g TaskManageServer) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksReply, error) {
	userID, err := application.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID from context: %v", err)
	}
	tasks, err := g.taskService.GetTasks(ctx, userID)
	if err != nil {
		return nil, mapError(err, g.logger)
	}

	reply := make([]*GetTasksReply_Task, len(tasks))
	for i, task := range tasks {
		reply[i] = &GetTasksReply_Task{
			Id:          int32(task.ID),
			Description: task.Description,
			Done:        task.Done,
		}
	}

	return &GetTasksReply{Tasks: reply}, nil
}

func mapError(err error, logger *slog.Logger) error {
	if err == nil {
		return nil
	}

	if logger != nil {
		logger.Error("Domain error", slog.String("error", err.Error()))
	}

	switch {
	case errors.Is(err, domain.ErrDescriptionRequired),
		errors.Is(err, domain.ErrDescriptionTooLong),
		errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrStorageFailure):
		return status.Error(codes.Internal, "internal server error")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
