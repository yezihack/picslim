package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	port   int
	logger *zap.Logger
	srv    *http.Server
}

func NewServer(port int, logger *zap.Logger) *Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/api/v1/system/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/system/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "0.1.0-dev"})
	})

	return &Server{
		port:   port,
		logger: logger,
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("local http server starting", zap.Int("port", s.port))
	err := s.srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.logger.Info("local http server stopping")
	return s.srv.Shutdown(ctx)
}
