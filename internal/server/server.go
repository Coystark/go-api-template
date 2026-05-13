package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authhttp "github.com/caiohenrique/go-api-template/internal/features/auth/adapters/http"
	orderhttp "github.com/caiohenrique/go-api-template/internal/features/order/adapters/http"
	userhttp "github.com/caiohenrique/go-api-template/internal/features/user/adapters/http"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server encapsulates the Gin engine and HTTP lifecycle.
type Server struct {
	engine *gin.Engine
	log    *slog.Logger
}

// New mounts global middlewares and registers feature routes.
func New(
	log *slog.Logger,
	userHandler *userhttp.Handler,
	orderHandler *orderhttp.Handler,
	authHandler *authhttp.Handler,
	authMiddleware gin.HandlerFunc,
) *Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	root := r.Group("")
	userHandler.RegisterRoutes(root, authMiddleware)
	orderHandler.RegisterRoutes(root, authMiddleware)
	authHandler.RegisterRoutes(root)

	return &Server{engine: r, log: log}
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

// Run starts the HTTP server and shuts down gracefully when ctx is canceled.
func (s *Server) Run(ctx context.Context, addr string, shutdownTimeout time.Duration) error {
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           s.engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		if err := <-errCh; err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
