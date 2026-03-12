package grpcserver

import (
	"context"
	"myproject/application"
	"myproject/domain"
)

type TaskManageServer struct {
	UnimplementedTaskManagerServer
	store       domain.Storage
	authService *application.AuthService
	taskService *application.Service
}

func NewTaskManageServer(store domain.Storage, authService *application.AuthService, taskService *application.Service) *TaskManageServer {
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

	return &CreateTaskReply{TaskId: 1}, nil
}

func (g TaskManageServer) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksReply, error) {
	return &GetTasksReply{Tasks: []*GetTasksReply_Task{
		{Id: 1, Description: "task 1", Done: false},
	}}, nil
}
