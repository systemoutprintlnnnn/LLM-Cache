package services

import (
	"context"

	"llm-cache/internal/domain/models"
)

// VectorService 向量服务业务层接口
// 协调 EmbeddingService 和 VectorRepository，提供高级业务功能
type VectorService interface {
	// SearchCache 搜索语义缓存
	// 完整的语义缓存查询流程：文本向量化 + 相似度搜索
	SearchCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error)

	// StoreCache 存储查询和响应到缓存
	// 将用户查询和LLM响应存储到向量缓存中
	StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error)

	// DeleteCache 删除缓存项
	// 从向量缓存中删除指定的缓存项
	DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error)

	// SelectBestResult 选择最优结果
	// 从候选结果中选择最符合查询意图的单个结果
	//
	// 参数:
	//   ctx: 上下文
	//   results: 候选结果列表
	//   query: 查询请求
	//   strategy: 选择策略（如 "first", "highest_score", "temperature_softmax"）
	//
	// 返回:
	//   *models.VectorSearchResult: 选中的最优结果
	//   error: 错误信息
	SelectBestResult(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, strategy string) (*models.VectorSearchResult, error)
}
