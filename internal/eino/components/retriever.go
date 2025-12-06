// Package components 提供 Eino 组件的工厂函数
package components

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	es8retriever "github.com/cloudwego/eino-ext/components/retriever/es8"
	es8searchmode "github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	qdrantretriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	redisretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	qdrantClient "github.com/qdrant/go-client/qdrant"
	redis "github.com/redis/go-redis/v9"

	"llm-cache/internal/eino/config"
)

// NewRetriever 根据配置创建并返回一个 Eino Retriever 实例。
// 支持 Qdrant, Milvus, Redis, Elasticsearch, VikingDB 等多种后端。
// 参数 ctx: 上下文对象。
// 参数 cfg: Retriever 配置，包含后端类型、连接信息、TopK 和阈值等。
// 参数 embedder: 用于生成查询向量的 Embedder 实例，在检索过程中需要使用它来向量化查询文本。
// 返回: 初始化后的 Retriever 实例，如果后端不支持或初始化失败则返回错误。
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
	client, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port),
		Username: cfg.Milvus.Username,
		Password: cfg.Milvus.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	milvusCfg := &milvusretriever.RetrieverConfig{
		Client:         client,
		Collection:     cfg.Collection,
		VectorField:    cfg.Milvus.VectorField,
		OutputFields:   cfg.Milvus.OutputFields,
		TopK:           cfg.TopK,
		Embedding:      embedder,
		Partition:      cfg.Milvus.Partition,
		ScoreThreshold: cfg.ScoreThreshold,
	}

	if cfg.Milvus.MetricType != "" {
		milvusCfg.MetricType = parseMilvusMetricType(cfg.Milvus.MetricType)
	}

	return milvusretriever.NewRetriever(ctx, milvusCfg)
}

// newRedisRetriever 创建 Redis Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/redis 依赖
func newRedisRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Protocol: 2,
	})

	vectorField := cfg.Redis.VectorField
	if vectorField == "" {
		vectorField = "vector_content"
	}

	returnFields := cfg.Redis.ReturnFields
	if len(returnFields) == 0 {
		returnFields = []string{"content", vectorField, "metadata"}
	} else {
		// ensure metadata field is always returned
		seen := make(map[string]bool)
		for _, f := range returnFields {
			seen[f] = true
		}
		if !seen["metadata"] {
			returnFields = append(returnFields, "metadata")
		}
	}

	redisCfg := &redisretriever.RetrieverConfig{
		Client:            rdb,
		Index:             cfg.Redis.Index,
		VectorField:       vectorField,
		TopK:              cfg.TopK,
		Embedding:         embedder,
		ReturnFields:      returnFields,
		DistanceThreshold: distanceThresholdFromScore(cfg.ScoreThreshold),
		DocumentConverter: buildRedisDocumentConverter(vectorField),
	}

	return redisretriever.NewRetriever(ctx, redisCfg)
}

// newES8Retriever 创建 Elasticsearch Retriever
// 注意：需要添加 github.com/cloudwego/eino-ext/components/retriever/es8 依赖
func newES8Retriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.ES8.Addresses,
		Username:  cfg.ES8.Username,
		Password:  cfg.ES8.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	vectorField := cfg.ES8.VectorField
	if vectorField == "" {
		vectorField = "vector"
	}

	searchMode := buildESSearchMode(cfg, vectorField)
	resultParser := buildESResultParser()

	esCfg := &es8retriever.RetrieverConfig{
		Client:         esClient,
		Index:          cfg.ES8.Index,
		TopK:           cfg.TopK,
		ScoreThreshold: thresholdPtr(cfg.ScoreThreshold),
		Embedding:      embedder,
		SearchMode:     searchMode,
		ResultParser:   resultParser,
	}

	return es8retriever.NewRetriever(ctx, esCfg)
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

// GetQdrantClient 创建并返回一个 Qdrant 客户端实例。
// 用于直接执行 Qdrant 操作（如删除缓存项），绕过 Eino Retriever 接口。
// 参数 cfg: Retriever 配置（包含 Qdrant 连接信息）。
// 返回: Qdrant 客户端或错误。
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

// parseMilvusMetricType maps a config string to milvus entity metric type.
func parseMilvusMetricType(metric string) entity.MetricType {
	switch strings.ToLower(metric) {
	case "ip", "inner", "inner_product":
		return entity.IP
	case "l2", "euclid", "euclidean":
		return entity.L2
	case "hamming":
		return entity.HAMMING
	case "cosine":
		return entity.COSINE
	default:
		return entity.MetricType("")
	}
}

func distanceThresholdFromScore(score float64) *float64 {
	if score <= 0 {
		return nil
	}
	return &score
}

// buildRedisDocumentConverter parses metadata json back into MetaData.
func buildRedisDocumentConverter(vectorField string) func(ctx context.Context, doc redis.Document) (*schema.Document, error) {
	return func(ctx context.Context, doc redis.Document) (*schema.Document, error) {
		resp := &schema.Document{
			ID:       doc.ID,
			MetaData: map[string]any{},
		}

		if val, ok := doc.Fields["content"]; ok {
			resp.Content = val
		}

		if v, ok := doc.Fields[vectorField]; ok {
			resp.WithDenseVector(redisretriever.Bytes2Vector([]byte(v)))
		}

		if metaStr, ok := doc.Fields["metadata"]; ok && metaStr != "" {
			var metadata map[string]any
			if err := json.Unmarshal([]byte(metaStr), &metadata); err == nil {
				for k, v := range metadata {
					resp.MetaData[k] = v
				}
			}
		}

		// capture any additional return fields
		for k, v := range doc.Fields {
			if k != "content" && k != vectorField && k != "metadata" {
				resp.MetaData[k] = v
			}
		}

		return resp, nil
	}
}

// buildESSearchMode builds an ES8 search mode according to config.
func buildESSearchMode(cfg *config.RetrieverConfig, vectorField string) es8retriever.SearchMode {
	switch strings.ToLower(cfg.ES8.SearchMode) {
	case "hybrid":
		queryField := cfg.ES8.TextField
		if queryField == "" {
			queryField = "content"
		}
		return es8searchmode.SearchModeApproximate(&es8searchmode.ApproximateConfig{
			QueryFieldName:  queryField,
			VectorFieldName: vectorField,
			Hybrid:          true,
		})
	default:
		return es8searchmode.SearchModeDenseVectorSimilarity(
			es8searchmode.DenseVectorSimilarityTypeCosineSimilarity, vectorField)
	}
}

// buildESResultParser converts ES hits back to schema.Document with metadata.
func buildESResultParser() func(ctx context.Context, hit types.Hit) (*schema.Document, error) {
	return func(ctx context.Context, hit types.Hit) (*schema.Document, error) {
		doc := &schema.Document{
			MetaData: map[string]any{},
		}

		if hit.Id_ != nil {
			doc.ID = *hit.Id_
		}
		if hit.Score_ != nil {
			doc.WithScore(float64(*hit.Score_))
		}

		if len(hit.Source_) > 0 {
			var src map[string]any
			if err := json.Unmarshal(hit.Source_, &src); err != nil {
				return nil, fmt.Errorf("parse es source: %w", err)
			}

			if content, ok := src["content"].(string); ok {
				doc.Content = content
			}

			if metadataRaw, ok := src["metadata"]; ok {
				if metadataMap, ok := metadataRaw.(map[string]any); ok {
					for k, v := range metadataMap {
						doc.MetaData[k] = v
					}
				}
			}

			for k, v := range src {
				if k == "content" || k == "metadata" {
					continue
				}
				doc.MetaData[k] = v
			}
		}

		return doc, nil
	}
}

func thresholdPtr(score float64) *float64 {
	if score <= 0 {
		return nil
	}
	return &score
}
