package services

import (
	"context"

	"llm-cache/internal/domain/models"
)

// EmbeddingService 向量嵌入服务接口
// 提供单条与批量文本向量化能力
type EmbeddingService interface {
	// GenerateEmbedding 生成单个文本向量
	GenerateEmbedding(ctx context.Context, request *models.VectorProcessingRequest) (*models.VectorProcessingResult, error)

	// GenerateBatchEmbeddings 批量生成文本向量
	GenerateBatchEmbeddings(ctx context.Context, requests []*models.VectorProcessingRequest) ([]*models.VectorProcessingResult, error)

	// Close 释放资源（可选）
	Close() error
}
