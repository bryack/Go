package grpcserver

import (
	"context"
	"fmt"
	"myproject/internal/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Driver struct {
	Addr string
}

func (d Driver) Register(email, password string) error {
	conn, err := grpc.Dial(d.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := NewTaskManagerClient(conn)
	_, err = client.Register(context.Background(), &RegisterRequest{
		Email:    email,
		Password: password,
	})
	return err
}

func (d Driver) Login(email, password string) (token string, err error) {

	return "", fmt.Errorf("not implemented")
}

func (d Driver) CreateTask(token, description string) (taskID int, err error) {

	return 0, fmt.Errorf("not implemented")
}

func (d Driver) GetTasks(token string) ([]domain.Task, error) {

	return []domain.Task{}, fmt.Errorf("not implemented")
}
