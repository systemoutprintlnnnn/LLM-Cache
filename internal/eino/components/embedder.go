// Package components 提供 Eino 组件的工厂函数
package components

import (
	"context"
	"fmt"
	"time"

	openaiembed "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"

	"llm-cache/internal/eino/config"
)

// NewEmbedder 根据配置创建并返回一个 Eino Embedder 实例。
// 支持 OpenAI, ARK, Ollama, Dashscope, Qianfan, Tencentcloud 等多种提供商。
// 参数 ctx: 上下文对象。
// 参数 cfg: Embedder 配置，包含提供商类型、API 密钥、模型名称等。
// 返回: 初始化后的 Embedder 实例，如果提供商不支持或初始化失败则返回错误。
func NewEmbedder(ctx context.Context, cfg *config.EmbedderConfig) (embedding.Embedder, error) {
	timeout := time.Duration(cfg.Timeout) * time.Second

	switch cfg.Provider {
	case "openai":
		return newOpenAIEmbedder(ctx, cfg, timeout)
	case "ark":
		return newARKEmbedder(ctx, cfg, timeout)
	case "ollama":
		return newOllamaEmbedder(ctx, cfg, timeout)
	case "dashscope":
		return newDashscopeEmbedder(ctx, cfg, timeout)
	case "qianfan":
		return newQianfanEmbedder(ctx, cfg)
	case "tencentcloud":
		return newTencentcloudEmbedder(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.Provider)
	}
}

// newOpenAIEmbedder 创建 OpenAI Embedder
func newOpenAIEmbedder(ctx context.Context, cfg *config.EmbedderConfig, timeout time.Duration) (embedding.Embedder, error) {
	embedCfg := &openaiembed.EmbeddingConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		Timeout: timeout,
	}

	// 设置 BaseURL（如果提供）
	if cfg.BaseURL != "" {
		embedCfg.BaseURL = cfg.BaseURL
	}

	// Azure OpenAI 配置
	if cfg.ByAzure {
		embedCfg.ByAzure = true
		embedCfg.APIVersion = cfg.APIVersion
	}

	// 设置维度（如果提供）
	if cfg.Dimensions != nil {
		embedCfg.Dimensions = cfg.Dimensions
	}

	return openaiembed.NewEmbedder(ctx, embedCfg)
}

// newARKEmbedder 创建 ARK (火山引擎) Embedder
// 注意：需要添加 github.com/cloudwego/eino-ext/components/embedding/ark 依赖
func newARKEmbedder(ctx context.Context, cfg *config.EmbedderConfig, timeout time.Duration) (embedding.Embedder, error) {
	// ARK Embedder 需要额外导入，这里返回错误提示
	// 如果需要使用 ARK，请添加对应依赖并取消注释以下代码
	/*
		import arkembed "github.com/cloudwego/eino-ext/components/embedding/ark"

		return arkembed.NewEmbedder(ctx, &arkembed.EmbeddingConfig{
			APIKey:     cfg.APIKey,
			AccessKey:  cfg.AccessKey,
			SecretKey:  cfg.SecretKey,
			Model:      cfg.Model,
			BaseURL:    cfg.BaseURL,
			Region:     cfg.Region,
			Timeout:    &timeout,
			RetryTimes: cfg.RetryTimes,
		})
	*/
	return nil, fmt.Errorf("ARK embedding provider is not enabled. Please add github.com/cloudwego/eino-ext/components/embedding/ark dependency")
}

// newOllamaEmbedder 创建 Ollama Embedder
// 注意：需要添加 github.com/cloudwego/eino-ext/components/embedding/ollama 依赖
func newOllamaEmbedder(ctx context.Context, cfg *config.EmbedderConfig, timeout time.Duration) (embedding.Embedder, error) {
	// Ollama Embedder 需要额外导入
	/*
		import ollamaembed "github.com/cloudwego/eino-ext/components/embedding/ollama"

		return ollamaembed.NewEmbedder(ctx, &ollamaembed.EmbeddingConfig{
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
			Timeout: timeout,
		})
	*/
	return nil, fmt.Errorf("Ollama embedding provider is not enabled. Please add github.com/cloudwego/eino-ext/components/embedding/ollama dependency")
}

// newDashscopeEmbedder 创建 Dashscope (阿里云) Embedder
// 注意：需要添加 github.com/cloudwego/eino-ext/components/embedding/dashscope 依赖
func newDashscopeEmbedder(ctx context.Context, cfg *config.EmbedderConfig, timeout time.Duration) (embedding.Embedder, error) {
	// Dashscope Embedder 需要额外导入
	/*
		import dashscopeembed "github.com/cloudwego/eino-ext/components/embedding/dashscope"

		return dashscopeembed.NewEmbedder(ctx, &dashscopeembed.EmbeddingConfig{
			APIKey:     cfg.APIKey,
			Model:      cfg.Model,
			Timeout:    timeout,
			Dimensions: cfg.Dimensions,
		})
	*/
	return nil, fmt.Errorf("Dashscope embedding provider is not enabled. Please add github.com/cloudwego/eino-ext/components/embedding/dashscope dependency")
}

// newQianfanEmbedder 创建 Qianfan (百度千帆) Embedder
// 注意：需要添加 github.com/cloudwego/eino-ext/components/embedding/qianfan 依赖
func newQianfanEmbedder(ctx context.Context, cfg *config.EmbedderConfig) (embedding.Embedder, error) {
	// Qianfan Embedder 需要额外导入
	/*
		import qianfanembed "github.com/cloudwego/eino-ext/components/embedding/qianfan"

		// Qianfan 使用单例配置
		qcfg := qianfanembed.GetQianfanSingletonConfig()
		qcfg.AccessKey = cfg.AccessKey
		qcfg.SecretKey = cfg.SecretKey

		return qianfanembed.NewEmbedder(ctx, &qianfanembed.EmbeddingConfig{
			Model: cfg.Model,
		})
	*/
	return nil, fmt.Errorf("Qianfan embedding provider is not enabled. Please add github.com/cloudwego/eino-ext/components/embedding/qianfan dependency")
}

// newTencentcloudEmbedder 创建 Tencentcloud (腾讯云) Embedder
// 注意：需要添加 github.com/cloudwego/eino-ext/components/embedding/tencentcloud 依赖
func newTencentcloudEmbedder(ctx context.Context, cfg *config.EmbedderConfig) (embedding.Embedder, error) {
	// Tencentcloud Embedder 需要额外导入
	/*
		import tencentcloudembed "github.com/cloudwego/eino-ext/components/embedding/tencentcloud"

		return tencentcloudembed.NewEmbedder(ctx, &tencentcloudembed.EmbeddingConfig{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
			Region:    cfg.Region,
		})
	*/
	return nil, fmt.Errorf("Tencentcloud embedding provider is not enabled. Please add github.com/cloudwego/eino-ext/components/embedding/tencentcloud dependency")
}
