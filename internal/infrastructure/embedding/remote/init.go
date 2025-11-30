package remote

import (
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"llm-cache/configs"
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
)

// NewRemoteEmbeddingService 创建新的远程嵌入模型服务
// 根据配置初始化OpenAI客户端并返回实现了EmbeddingService接口的服务实例
func NewRemoteEmbeddingService(config *configs.RemoteEmbedding, log logger.Logger) (services.EmbeddingService, error) {
	if config == nil {
		return nil, fmt.Errorf("remote embedding config is required")
	}

	if log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid remote embedding config: %w", err)
	}

	// 创建OpenAI客户端选项
	opts := []option.RequestOption{
		option.WithBaseURL(config.APIEndpoint),
		option.WithRequestTimeout(config.Timeout),
		option.WithMaxRetries(config.MaxRetries),
	}

	// 添加自定义请求头
	for key, value := range config.Headers {
		opts = append(opts, option.WithHeader(key, value))
	}

	// 创建OpenAI客户端
	client := openai.NewClient(opts...)

	service := &RemoteEmbeddingService{
		client: client,
		config: config,
		logger: log,
	}

	log.Info("远程嵌入模型服务初始化成功",
		"model", config.ModelName,
		"endpoint", config.APIEndpoint)

	return service, nil
}
