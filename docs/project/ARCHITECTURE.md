# LLM-Cache 架构文档

## 项目概览

LLM-Cache 是一个基于 **CloudWeGo Eino 框架** 和 Go 语言开发的智能语义缓存系统。通过 Eino 的流程编排能力，实现了灵活、可扩展的缓存查询和存储流程。

### 设计原则

1. **直接使用 Eino 类型** - 业务代码直接依赖 `embedding.Embedder`、`retriever.Retriever`、`compose.Runnable` 等 Eino 原生类型
2. **Graph 流程编排** - 使用 Eino Graph 实现业务流程，支持条件分支和并行执行
3. **组件化设计** - 通过工厂模式创建可替换的组件（Embedder、Retriever、Indexer）
4. **可观测性优先** - 内置 Callback 机制，支持日志、指标、追踪

## 技术栈

| 类别 | 技术选型 | 说明 |
|------|----------|------|
| 编程语言 | Go 1.23+ | 高性能并发支持 |
| 核心框架 | CloudWeGo Eino v0.7.3 | LLM 应用开发框架 |
| Web 框架 | Gin 1.10.1 | HTTP 路由和中间件 |
| 向量数据库 | Qdrant / Milvus / Redis / ES8 | 向量存储和检索 |
| Embedding | OpenAI / ARK / Ollama / Dashscope | 文本向量化服务 |
| 配置管理 | YAML + 环境变量 | 灵活的配置方式 |
| 日志系统 | log/slog | Go 标准库结构化日志 |

## 项目结构

```
llm-cache/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口，Eino 组件初始化
├── configs/
│   ├── config.go                   # 配置结构体定义
│   └── loader.go                   # 配置加载器
├── internal/
│   ├── app/
│   │   ├── handlers/
│   │   │   └── cache_handler.go    # HTTP Handler，依赖 compose.Runnable
│   │   ├── middleware/
│   │   │   └── logging.go          # HTTP 日志中间件
│   │   └── server/
│   │       ├── routes.go           # 路由配置
│   │       └── server.go           # HTTP 服务器
│   ├── domain/
│   │   └── models/
│   │       ├── cache.go            # 缓存领域模型
│   │       ├── request.go          # 请求模型
│   │       └── vector.go           # 向量模型
│   └── eino/                       # Eino 框架集成
│       ├── callbacks/              # Callback 处理器
│       │   ├── factory.go          # Callback 工厂
│       │   ├── logging.go          # 日志回调
│       │   ├── metrics.go          # 指标回调
│       │   └── tracing.go          # 追踪回调
│       ├── components/             # Eino 组件工厂
│       │   ├── embedder.go         # Embedder 工厂
│       │   ├── retriever.go        # Retriever 工厂
│       │   └── indexer.go          # Indexer 工厂
│       ├── config/
│       │   └── config.go           # Eino 配置结构
│       ├── flows/                  # 业务 Graph 流程
│       │   ├── cache_query.go      # 缓存查询 Graph
│       │   ├── cache_store.go      # 缓存存储 Graph
│       │   └── cache_delete.go     # 缓存删除服务
│       └── nodes/                  # Lambda 节点
│           ├── preprocessing.go    # 查询预处理
│           ├── postprocessing.go   # 结果后处理
│           ├── quality_check.go    # 质量检查
│           └── result_select.go    # 结果选择
├── pkg/
│   ├── logger/
│   │   └── logger.go               # 日志工具
│   └── status/
│       └── codes.go                # 状态码定义
├── docs/
│   └── project/
│       ├── ARCHITECTURE.md         # 架构文档（本文件）
│       └── EINO_INTEGRATION_PLAN.md # Eino 集成方案
├── go.mod
└── go.sum
```

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         客户端应用层                                  │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                    API 接口层 (Gin HTTP)                             │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │  CacheHandler                                                  │  │
│  │  - queryRunner: compose.Runnable[*QueryInput, *QueryOutput]   │  │
│  │  - storeRunner: compose.Runnable[*StoreInput, *StoreOutput]   │  │
│  │  - deleteService: *CacheDeleteService                         │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                   Eino Graph 流程编排层                              │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │ CacheQueryGraph                                              │    │
│  │ START → Preprocess → Retrieve → Select → Postprocess → END  │    │
│  └─────────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │ CacheStoreGraph                                              │    │
│  │ START → QualityCheck → Branch → Embed → Index → END         │    │
│  │                         ↓                                    │    │
│  │                    Reject (if failed)                        │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                     Eino 组件层                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │  Embedder   │  │  Retriever  │  │   Indexer   │                  │
│  │  (OpenAI)   │  │  (Qdrant)   │  │  (Qdrant)   │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                     数据存储层                                       │
│         Qdrant  |  Milvus  |  Redis  |  Elasticsearch               │
└─────────────────────────────────────────────────────────────────────┘
```

### 层次说明

| 层次 | 职责 | 主要组件 |
|------|------|----------|
| API 接口层 | HTTP 请求处理、参数验证、响应格式化 | CacheHandler, Gin Router |
| Graph 流程层 | 业务流程编排、条件分支、节点执行 | CacheQueryGraph, CacheStoreGraph |
| 组件层 | 向量化、检索、索引等原子操作 | Embedder, Retriever, Indexer |
| 数据存储层 | 向量数据持久化和检索 | Qdrant, Milvus, Redis, ES8 |

## 核心模块详解

### 1. Eino 组件工厂 (internal/eino/components/)

组件工厂负责根据配置创建 Eino 原生组件实例。

#### Embedder 工厂

```go
// internal/eino/components/embedder.go

// NewEmbedder 创建 Eino Embedder 实例
func NewEmbedder(ctx context.Context, cfg *config.EmbedderConfig) (embedding.Embedder, error) {
    switch cfg.Provider {
    case "openai":
        return openaiembed.NewEmbedder(ctx, &openaiembed.EmbeddingConfig{
            APIKey:  cfg.APIKey,
            BaseURL: cfg.BaseURL,
            Model:   cfg.Model,
            Timeout: time.Duration(cfg.Timeout) * time.Second,
        })
    case "ark":
        return ark.NewEmbedder(ctx, &ark.EmbeddingConfig{...})
    case "ollama":
        return ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{...})
    // ... 其他提供商
    }
}
```

**支持的 Embedding 提供商**:
- `openai` - OpenAI API
- `ark` - 火山引擎 ARK
- `ollama` - 本地 Ollama
- `dashscope` - 阿里云 Dashscope
- `qianfan` - 百度千帆
- `tencentcloud` - 腾讯云

#### Retriever 工厂

```go
// internal/eino/components/retriever.go

// NewRetriever 创建 Eino Retriever 实例
func NewRetriever(ctx context.Context, cfg *config.RetrieverConfig, embedder embedding.Embedder) (retriever.Retriever, error) {
    switch cfg.Provider {
    case "qdrant":
        client, _ := qdrantClient.NewClient(&qdrantClient.Config{
            Host: cfg.Qdrant.Host,
            Port: cfg.Qdrant.Port,
        })
        return qdrantretriever.NewRetriever(ctx, &qdrantretriever.Config{
            Client:     client,
            Collection: cfg.Collection,
            Embedding:  embedder,
            TopK:       cfg.TopK,
        })
    // ... 其他提供商
    }
}
```

**支持的向量数据库**:
- `qdrant` - Qdrant
- `milvus` - Milvus
- `redis` - Redis Stack
- `es8` - Elasticsearch 8

### 2. Lambda 节点 (internal/eino/nodes/)

Lambda 节点封装了业务逻辑，作为 Graph 中的处理单元。

#### 预处理节点

```go
// internal/eino/nodes/preprocessing.go

// PreprocessInput 预处理输入
type PreprocessInput struct {
    Query    string
    UserType string
}

// PreprocessOutput 预处理输出
type PreprocessOutput struct {
    Query    string
    UserType string
}

// PreprocessQuery 查询预处理
func PreprocessQuery(ctx context.Context, input *PreprocessInput) (*PreprocessOutput, error) {
    query := input.Query
    
    // 1. 去除首尾空白
    query = strings.TrimSpace(query)
    
    // 2. 规范化空白字符
    query = normalizeWhitespace(query)
    
    // 3. 移除控制字符
    query = removeControlChars(query)
    
    return &PreprocessOutput{
        Query:    query,
        UserType: input.UserType,
    }, nil
}
```

#### 质量检查节点

```go
// internal/eino/nodes/quality_check.go

// QualityChecker 质量检查器
type QualityChecker struct {
    cfg *config.QualityConfig
}

// Check 执行质量检查
func (c *QualityChecker) Check(ctx context.Context, input *QualityCheckInput) (*QualityCheckResult, error) {
    // 1. 检查问题长度
    if len(input.Question) < c.cfg.MinQuestionLength {
        return &QualityCheckResult{Passed: false, Reason: "question too short"}, nil
    }
    
    // 2. 检查答案长度
    if len(input.Answer) < c.cfg.MinAnswerLength {
        return &QualityCheckResult{Passed: false, Reason: "answer too short"}, nil
    }
    
    // 3. 检查黑名单
    if containsBlacklistWords(input.Question) || containsBlacklistWords(input.Answer) {
        return &QualityCheckResult{Passed: false, Reason: "contains blacklisted content"}, nil
    }
    
    // 4. 计算质量分数
    score := calculateQualityScore(input.Question, input.Answer)
    if score < c.cfg.ScoreThreshold {
        return &QualityCheckResult{Passed: false, Reason: "quality score below threshold"}, nil
    }
    
    return &QualityCheckResult{Passed: true}, nil
}
```

#### 结果选择节点

```go
// internal/eino/nodes/result_select.go

// ResultSelector 结果选择器
type ResultSelector struct {
    strategy    string  // first, highest_score, temperature_softmax
    temperature float64
}

// Select 选择最佳结果
func (s *ResultSelector) Select(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
    if len(docs) == 0 {
        return nil, nil
    }
    
    switch s.strategy {
    case "first":
        return docs[0], nil
    case "highest_score":
        return s.selectHighestScore(docs), nil
    case "temperature_softmax":
        return s.selectBySoftmax(docs), nil
    default:
        return docs[0], nil
    }
}
```

**支持的选择策略**:
- `first` - 返回第一个结果
- `highest_score` - 返回最高分结果
- `temperature_softmax` - 基于温度的 Softmax 采样

### 3. 业务 Graph (internal/eino/flows/)

#### 缓存查询 Graph

```go
// internal/eino/flows/cache_query.go

// CacheQueryGraph 缓存查询 Graph
type CacheQueryGraph struct {
    embedder  embedding.Embedder
    retriever retriever.Retriever
    cfg       *config.QueryConfig
}

// Compile 编译 Graph 为可执行的 Runnable
func (g *CacheQueryGraph) Compile(ctx context.Context) (compose.Runnable[*CacheQueryInput, *CacheQueryOutput], error) {
    graph := compose.NewGraph[*CacheQueryInput, *CacheQueryOutput]()
    
    // 添加节点
    graph.AddLambdaNode("preprocess", compose.InvokableLambda(nodes.PreprocessQuery))
    graph.AddRetrieverNode("retrieve", g.retriever)
    graph.AddLambdaNode("select", compose.InvokableLambda(nodes.NewResultSelector(g.cfg.SelectionStrategy, g.cfg.Temperature)))
    graph.AddLambdaNode("postprocess", compose.InvokableLambda(nodes.PostprocessResult))
    
    // 设置边
    graph.AddEdge(compose.START, "preprocess")
    graph.AddEdge("preprocess", "retrieve")
    graph.AddEdge("retrieve", "select")
    graph.AddEdge("select", "postprocess")
    graph.AddEdge("postprocess", compose.END)
    
    return graph.Compile(ctx, compose.WithGraphName("cache_query"))
}
```

**查询流程**:

```
┌───────┐   ┌────────────┐   ┌──────────┐   ┌────────┐   ┌─────────────┐   ┌─────┐
│ START │ → │ Preprocess │ → │ Retrieve │ → │ Select │ → │ Postprocess │ → │ END │
└───────┘   └────────────┘   └──────────┘   └────────┘   └─────────────┘   └─────┘
```

#### 缓存存储 Graph

```go
// internal/eino/flows/cache_store.go

// CacheStoreGraph 缓存存储 Graph
type CacheStoreGraph struct {
    embedder embedding.Embedder
    indexer  indexer.Indexer
    cfg      *config.StoreConfig
    quality  *config.QualityConfig
}

// Compile 编译 Graph 为可执行的 Runnable
func (g *CacheStoreGraph) Compile(ctx context.Context) (compose.Runnable[*CacheStoreInput, *CacheStoreOutput], error) {
    graph := compose.NewGraph[*CacheStoreInput, *CacheStoreOutput]()
    
    // 添加节点
    checker := nodes.NewQualityChecker(g.quality)
    graph.AddLambdaNode("quality_check", compose.InvokableLambda(checker.Check))
    graph.AddLambdaNode("embed", compose.InvokableLambda(g.embedQuestion))
    graph.AddLambdaNode("index", compose.InvokableLambda(g.indexDocument))
    graph.AddLambdaNode("reject", compose.InvokableLambda(g.rejectStore))
    
    // 设置边和条件分支
    graph.AddEdge(compose.START, "quality_check")
    graph.AddBranch("quality_check", compose.NewGraphBranch(
        func(ctx context.Context, result *nodes.QualityCheckResult) (string, error) {
            if result.Passed {
                return "embed", nil
            }
            return "reject", nil
        },
        map[string]bool{"embed": true, "reject": true},
    ))
    graph.AddEdge("embed", "index")
    graph.AddEdge("index", compose.END)
    graph.AddEdge("reject", compose.END)
    
    return graph.Compile(ctx, compose.WithGraphName("cache_store"))
}
```

**存储流程**:

```
┌───────┐   ┌───────────────┐   ┌─────────┐   ┌───────┐   ┌─────┐
│ START │ → │ QualityCheck  │ → │  Embed  │ → │ Index │ → │ END │
└───────┘   └───────────────┘   └─────────┘   └───────┘   └─────┘
                    │
                    ↓ (if failed)
              ┌──────────┐
              │  Reject  │ → END
              └──────────┘
```

### 4. Callback 处理器 (internal/eino/callbacks/)

Callback 处理器提供可观测性支持，在组件执行的各个阶段被调用。

#### Callback 接口

```go
// Eino Callback 接口
type Handler interface {
    OnStart(ctx context.Context, info *RunInfo, input CallbackInput) context.Context
    OnEnd(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context
    OnError(ctx context.Context, info *RunInfo, err error) context.Context
    OnStartWithStreamInput(ctx context.Context, info *RunInfo, input *StreamReader[CallbackInput]) context.Context
    OnEndWithStreamOutput(ctx context.Context, info *RunInfo, output *StreamReader[CallbackOutput]) context.Context
}
```

#### 日志回调

```go
// internal/eino/callbacks/logging.go

// LoggingHandler 日志回调处理器
type LoggingHandler struct {
    logger logger.Logger
    cfg    *config.LoggingCallbackConfig
}

// OnStart 组件开始执行时调用
func (h *LoggingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
    h.logger.InfoContext(ctx, "组件开始执行",
        "component", info.Component,
        "name", info.Name,
        "type", info.Type,
    )
    return context.WithValue(ctx, startTimeKey, time.Now())
}

// OnEnd 组件执行完成时调用
func (h *LoggingHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
    startTime, _ := ctx.Value(startTimeKey).(time.Time)
    duration := time.Since(startTime)
    
    h.logger.InfoContext(ctx, "组件执行完成",
        "component", info.Component,
        "name", info.Name,
        "duration_ms", duration.Milliseconds(),
    )
    return ctx
}
```

#### 指标回调

```go
// internal/eino/callbacks/metrics.go

// MetricsHandler 指标回调处理器
type MetricsHandler struct {
    cfg     *config.MetricsCallbackConfig
    metrics *MetricsCollector
}

// GetMetrics 获取当前指标
func (h *MetricsHandler) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "total_calls":      h.metrics.TotalCalls,
        "successful_calls": h.metrics.SuccessfulCalls,
        "failed_calls":     h.metrics.FailedCalls,
        "avg_latency_ms":   h.metrics.TotalLatencyMs / h.metrics.SuccessfulCalls,
        "component_stats":  h.metrics.ComponentLatency,
    }
}
```

#### Callback 工厂

```go
// internal/eino/callbacks/factory.go

// Factory Callback 工厂
type Factory struct {
    cfg    *config.CallbacksConfig
    logger logger.Logger
}

// CreateHandlers 创建所有启用的 Callback 处理器
func (f *Factory) CreateHandlers() []callbacks.Handler {
    handlers := make([]callbacks.Handler, 0)
    
    if f.cfg.Logging.Enabled {
        handlers = append(handlers, NewLoggingHandler(f.logger, &f.cfg.Logging))
    }
    if f.cfg.Metrics.Enabled {
        handlers = append(handlers, NewMetricsHandler(&f.cfg.Metrics))
    }
    if f.cfg.Tracing.Enabled {
        handlers = append(handlers, NewTracingHandler(&f.cfg.Tracing, f.logger))
    }
    // Langfuse, APMPlus, Cozeloop 等外部集成
    
    return handlers
}
```

## 数据流程

### 缓存查询流程

```
1. 客户端发送查询请求
   POST /v1/cache/search
   {"question": "什么是机器学习?", "user_type": "default"}

2. CacheHandler 接收请求
   - 参数验证
   - 构建 CacheQueryInput

3. queryRunner.Invoke(ctx, input)
   - Preprocess: 清洗查询文本
   - Retrieve: 向量检索 Top-K 结果
   - Select: 选择最佳匹配
   - Postprocess: 格式化输出

4. 返回响应
   {"hit": true, "answer": "...", "score": 0.95}
```

### 缓存存储流程

```
1. 客户端发送存储请求
   POST /v1/cache/store
   {"question": "...", "answer": "...", "user_type": "default"}

2. CacheHandler 接收请求
   - 参数验证
   - 构建 CacheStoreInput

3. storeRunner.Invoke(ctx, input)
   - QualityCheck: 质量检查
     - 通过 → 继续
     - 失败 → 拒绝存储
   - Embed: 生成向量
   - Index: 存储到向量数据库

4. 返回响应
   {"success": true, "cache_id": "..."}
```

## 配置说明

### Eino 配置结构

```go
// internal/eino/config/config.go

// EinoConfig Eino 框架相关配置
type EinoConfig struct {
    Embedder   EmbedderConfig   `yaml:"embedder"`
    Retriever  RetrieverConfig  `yaml:"retriever"`
    Indexer    IndexerConfig    `yaml:"indexer"`
    Query      QueryConfig      `yaml:"query"`
    Store      StoreConfig      `yaml:"store"`
    Quality    QualityConfig    `yaml:"quality"`
    Callbacks  CallbacksConfig  `yaml:"callbacks"`
}
```

### 配置示例

```yaml
# configs/config.yaml
eino:
  embedder:
    provider: "openai"
    api_key: "${OPENAI_API_KEY}"
    model: "text-embedding-3-small"
    timeout: 30
  
  retriever:
    provider: "qdrant"
    collection: "llm_cache"
    top_k: 5
    score_threshold: 0.7
    qdrant:
      host: "localhost"
      port: 6334
  
  indexer:
    provider: "qdrant"
    collection: "llm_cache"
    vector_size: 1536
    qdrant:
      host: "localhost"
      port: 6334
      distance: "Cosine"
  
  query:
    preprocess_enabled: true
    postprocess_enabled: true
    selection_strategy: "highest_score"
    temperature: 0.7
  
  store:
    quality_check_enabled: true
    min_question_length: 5
    min_answer_length: 10
  
  callbacks:
    logging:
      enabled: true
      level: "info"
    metrics:
      enabled: true
```

## 扩展指南

### 添加新的 Embedding 提供商

1. 在 `internal/eino/components/embedder.go` 中添加新的 case：

```go
case "new_provider":
    return newprovider.NewEmbedder(ctx, &newprovider.EmbeddingConfig{
        // 配置参数
    })
```

2. 在 `internal/eino/config/config.go` 中添加相应的配置字段。

### 添加新的向量数据库

1. 在 `internal/eino/components/retriever.go` 和 `indexer.go` 中添加新的 case。
2. 确保使用 `eino-ext` 提供的组件，或实现 `retriever.Retriever` 和 `indexer.Indexer` 接口。

### 添加新的 Lambda 节点

1. 在 `internal/eino/nodes/` 中创建新的节点文件。
2. 实现节点函数，签名为 `func(context.Context, InputType) (OutputType, error)`。
3. 在 Graph 中添加节点：`graph.AddLambdaNode("name", compose.InvokableLambda(nodeFunc))`。

### 添加新的 Callback

1. 在 `internal/eino/callbacks/` 中创建新的 callback 文件。
2. 实现 `callbacks.Handler` 接口。
3. 在 `factory.go` 中的 `CreateHandlers` 方法中添加创建逻辑。

## 性能优化

### 向量检索优化

- 使用 HNSW 索引加速 ANN 检索
- 合理设置 `top_k` 和 `score_threshold` 减少无效结果
- 对于高并发场景，使用连接池

### Embedding 优化

- 批量处理请求，减少 API 调用次数
- 使用本地 Embedding 模型（Ollama）降低延迟
- 缓存常见查询的向量结果

### Graph 执行优化

- 使用并行节点处理独立任务
- 合理设置 Timeout 避免长时间阻塞
- 使用 Callback 监控性能瓶颈

## 测试策略

### 单元测试

```bash
# 运行所有单元测试
go test ./...

# 运行 Eino 节点测试
go test ./internal/eino/nodes/... -v

# 覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 集成测试

```bash
# 启动本地 Qdrant
docker run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant:latest

# 运行集成测试
go test -tags=integration ./test/...
```

## 参考文档

- [CloudWeGo Eino 框架](https://github.com/cloudwego/eino)
- [Eino 扩展组件](https://github.com/cloudwego/eino-ext)
- [Qdrant 文档](https://qdrant.tech/documentation/)
- [Eino 集成方案](./EINO_INTEGRATION_PLAN.md)

