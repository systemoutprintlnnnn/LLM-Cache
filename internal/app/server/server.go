package server

import (
	"context"
	"fmt"
	"llm-cache/configs"
	"llm-cache/internal/app/handlers"
	"llm-cache/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器结构体
// 负责整个服务器的生命周期管理，包括初始化、启动、运行和优雅关闭
type Server struct {
	config       *configs.ServerConfig  // 服务器配置
	httpServer   *http.Server           // HTTP服务器实例
	engine       *gin.Engine            // Gin引擎
	cacheHandler *handlers.CacheHandler // 缓存处理器
	logger       logger.Logger          // 日志器
}

// NewServer 创建新的HTTP服务器实例
// config: 服务器配置
// cacheHandler: 缓存处理器
// log: 日志器
func NewServer(config *configs.ServerConfig, cacheHandler *handlers.CacheHandler, log logger.Logger) *Server {
	// 根据配置设置Gin模式
	if config.Host == "0.0.0.0" || config.Host == "" {
		gin.SetMode(gin.ReleaseMode) // 生产模式
	} else {
		gin.SetMode(gin.DebugMode) // 开发模式
	}

	// 创建Gin引擎
	engine := gin.New()

	return &Server{
		config:       config,
		engine:       engine,
		cacheHandler: cacheHandler,
		logger:       log,
	}
}

// Start 启动HTTP服务器
// ctx: 上下文，用于控制服务器启动过程
func (s *Server) Start(ctx context.Context) error {

	// 设置路由
	SetupRoutes(s.engine, s.cacheHandler, s.logger)

	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.logger.InfoContext(ctx, "HTTP服务器初始化完成",
		"addr", s.httpServer.Addr,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"idle_timeout", s.config.IdleTimeout)

	// 启动服务器（非阻塞）
	go func() {
		s.logger.InfoContext(ctx, "HTTP服务器开始监听", "addr", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "HTTP服务器启动失败", "error", err.Error())
		}
	}()

	return nil
}

// Shutdown 优雅关闭服务器
// ctx: 上下文，用于控制关闭过程的超时
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "开始执行HTTP服务器优雅关闭")

	// 创建带超时的关闭上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.GracefulShutdownTimeout)
	defer cancel()

	// 执行优雅关闭
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(ctx, "HTTP服务器优雅关闭失败，强制关闭", "error", err.Error())
		return fmt.Errorf("HTTP服务器关闭失败: %w", err)
	}

	s.logger.InfoContext(ctx, "HTTP服务器优雅关闭完成")
	return nil
}
