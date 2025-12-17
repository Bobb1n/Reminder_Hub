package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"reminder-hub/pkg/logger"

	"auth/internal/usecase"
)

const (
	httpStatusMethodNotAllowed = 405
	httpStatusOK               = 200
	httpStatusNotFound         = 404
	gracefulShutdownTimeout    = 30 * time.Second
)

type Server struct {
	httpServer *http.Server
	handlers   *AuthHandlers
	logger     *logger.CurrentLogger
}

func NewServer(port int, authUsecase usecase.AuthUsecase, log *logger.CurrentLogger) *Server {
	handlers := NewAuthHandlers(authUsecase, log)

	router := setupRouter(handlers)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		handlers:   handlers,
		logger:     log,
	}
}

func setupRouter(handlers *AuthHandlers) *gin.Engine {
	r := gin.Default()

	// Список известных путей для проверки 405
	knownPaths := map[string]map[string]bool{
		"/auth/register": {"POST": true},
		"/auth/login":    {"POST": true},
		"/auth/me":       {"GET": true},
		"/auth/validate": {"POST": true},
		"/auth/logout":   {"POST": true},
		"/health":        {"GET": true},
	}

	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		if allowedMethods, exists := knownPaths[path]; exists {
			if !allowedMethods[method] {
				c.JSON(httpStatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
				c.Abort()
				return
			}
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(httpStatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	auth.POST("/register", handlers.Register)      // Регистрация нового пользователя
	auth.POST("/login", handlers.Login)            // Авторизация, выдача JWT токена
	auth.GET("/me", handlers.GetCurrentUser)       // Информация о текущем пользователе
	auth.POST("/validate", handlers.ValidateToken) // Внутренний endpoint для других сервисов
	auth.POST("/logout", handlers.Logout)          // Выход из системы (демонстрирует работу с blacklist)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(httpStatusNotFound, gin.H{"error": "Route not found"})
	})

	return r
}

func (s *Server) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(ctx, "Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	s.logger.Info(ctx, "Server started", "addr", s.httpServer.Addr)

	<-quit
	s.logger.Info(ctx, "Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error(ctx, "Server forced to shutdown", "error", err)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info(ctx, "Server exited gracefully")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
