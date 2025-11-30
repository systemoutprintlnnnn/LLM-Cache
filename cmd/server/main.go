package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/eino/compose"

	"llm-cache/configs"
	"llm-cache/internal/app/handlers"
	"llm-cache/internal/app/server"
	"llm-cache/internal/eino/components"
	einoconfig "llm-cache/internal/eino/config"
	"llm-cache/internal/eino/flows"
	"llm-cache/pkg/logger"
)

// main 主函数 - 应用程序入口点
func main() {
	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建早期logger（使用默认配置）
	earlyLogger := logger.Default()

	// 初始化应用程序
	if err := initializeApplication(ctx, earlyLogger); err != nil {
		earlyLogger.ErrorContext(ctx, "应用程序初始化失败", "error", err)
		os.Exit(1)
	}
}

// initializeApplication 初始化应用程序
func initializeApplication(ctx context.Context, earlyLogger logger.Logger) error {

	// 1. 加载配置
	config, err := configs.Load(ctx)
	if err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	earlyLogger.InfoContext(ctx, "配置加载成功",
		"server_port", config.Server.Port,
		"eino_embedder_provider", config.Eino.Embedder.Provider,
		"eino_retriever_provider", config.Eino.Retriever.Provider)

	// 2. 初始化日志服务
	appLogger, err := initializeLogger(config.Logging)
	if err != nil {
		return fmt.Errorf("日志服务初始化失败: %w", err)
	}

	appLogger.InfoContext(ctx, "日志服务初始化完成")

	// 3. 初始化 Eino 组件
	queryRunner, storeRunner, deleteService, err := initializeEinoComponents(ctx, &config.Eino, appLogger)
	if err != nil {
		return fmt.Errorf("Eino 组件初始化失败: %w", err)
	}
	appLogger.InfoContext(ctx, "Eino 组件初始化完成")

	// 4. 初始化应用层
	cacheHandler := handlers.NewCacheHandler(queryRunner, storeRunner, deleteService, appLogger)
	httpServer := server.NewServer(&config.Server, cacheHandler, appLogger)

	// 5. 启动服务并等待停止信号
	return runApplication(ctx, httpServer, appLogger)
}

// initializeLogger 初始化日志服务
func initializeLogger(config configs.LoggingConfig) (logger.Logger, error) {
	// 解析日志级别
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 创建日志配置
	loggerConfig := logger.Config{
		Level:  level,
		Output: config.Output,
	}

	if config.Output == "file" {
		loggerConfig.FilePath = config.FilePath
	}

	return logger.New(loggerConfig), nil
}

// initializeEinoComponents 初始化 Eino 组件
func initializeEinoComponents(
	ctx context.Context,
	einoCfg *einoconfig.EinoConfig,
	log logger.Logger,
) (
	compose.Runnable[*flows.CacheQueryInput, *flows.CacheQueryOutput],
	compose.Runnable[*flows.CacheStoreInput, *flows.CacheStoreOutput],
	*flows.CacheDeleteService,
	error,
) {
	// 1. 创建 Embedder
	log.InfoContext(ctx, "正在初始化 Embedder",
		"provider", einoCfg.Embedder.Provider,
		"model", einoCfg.Embedder.Model)

	embedder, err := components.NewEmbedder(ctx, &einoCfg.Embedder)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Embedder 初始化失败: %w", err)
	}
	log.InfoContext(ctx, "Embedder 初始化成功")

	// 2. 创建 Retriever
	log.InfoContext(ctx, "正在初始化 Retriever",
		"provider", einoCfg.Retriever.Provider,
		"collection", einoCfg.Retriever.Collection)

	retriever, err := components.NewRetriever(ctx, &einoCfg.Retriever, embedder)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Retriever 初始化失败: %w", err)
	}
	log.InfoContext(ctx, "Retriever 初始化成功")

	// 3. 创建 Indexer
	log.InfoContext(ctx, "正在初始化 Indexer",
		"provider", einoCfg.Indexer.Provider,
		"collection", einoCfg.Indexer.Collection)

	indexer, err := components.NewIndexer(ctx, &einoCfg.Indexer, embedder)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Indexer 初始化失败: %w", err)
	}
	log.InfoContext(ctx, "Indexer 初始化成功")

	// 4. 创建 Query Graph 并编译
	queryGraph := flows.NewCacheQueryGraph(embedder, retriever, &einoCfg.Query)
	queryRunner, err := queryGraph.Compile(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Query Graph 编译失败: %w", err)
	}
	log.InfoContext(ctx, "Query Graph 编译成功")

	// 5. 创建 Store Graph 并编译
	storeGraph := flows.NewCacheStoreGraph(embedder, indexer, &einoCfg.Store, &einoCfg.Quality)
	storeRunner, err := storeGraph.Compile(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Store Graph 编译失败: %w", err)
	}
	log.InfoContext(ctx, "Store Graph 编译成功")

	// 6. 创建 Delete Service
	deleteService, err := flows.NewCacheDeleteService(&einoCfg.Retriever)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Delete Service 初始化失败: %w", err)
	}
	log.InfoContext(ctx, "Delete Service 创建成功")

	return queryRunner, storeRunner, deleteService, nil
}

// runApplication 运行应用程序，监听停止信号
// 此函数会阻塞直到收到停止信号、服务器错误或上下文取消
func runApplication(ctx context.Context, httpServer *server.Server, log logger.Logger) error {
	// 创建错误通道 - 用于接收服务器运行时错误
	errChan := make(chan error, 1)

	// 创建信号通道
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动HTTP服务器（非阻塞）
	// 服务器的运行时错误会通过 errChan 传递
	httpServer.Start(ctx, errChan)

	// 等待停止信号、服务器错误或上下文取消
	select {
	case err := <-errChan:
		log.ErrorContext(ctx, "服务器运行错误", "error", err)
		return err

	case sig := <-signalChan:
		log.InfoContext(ctx, "收到停止信号，开始优雅关闭", "signal", sig.String())
		return gracefulShutdown(ctx, httpServer, log)

	case <-ctx.Done():
		log.InfoContext(ctx, "上下文取消，开始优雅关闭")
		return gracefulShutdown(ctx, httpServer, log)
	}
}

// gracefulShutdown 执行优雅关闭
func gracefulShutdown(ctx context.Context, httpServer *server.Server, log logger.Logger) error {
	log.InfoContext(ctx, "开始执行优雅关闭流程")

	// 创建带超时的关闭上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行HTTP服务器优雅关闭
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.ErrorContext(ctx, "HTTP服务器关闭失败", "error", err)
		return fmt.Errorf("HTTP服务器关闭失败: %w", err)
	}

	log.InfoContext(ctx, "优雅关闭完成")
	return nil
}
