package repositories

import (
	"context"

	"llm-cache/internal/domain/models"
)

// VectorRepository 向量数据库仓储接口
// 专门负责向量的存储、检索和管理
type VectorRepository interface {
	// Store 存储单个向量数据
	// 将单个向量数据存储到向量数据库
	Store(ctx context.Context, request *models.VectorStoreRequest) (*models.VectorStoreResponse, error)

	// BatchStore 存储向量数据
	// 将向量数据存储到向量数据库
	BatchStore(ctx context.Context, request *models.VectorBatchStoreRequest) (*models.VectorBatchStoreResponse, error)

	// Search 向量相似性搜索
	// 在向量数据库中搜索相似向量
	Search(ctx context.Context, request *models.VectorSearchRequest) (*models.VectorSearchResponse, error)

	// Delete 删除向量数据
	// 根据ID删除向量记录
	Delete(ctx context.Context, ids []string, userType string) (*models.CacheDeleteResult, error)

	// GetByID 根据ID获取向量
	// 获取指定ID的向量数据
	GetByID(ctx context.Context, id string) (*models.Vector, error)
}
