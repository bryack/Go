package main

import (
	"context"
	"log"
	"log/slog"
	"myproject/adapters/storage"
	"myproject/config"
	"myproject/logger"
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
	l.Info("Logger initialized successfully",
		slog.String("level", cfg.LogConfig.Level),
		slog.String("format", cfg.LogConfig.Format),
		slog.String("output", cfg.LogConfig.Output),
		slog.String("service_name", cfg.LogConfig.ServiceName),
	)

	store, err := storage.NewDatabaseStorage(cfg.DatabaseConfig.Path, l)
	if err != nil {
		l.Error("Failed to initialize database",
			slog.String("operation", "database_init"),
			slog.String("path", cfg.DatabaseConfig.Path),
			slog.String("error", err.Error()),
		)
		log.Fatal(err)
	}

	app, err := NewApp(cfg, l, store)
	if err != nil {
		log.Fatal(err)
	}
	if err = app.Run(context.Background()); err != nil {
		l.Error("application error", slog.String("error", err.Error()))
		log.Fatal(err)
	}
}
