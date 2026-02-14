package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"myproject/auth"
	"myproject/cmd/server/config"
	"myproject/logger"
	"myproject/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	cfg, v, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Check if --show-config flag was set
	if pflag.Lookup("show-config").Changed && pflag.Lookup("show-config").Value.String() == "true" {
		config.ShowConfig(cfg, v)
		os.Exit(0)
	}

	l, err := logger.NewLogger(&cfg.LogConfig)
	if err != nil {
		log.Fatal("Failed to create logger: ", err)
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
		log.Fatal("Failed to initialize database storage:", err)
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

	tasksServer := NewTasksServer(s, authService, authMiddleware, l)

	endpointsList := []string{
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

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-shutdownChan
		l.Info("Shutdown signal received",
			slog.String("signal", sig.String()),
		)

		go func() {
			<-shutdownChan
			l.Warn("Force shutdown signal received, exiting immediately")
			s.Close()
			os.Exit(1)
		}()

		shutdownStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerConfig.ShutdownTimeout)
		defer cancel()

		exitCode := 0
		if err := server.Shutdown(ctx); err != nil {
			exitCode = 1
			if errors.Is(err, context.DeadlineExceeded) {
				l.Warn("Graceful shutdown timed out",
					slog.Duration("shutdown_timeout", cfg.ServerConfig.ShutdownTimeout),
					slog.Duration(logger.FieldDuration, time.Since(shutdownStart)),
					slog.String(logger.FieldError, context.DeadlineExceeded.Error()),
				)
			} else {
				l.Error("Server shutdown failed",
					slog.Duration(logger.FieldDuration, time.Since(shutdownStart)),
					slog.String(logger.FieldError, err.Error()),
				)
			}
		} else {
			l.Info("Server shutdown completed successfully",
				slog.Duration(logger.FieldDuration, time.Since(shutdownStart)),
				slog.String("status", "success"),
			)
		}

		if err := s.Close(); err != nil {
			exitCode = 1
		}
		os.Exit(exitCode)
	}()

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		l.Error("Fatal server error",
			slog.String(logger.FieldError, err.Error()),
		)
		os.Exit(1)
	}
}
