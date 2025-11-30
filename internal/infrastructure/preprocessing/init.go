package preprocessing

import (
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
)

// Factory 请求预处理服务工厂
type Factory struct {
	logger logger.Logger
}

// NewFactory 创建请求预处理服务工厂
func NewFactory(log logger.Logger) *Factory {
	return &Factory{
		logger: log,
	}
}

// CreateRequestPreprocessingService 创建请求预处理服务
func (f *Factory) CreateRequestPreprocessingService(config *Config) services.RequestPreprocessingService {
	if config == nil {
		config = DefaultConfig()
	}

	return NewDefaultRequestPreprocessingService(config, f.logger)
}

// ValidateConfig 验证配置有效性
func (f *Factory) ValidateConfig(config *Config) error {
	if config == nil {
		return nil // 空配置将使用默认值
	}
	return config.Validate()
}
