// Package flows 提供 Eino Graph 流程定义
package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"llm-cache/internal/eino/config"
	"llm-cache/internal/eino/nodes"
)

// CacheStoreInput 存储输入
type CacheStoreInput struct {
	Question   string         `json:"question"`
	Answer     string         `json:"answer"`
	UserType   string         `json:"user_type"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	ForceWrite bool           `json:"force_write,omitempty"`
}

// CacheStoreOutput 存储输出
type CacheStoreOutput struct {
	Success  bool   `json:"success"`
	CacheID  string `json:"cache_id,omitempty"`
	Rejected bool   `json:"rejected,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// EmbeddingResult 嵌入结果（内部使用）
type EmbeddingResult struct {
	Question string
	Answer   string
	UserType string
	Metadata map[string]any
	Vector   []float64
	Rejected bool
	Reason   string
}

// CacheStoreGraph 缓存存储 Graph
type CacheStoreGraph struct {
	embedder         embedding.Embedder
	indexer          indexer.Indexer
	cfg              *config.StoreConfig
	quality          *config.QualityConfig
	callbackHandlers []callbacks.Handler
}

// NewCacheStoreGraph 创建缓存存储 Graph
func NewCacheStoreGraph(
	embedder embedding.Embedder,
	idx indexer.Indexer,
	cfg *config.StoreConfig,
	quality *config.QualityConfig,
	callbackHandlers ...callbacks.Handler,
) *CacheStoreGraph {
	return &CacheStoreGraph{
		embedder:         embedder,
		indexer:          idx,
		cfg:              cfg,
		quality:          quality,
		callbackHandlers: callbackHandlers,
	}
}

// Compile 编译 Graph 为 Runnable
func (g *CacheStoreGraph) Compile(ctx context.Context) (compose.Runnable[*CacheStoreInput, *CacheStoreOutput], error) {
	graph := compose.NewGraph[*CacheStoreInput, *CacheStoreOutput]()

	// 1. 添加质量检查节点
	qualityChecker := nodes.NewQualityChecker(g.quality)
	qualityNode := compose.InvokableLambda(func(ctx context.Context, input *CacheStoreInput) (*nodes.QualityCheckResult, error) {
		return qualityChecker.Check(ctx, &nodes.QualityCheckInput{
			Question:   input.Question,
			Answer:     input.Answer,
			UserType:   input.UserType,
			Metadata:   input.Metadata,
			ForceWrite: input.ForceWrite,
		})
	})
	if err := graph.AddLambdaNode("quality_check", qualityNode); err != nil {
		return nil, fmt.Errorf("add quality_check node: %w", err)
	}

	// 2. 添加 Embedding 节点（处理质量检查结果）
	embeddingNode := compose.InvokableLambda(func(ctx context.Context, result *nodes.QualityCheckResult) (*EmbeddingResult, error) {
		if !result.Passed {
			return &EmbeddingResult{
				Rejected: true,
				Reason:   result.Reason,
			}, nil
		}

		vectors, err := g.embedder.EmbedStrings(ctx, []string{result.Question})
		if err != nil {
			return nil, fmt.Errorf("embed question: %w", err)
		}
		if len(vectors) == 0 {
			return nil, fmt.Errorf("no embedding generated")
		}

		return &EmbeddingResult{
			Question: result.Question,
			Answer:   result.Answer,
			UserType: result.UserType,
			Metadata: result.Metadata,
			Vector:   vectors[0],
		}, nil
	})
	if err := graph.AddLambdaNode("embedding", embeddingNode); err != nil {
		return nil, fmt.Errorf("add embedding node: %w", err)
	}

	// 3. 添加 Index 节点
	indexNode := compose.InvokableLambda(func(ctx context.Context, result *EmbeddingResult) (*CacheStoreOutput, error) {
		cacheID := generateCacheID()

		doc := &schema.Document{
			ID:      cacheID,
			Content: result.Question,
			MetaData: map[string]any{
				"question":   result.Question,
				"answer":     result.Answer,
				"user_type":  result.UserType,
				"created_at": time.Now().Unix(),
			},
		}

		// 合并自定义 Metadata
		for k, v := range result.Metadata {
			doc.MetaData[k] = v
		}

		ids, err := g.indexer.Store(ctx, []*schema.Document{doc})
		if err != nil {
			return nil, fmt.Errorf("store document: %w", err)
		}

		return &CacheStoreOutput{
			Success: true,
			CacheID: ids[0],
		}, nil
	})
	if err := graph.AddLambdaNode("index_node", indexNode); err != nil {
		return nil, fmt.Errorf("add index node: %w", err)
	}

	// 4. 添加拒绝节点
	rejectNode := compose.InvokableLambda(func(ctx context.Context, result *EmbeddingResult) (*CacheStoreOutput, error) {
		return &CacheStoreOutput{
			Success:  false,
			Rejected: true,
			Reason:   result.Reason,
		}, nil
	})
	if err := graph.AddLambdaNode("reject_node", rejectNode); err != nil {
		return nil, fmt.Errorf("add reject node: %w", err)
	}

	// 5. 添加条件分支（需在相关节点创建后）
	branch := compose.NewGraphBranch(func(ctx context.Context, result *EmbeddingResult) (string, error) {
		if result.Rejected {
			return "reject_node", nil
		}
		return "index_node", nil
	}, map[string]bool{
		"index_node":  true,
		"reject_node": true,
	})

	if err := graph.AddBranch("embedding", branch); err != nil {
		return nil, fmt.Errorf("add branch: %w", err)
	}

	// 6. 连接节点
	if err := graph.AddEdge(compose.START, "quality_check"); err != nil {
		return nil, fmt.Errorf("add edge START->quality_check: %w", err)
	}
	if err := graph.AddEdge("quality_check", "embedding"); err != nil {
		return nil, fmt.Errorf("add edge quality_check->embedding: %w", err)
	}
	// Branch 已经定义了 embedding 到 index_node/reject_node 的连接
	if err := graph.AddEdge("index_node", compose.END); err != nil {
		return nil, fmt.Errorf("add edge index_node->END: %w", err)
	}
	if err := graph.AddEdge("reject_node", compose.END); err != nil {
		return nil, fmt.Errorf("add edge reject_node->END: %w", err)
	}

	// 编译 Graph（带 Callback 处理器）
	compileOpts := []compose.GraphCompileOption{
		compose.WithGraphName("cache_store"),
	}
	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, err
	}

	// 如果有 Callback 处理器，通过 RunOption 注入
	// Eino 的 Callback 在运行时通过 RunOption 传入
	_ = g.callbackHandlers // 在 Invoke 时使用

	return runnable, nil
}

// Run 执行存储
func (g *CacheStoreGraph) Run(ctx context.Context, input *CacheStoreInput) (*CacheStoreOutput, error) {
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}

	return runnable.Invoke(ctx, input)
}

// generateCacheID 生成缓存 ID
func generateCacheID() string {
	return uuid.New().String()
}
