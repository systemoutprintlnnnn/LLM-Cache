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

// Server 定义了 HTTP 服务器的核心结构。
// 它管理服务器的生命周期，包括配置、Gin 引擎、路由处理、缓存处理器以及优雅关闭等。
type Server struct {
	config       *configs.ServerConfig  // 服务器配置
	httpServer   *http.Server           // HTTP服务器实例
	engine       *gin.Engine            // Gin引擎
	cacheHandler *handlers.CacheHandler // 缓存处理器
	logger       logger.Logger          // 日志器
}

// NewServer 创建并初始化一个新的 HTTP 服务器实例。
// 根据配置设置 Gin 模式（Debug/Release），并注入缓存处理器和日志组件。
// 参数 config: 服务器配置。
// 参数 cacheHandler: 核心业务处理器。
// 参数 log: 日志记录器。
// 返回: 初始化后的 Server 指针。
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

// Start 启动 HTTP 服务器并开始监听请求（异步执行）。
// 它会在后台 goroutine 中运行 ListenAndServe。
// 参数 ctx: 上下文对象。
// 参数 errChan: 用于接收服务器运行时错误的通道。调用者应监听此通道。
func (s *Server) Start(ctx context.Context, errChan chan<- error) {
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
			s.logger.ErrorContext(ctx, "HTTP服务器运行错误", "error", err.Error())
			// 将错误传递给调用者
			if errChan != nil {
				errChan <- fmt.Errorf("HTTP服务器运行错误: %w", err)
			}
		}
	}()
}

// Shutdown 执行服务器的优雅关闭流程。
// 它会等待当前正在处理的请求完成，或者直到上下文超时。
// 参数 ctx: 控制关闭流程超时的上下文。
// 返回: 关闭过程中的错误（如果有）。
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
