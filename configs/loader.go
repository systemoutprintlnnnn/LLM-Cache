package configs

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	einoconfig "llm-cache/internal/eino/config"
)

// Load 加载并验证应用程序配置。
// 它按照以下优先级顺序加载配置：
// 1. 默认配置
// 2. 配置文件（config.yaml，支持多个搜索路径）
// 3. 环境变量（覆盖配置文件中的值）
//
// 参数 ctx: 上下文对象。
// 返回加载并验证后的 Config 指针，如果出错则返回 error。
func Load(ctx context.Context) (*Config, error) {
	// 加载 .env 文件（如果存在）
	// 忽略错误，因为 .env 文件是可选的
	_ = godotenv.Load()

	config := DefaultConfig()

	// 尝试加载配置文件
	configPaths := []string{
		"configs/config.yaml",
		"config.yaml",
		"/etc/llm-cache/config.yaml",
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, err
			}
			break
		}
	}

	// 从环境变量覆盖配置
	loadFromEnv(config)

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// DefaultConfig 创建并返回一个包含默认值的 Config 对象。
// 默认值覆盖了服务器、数据库、嵌入模型、日志、缓存和 Eino 框架的常用配置。
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:                    "0.0.0.0",
			Port:                    8080,
			ReadTimeout:             30 * time.Second,
			WriteTimeout:            30 * time.Second,
			IdleTimeout:             60 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
			MaxConnections:          1000,
		},
		Database: DatabaseConfig{
			Type: "qdrant",
			Qdrant: QdrantConfig{
				Host:           "localhost",
				Port:           6334,
				CollectionName: "llm_cache",
				VectorSize:     1536,
				Distance:       "cosine",
				Timeout:        30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     time.Second,
			},
		},
		Embedding: EmbeddingConfig{
			Type: "remote",
			Remote: RemoteEmbedding{
				APIEndpoint: "https://api.openai.com/v1/embeddings",
				ModelName:   "text-embedding-3-small",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				RetryDelay:  time.Second,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "stdout",
		},
		Cache: CacheConfig{
			SimilarityThreshold: 0.8,
			TopK:                5,
			TTL:                 24 * time.Hour,
			MaxCacheSize:        10000,
			EnableAsyncUpdate:   false,
		},
		Quality: QualityConfig{
			Enabled:   true,
			Threshold: 0.5,
		},
		Eino: *einoconfig.DefaultEinoConfig(),
	}
}

// loadFromEnv 从环境变量中读取配置并覆盖 Config 中的值。
// 支持 LLM_CACHE_PORT, QDRANT_HOST, OPENAI_API_KEY 等环境变量。
func loadFromEnv(config *Config) {
	// Server 配置
	if port := os.Getenv("LLM_CACHE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 && p <= 65535 {
			config.Server.Port = p
		}
	}

	// Qdrant 配置
	if host := os.Getenv("QDRANT_HOST"); host != "" {
		config.Database.Qdrant.Host = host
		config.Eino.Retriever.Qdrant.Host = host
		config.Eino.Indexer.Qdrant.Host = host
	}

	// OpenAI 配置
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.Embedding.Remote.APIKey = apiKey
		config.Eino.Embedder.APIKey = apiKey
	}

	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.Embedding.Remote.APIEndpoint = baseURL
		config.Eino.Embedder.BaseURL = baseURL
	}
}
