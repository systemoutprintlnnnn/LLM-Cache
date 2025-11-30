package vector

import (
	"fmt"

	"llm-cache/internal/domain/repositories"
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
)

// VectorServiceFactory 向量服务工厂
type VectorServiceFactory struct {
	logger logger.Logger
}

// NewVectorServiceFactory 创建向量服务工厂实例
func NewVectorServiceFactory(log logger.Logger) *VectorServiceFactory {
	return &VectorServiceFactory{
		logger: log,
	}
}

// CreateVectorService 创建向量服务实例
// 根据配置和依赖创建VectorService的具体实现
func (f *VectorServiceFactory) CreateVectorService(
	embeddingService services.EmbeddingService,
	vectorRepository repositories.VectorRepository,
	config *VectorServiceConfig,
) (services.VectorService, error) {

	// 验证必要依赖
	if embeddingService == nil {
		return nil, fmt.Errorf("EmbeddingService不能为空")
	}

	if vectorRepository == nil {
		return nil, fmt.Errorf("VectorRepository不能为空")
	}

	// 验证配置
	if config != nil {
		if err := f.validateVectorServiceConfig(config); err != nil {
			return nil, fmt.Errorf("向量服务配置验证失败: %w", err)
		}
	}

	// 创建向量服务实例
	vectorService := NewDefaultVectorService(
		embeddingService,
		vectorRepository,
		config,
		f.logger,
	)

	f.logger.Info("向量服务实例创建成功",
		"collection_name", getCollectionName(config),
		"default_top_k", getDefaultTopK(config),
		"similarity_threshold", getDefaultSimilarityThreshold(config),
	)

	return vectorService, nil
}

// validateVectorServiceConfig 验证向量服务配置
func (f *VectorServiceFactory) validateVectorServiceConfig(config *VectorServiceConfig) error {
	if config.DefaultCollectionName == "" {
		return fmt.Errorf("默认集合名称不能为空")
	}

	if config.DefaultTopK <= 0 {
		return fmt.Errorf("默认TopK必须大于0，当前值: %d", config.DefaultTopK)
	}

	if config.DefaultTopK > 1000 {
		return fmt.Errorf("默认TopK不能超过1000，当前值: %d", config.DefaultTopK)
	}

	if config.DefaultSimilarityThreshold < 0.0 || config.DefaultSimilarityThreshold > 1.0 {
		return fmt.Errorf("默认相似度阈值必须在[0.0, 1.0]范围内，当前值: %f", config.DefaultSimilarityThreshold)
	}

	if config.MaxBatchSize <= 0 {
		return fmt.Errorf("最大批量大小必须大于0，当前值: %d", config.MaxBatchSize)
	}

	if config.MaxBatchSize > 10000 {
		return fmt.Errorf("最大批量大小不能超过10000，当前值: %d", config.MaxBatchSize)
	}

	if config.RequestTimeout <= 0 {
		return fmt.Errorf("请求超时时间必须大于0秒，当前值: %d", config.RequestTimeout)
	}

	if config.RequestTimeout > 300 {
		return fmt.Errorf("请求超时时间不能超过300秒，当前值: %d", config.RequestTimeout)
	}

	return nil
}

// CreateVectorServiceWithDefaults 使用默认配置创建向量服务实例
func (f *VectorServiceFactory) CreateVectorServiceWithDefaults(
	embeddingService services.EmbeddingService,
	vectorRepository repositories.VectorRepository,
) (services.VectorService, error) {
	return f.CreateVectorService(
		embeddingService,
		vectorRepository,
		DefaultVectorServiceConfig(),
	)
}

// 辅助函数：获取配置值（如果配置为nil则使用默认值）

func getCollectionName(config *VectorServiceConfig) string {
	if config == nil {
		return DefaultVectorServiceConfig().DefaultCollectionName
	}
	return config.DefaultCollectionName
}

func getDefaultTopK(config *VectorServiceConfig) int {
	if config == nil {
		return DefaultVectorServiceConfig().DefaultTopK
	}
	return config.DefaultTopK
}

func getDefaultSimilarityThreshold(config *VectorServiceConfig) float64 {
	if config == nil {
		return DefaultVectorServiceConfig().DefaultSimilarityThreshold
	}
	return config.DefaultSimilarityThreshold
}

// VectorServiceBuilder 向量服务构建器（提供更灵活的构建方式）
type VectorServiceBuilder struct {
	embeddingService services.EmbeddingService
	vectorRepository repositories.VectorRepository
	config           *VectorServiceConfig
	logger           logger.Logger
}

// NewVectorServiceBuilder 创建向量服务构建器
func NewVectorServiceBuilder() *VectorServiceBuilder {
	return &VectorServiceBuilder{
		config: DefaultVectorServiceConfig(),
	}
}

// WithEmbeddingService 设置嵌入服务
func (b *VectorServiceBuilder) WithEmbeddingService(service services.EmbeddingService) *VectorServiceBuilder {
	b.embeddingService = service
	return b
}

// WithVectorRepository 设置向量仓储
func (b *VectorServiceBuilder) WithVectorRepository(repo repositories.VectorRepository) *VectorServiceBuilder {
	b.vectorRepository = repo
	return b
}

// WithConfig 设置配置
func (b *VectorServiceBuilder) WithConfig(config *VectorServiceConfig) *VectorServiceBuilder {
	b.config = config
	return b
}

// WithLogger 设置日志器
func (b *VectorServiceBuilder) WithLogger(log logger.Logger) *VectorServiceBuilder {
	b.logger = log
	return b
}

// WithCollectionName 设置集合名称
func (b *VectorServiceBuilder) WithCollectionName(name string) *VectorServiceBuilder {
	if b.config == nil {
		b.config = DefaultVectorServiceConfig()
	}
	b.config.DefaultCollectionName = name
	return b
}

// WithDefaultTopK 设置默认TopK
func (b *VectorServiceBuilder) WithDefaultTopK(topK int) *VectorServiceBuilder {
	if b.config == nil {
		b.config = DefaultVectorServiceConfig()
	}
	b.config.DefaultTopK = topK
	return b
}

// WithSimilarityThreshold 设置相似度阈值
func (b *VectorServiceBuilder) WithSimilarityThreshold(threshold float64) *VectorServiceBuilder {
	if b.config == nil {
		b.config = DefaultVectorServiceConfig()
	}
	b.config.DefaultSimilarityThreshold = threshold
	return b
}

// WithNormalization 设置是否启用向量归一化
func (b *VectorServiceBuilder) WithNormalization(enabled bool) *VectorServiceBuilder {
	if b.config == nil {
		b.config = DefaultVectorServiceConfig()
	}
	b.config.EnableNormalization = enabled
	return b
}

// Build 构建向量服务实例
func (b *VectorServiceBuilder) Build() (services.VectorService, error) {
	if b.embeddingService == nil {
		return nil, fmt.Errorf("EmbeddingService必须设置")
	}

	if b.vectorRepository == nil {
		return nil, fmt.Errorf("VectorRepository必须设置")
	}

	if b.logger == nil {
		return nil, fmt.Errorf("Logger必须设置")
	}

	factory := NewVectorServiceFactory(b.logger)
	return factory.CreateVectorService(
		b.embeddingService,
		b.vectorRepository,
		b.config,
	)
}
