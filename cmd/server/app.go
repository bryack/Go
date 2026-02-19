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
	"myproject/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
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

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	server  *http.Server
	storage *storage.DatabaseStorage
}

func NewApp() (*App, error) {
	cfg, v, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to load config: %w", err)
	}

	// Check if --show-config flag was set
	if pflag.Lookup("show-config").Changed && pflag.Lookup("show-config").Value.String() == "true" {
		config.ShowConfig(cfg, v)
		os.Exit(0)
	}

	l, err := logger.NewLogger(&cfg.LogConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logger: %w", err)
	}

	l.Info("Logger initialized successfully",
		slog.String("level", cfg.LogConfig.Level),
		slog.String("format", cfg.LogConfig.Format),
		slog.String("output", cfg.LogConfig.Output),
		slog.String("service_name", cfg.LogConfig.ServiceName),
	)

	s, err := storage.NewDatabaseStorage(cfg.DatabaseConfig.Path, l)
	if err != nil {
		l.Error("Failed to initialize database",
			slog.String("operation", "database_init"),
			slog.String("path", cfg.DatabaseConfig.Path),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("Failed to initialize database storage: %w", err)
	}

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
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  2 * time.Second,
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

	shutdownCtx, cansel := context.WithTimeout(
		context.Background(),
		a.cfg.ServerConfig.ShutdownTimeout,
	)

	defer cansel()

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
