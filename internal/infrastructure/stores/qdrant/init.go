package qdrant

import (
	"context"
	"fmt"
	"log/slog"

	"llm-cache/configs"
	"llm-cache/internal/domain/repositories"
)

// QdrantVectorStoreFactory Qdrant向量存储工厂
type QdrantVectorStoreFactory struct {
	logger *slog.Logger
}

// NewQdrantVectorStoreFactory 创建Qdrant向量存储工厂
func NewQdrantVectorStoreFactory(logger *slog.Logger) *QdrantVectorStoreFactory {
	if logger == nil {
		logger = slog.Default()
	}

	return &QdrantVectorStoreFactory{
		logger: logger,
	}
}

// CreateVectorRepository 创建Qdrant向量仓储实例
// 根据配置创建Qdrant向量存储实现
func (f *QdrantVectorStoreFactory) CreateVectorRepository(ctx context.Context, config *configs.QdrantConfig) (repositories.VectorRepository, error) {
	if config == nil {
		return nil, fmt.Errorf("qdrant config cannot be nil")
	}

	store, err := NewQdrantVectorStore(ctx, config, f.logger)
	if err != nil {
		f.logger.ErrorContext(ctx, "Qdrant向量存储初始化失败", "error", err)
		return nil, fmt.Errorf("failed to create qdrant vector store: %w", err)
	}

	return store, nil
}

// ValidateQdrantConfiguration 验证Qdrant配置
func ValidateQdrantConfiguration(config *configs.QdrantConfig) error {
	if config == nil {
		return fmt.Errorf("qdrant config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("qdrant config validation failed: %w", err)
	}

	return nil
}
