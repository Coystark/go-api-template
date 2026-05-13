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
	authhttp "github.com/caiohenrique/go-api-template/internal/features/auth/adapters/http"
	authjwt "github.com/caiohenrique/go-api-template/internal/features/auth/adapters/jwt"
	authapp "github.com/caiohenrique/go-api-template/internal/features/auth/app"
	orderhttp "github.com/caiohenrique/go-api-template/internal/features/order/adapters/http"
	orderrepo "github.com/caiohenrique/go-api-template/internal/features/order/adapters/repo"
	orderapp "github.com/caiohenrique/go-api-template/internal/features/order/app"
	userhttp "github.com/caiohenrique/go-api-template/internal/features/user/adapters/http"
	userrepo "github.com/caiohenrique/go-api-template/internal/features/user/adapters/repo"
	userapp "github.com/caiohenrique/go-api-template/internal/features/user/app"
	"github.com/caiohenrique/go-api-template/internal/platform/config"
	bcryptadapter "github.com/caiohenrique/go-api-template/internal/platform/crypto/bcrypt"
	"github.com/caiohenrique/go-api-template/internal/platform/database"
	"github.com/caiohenrique/go-api-template/internal/platform/logger"
	asynqq "github.com/caiohenrique/go-api-template/internal/platform/queue/asynq"
	"github.com/caiohenrique/go-api-template/internal/platform/validator"
	"github.com/caiohenrique/go-api-template/internal/server"
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
	asynqClient := asynqq.NewClient(cfg.RedisAddr)
	defer func() {
		if err := asynqClient.Close(); err != nil {
			log.Error("close asynq client", "error", err)
		}
	}()

	userRepo := userrepo.NewRepository(db)
	mailer := asynqq.NewMailer(asynqClient)
	hasher := bcryptadapter.New()
	userSvc := userapp.NewService(userRepo, val, mailer, hasher)
	userHandler := userhttp.NewHandler(userSvc)

	orderRepo := orderrepo.NewRepository(db)
	orderSvc := orderapp.NewService(orderRepo, val)
	orderHandler := orderhttp.NewHandler(orderSvc)

	issuer := authjwt.NewIssuer(cfg.JWTSecret, jwtTTL)
	parser := authjwt.NewParser(cfg.JWTSecret)
	authSvc := authapp.NewService(userRepo, issuer, hasher, val)
	authHandler := authhttp.NewHandler(authSvc)
	authMw := authhttp.NewMiddleware(parser)

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
