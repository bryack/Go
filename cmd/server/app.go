package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"myproject/adapters/storage"
	"myproject/adapters/webserver"
	"myproject/auth"
	"myproject/cmd/server/config"
	"myproject/internal/domain"
	"net/http"
	"os/signal"
	"syscall"
)

var endpointsList = []string{
	"GET /",
	"GET /health",
	"GET /tasks",
	"POST /tasks",
	"GET /tasks/{id}",
	"PUT /tasks/{id}",
	"DELETE /tasks/{id}",
	"POST /register",
	"POST /login",
}

type appStorage interface {
	domain.Storage
	storage.UserStorage
}

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	server  *http.Server
	storage appStorage
}

func NewApp(cfg *config.Config, l *slog.Logger, s appStorage) (*App, error) {
	jwtService := auth.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.Expiration)
	authService := auth.NewService(s, jwtService, l)
	authMiddleware := auth.NewAuthMiddleware(jwtService, l)

	l.Info("Database storage initialized",
		slog.String("path", cfg.DatabaseConfig.Path),
	)

	l.Info("Authentication system initialized",
		slog.Duration("expiration", cfg.JWTConfig.Expiration),
	)

	tasksServer := webserver.NewTasksServer(s, authService, authMiddleware, l)

	l.Info("HTTP Server initialized",
		slog.String("server_address", fmt.Sprintf("http://%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)),
		slog.Any("endpoints", endpointsList),
		slog.Duration("shutdown_timeout", cfg.ServerConfig.ShutdownTimeout),
	)

	address := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	server := &http.Server{
		Addr:         address,
		Handler:      tasksServer,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
		IdleTimeout:  cfg.ServerConfig.IdleTimeout,
	}

	return &App{
		cfg:     cfg,
		logger:  l,
		server:  server,
		storage: s,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info("starting server", slog.String("server_address", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("server shutdown: %w", err))
	}

	if err := a.storage.Close(); err != nil {
		errs = append(errs, fmt.Errorf("storage close: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	a.logger.Info("shutdown complete")
	return nil
}
