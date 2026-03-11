package main

import (
	"context"
	"fmt"
	"log"
	"myproject/adapters/grpcserver"
	"net"

	"google.golang.org/grpc"
)

func main() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	grpcserver.RegisterTaskManagerServer(s, &TaskManagerServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

type TaskManagerServer struct {
	grpcserver.UnimplementedTaskManagerServer
}

func (g TaskManagerServer) Register(ctx context.Context, request *grpcserver.RegisterRequest) (*grpcserver.RegisterReply, error) {
	return &grpcserver.RegisterReply{Token: "fixme"}, nil
}
