// Package components 提供 Eino 组件的工厂函数
package components

import (
	"context"
	"encoding/json"
	"fmt"

	es8retriever "github.com/cloudwego/eino-ext/components/retriever/es8"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	qdrantretriever "github.com/cloudwego/eino-ext/components/retriever/qdrant"
	redisretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	vikingdbretriever "github.com/cloudwego/eino-ext/components/retriever/volc_vikingdb"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	qdrantClient "github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"

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
func newMilvusRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	// 创建 Milvus 客户端
	client, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port),
		Username: cfg.Milvus.Username,
		Password: cfg.Milvus.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	// 创建 Retriever 配置
	retrieverCfg := &milvusretriever.RetrieverConfig{
		Client:         client,
		Collection:     cfg.Collection,
		VectorField:    cfg.Milvus.VectorField,
		OutputFields:   cfg.Milvus.OutputFields,
		TopK:           cfg.TopK,
		ScoreThreshold: cfg.ScoreThreshold,
		Embedding:      embedder,
	}

	return milvusretriever.NewRetriever(ctx, retrieverCfg)
}

// newRedisRetriever 创建 Redis Retriever
func newRedisRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	// 创建 Redis 客户端
	// 注意：需要配置 Protocol: 2 和 UnstableResp3: true 以支持 FT.SEARCH
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		Protocol:     2,
		UnstableResp3: true,
	})

	// 创建 Retriever 配置
	retrieverCfg := &redisretriever.RetrieverConfig{
		Client:       rdb,
		Index:        cfg.Redis.Index,
		VectorField:  cfg.Redis.VectorField,
		TopK:         cfg.TopK,
		Embedding:    embedder,
		ReturnFields: cfg.Redis.ReturnFields,
	}

	// 设置距离阈值（用于范围搜索）
	if cfg.ScoreThreshold > 0 {
		threshold := cfg.ScoreThreshold
		retrieverCfg.DistanceThreshold = &threshold
	}

	return redisretriever.NewRetriever(ctx, retrieverCfg)
}

// newES8Retriever 创建 Elasticsearch 8 Retriever
func newES8Retriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	// 创建 Elasticsearch 客户端
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.ES8.Addresses,
		Username:  cfg.ES8.Username,
		Password:  cfg.ES8.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 解析搜索模式
	searchMode := parseES8SearchMode(cfg.ES8.SearchMode, cfg.ES8.VectorField)

	// 创建 Retriever 配置
	retrieverCfg := &es8retriever.RetrieverConfig{
		Client:       esClient,
		Index:        cfg.ES8.Index,
		TopK:         cfg.TopK,
		Embedding:    embedder,
		SearchMode:   searchMode,
		ResultParser: defaultES8ResultParser,
	}

	// 设置相似度阈值
	if cfg.ScoreThreshold > 0 {
		threshold := cfg.ScoreThreshold
		retrieverCfg.ScoreThreshold = &threshold
	}

	return es8retriever.NewRetriever(ctx, retrieverCfg)
}

// knnSearchMode 实现 es8retriever.SearchMode 接口，用于 KNN 向量搜索
type knnSearchMode struct {
	vectorField string
}

// BuildRequest 构建 ES8 KNN 搜索请求
func (m *knnSearchMode) BuildRequest(ctx context.Context, conf *es8retriever.RetrieverConfig, query string, opts ...retriever.Option) (*search.Request, error) {
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK:      &conf.TopK,
		Embedding: conf.Embedding,
	}, opts...)

	if options.Embedding == nil {
		return nil, fmt.Errorf("embedding is required for KNN search")
	}

	// 生成查询向量
	vectors, err := options.Embedding.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("embedding returned empty vector")
	}

	// 转换为 float32 切片
	vector := make([]float32, len(vectors[0]))
	for i, v := range vectors[0] {
		vector[i] = float32(v)
	}

	// 构建 KNN 查询
	topK := conf.TopK
	knn := types.KnnSearch{
		Field:       m.vectorField,
		QueryVector: vector,
		K:           &topK,
		NumCandidates: &topK,
	}

	return &search.Request{
		Knn: []types.KnnSearch{knn},
	}, nil
}

// parseES8SearchMode 解析 ES8 搜索模式
func parseES8SearchMode(mode string, vectorField string) es8retriever.SearchMode {
	if vectorField == "" {
		vectorField = "vector"
	}
	// 使用自定义的 KNN 搜索模式
	return &knnSearchMode{vectorField: vectorField}
}

// defaultES8ResultParser 默认的 ES8 结果解析器
func defaultES8ResultParser(ctx context.Context, hit types.Hit) (*schema.Document, error) {
	doc := &schema.Document{
		MetaData: make(map[string]any),
	}

	// 解析 ID
	if hit.Id_ != nil {
		doc.ID = *hit.Id_
	}

	// 解析 Source (json.RawMessage 类型)
	if len(hit.Source_) > 0 {
		var source map[string]any
		if err := json.Unmarshal(hit.Source_, &source); err == nil {
			// 尝试解析 content 字段
			if content, ok := source["content"]; ok {
				if str, ok := content.(string); ok {
					doc.Content = str
				}
			}

			// 将其他字段作为元数据
			for k, v := range source {
				if k != "content" && k != "vector" {
					doc.MetaData[k] = v
				}
			}
		}
	}

	// 解析分数
	if hit.Score_ != nil {
		doc.MetaData["_score"] = float64(*hit.Score_)
	}

	return doc, nil
}

// newVikingDBRetriever 创建 VikingDB Retriever
func newVikingDBRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
	// 创建 Retriever 配置
	retrieverCfg := &vikingdbretriever.RetrieverConfig{
		Host:              cfg.VikingDB.Host,
		Region:            cfg.VikingDB.Region,
		AK:                cfg.VikingDB.AK,
		SK:                cfg.VikingDB.SK,
		Scheme:            cfg.VikingDB.Scheme,
		ConnectionTimeout: cfg.VikingDB.ConnectionTimeout,
		Collection:        cfg.Collection,
		Index:             cfg.VikingDB.Index,
		Partition:         cfg.VikingDB.Partition,
		WithMultiModal:    cfg.VikingDB.WithMultiModal,
	}

	// 设置 TopK
	if cfg.TopK > 0 {
		topK := cfg.TopK
		retrieverCfg.TopK = &topK
	}

	// 设置相似度阈值
	if cfg.ScoreThreshold > 0 {
		threshold := cfg.ScoreThreshold
		retrieverCfg.ScoreThreshold = &threshold
	}

	// 配置 Embedding
	if !cfg.VikingDB.WithMultiModal {
		if cfg.VikingDB.UseBuiltinEmbedding {
			retrieverCfg.EmbeddingConfig = vikingdbretriever.EmbeddingConfig{
				UseBuiltin:  true,
				ModelName:   cfg.VikingDB.EmbeddingModelName,
				UseSparse:   cfg.VikingDB.UseSparse,
				DenseWeight: cfg.VikingDB.DenseWeight,
			}
		} else {
			retrieverCfg.EmbeddingConfig = vikingdbretriever.EmbeddingConfig{
				UseBuiltin: false,
				Embedding:  embedder,
			}
		}
	}

	return vikingdbretriever.NewRetriever(ctx, retrieverCfg)
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
