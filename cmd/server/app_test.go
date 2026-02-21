package main

import (
	"context"
	"fmt"
	"myproject/adapters/storage"
	"myproject/auth"
	"myproject/cmd/server/config"
	"myproject/internal/domain"
	"myproject/logger"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type slowStorage struct {
	domain.AppStorage
	delay time.Duration
}

func (s *slowStorage) LoadTasks(userID int) ([]domain.Task, error) {
	time.Sleep(s.delay)
	return s.AppStorage.LoadTasks(userID)
}

func TestApp_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping graceful shutdown test in short mode")
	}

	os.Setenv("TASKMANAGER_JWT_SECRET", "test-only-secret-min32chars-long")

	cfg, _, err := config.LoadConfig()
	require.NoError(t, err)
	cfg.ServerConfig.Port = 8888

	l, err := logger.NewLogger(&logger.Config{
		Level:       "error",
		Format:      "text",
		Output:      "stderr",
		ServiceName: "test-service",
		Environment: "test",
	})
	assert.NoError(t, err)

	dbPath := filepath.Join(t.TempDir(), "/test.db")
	db, err := storage.NewDatabaseStorage(dbPath, l)
	require.NoError(t, err)

	slowDB := &slowStorage{
		AppStorage: db,
		delay:      2 * time.Second,
	}

	app, err := NewApp(cfg, l, slowDB)
	require.NoError(t, err)

	runCtx, cancelRun := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)

	go func() {
		serverDone <- app.Run(runCtx)
	}()

	_, err = storage.Retry(func() (bool, error) {
		response, err := http.Get("http://localhost:8888/health")
		if err != nil {
			return false, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return false, fmt.Errorf("server returned status: %d", response.StatusCode)
		}
		return true, nil
	}, 10)
	require.NoError(t, err)

	jwtService := auth.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.Expiration)
	token, err := jwtService.GenerateToken(1)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8888/tasks", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	requestFinished := make(chan struct{})
	go func() {
		client := http.Client{}
		response, err := client.Do(req)
		assert.NoError(t, err)

		if err == nil {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			response.Body.Close()
		}
		close(requestFinished)
	}()

	time.Sleep(500 * time.Millisecond)

	t.Log("Triggering graceful shutdown...")
	cancelRun()

	select {
	case <-requestFinished:
		t.Log("Request finished successfully during shutdown.")
	case <-time.After(5 * time.Second):
		t.Fatal("Test failed: Server did not wait for in-flight request")
	}

	err = <-serverDone
	assert.NoError(t, err)
}
