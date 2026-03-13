package grpcserver

import (
	"context"
	"fmt"
	"myproject/application"
	"myproject/domain"
)

type TaskManageServer struct {
	UnimplementedTaskManagerServer
	store       domain.Storage
	authService domain.AuthService
	taskService domain.TaskService
}

func NewTaskManageServer(store domain.Storage, authService domain.AuthService, taskService domain.TaskService) *TaskManageServer {
	return &TaskManageServer{
		store:       store,
		authService: authService,
		taskService: taskService,
	}
}

func (g TaskManageServer) Register(ctx context.Context, request *RegisterRequest) (*RegisterReply, error) {
	token, err := g.authService.Register(request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	return &RegisterReply{Token: token}, nil
}

func (g TaskManageServer) Login(ctx context.Context, request *LoginRequest) (*LoginReply, error) {
	token, err := g.authService.Login(request.Email, request.Password)
	if err != nil {
		return nil, err
	}
	return &LoginReply{Token: token}, nil
}

func (g TaskManageServer) CreateTask(ctx context.Context, request *CreateTaskRequest) (*CreateTaskReply, error) {
	userID, err := application.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}

	task, err := g.taskService.CreateTask(request.Description, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task for user ID %d: %w", userID, err)
	}

	return &CreateTaskReply{TaskId: int32(task.ID)}, nil
}

func (g TaskManageServer) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksReply, error) {
	userID, err := application.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}
	tasks, err := g.taskService.GetTasks(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for user ID %d: %w", userID, err)
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
