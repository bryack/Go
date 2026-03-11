package grpcserver

import (
	"context"
	"fmt"
	"myproject/internal/domain"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Driver struct {
	Addr           string
	connectionOnce sync.Once
	conn           *grpc.ClientConn
	client         TaskManagerClient
}

func (d *Driver) getClient() (TaskManagerClient, error) {
	var err error
	d.connectionOnce.Do(func() {
		d.conn, err = grpc.NewClient(d.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		d.client = NewTaskManagerClient(d.conn)
	})
	return d.client, err
}

func (d *Driver) Register(email, password string) error {
	client, err := d.getClient()
	if err != nil {
		return fmt.Errorf("failed to get task manager client: %w", err)
	}
	_, err = client.Register(context.Background(), &RegisterRequest{
		Email:    email,
		Password: password,
	})
	return err
}

func (d *Driver) Login(email, password string) (token string, err error) {
	client, err := d.getClient()
	if err != nil {
		return "", fmt.Errorf("failed to get task manager client: %w", err)
	}

	reply, err := client.Login(context.Background(), &LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	return reply.Token, nil
}

func (d *Driver) CreateTask(token, description string) (taskID int, err error) {
	client, err := d.getClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get task manager client: %w", err)
	}

	reply, err := client.CreateTask(context.Background(), &CreateTaskRequest{
		Token:       token,
		Description: description,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	return int(reply.TaskId), nil
}

func (d *Driver) GetTasks(token string) ([]domain.Task, error) {
	client, err := d.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get task manager client: %w", err)
	}
	reply, err := client.GetTasks(context.Background(), &GetTasksRequest{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	tasks := make([]domain.Task, len(reply.Tasks))
	for i, task := range reply.Tasks {
		tasks[i] = domain.Task{
			ID:          int(task.Id),
			Description: task.Description,
			Done:        task.Done,
		}
	}

	return tasks, nil
}
