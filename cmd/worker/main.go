package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/caiohenrique/go-api-template/internal/platform/config"
	"github.com/caiohenrique/go-api-template/internal/platform/logger"
	"github.com/caiohenrique/go-api-template/internal/platform/queue"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	srv := queue.NewServer(cfg.RedisAddr)
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskTypeWelcomeEmail, handleWelcomeEmail)

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Run(mux); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-sigCtx.Done():
		slog.Info("worker shutting down")
		srv.Shutdown()
		if err := <-errCh; err != nil && !errors.Is(err, asynq.ErrServerClosed) {
			slog.Error("asynq server", "error", err)
		}
	case err := <-errCh:
		if err != nil {
			slog.Error("asynq server", "error", err)
			os.Exit(1)
		}
	}
	slog.Info("worker shutdown complete")
}

func handleWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	var p queue.WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	slog.InfoContext(ctx, "welcome email job", "email", p.Email, "message", "would send welcome email")
	return nil
}
