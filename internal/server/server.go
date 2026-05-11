package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/caiohenrique/go-api-template/internal/auth"
	"github.com/caiohenrique/go-api-template/internal/order"
	"github.com/caiohenrique/go-api-template/internal/user"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server encapsula o engine Gin e o ciclo de vida HTTP.
type Server struct {
	engine *gin.Engine
	log    *slog.Logger
}

// New monta middlewares globais e registra rotas das features.
func New(
	log *slog.Logger,
	userHandler *user.Handler,
	orderHandler *order.Handler,
	authHandler *auth.Handler,
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

// Run inicia o HTTP server e encerra com shutdown gracioso quando ctx é cancelado.
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
