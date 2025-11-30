package services

import (
	"context"
	"llm-cache/internal/domain/models"
	"llm-cache/pkg/status"
)

// RecallPostprocessingService 召回后处理服务接口
// 负责对向量检索召回的结果进行优化处理，确保返回最相关和高质量的结果
type RecallPostprocessingService interface {
	// ProcessRecallResults 处理向量检索的召回结果
	// 对原始检索结果进行后处理，包括去重、排序、质量筛选等操作
	//
	// 参数:
	//   ctx: 上下文，用于控制请求生命周期和传递追踪信息
	//   results: 向量检索的原始结果列表
	//   originalQuery: 原始查询请求
	//
	// 返回:
	//   []*models.VectorSearchResult: 处理后的结果列表
	//   status.StatusCode: 处理状态码
	//   error: 错误信息
	ProcessRecallResults(ctx context.Context, results []*models.VectorSearchResult, originalQuery *models.CacheQuery) ([]*models.VectorSearchResult, status.StatusCode, error)

	// FormatResult 格式化结果
	// 将向量搜索结果转换为缓存查询结果格式
	//
	// 参数:
	//   ctx: 上下文
	//   result: 向量搜索结果
	//   includeMetadata: 是否包含元数据
	//
	// 返回:
	//   *models.CacheResult: 格式化后的缓存结果
	//   status.StatusCode: 处理状态码
	//   error: 错误信息
	FormatResult(ctx context.Context, result *models.VectorSearchResult, includeMetadata bool) (*models.CacheResult, status.StatusCode, error)
}

// ResultFormatter 结果格式化器接口
type ResultFormatter interface {
	// FormatCacheResult 格式化为缓存结果
	//
	// 参数:
	//   ctx: 上下文
	//   vectorResult: 向量搜索结果
	//   options: 格式化选项
	//
	// 返回:
	//   *models.CacheResult: 缓存结果
	//   error: 错误信息
	FormatCacheResult(ctx context.Context, vectorResult *models.VectorSearchResult) (*models.CacheResult, error)

	// ExtractAnswer 从结果中提取答案
	//
	// 参数:
	//   result: 向量搜索结果
	//
	// 返回:
	//   string: 提取的答案
	//   error: 错误信息
	ExtractAnswer(result *models.VectorSearchResult) (string, error)

	// ExtractMetadata 提取元数据
	//
	// 参数:
	//   result: 向量搜索结果
	//
	// 返回:
	//   *models.CacheMetadata: 提取的元数据
	//   error: 错误信息
	ExtractMetadata(result *models.VectorSearchResult) (*models.CacheMetadata, error)
}
