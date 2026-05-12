// @title						go-api-template API
// @version						1.0
// @description					API HTTP do template. Em produção, ajuste o host conforme o deployment.
// @host						localhost:8080
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description					Digite "Bearer {token}"
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/caiohenrique/go-api-template/docs"
	"github.com/caiohenrique/go-api-template/internal/features/auth"
	"github.com/caiohenrique/go-api-template/internal/features/order"
	"github.com/caiohenrique/go-api-template/internal/platform/config"
	"github.com/caiohenrique/go-api-template/internal/platform/database"
	"github.com/caiohenrique/go-api-template/internal/platform/logger"
	"github.com/caiohenrique/go-api-template/internal/platform/queue"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/caiohenrique/go-api-template/internal/server"
	"github.com/caiohenrique/go-api-template/internal/features/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	docs.SwaggerInfo.Host = cfg.SwaggerSpecHost()
	if cfg.SwaggerHTTPS {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	log := logger.New(cfg.LogLevel)

	jwtTTL, err := cfg.JWTExpiration()
	if err != nil {
		log.Error("jwt expiration", "error", err)
		os.Exit(1)
	}
	shutdownTTL, err := cfg.ShutdownTimeout()
	if err != nil {
		log.Error("shutdown timeout", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, closeDB, err := database.Open(ctx, cfg)
	if err != nil {
		log.Error("database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := closeDB(); err != nil {
			log.Error("close database", "error", err)
		}
	}()

	val := validator.New()
	asynqClient := queue.NewClient(cfg.RedisAddr)
	defer func() {
		if err := asynqClient.Close(); err != nil {
			log.Error("close asynq client", "error", err)
		}
	}()

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, val, asynqClient)
	userHandler := user.NewHandler(userSvc)

	orderRepo := order.NewRepository(db)
	orderSvc := order.NewService(orderRepo, val)
	orderHandler := order.NewHandler(orderSvc)

	tokenMgr := auth.NewTokenManager(cfg.JWTSecret, jwtTTL)
	authSvc := auth.NewService(userRepo, tokenMgr, val)
	authHandler := auth.NewHandler(authSvc)
	authMw := auth.NewMiddleware(tokenMgr)

	srv := server.New(log, userHandler, orderHandler, authHandler, authMw)
	addr := ":" + cfg.AppPort

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("api listening", "addr", addr)
	if err := srv.Run(sigCtx, addr, shutdownTTL); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
	log.Info("api shutdown complete")
}
