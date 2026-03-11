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

	return "", fmt.Errorf("not implemented")
}

func (d *Driver) CreateTask(token, description string) (taskID int, err error) {

	return 0, fmt.Errorf("not implemented")
}

func (d *Driver) GetTasks(token string) ([]domain.Task, error) {

	return []domain.Task{}, fmt.Errorf("not implemented")
}
