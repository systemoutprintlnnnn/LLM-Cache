package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-cache/configs"
	"llm-cache/internal/app/handlers"
	"llm-cache/internal/app/server"
	"llm-cache/internal/domain/repositories"
	"llm-cache/internal/domain/services"
	"llm-cache/internal/infrastructure/cache"
	"llm-cache/internal/infrastructure/embedding/remote"
	"llm-cache/internal/infrastructure/postprocessing"
	"llm-cache/internal/infrastructure/preprocessing"
	"llm-cache/internal/infrastructure/quality"
	"llm-cache/internal/infrastructure/stores/qdrant"
	"llm-cache/internal/infrastructure/vector"
	"llm-cache/pkg/logger"
)

// main 主函数 - 应用程序入口点
func main() {
	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建早期logger（使用默认配置）
	earlyLogger := logger.Default()

	// 第一阶段：配置初始化
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
		"database_type", config.Database.Type,
		"embedding_type", config.Embedding.Type)

	// 2. 初始化日志服务
	appLogger, err := initializeLogger(config.Logging)
	if err != nil {
		return fmt.Errorf("日志服务初始化失败: %w", err)
	}

	appLogger.InfoContext(ctx, "日志服务初始化完成")

	// 3. 初始化外部依赖
	vectorRepo, embeddingService, err := initializeInfrastructure(ctx, config, appLogger)
	if err != nil {
		return fmt.Errorf("外部依赖初始化失败: %w", err)
	}
	appLogger.InfoContext(ctx, "外部依赖初始化完成")

	// 4. 初始化业务服务
	cacheService, err := initializeServices(ctx, config, vectorRepo, embeddingService, appLogger)
	if err != nil {
		return fmt.Errorf("业务服务初始化失败: %w", err)
	}
	appLogger.InfoContext(ctx, "业务服务初始化完成")

	// 5. 初始化应用层
	cacheHandler := handlers.NewCacheHandler(cacheService, appLogger)
	httpServer := server.NewServer(&config.Server, cacheHandler, appLogger)

	// 6. 启动服务并等待停止信号
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

// initializeInfrastructure 初始化基础设施层
func initializeInfrastructure(ctx context.Context, config *configs.Config, log logger.Logger) (repositories.VectorRepository, services.EmbeddingService, error) {
	// 初始化向量存储
	var vectorRepo repositories.VectorRepository
	var err error

	switch config.Database.Type {
	case "qdrant":
		factory := qdrant.NewQdrantVectorStoreFactory(log.SlogLogger())
		vectorRepo, err = factory.CreateVectorRepository(ctx, &config.Database.Qdrant)
		if err != nil {
			return nil, nil, fmt.Errorf("Qdrant向量存储初始化失败: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("不支持的数据库类型: %s", config.Database.Type)
	}

	// 初始化嵌入服务
	var embeddingService services.EmbeddingService

	switch config.Embedding.Type {
	case "remote":
		embeddingService, err = remote.NewRemoteEmbeddingService(&config.Embedding.Remote, log)
		if err != nil {
			return nil, nil, fmt.Errorf("远程嵌入服务初始化失败: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("不支持的嵌入服务类型: %s", config.Embedding.Type)
	}

	return vectorRepo, embeddingService, nil
}

// initializeServices 初始化业务服务
func initializeServices(
	ctx context.Context,
	config *configs.Config,
	vectorRepo repositories.VectorRepository,
	embeddingService services.EmbeddingService,
	log logger.Logger,
) (services.CacheService, error) {

	// 初始化VectorService
	vectorServiceFactory := vector.NewVectorServiceFactory(log)
	vectorService, err := vectorServiceFactory.CreateVectorService(
		embeddingService,
		vectorRepo,
		vector.DefaultVectorServiceConfig(),
	)
	if err != nil {
		log.ErrorContext(ctx, "向量服务初始化失败", "error", err)
		return nil, fmt.Errorf("向量服务初始化失败: %w", err)
	}
	log.InfoContext(ctx, "向量服务初始化成功")

	var preprocessingService services.RequestPreprocessingService
	var postprocessingService services.RecallPostprocessingService
	var qualityService services.QualityService

	// 初始化请求预处理服务
	if preprocessingFactory := preprocessing.NewFactory(log); preprocessingFactory != nil {
		preprocessingService = preprocessingFactory.CreateRequestPreprocessingService(nil)
		log.InfoContext(ctx, "请求预处理服务初始化成功")
	} else {
		log.WarnContext(ctx, "请求预处理服务初始化失败，将禁用预处理功能")
	}

	// 初始化召回后处理服务
	if postprocessingFactory := postprocessing.NewFactory(log); postprocessingFactory != nil {
		postprocessingService = postprocessingFactory.CreateRecallPostprocessingService(nil)
		log.InfoContext(ctx, "召回后处理服务初始化成功")
	} else {
		log.WarnContext(ctx, "召回后处理服务初始化失败，将禁用后处理功能")
	}

	// 初始化质量评估服务
	if qualityFactory := quality.NewQualityServiceFactory(log); qualityFactory != nil {
		// 使用默认质量评估配置
		qualityConfig := createDefaultQualityConfig()
		qualityService, err = qualityFactory.CreateQualityService(qualityConfig)
		if err != nil {
			log.WarnContext(ctx, "质量评估服务初始化失败", "error", err)
			qualityService = nil
		}
	}

	// 创建CacheService 核心缓存服务
	cacheFactory := cache.NewCacheServiceFactory(log)

	cacheService, err := cacheFactory.CreateCacheService(
		vectorService,
		embeddingService,
		vectorRepo,
		preprocessingService,
		postprocessingService,
		qualityService,
		cache.DefaultConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("核心缓存服务创建失败: %w", err)
	}

	log.InfoContext(ctx, "基础缓存服务创建成功")
	return cacheService, nil
}

// createDefaultQualityConfig 创建默认质量评估配置
func createDefaultQualityConfig() *services.QualityConfig {
	return &services.QualityConfig{
		Strategies:           []string{"format", "relevance"},
		StrategyWeights:      map[string]float64{"format": 0.3, "relevance": 0.7},
		UserTypeThresholds:   map[string]float64{"default": 0.7},
		DefaultThreshold:     0.7,
		MinQuestionLength:    5,
		MaxQuestionLength:    1000,
		MinAnswerLength:      10,
		MaxAnswerLength:      10000,
		BlacklistKeywords:    []string{},
		BlacklistPatterns:    []string{},
		EnableBlacklistCheck: true,
		Timeout:              30,
	}
}

// runApplication 运行应用程序，监听停止信号
func runApplication(ctx context.Context, httpServer *server.Server, log logger.Logger) error {
	// 创建错误通道
	errChan := make(chan error, 1)

	// 创建信号通道
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动HTTP服务器（非阻塞）
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			errChan <- fmt.Errorf("HTTP服务器启动失败: %w", err)
		}
	}()

	// 等待停止信号或错误
	select {
	case err := <-errChan:
		log.ErrorContext(ctx, "服务器运行错误", "error", err)
		return err

	case sig := <-signalChan:
		log.InfoContext(ctx, "收到停止信号，开始优雅关闭", "signal", sig.String())

		// 执行优雅关闭
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
