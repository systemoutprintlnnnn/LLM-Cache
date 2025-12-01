// Package components 提供 Eino 组件的工厂函数
package components

import (
	"context"
	"fmt"

	qdrantretriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	qdrantClient "github.com/qdrant/go-client/qdrant"

	"llm-cache/internal/eino/config"
)

// NewRetriever 根据配置创建 Eino Retriever 实例
func NewRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	switch cfg.Provider {
	case "qdrant":
		return newQdrantRetriever(ctx, cfg, embedder)
	case "milvus":
		return newMilvusRetriever(ctx, cfg, embedder)
	case "redis":
		return newRedisRetriever(ctx, cfg, embedder)
	case "es8":
		return newES8Retriever(ctx, cfg, embedder)
	case "vikingdb":
		return newVikingDBRetriever(ctx, cfg, embedder)
	default:
		return nil, fmt.Errorf("unsupported retriever provider: %s", cfg.Provider)
	}
}

// newQdrantRetriever 创建 Qdrant Retriever
func newQdrantRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	// 创建 Qdrant 客户端
	clientCfg := &qdrantClient.Config{
		Host: cfg.Qdrant.Host,
		Port: cfg.Qdrant.Port,
	}

	if cfg.Qdrant.APIKey != "" {
		clientCfg.APIKey = cfg.Qdrant.APIKey
	}

	if cfg.Qdrant.UseTLS {
		clientCfg.UseTLS = true
	}

	client, err := qdrantClient.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	// 创建 Retriever 配置
	retrieverCfg := &qdrantretriever.Config{
		Client:     client,
		Collection: cfg.Collection,
		Embedding:  embedder,
		TopK:       cfg.TopK,
	}

	// 设置相似度阈值
	if cfg.ScoreThreshold > 0 {
		threshold := cfg.ScoreThreshold
		retrieverCfg.ScoreThreshold = &threshold
	}

	return qdrantretriever.NewRetriever(ctx, retrieverCfg)
}

// newMilvusRetriever 创建 Milvus Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/milvus 依赖
func newMilvusRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	/*
		import (
			milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
			milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
		)

		client, err := milvusClient.NewClient(ctx, milvusClient.Config{
			Address:  fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port),
			Username: cfg.Milvus.Username,
			Password: cfg.Milvus.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create milvus client: %w", err)
		}

		return milvusretriever.NewRetriever(ctx, &milvusretriever.RetrieverConfig{
			Client:       client,
			Collection:   cfg.Collection,
			VectorField:  cfg.Milvus.VectorField,
			OutputFields: cfg.Milvus.OutputFields,
			MetricType:   cfg.Milvus.MetricType,
			TopK:         cfg.TopK,
			Embedding:    embedder,
		})
	*/
	return nil, fmt.Errorf("Milvus retriever is not enabled. Please add github.com/cloudwego/eino-ext/components/retriever/milvus dependency")
}

// newRedisRetriever 创建 Redis Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/redis 依赖
func newRedisRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	/*
		import (
			redisretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
			"github.com/redis/go-redis/v9"
		)

		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		return redisretriever.NewRetriever(ctx, &redisretriever.RetrieverConfig{
			Client:            rdb,
			Index:             cfg.Redis.Index,
			VectorField:       cfg.Redis.VectorField,
			TopK:              cfg.TopK,
			DistanceThreshold: &cfg.ScoreThreshold,
			Embedding:         embedder,
			ReturnFields:      cfg.Redis.ReturnFields,
		})
	*/
	return nil, fmt.Errorf("Redis retriever is not enabled. Please add github.com/cloudwego/eino-ext/components/retriever/redis dependency")
}

// newES8Retriever 创建 Elasticsearch Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/es8 依赖
func newES8Retriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	/*
		import (
			es8retriever "github.com/cloudwego/eino-ext/components/retriever/es8"
			"github.com/elastic/go-elasticsearch/v8"
		)

		esClient, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: cfg.ES8.Addresses,
			Username:  cfg.ES8.Username,
			Password:  cfg.ES8.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
		}

		return es8retriever.NewRetriever(ctx, &es8retriever.RetrieverConfig{
			Client:         esClient,
			Index:          cfg.ES8.Index,
			TopK:           cfg.TopK,
			ScoreThreshold: &cfg.ScoreThreshold,
			Embedding:      embedder,
			SearchMode:     cfg.ES8.SearchMode,
			VectorField:    cfg.ES8.VectorField,
		})
	*/
	return nil, fmt.Errorf("Elasticsearch retriever is not enabled. Please add github.com/cloudwego/eino-ext/components/retriever/es8 dependency")
}

// newVikingDBRetriever 创建 VikingDB Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/vikingdb 依赖
func newVikingDBRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	/*
		import vikingdbretriever "github.com/cloudwego/eino-ext/components/retriever/vikingdb"

		return vikingdbretriever.NewRetriever(ctx, &vikingdbretriever.RetrieverConfig{
			Collection: collection,
			Index:      index,
			TopK:       cfg.TopK,
			Embedding:  embedder,
			ScoreThreshold: &cfg.ScoreThreshold,
		})
	*/
	return nil, fmt.Errorf("VikingDB retriever is not enabled. Please add github.com/cloudwego/eino-ext/components/retriever/vikingdb dependency")
}

// GetQdrantClient 获取 Qdrant 客户端（用于删除等操作）
func GetQdrantClient(cfg *config.RetrieverConfig) (*qdrantClient.Client, error) {
	clientCfg := &qdrantClient.Config{
		Host: cfg.Qdrant.Host,
		Port: cfg.Qdrant.Port,
	}

	if cfg.Qdrant.APIKey != "" {
		clientCfg.APIKey = cfg.Qdrant.APIKey
	}

	if cfg.Qdrant.UseTLS {
		clientCfg.UseTLS = true
	}

	return qdrantClient.NewClient(clientCfg)
}
