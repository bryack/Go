package grpcserver

import (
	"context"
)

type TaskManageServer struct {
	UnimplementedTaskManagerServer
}

func (g TaskManageServer) Register(ctx context.Context, request *RegisterRequest) (*RegisterReply, error) {
	return &RegisterReply{Token: "fixme"}, nil
}

func (g TaskManageServer) Login(ctx context.Context, request *LoginRequest) (*LoginReply, error) {
	return &LoginReply{Token: "fixme"}, nil
}

func (g TaskManageServer) CreateTask(ctx context.Context, request *CreateTaskRequest) (*CreateTaskReply, error) {
	return &CreateTaskReply{TaskId: 1}, nil
}

func (g TaskManageServer) GetTasks(ctx context.Context, request *GetTasksRequest) (*GetTasksReply, error) {
	return &GetTasksReply{Tasks: []*GetTasksReply_Task{
		{Id: 1, Description: "task 1", Done: false},
	}}, nil
}
