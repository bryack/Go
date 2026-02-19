package main

import (
	"context"
	"log"
	"log/slog"
)

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		app.logger.Error("application error", slog.String("error", err.Error()))
	}
}
