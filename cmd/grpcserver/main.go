package main

import (
	"fmt"
	"log"
	"myproject/adapters/auth"
	"myproject/adapters/grpcserver"
	"myproject/adapters/storage"
	"myproject/application"
	"myproject/cmd/server/config"
	"myproject/logger"
	"net"

	"google.golang.org/grpc"
)

func main() {
	cfg, _, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.NewLogger(&cfg.LogConfig)
	if err != nil {
		log.Fatal(err)
	}
	store, err := storage.NewDatabaseStorage(cfg.DatabaseConfig.Path, l)
	if err != nil {
		log.Fatal(err)
	}
	jwtService := auth.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.Expiration)
	authService := auth.NewService(store, jwtService, l)
	taskService := application.NewService(store)
	grpcServer := grpcserver.NewTaskManageServer(store, authService, taskService)

	authInterceptor := grpcserver.NewAuthInterceptor(jwtService, l)

	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
	)
	grpcserver.RegisterTaskManagerServer(s, grpcServer)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
