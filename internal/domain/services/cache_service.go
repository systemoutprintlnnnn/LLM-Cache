package services

import (
	"context"

	"llm-cache/internal/domain/models"
)

// CacheService 缓存服务接口，负责缓存的核心业务逻辑
// 提供高级的缓存操作，协调各个组件完成缓存查询、存储和管理
type CacheService interface {
	// QueryCache 查询缓存
	// 执行完整的缓存查询流程：请求预处理 -> 向量搜索 -> 后处理 -> 返回结果
	// ctx: 上下文
	// query: 缓存查询请求
	// 返回: 查询结果和错误信息
	QueryCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error)

	// StoreCache 存储缓存
	// 执行完整的缓存写入流程：质量评估 -> 向量生成 -> 数据存储
	// ctx: 上下文
	// request: 缓存写入请求
	// 返回: 写入结果和错误信息
	StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error)

	// DeleteCache 删除缓存
	// ctx: 上下文
	// request: 删除请求
	// 返回: 删除结果和错误信息
	DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error)

	// GetCacheByID 根据ID获取缓存项
	// ctx: 上下文
	// cacheID: 缓存项ID
	// userType: 用户类型，用于权限验证
	// 返回: 缓存项和错误信息
	GetCacheByID(ctx context.Context, cacheID, userType string) (*models.CacheItem, error)

	// GetCacheStatistics 获取缓存系统统计信息
	// ctx: 上下文
	// 返回: 统计信息映射和错误信息
	GetCacheStatistics(ctx context.Context) (map[string]interface{}, error)

	// GetCacheHealth 获取缓存系统健康状态
	// ctx: 上下文
	// 返回: 健康状态信息映射和错误信息
	GetCacheHealth(ctx context.Context) (map[string]interface{}, error)
}
