// Package flows 提供 Eino Graph 流程定义
package flows

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"llm-cache/internal/eino/config"
	"llm-cache/internal/eino/nodes"
)

// CacheQueryInput 查询输入
type CacheQueryInput struct {
	Query          string  `json:"query"`
	UserType       string  `json:"user_type"`
	TopK           int     `json:"top_k,omitempty"`
	ScoreThreshold float64 `json:"score_threshold,omitempty"`
}

// CacheQueryOutput 查询输出
type CacheQueryOutput struct {
	Hit      bool           `json:"hit"`
	Question string         `json:"question,omitempty"`
	Answer   string         `json:"answer,omitempty"`
	Score    float64        `json:"score,omitempty"`
	CacheID  string         `json:"cache_id,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CacheQueryGraph 缓存查询 Graph
type CacheQueryGraph struct {
	embedder         embedding.Embedder
	retriever        retriever.Retriever
	cfg              *config.QueryConfig
	callbackHandlers []callbacks.Handler
}

// NewCacheQueryGraph 创建缓存查询 Graph
func NewCacheQueryGraph(
	embedder embedding.Embedder,
	ret retriever.Retriever,
	cfg *config.QueryConfig,
	callbackHandlers ...callbacks.Handler,
) *CacheQueryGraph {
	return &CacheQueryGraph{
		embedder:         embedder,
		retriever:        ret,
		cfg:              cfg,
		callbackHandlers: callbackHandlers,
	}
}

// Compile 编译 Graph 为 Runnable
func (g *CacheQueryGraph) Compile(ctx context.Context) (compose.Runnable[*CacheQueryInput, *CacheQueryOutput], error) {
	graph := compose.NewGraph[*CacheQueryInput, *CacheQueryOutput]()

	// 1. 添加预处理节点
	preprocessNode := compose.InvokableLambda(func(ctx context.Context, input *CacheQueryInput) (string, error) {
		if g.cfg.PreprocessEnabled {
			return nodes.PreprocessQueryToString(ctx, input.Query)
		}
		return input.Query, nil
	})
	if err := graph.AddLambdaNode("preprocess", preprocessNode); err != nil {
		return nil, fmt.Errorf("add preprocess node: %w", err)
	}

	// 2. 添加 Retriever 节点
	if err := graph.AddRetrieverNode("retrieve", g.retriever); err != nil {
		return nil, fmt.Errorf("add retriever node: %w", err)
	}

	// 3. 添加结果选择节点
	selector := nodes.NewResultSelector(g.cfg.SelectionStrategy, g.cfg.Temperature)
	selectNode := compose.InvokableLambda(selector.Select)
	if err := graph.AddLambdaNode("select", selectNode); err != nil {
		return nil, fmt.Errorf("add select node: %w", err)
	}

	// 4. 添加后处理节点
	postprocessNode := compose.InvokableLambda(func(ctx context.Context, doc *schema.Document) (*CacheQueryOutput, error) {
		if doc == nil {
			return &CacheQueryOutput{Hit: false}, nil
		}

		output := &CacheQueryOutput{
			Hit:      true,
			CacheID:  doc.ID,
			Metadata: doc.MetaData,
		}

		// 从 MetaData 提取问答
		if question, ok := doc.MetaData["question"].(string); ok {
			output.Question = question
		}
		if answer, ok := doc.MetaData["answer"].(string); ok {
			output.Answer = answer
		}
		if score, ok := doc.MetaData["score"].(float64); ok {
			output.Score = score
		}

		return output, nil
	})
	if err := graph.AddLambdaNode("postprocess", postprocessNode); err != nil {
		return nil, fmt.Errorf("add postprocess node: %w", err)
	}

	// 5. 连接节点
	if err := graph.AddEdge(compose.START, "preprocess"); err != nil {
		return nil, fmt.Errorf("add edge START->preprocess: %w", err)
	}
	if err := graph.AddEdge("preprocess", "retrieve"); err != nil {
		return nil, fmt.Errorf("add edge preprocess->retrieve: %w", err)
	}
	if err := graph.AddEdge("retrieve", "select"); err != nil {
		return nil, fmt.Errorf("add edge retrieve->select: %w", err)
	}
	if err := graph.AddEdge("select", "postprocess"); err != nil {
		return nil, fmt.Errorf("add edge select->postprocess: %w", err)
	}
	if err := graph.AddEdge("postprocess", compose.END); err != nil {
		return nil, fmt.Errorf("add edge postprocess->END: %w", err)
	}

	// 编译 Graph（带 Callback 处理器）
	compileOpts := []compose.GraphCompileOption{
		compose.WithGraphName("cache_query"),
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

// Run 执行查询
func (g *CacheQueryGraph) Run(ctx context.Context, input *CacheQueryInput) (*CacheQueryOutput, error) {
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}

	return runnable.Invoke(ctx, input)
}
