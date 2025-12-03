// Package components 提供 Eino 组件的工厂函数
package components

import (
	"context"
	"fmt"

	qdrantindexer "github.com/cloudwego/eino-ext/components/indexer/qdrant"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	qdrantClient "github.com/qdrant/go-client/qdrant"

	"llm-cache/internal/eino/config"
)

// NewIndexer 根据配置创建并返回一个 Eino Indexer 实例。
// 支持 Qdrant, Milvus, Redis, Elasticsearch, VikingDB 等多种后端存储。
// 参数 ctx: 上下文对象。
// 参数 cfg: Indexer 配置，包含后端类型、连接信息和集合名称等。
// 参数 embedder: 用于生成向量的 Embedder 实例，在索引过程中需要使用它来向量化文档。
// 返回: 初始化后的 Indexer 实例，如果后端不支持或初始化失败则返回错误。
func NewIndexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
	switch cfg.Provider {
	case "qdrant":
		return newQdrantIndexer(ctx, cfg, embedder)
	case "milvus":
		return newMilvusIndexer(ctx, cfg, embedder)
	case "redis":
		return newRedisIndexer(ctx, cfg, embedder)
	case "es8":
		return newES8Indexer(ctx, cfg, embedder)
	case "vikingdb":
		return newVikingDBIndexer(ctx, cfg, embedder)
	default:
		return nil, fmt.Errorf("unsupported indexer provider: %s", cfg.Provider)
	}
}

// newQdrantIndexer 创建 Qdrant Indexer
func newQdrantIndexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
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

	// 解析距离类型
	distance := parseQdrantDistance(cfg.Qdrant.Distance)

	// 创建 Indexer 配置
	indexerCfg := &qdrantindexer.Config{
		Client:     client,
		Collection: cfg.Collection,
		VectorDim:  cfg.VectorSize,
		Distance:   distance,
		Embedding:  embedder,
	}

	return qdrantindexer.NewIndexer(ctx, indexerCfg)
}

// parseQdrantDistance 解析 Qdrant 距离类型
func parseQdrantDistance(dist string) qdrantClient.Distance {
	switch dist {
	case "Cosine", "cosine":
		return qdrantClient.Distance_Cosine
	case "Euclid", "euclid", "euclidean":
		return qdrantClient.Distance_Euclid
	case "Dot", "dot":
		return qdrantClient.Distance_Dot
	case "Manhattan", "manhattan":
		return qdrantClient.Distance_Manhattan
	default:
		return qdrantClient.Distance_Cosine
	}
}

// newMilvusIndexer 创建 Milvus Indexer
// 注意：需要添加 github.com/cloudwego/eino-ext/components/indexer/milvus 依赖
func newMilvusIndexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
	/*
		import (
			milvusindexer "github.com/cloudwego/eino-ext/components/indexer/milvus"
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

		return milvusindexer.NewIndexer(ctx, &milvusindexer.IndexerConfig{
			Client:      client,
			Collection:  cfg.Collection,
			Embedding:   embedder,
			VectorField: cfg.Milvus.VectorField,
		})
	*/
	return nil, fmt.Errorf("Milvus indexer is not enabled. Please add github.com/cloudwego/eino-ext/components/indexer/milvus dependency")
}

// newRedisIndexer 创建 Redis Indexer
// 注意：需要添加 github.com/cloudwego/eino-ext/components/indexer/redis 依赖
func newRedisIndexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
	/*
		import (
			redisindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
			"github.com/redis/go-redis/v9"
		)

		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		return redisindexer.NewIndexer(ctx, &redisindexer.IndexerConfig{
			Client:      rdb,
			Index:       cfg.Redis.Index,
			Prefix:      cfg.Redis.Prefix,
			VectorField: cfg.Redis.VectorField,
			Embedding:   embedder,
		})
	*/
	return nil, fmt.Errorf("Redis indexer is not enabled. Please add github.com/cloudwego/eino-ext/components/indexer/redis dependency")
}

// newES8Indexer 创建 Elasticsearch Indexer
// 注意：需要添加 github.com/cloudwego/eino-ext/components/indexer/es8 依赖
func newES8Indexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
	/*
		import (
			es8indexer "github.com/cloudwego/eino-ext/components/indexer/es8"
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

		return es8indexer.NewIndexer(ctx, &es8indexer.IndexerConfig{
			Client:      esClient,
			Index:       cfg.ES8.Index,
			VectorField: cfg.ES8.VectorField,
			Embedding:   embedder,
		})
	*/
	return nil, fmt.Errorf("Elasticsearch indexer is not enabled. Please add github.com/cloudwego/eino-ext/components/indexer/es8 dependency")
}

// newVikingDBIndexer 创建 VikingDB Indexer
// 注意：需要添加 github.com/cloudwego/eino-ext/components/indexer/vikingdb 依赖
func newVikingDBIndexer(ctx context.Context, cfg *config.IndexerConfig, embedder embedding.Embedder) (indexer.Indexer, error) {
	/*
		import vikingdbindexer "github.com/cloudwego/eino-ext/components/indexer/vikingdb"

		return vikingdbindexer.NewIndexer(ctx, &vikingdbindexer.IndexerConfig{
			Collection: collection,
			Index:      index,
			Embedding:  embedder,
		})
	*/
	return nil, fmt.Errorf("VikingDB indexer is not enabled. Please add github.com/cloudwego/eino-ext/components/indexer/vikingdb dependency")
}
