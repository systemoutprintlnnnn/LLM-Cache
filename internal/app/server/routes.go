package server

import (
	"llm-cache/internal/app/handlers"
	"llm-cache/internal/app/middleware"
	"llm-cache/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置并注册 HTTP 服务器的所有路由规则。
// 它负责加载中间件，定义 API 版本分组，并将 URL 路径映射到相应的处理函数。
// 参数 engine: Gin 引擎实例。
// 参数 cacheHandler: 业务逻辑处理器。
// 参数 log: 日志记录器。
func SetupRoutes(engine *gin.Engine, cacheHandler *handlers.CacheHandler, log logger.Logger) {
	// 应用全局中间件
	setupMiddleware(engine, log)

	// 设置API路由组
	v1 := engine.Group("/v1")
	// 缓存相关路由
	cache := v1.Group("/cache")

	// 查询缓存 - POST方法，支持复杂查询条件
	cache.POST("/search", cacheHandler.QueryCache)
	// 存储缓存 - 将问答对存入语义缓存
	cache.POST("/store", cacheHandler.StoreCache)
	// 根据ID获取缓存项 - 支持查询参数：user_type, include_statistics
	cache.GET("/:cache_id", cacheHandler.GetCacheByID)
	// 删除单个缓存项 - 支持查询参数：user_type, force
	cache.DELETE("/:cache_id", cacheHandler.DeleteCache)
	// 批量删除缓存 - 请求体包含要删除的ID列表
	cache.DELETE("/batch", cacheHandler.BatchDeleteCache)
	// 获取缓存统计信息 - 支持查询参数：user_type, time_range
	cache.GET("/statistics", cacheHandler.GetCacheStatistics)
	// 健康检查 - 检查缓存服务状态
	cache.GET("/health", cacheHandler.HealthCheck)

}

// setupMiddleware 设置全局中间件
func setupMiddleware(engine *gin.Engine, log logger.Logger) {
	// 设置恢复中间件 - 捕获panic并返回500错误
	engine.Use(gin.Recovery())

	// 设置日志中间件 - 记录请求日志并生成请求ID
	loggingConfig := &middleware.LoggingConfig{
		// 跳过健康检查路径的日志记录，减少日志噪音
		SkipPaths: []string{
			"/v1/cache/health",
		},
		// 暂不记录请求和响应体，避免日志过大
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		Logger:              log,
	}
	engine.Use(middleware.LoggingMiddleware(loggingConfig))
}
