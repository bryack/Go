package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"myproject/adapters/auth"
	"myproject/adapters/grpcserver"
	"myproject/application"
	"myproject/config"
	"myproject/domain"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

type App struct {
	cfg             *config.Config
	logger          *slog.Logger
	server          *grpc.Server
	storage         domain.AppStorage
	shutdownTimeout time.Duration
}

func NewApp(cfg *config.Config, l *slog.Logger, store domain.AppStorage) (*App, error) {
	jwtService := auth.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.Expiration)
	authService := application.NewAuthService(store, jwtService, l)
	taskService := application.NewService(store)
	grpcSrv := grpcserver.NewTaskManageServer(authService, taskService, l)
	authInterceptor := grpcserver.NewAuthInterceptor(jwtService, l)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
	)
	grpcserver.RegisterTaskManagerServer(server, grpcSrv)

	return &App{
		cfg:             cfg,
		logger:          l,
		server:          server,
		storage:         store,
		shutdownTimeout: cfg.ServerConfig.ShutdownTimeout,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPCConfig.Port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	go func() {
		a.logger.Info("starting gRPC server", slog.Int("port", a.cfg.GRPCConfig.Port))
		if err := a.server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
	return a.shutdown()
}

func (a *App) shutdown() error {
	a.logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		a.cfg.ServerConfig.ShutdownTimeout,
	)
	defer cancel()

	var errs []error
	done := make(chan struct{})
	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		errs = append(errs, fmt.Errorf("gRPC shutdown timed out after %v", a.cfg.ServerConfig.ShutdownTimeout))
		a.server.Stop()
	}

	if err := a.storage.Close(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("failed storage close: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	a.logger.Info("shutdown complete")
	return nil
}
