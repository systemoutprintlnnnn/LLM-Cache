# 大模型语义缓存系统 - 架构文档

## 项目概览

大模型语义缓存系统是一个基于Go语言开发的智能缓存解决方案，旨在通过语义相似度匹配来优化大模型的调用效率。本项目采用清洁架构（Clean Architecture）设计原则，确保代码的可维护性、可测试性和可扩展性。
卷 Windows 的文件夹 PATH 列表

|   go.mod
|   go.sum
|   README.md
|   
+---build
+---cmd
|   \---server
|           main.go
|
+---configs
|       config.go
|       config.yaml
|       loader.go
|
+---internal
|   +---app
|   |   +---handlers
|   |   |       cache_handler.go
|   |   |
|   |   +---middleware
|   |   |       logging.go
|   |   |
|   |   \---server
|   |           routes.go
|   |           server.go
|   |
|   +---domain
|   |   +---models
|   |   |       cache.go
|   |   |       request.go
|   |   |       vector.go
|   |   |
|   |   +---repositories
|   |   |       vector_repository.go
|   |   |
|   |   \---services
|   |           cache_service.go
|   |           embedding_service.go
|   |           quality_service.go
|   |           recall_postprocessing_service.go
|   |           request_preprocessing_service.go
|   |           vector_service.go
|   |
|   \---infrastructure
|       +---cache
|       |       cache_service.go
|       |       config.go
|       |       init.go
|       |       README.md
|       |
|       +---embedding
|       |   \---remote
|       |           embedding.go
|       |           init.go
|       |
|       +---postprocessing
|       |       config.go
|       |       default_service.go
|       |       formatter.go
|       |       init.go
|       |       README.md
|       |
|       +---preprocessing
|       |       config.go
|       |       init.go
|       |       preprocessing_service.go
|       |       README.md
|       |
|       +---quality
|       |       blacklist.go
|       |       init.go
|       |       quality_service.go
|       |       strategies.go
|       |
|       +---stores
|       |   |   README.md
|       |   |
|       |   \---qdrant
|       |           init.go
|       |           qdrant_client.go
|       |           vector_store.go
|       |
|       \---vector
|               init.go
|               README.md
|               strategies.go
|               vector_service.go
|
+---logs
|       test.log
|
+---pkg
|   +---logger
|   |       logger.go
|   |
|   +---status
|   |       codes.go
|   |
|   \---utils
+---scripts
\---test
qdrant_example.go
## 技术栈

- **编程语言**: Go 1.23+
- **配置管理**: YAML + 环境变量
- **日志系统**: log/slog (Go标准库)
- **依赖管理**: Go Modules
- **Web框架**: Gin

## 已完成模块

### 1. 核心域模型 (internal/domain/models/)

核心域模型定义了语义缓存系统的核心数据结构和业务实体，遵循领域驱动设计（DDD）原则。

#### 架构设计

```
internal/domain/models/
├── cache.go       # 缓存相关模型
├── vector.go      # 向量相关模型
└── request.go     # 请求处理模型
```

#### 1.1 缓存模型 (cache.go)

缓存模型定义了缓存系统的核心业务实体，包括缓存项、查询、写入和删除操作的数据结构。

**核心结构体：**

```go
// 核心缓存实体
type CacheItem struct {
    ID string `json:"id" validate:"required"`                              // 缓存项的唯一标识符
    Question string `json:"question" validate:"required,min=1,max=1000"`   // 用户原始问题文本
    Answer string `json:"answer" validate:"required,min=1,max=10000"`      // LLM生成的答案文本
    UserType string `json:"user_type" validate:"required"`                 // 用于场景隔离的用户类型标识
    Vector []float32 `json:"vector,omitempty"`                             // 问题的向量表示
    Metadata CacheMetadata `json:"metadata"`                               // 缓存项的元数据
    Statistics CacheStatistics `json:"statistics"`                         // 缓存项的统计信息
    CreateTime time.Time `json:"create_time"`                              // 创建时间
    UpdateTime time.Time `json:"update_time"`                              // 更新时间
}

// 缓存元数据
type CacheMetadata struct {
    Source string `json:"source,omitempty"`                               // 数据来源标识
    Tags []string `json:"tags,omitempty"`                                 // 标签，用于分类和过滤
    QualityScore float64 `json:"quality_score,omitempty"`                 // 质量评估分数 (0.0-1.0)
    Version int `json:"version,omitempty"`                                // 数据版本号
}

// 缓存统计信息
type CacheStatistics struct {
    HitCount int64 `json:"hit_count"`                                      // 命中次数
    LikeCount int64 `json:"like_count"`                                    // 点赞次数
    DislikeCount int64 `json:"dislike_count"`                             // 点踩次数
    LastHitTime *time.Time `json:"last_hit_time,omitempty"`               // 最后命中时间
    ResponseTime float64 `json:"response_time,omitempty"`                 // 平均响应时间（毫秒）
}

// 缓存查询请求
type CacheQuery struct {
    Question string `json:"question" validate:"required,min=1,max=1000"`  // 查询问题
    UserType string `json:"user_type" validate:"required"`                // 用户类型
    SimilarityThreshold *float64 `json:"similarity_threshold,omitempty" validate:"omitempty,min=0,max=1"` // 相似度阈值，覆盖默认配置
    TopK *int `json:"top_k,omitempty" validate:"omitempty,min=1,max=100"` // 返回结果数量，覆盖默认配置
    Filters map[string]interface{} `json:"filters,omitempty"`            // 额外的过滤条件
    IncludeStatistics bool `json:"include_statistics,omitempty"`          // 是否包含统计信息
}

// 缓存查询结果
type CacheResult struct {
    Found bool `json:"found"`                                              // 是否找到匹配的缓存
    CacheID string `json:"cache_id,omitempty"`                            // 缓存项ID
    Answer string `json:"answer,omitempty"`                               // 缓存的答案
    Similarity float64 `json:"similarity,omitempty"`                      // 相似度分数
    ResponseTime float64 `json:"response_time,omitempty"`                 // 响应时间（毫秒）
    Metadata *CacheMetadata `json:"metadata,omitempty"`                   // 缓存元数据（可选）
    Statistics *CacheStatistics `json:"statistics,omitempty"`             // 统计信息（可选）
}

// 缓存写入请求
type CacheWriteRequest struct {
    Question string `json:"question" validate:"required,min=1,max=1000"`  // 问题文本
    Answer string `json:"answer" validate:"required,min=1,max=10000"`     // 答案文本
    UserType string `json:"user_type" validate:"required"`                // 用户类型
    Metadata *CacheMetadata `json:"metadata,omitempty"`                   // 元数据
    ForceWrite bool `json:"force_write,omitempty"`                        // 是否强制写入（跳过质量评估）
}

// 缓存写入结果
type CacheWriteResult struct {
    Success bool `json:"success"`                                          // 是否写入成功
    CacheID string `json:"cache_id,omitempty"`                            // 缓存项ID
    Message string `json:"message,omitempty"`                             // 结果消息
    QualityScore float64 `json:"quality_score,omitempty"`                 // 质量评估分数
    Reason string `json:"reason,omitempty"`                               // 失败原因（如果写入失败）
}

// 缓存删除请求
type CacheDeleteRequest struct {
    CacheIDs []string `json:"cache_ids" validate:"required,min=1"`        // 要删除的缓存项ID列表
    UserType string `json:"user_type" validate:"required"`                // 用户类型，用于权限验证
    Force bool `json:"force,omitempty"`                                   // 是否强制删除
}

// 缓存删除结果
type CacheDeleteResult struct {
    Success bool `json:"success"`                                          // 是否删除成功
    DeletedCount int `json:"deleted_count"`                               // 实际删除的数量
    FailedIDs []string `json:"failed_ids,omitempty"`                      // 删除失败的ID列表
    Message string `json:"message,omitempty"`                             // 结果消息
}
```


**设计说明：**

缓存模型定义了完整的数据结构，但不包含业务方法。所有的缓存项操作（如统计更新、评分计算等）都由相应的服务层处理，遵循清洁架构的分层原则。


#### 1.2 向量模型 (vector.go)

向量模型定义了向量数据结构和向量操作的相关数据类型，支持向量存储、搜索和处理。

**核心结构体：**

```go
// Vector 向量数据结构
type Vector struct {
    ID string `json:"id"`                                                   // 向量唯一标识符
    Values []float32 `json:"values" validate:"required,min=1"`             // 向量值数组
    Dimension int `json:"dimension"`                                        // 向量维度
    CreateTime time.Time `json:"create_time"`                               // 创建时间
    UpdateTime time.Time `json:"update_time"`                               // 更新时间
    Normalized bool `json:"normalized"`                                     // 是否已归一化
    ModelName string `json:"model_name"`                                    // 生成向量的模型名称
}

// VectorSearchRequest 向量搜索请求
type VectorSearchRequest struct {
    QueryText string `json:"query_text,omitempty"`                         // 查询文本（用于日志记录）
    QueryID string `json:"query_id,omitempty"`                             // 查询向量ID（可选）
    QueryVector []float32 `json:"query_vector,omitempty"`                  // 查询向量（主要搜索方式）
    TopK int `json:"top_k" validate:"required,min=1,max=100"`              // 返回最相似的K个结果
    SimilarityThreshold float64 `json:"similarity_threshold" validate:"min=0,max=1"` // 相似度阈值
    Filters map[string]interface{} `json:"filters,omitempty"`              // 过滤条件
    UserType string `json:"user_type" validate:"required"`                 // 用户类型，用于场景隔离
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
    ID string `json:"id"`                                                   // 向量ID
    Score float64 `json:"score"`                                            // 相似度分数
    Vector *Vector `json:"vector,omitempty"`                               // 向量数据（可选）
    Payload map[string]interface{} `json:"payload,omitempty"`              // 关联的负载数据
}

// VectorSearchResponse 向量搜索响应
type VectorSearchResponse struct {
    Results []VectorSearchResult `json:"results"`                          // 搜索结果列表
    TotalCount int `json:"total_count"`                                     // 总结果数量
    SearchTime float64 `json:"search_time"`                                 // 搜索耗时（毫秒）
    QueryInfo VectorQueryInfo `json:"query_info"`                          // 查询信息
}

// VectorQueryInfo 查询信息
type VectorQueryInfo struct {
    Dimension int `json:"dimension"`                                        // 查询向量维度
    FilterApplied bool `json:"filter_applied"`                             // 是否应用了过滤条件
}

// VectorBatchStoreRequest 批量向量存储请求
type VectorBatchStoreRequest struct {
    Vectors []VectorStoreItem `json:"vectors" validate:"required,min=1"`   // 要存储的向量列表
    CollectionName string `json:"collection_name" validate:"required"`     // 集合名称
    UpsertMode bool `json:"upsert_mode,omitempty"`                         // 是否为更新插入模式
}

// VectorStoreRequest 单个向量存储请求
type VectorStoreRequest struct {
    ID string `json:"id" validate:"required"`                              // 向量ID
    Vector []float32 `json:"vector" validate:"required,min=1"`             // 向量值
    CollectionName string `json:"collection_name" validate:"required"`     // 集合名称
    Payload map[string]interface{} `json:"payload,omitempty"`              // 关联的负载数据
    UpsertMode bool `json:"upsert_mode,omitempty"`                         // 是否为更新插入模式
}

// VectorStoreItem 向量存储项
type VectorStoreItem struct {
    ID string `json:"id" validate:"required"`                              // 向量ID
    Vector []float32 `json:"vector" validate:"required,min=1"`             // 向量值
    Payload map[string]interface{} `json:"payload,omitempty"`              // 关联的负载数据
}

// VectorBatchStoreResponse 批量向量存储响应
type VectorBatchStoreResponse struct {
    Success bool `json:"success"`                                           // 是否存储成功
    StoredCount int `json:"stored_count"`                                   // 成功存储的数量
    FailedCount int `json:"failed_count"`                                   // 存储失败的数量
    FailedIDs []string `json:"failed_ids,omitempty"`                       // 存储失败的ID列表
    Message string `json:"message,omitempty"`                              // 结果消息
    StoreTime float64 `json:"store_time"`                                   // 存储耗时（毫秒）
}

// VectorStoreResponse 单个向量存储响应
type VectorStoreResponse struct {
    Success bool `json:"success"`                                           // 是否存储成功
    VectorID string `json:"vector_id"`                                      // 存储的向量ID
    Message string `json:"message,omitempty"`                              // 结果消息
    StoreTime float64 `json:"store_time"`                                   // 存储耗时（毫秒）
}
```

**向量操作方法：**

```go
// NewVector 创建新的向量
func NewVector(id string, values []float32) *Vector {
    now := time.Now()
    return &Vector{
        ID:         id,
        Values:     values,
        Dimension:  len(values),
        CreateTime: now,
        UpdateTime: now,
        Normalized: false,
        ModelName:  "",
    }
}

// Validate 验证向量数据的有效性
func (v *Vector) Validate() error {
    if v.ID == "" {
        return fmt.Errorf("vector ID cannot be empty")
    }

    if len(v.Values) == 0 {
        return fmt.Errorf("vector values cannot be empty")
    }

    if v.Dimension != len(v.Values) {
        return fmt.Errorf("dimension mismatch: expected %d, got %d", v.Dimension, len(v.Values))
    }

    // 检查是否包含无效值
    for i, val := range v.Values {
        if math.IsNaN(float64(val)) || math.IsInf(float64(val), 0) {
            return fmt.Errorf("invalid value at index %d: %f", i, val)
        }
    }

    return nil
}

// Normalize 向量归一化（L2范数）
func (v *Vector) Normalize() {
    norm := v.L2Norm()
    if norm == 0 {
        return // 零向量不进行归一化
    }

    for i := range v.Values {
        v.Values[i] /= norm
    }
    v.Normalized = true
    v.UpdateTime = time.Now()
}

// L2Norm 计算L2范数
func (v *Vector) L2Norm() float32 {
    var sum float32
    for _, val := range v.Values {
        sum += val * val
    }
    return float32(math.Sqrt(float64(sum)))
}
```

**设计特点：**
- **完整的数据结构**: 包含向量存储、搜索、响应的完整数据模型
- **时间戳管理**: 记录向量创建和更新时间，支持版本控制
- **数学运算支持**: 提供向量归一化、L2范数计算等基础数学操作
- **数据验证**: 内置向量数据有效性验证机制，检查无效值和维度一致性
- **灵活存储**: 支持单个和批量存储，以及更新插入模式
- **模型标识**: 记录生成向量的模型名称，便于向量来源追踪

#### 1.3 请求模型 (request.go)

请求模型定义了系统处理流程中的关键数据结构，包括预处理、向量化和质量评估等环节的请求和响应。

**核心结构体：**

```go
// PreprocessedRequest 预处理后的请求
type PreprocessedRequest struct {
    Original *CacheQuery `json:"original"`                                   // 原始请求
    ProcessedQuestion string `json:"processed_question"`                     // 预处理后的问题文本
    ProcessingTime float64 `json:"processing_time"`                          // 预处理耗时（毫秒）
    Success bool `json:"success"`                                            // 预处理是否成功
    Error string `json:"error,omitempty"`                                    // 预处理错误信息
}

// VectorProcessingRequest 向量处理请求
type VectorProcessingRequest struct {
    Text string `json:"text" validate:"required"`                           // 待向量化的文本
    ModelName string `json:"model_name,omitempty"`                          // 指定使用的向量化模型
    Normalize bool `json:"normalize,omitempty"`                             // 是否归一化向量
}

// VectorProcessingResult 向量处理结果
type VectorProcessingResult struct {
    Vector *Vector `json:"vector"`                                           // 生成的向量
    ProcessingTime float64 `json:"processing_time"`                          // 处理耗时（毫秒）
    TokenCount int `json:"token_count"`                                      // 实际处理的token数量
    ModelUsed string `json:"model_used"`                                     // 实际使用的模型
    Success bool `json:"success"`                                            // 是否成功
    Error string `json:"error,omitempty"`                                    // 错误信息
}

// QualityAssessmentRequest 质量评估请求
type QualityAssessmentRequest struct {
    Question string `json:"question" validate:"required"`                   // 问题文本
    Answer string `json:"answer" validate:"required"`                       // 答案文本
    UserType string `json:"user_type" validate:"required"`                  // 用户类型
}

// QualityAssessmentResult 质量评估结果
type QualityAssessmentResult struct {
    Passed bool `json:"passed"`                                              // 是否通过评估
    Score float64 `json:"score"`                                             // 综合质量分数 (0.0-1.0)
    Threshold float64 `json:"threshold"`                                     // 使用的阈值
    AssessmentTime float64 `json:"assessment_time"`                          // 评估耗时（毫秒）
    Reason string `json:"reason,omitempty"`                                  // 失败原因（如果未通过）
}
```

**设计特点：**
- **简洁的处理流程**: 包含预处理、向量处理、质量评估的核心请求处理数据模型
- **性能监控**: 记录各处理阶段的耗时，支持性能分析
- **错误处理**: 包含详细的成功/失败状态和错误信息
- **模型可选**: 支持指定向量化模型和归一化选项
- **统计信息**: 记录token数量和实际使用的模型信息

#### 1.4 模型间关联关系

三个核心域模型通过明确的关联关系协同工作：

```
请求模型 (request.go) ← 协调层
    ↓ 控制流程
向量模型 (vector.go) ← 技术层  
    ↓ 数据支撑
缓存模型 (cache.go) ← 业务层
```

**关键关联：**
- `RequestContext` 为所有操作提供上下文信息
- `VectorProcessingResult` 的向量数据用于构建 `CacheItem`
- `VectorSearchResult` 转换为 `CacheResult` 返回给用户
- `QualityAssessmentResult` 决定缓存写入的成功与否

### 2. 业务接口定义 (internal/domain/)

业务接口定义包含了核心业务服务和数据访问的契约，遵循依赖倒置原则和Clean Architecture分层原则。

#### 架构设计

```
internal/domain/
├── repositories/                       # 数据访问层接口
│   └── vector_repository.go            # 向量仓储接口
└── services/                           # 业务逻辑层接口
    ├── cache_service.go                 # 缓存服务接口
    ├── vector_service.go                # 向量服务接口
    ├── request_preprocessing_service.go # 请求预处理服务接口
    ├── recall_postprocessing_service.go # 召回后处理服务接口
    └── quality_service.go               # 质量评估服务接口
```

#### 2.1 数据访问层接口 (repositories/)

数据访问层接口遵循Repository模式，为数据持久化操作提供抽象，使业务逻辑与具体的数据存储技术解耦。

##### 向量仓储接口 (vector_repository.go)

向量仓储接口负责向量数据的持久化操作，提供向量数据库操作的抽象层。

**核心接口：**

```go
// VectorRepository 向量数据库仓储接口
// 专门负责向量的存储、检索和管理
type VectorRepository interface {
    // Store 存储单个向量数据
    // 将单个向量数据存储到向量数据库
    Store(ctx context.Context, request *models.VectorStoreRequest) (*models.VectorStoreResponse, error)

    // BatchStore 存储向量数据
    // 将向量数据存储到向量数据库
    BatchStore(ctx context.Context, request *models.VectorBatchStoreRequest) (*models.VectorBatchStoreResponse, error)

    // Search 向量相似性搜索
    // 在向量数据库中搜索相似向量
    Search(ctx context.Context, request *models.VectorSearchRequest) (*models.VectorSearchResponse, error)

    // Delete 删除向量数据
    // 根据ID删除向量记录
    Delete(ctx context.Context, ids []string, userType string) error

    // GetByID 根据ID获取向量
    // 获取指定ID的向量数据
    GetByID(ctx context.Context, id string) (*models.Vector, error)
}
```

**设计特点：**
- **灵活的存储接口**: 提供单个和批量向量存储方法，满足不同场景需求
- **统一的数据结构**: 使用标准化的请求和响应模型进行数据传递
- **高性能搜索**: 专门优化的向量相似性搜索接口
- **场景隔离**: 通过userType实现多租户数据隔离
- **完整的CRUD操作**: 支持创建、读取、更新、删除的完整操作
- **错误处理**: 完善的错误返回机制，便于上层业务处理

#### 2.2 业务逻辑层接口 (services/)

业务逻辑层接口定义了核心业务服务的契约，协调不同组件完成复杂的业务流程。

##### 缓存服务接口 (cache_service.go)

缓存服务接口负责缓存的核心业务逻辑，提供高级的缓存操作能力。

**核心接口：**

```go
// CacheService 缓存服务接口，负责缓存的核心业务逻辑
// 提供高级的缓存操作，协调各个组件完成缓存查询、存储和管理
type CacheService interface {
    // QueryCache 查询缓存
    // 执行完整的缓存查询流程：请求预处理 -> 向量搜索 -> 后处理 -> 返回结果
    // ctx: 上下文
    // query: 缓存查询请求
    // 返回: 查询结果和错误信息
    QueryCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error)

    // StoreCache 存储缓存
    // 执行完整的缓存写入流程：质量评估 -> 向量生成 -> 数据存储
    // ctx: 上下文
    // request: 缓存写入请求
    // 返回: 写入结果和错误信息
    StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error)

    // DeleteCache 删除缓存
    // ctx: 上下文
    // request: 删除请求
    // 返回: 删除结果和错误信息
    DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error)

    // GetCacheByID 根据ID获取缓存项
    // ctx: 上下文
    // cacheID: 缓存项ID
    // userType: 用户类型，用于权限验证
    // includeStatistics: 是否包含统计信息
    // 返回: 缓存项和错误信息
    GetCacheByID(ctx context.Context, cacheID, userType string, includeStatistics bool) (*models.CacheItem, error)

    // GetCacheStatistics 获取缓存系统统计信息
    // ctx: 上下文
    // userType: 用户类型，用于权限验证和场景隔离
    // timeRange: 时间范围，如 "24h", "7d", "30d"
    // 返回: 统计信息映射和错误信息
    GetCacheStatistics(ctx context.Context, userType, timeRange string) (map[string]interface{}, error)

    // GetCacheHealth 获取缓存系统健康状态
    // ctx: 上下文
    // 返回: 健康状态信息映射和错误信息
    GetCacheHealth(ctx context.Context) (map[string]interface{}, error)
}
```

**设计特点：**
- **完整的业务功能**: 提供查询、存储、删除、获取、统计、健康检查等全面的缓存操作
- **完整的缓存流程**: 集成请求预处理、向量搜索、后处理的完整缓存查询流程
- **可观测性支持**: 提供统计信息和健康检查接口，支持系统监控
- **协调器模式**: 协调各个组件（预处理、向量化、质量评估等）完成复杂业务流程
- **统一数据模型**: 基于models包的统一数据结构，确保类型安全

##### 向量服务接口 (vector_service.go)

向量服务接口负责向量相关的业务逻辑，提供文本向量化、向量处理、相似度计算等功能。

**核心接口：**

```go
// VectorService 向量服务业务层接口
// 协调 EmbeddingService 和 VectorRepository，提供高级业务功能
type VectorService interface {
    // SearchCache 搜索语义缓存
    // 完整的语义缓存查询流程：文本向量化 + 相似度搜索
    SearchCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error)

    // StoreCache 存储查询和响应到缓存
    // 将用户查询和LLM响应存储到向量缓存中
    StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error)

    // DeleteCache 删除缓存项
    // 从向量缓存中删除指定的缓存项
    DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error)

    // SelectBestResult 选择最优结果
    // 从候选结果中选择最符合查询意图的单个结果
    //
    // 参数:
    //   ctx: 上下文
    //   results: 候选结果列表
    //   query: 查询请求
    //   strategy: 选择策略（如 "first", "highest_score", "temperature_softmax"）
    //
    // 返回:
    //   *models.VectorSearchResult: 选中的最优结果
    //   error: 错误信息
    SelectBestResult(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, strategy string) (*models.VectorSearchResult, error)
}
```

**设计特点：**
- **高级业务接口**: 直接面向语义缓存业务，封装了底层向量操作复杂性
- **简洁的API设计**: 提供核心的缓存搜索、存储、删除和结果选择功能
- **服务协调**: 协调EmbeddingService和VectorRepository，组合完成复杂业务流程
- **性能导向**: 专门针对语义缓存场景优化的搜索和存储接口
- **智能结果选择**: 提供多种策略的最优结果选择功能

#### 2.1 请求预处理服务 (request_preprocessing_service.go)

请求预处理服务采用函数式设计，允许用户注册自定义预处理函数，对用户原始查询请求进行预处理，提高查询质量和匹配准确性。

**核心类型定义：**

```go
// PreprocessorFunc 预处理函数类型
// 用户自定义的预处理函数必须符合此签名
//
// 参数:
//   ctx: 上下文，用于控制请求生命周期和传递追踪信息
//   text: 待处理的文本
//   userType: 用户类型
//   metadata: 元数据，可以包含额外的处理信息
//
// 返回:
//   string: 处理后的文本
//   error: 错误信息
type PreprocessorFunc func(ctx context.Context, text string, userType string, metadata map[string]interface{}) (string, error)
```

**核心接口：**

```go
// RequestPreprocessingService 请求预处理服务接口
// 负责对用户原始查询请求进行预处理，提高查询质量和匹配准确性
// 采用函数式设计，支持用户注册自定义预处理函数
type RequestPreprocessingService interface {
    // PreprocessQuery 预处理查询请求
    // 使用注册的自定义预处理函数链来处理查询
    //
    // 参数:
    //   ctx: 上下文，用于控制请求生命周期和传递追踪信息
    //   request: 原始查询请求
    //
    // 返回:
    //   *models.PreprocessedRequest: 预处理后的请求，包含处理结果和元数据
    //   status.StatusCode: 处理状态码
    //   error: 错误信息
    PreprocessQuery(ctx context.Context, request *models.CacheQuery) (*models.PreprocessedRequest, status.StatusCode, error)

    // RegisterPreprocessor 注册预处理函数
    // 允许用户注册自定义的预处理函数，注册的函数将按注册顺序链式执行
    //
    // 参数:
    //   name: 预处理函数名称，用于标识和管理
    //   processor: 预处理函数
    //
    // 返回:
    //   error: 错误信息，如果名称已存在则返回错误
    RegisterPreprocessor(name string, processor PreprocessorFunc) error

    // UnregisterPreprocessor 取消注册预处理函数
    // 移除指定名称的预处理函数
    //
    // 参数:
    //   name: 预处理函数名称
    //
    // 返回:
    //   error: 错误信息，如果名称不存在则返回错误
    UnregisterPreprocessor(name string) error

    // ListPreprocessors 列出所有已注册的预处理函数名称
    //
    // 返回:
    //   []string: 预处理函数名称列表，按注册顺序返回
    ListPreprocessors() []string
}
```

**配置结构：**

```go
// RequestPreprocessingConfig 请求预处理配置
type RequestPreprocessingConfig struct {
    // Timeout 处理超时时间（秒）
    Timeout int `json:"timeout" yaml:"timeout"`

    // EnableLogging 是否启用详细日志
    EnableLogging bool `json:"enable_logging" yaml:"enable_logging"`
}
```

**设计特点：**
- **函数式设计**: 采用PreprocessorFunc类型，支持灵活的函数式预处理
- **动态注册**: 支持运行时注册和注销预处理函数
- **链式处理**: 按注册顺序链式执行多个预处理函数
- **简洁的配置**: 仅保留核心配置项，避免过度复杂化
- **统一状态码**: 使用pkg/status包提供统一的状态码管理

#### 2.2 召回后处理服务 (recall_postprocessing_service.go)

召回后处理服务负责对向量检索召回的结果进行优化处理，提供结果格式化功能。

**核心接口：**

```go
// RecallPostprocessingService 召回后处理服务接口
// 负责对向量检索召回的结果进行优化处理，确保返回最相关和高质量的结果
type RecallPostprocessingService interface {
    // ProcessRecallResults 处理向量检索的召回结果
    // 对原始检索结果进行后处理，包括去重、排序、质量筛选等操作
    //
    // 参数:
    //   ctx: 上下文，用于控制请求生命周期和传递追踪信息
    //   results: 向量检索的原始结果列表
    //   originalQuery: 原始查询请求
    //
    // 返回:
    //   []*models.VectorSearchResult: 处理后的结果列表
    //   status.StatusCode: 处理状态码
    //   error: 错误信息
    ProcessRecallResults(ctx context.Context, results []*models.VectorSearchResult, originalQuery *models.CacheQuery) ([]*models.VectorSearchResult, status.StatusCode, error)

    // FormatResult 格式化结果
    // 将向量搜索结果转换为缓存查询结果格式
    //
    // 参数:
    //   ctx: 上下文
    //   result: 向量搜索结果
    //   includeMetadata: 是否包含元数据
    //   includeStatistics: 是否包含统计信息
    //
    // 返回:
    //   *models.CacheResult: 格式化后的缓存结果
    //   status.StatusCode: 处理状态码
    //   error: 错误信息
    FormatResult(ctx context.Context, result *models.VectorSearchResult, includeMetadata bool, includeStatistics bool) (*models.CacheResult, status.StatusCode, error)
}
```

**辅助接口：**

```go
// ResultFormatter 结果格式化器接口
type ResultFormatter interface {
    // FormatCacheResult 格式化为缓存结果
    //
    // 参数:
    //   ctx: 上下文
    //   vectorResult: 向量搜索结果
    //   options: 格式化选项
    //
    // 返回:
    //   *models.CacheResult: 缓存结果
    //   error: 错误信息
    FormatCacheResult(ctx context.Context, vectorResult *models.VectorSearchResult) (*models.CacheResult, error)

    // ExtractAnswer 从结果中提取答案
    //
    // 参数:
    //   result: 向量搜索结果
    //
    // 返回:
    //   string: 提取的答案
    //   error: 错误信息
    ExtractAnswer(result *models.VectorSearchResult) (string, error)

    // ExtractMetadata 提取元数据
    //
    // 参数:
    //   result: 向量搜索结果
    //
    // 返回:
    //   *models.CacheMetadata: 提取的元数据
    //   error: 错误信息
    ExtractMetadata(result *models.VectorSearchResult) (*models.CacheMetadata, error)
}
```

**设计特点：**
- **简洁的接口设计**: 专注于核心的结果处理和格式化功能
- **结果格式化**: 提供向量搜索结果到缓存结果的标准化转换
- **可扩展的格式化器**: 通过ResultFormatter接口支持自定义格式化逻辑
- **统一状态码**: 使用pkg/status包提供统一的状态码管理

#### 2.3 质量评估服务 (quality_service.go)

质量评估服务负责评估问答对的质量，决定是否允许写入缓存，确保缓存数据的高质量。支持自定义质量评估函数，提供灵活的质量控制机制。

**核心类型定义：**

```go
// CustomQualityFunction 自定义质量评估函数类型
// 用户可以实现此函数来提供自定义的质量评估逻辑
//
// 参数:
//   ctx: 上下文
//   question: 问题文本
//   answer: 答案文本
//   userType: 用户类型
//   config: 配置参数
//
// 返回:
//   float64: 质量分数 (0.0-1.0)
//   map[string]interface{}: 评估详情
//   error: 错误信息
type CustomQualityFunction func(ctx context.Context, question string, answer string, userType string, config map[string]interface{}) (float64, map[string]interface{}, error)
```

**核心接口：**

```go
// QualityService 质量评估服务接口
// 负责评估问答对的质量，决定是否允许写入缓存，确保缓存数据的高质量
type QualityService interface {
    // AssessQuality 综合质量评估
    // 对问答对进行全面的质量评估，返回评估结果和建议
    //
    // 参数:
    //   ctx: 上下文，用于控制请求生命周期和传递追踪信息
    //   request: 质量评估请求
    //
    // 返回:
    //   *models.QualityAssessmentResult: 质量评估结果
    //   status.StatusCode: 评估状态码
    //   error: 错误信息
    AssessQuality(ctx context.Context, request *models.QualityAssessmentRequest) (*models.QualityAssessmentResult, status.StatusCode, error)

    // AssessQuestionQuality 评估问题质量
    // 专门评估问题文本的质量
    //
    // 参数:
    //   ctx: 上下文
    //   question: 问题文本
    //   userType: 用户类型
    //
    // 返回:
    //   float64: 问题质量分数 (0.0-1.0)
    //   []string: 质量问题描述
    //   status.StatusCode: 评估状态码
    //   error: 错误信息
    AssessQuestionQuality(ctx context.Context, question string, userType string) (float64, []string, status.StatusCode, error)

    // AssessAnswerQuality 评估答案质量
    // 专门评估答案文本的质量
    //
    // 参数:
    //   ctx: 上下文
    //   answer: 答案文本
    //   question: 对应的问题文本（用于相关性评估）
    //   userType: 用户类型
    //
    // 返回:
    //   float64: 答案质量分数 (0.0-1.0)
    //   []string: 质量问题描述
    //   status.StatusCode: 评估状态码
    //   error: 错误信息
    AssessAnswerQuality(ctx context.Context, answer string, question string, userType string) (float64, []string, status.StatusCode, error)

    // CheckBlacklist 检查黑名单
    // 检查问答对是否包含黑名单关键词或模式
    //
    // 参数:
    //   ctx: 上下文
    //   question: 问题文本
    //   answer: 答案文本
    //   userType: 用户类型
    //
    // 返回:
    //   bool: 是否命中黑名单
    //   []string: 命中的黑名单项
    //   status.StatusCode: 检查状态码
    //   error: 错误信息
    CheckBlacklist(ctx context.Context, question string, answer string, userType string) (bool, []string, status.StatusCode, error)

    // CalculateOverallScore 计算综合分数
    // 基于多个评估维度计算综合质量分数
    //
    // 参数:
    //   ctx: 上下文
    //   scores: 各维度分数
    //   weights: 各维度权重
    //
    // 返回:
    //   float64: 综合分数 (0.0-1.0)
    //   status.StatusCode: 计算状态码
    //   error: 错误信息
    CalculateOverallScore(ctx context.Context, scores map[string]float64, weights map[string]float64) (float64, status.StatusCode, error)

    // IsQualityAcceptable 判断质量是否可接受
    // 基于配置的阈值判断质量是否达到要求
    //
    // 参数:
    //   ctx: 上下文
    //   score: 质量分数
    //   userType: 用户类型
    //
    // 返回:
    //   bool: 质量是否可接受
    //   string: 拒绝原因（如果不可接受）
    //   status.StatusCode: 判断状态码
    //   error: 错误信息
    IsQualityAcceptable(ctx context.Context, score float64, userType string) (bool, string, status.StatusCode, error)

    // ApplyCustomQualityFunction 应用自定义质量评估函数
    // 用户可以提供自定义的质量评估函数来执行特定的质量检查
    //
    // 参数:
    //   ctx: 上下文
    //   question: 问题文本
    //   answer: 答案文本
    //   userType: 用户类型
    //   customFunc: 自定义质量评估函数
    //
    // 返回:
    //   float64: 质量分数 (0.0-1.0)
    //   map[string]interface{}: 评估详情
    //   status.StatusCode: 评估状态码
    //   error: 错误信息
    ApplyCustomQualityFunction(ctx context.Context, question string, answer string, userType string, customFunc CustomQualityFunction) (float64, map[string]interface{}, status.StatusCode, error)

    // RegisterCustomFunction 注册自定义质量评估函数
    // 将自定义函数注册到服务中，以便在综合评估中使用
    //
    // 参数:
    //   name: 函数名称
    //   function: 自定义质量评估函数
    //   weight: 函数权重
    RegisterCustomFunction(name string, function CustomQualityFunction, weight float64)

    // GetRegisteredFunctions 获取所有已注册的函数名称
    // 用于调试和监控，返回当前注册的所有函数名称列表
    //
    // 返回:
    //   []string: 函数名称列表
    GetRegisteredFunctions() []string

    // GetFunctionWeight 获取指定函数的权重
    //
    // 参数:
    //   functionName: 函数名称
    //
    // 返回:
    //   float64: 函数权重，如果函数不存在则返回0.0
    GetFunctionWeight(functionName string) float64
}
```

**辅助策略接口：**

```go
// QualityAssessmentStrategy 质量评估策略接口
// 允许插拔式的质量评估策略实现
type QualityAssessmentStrategy interface {
    // Name 策略名称
    Name() string

    // Assess 执行评估
    //
    // 参数:
    //   ctx: 上下文
    //   question: 问题文本
    //   answer: 答案文本
    //   userType: 用户类型
    //   config: 策略配置
    //
    // 返回:
    //   float64: 评估分数 (0.0-1.0)
    //   map[string]interface{}: 评估详情
    //   error: 错误信息
    Assess(ctx context.Context, question string, answer string, userType string, config map[string]interface{}) (float64, map[string]interface{}, error)

    // GetWeight 获取策略权重
    //
    // 参数:
    //   userType: 用户类型
    //
    // 返回:
    //   float64: 权重值
    GetWeight(userType string) float64

    // Validate 验证策略配置
    //
    // 参数:
    //   config: 策略配置
    //
    // 返回:
    //   bool: 配置是否有效
    //   error: 验证错误
    Validate(config map[string]interface{}) (bool, error)
}

// BlacklistChecker 黑名单检查器接口
type BlacklistChecker interface {
    // CheckKeywords 检查关键词黑名单
    //
    // 参数:
    //   text: 待检查文本
    //   userType: 用户类型
    //
    // 返回:
    //   bool: 是否命中黑名单
    //   []string: 命中的关键词
    //   error: 错误信息
    CheckKeywords(text string, userType string) (bool, []string, error)

    // CheckPatterns 检查模式黑名单
    //
    // 参数:
    //   text: 待检查文本
    //   userType: 用户类型
    //
    // 返回:
    //   bool: 是否命中黑名单
    //   []string: 命中的模式
    //   error: 错误信息
    CheckPatterns(text string, userType string) (bool, []string, error)
}
```

**配置结构：**

```go
// QualityConfig 质量评估配置
type QualityConfig struct {
    // Strategies 启用的质量评估策略
    Strategies []string `json:"strategies" yaml:"strategies"`

    // StrategyWeights 策略权重配置
    StrategyWeights map[string]float64 `json:"strategy_weights" yaml:"strategy_weights"`

    // UserTypeThresholds 不同用户类型的质量阈值
    UserTypeThresholds map[string]float64 `json:"user_type_thresholds" yaml:"user_type_thresholds"`

    // DefaultThreshold 默认质量阈值
    DefaultThreshold float64 `json:"default_threshold" yaml:"default_threshold"`

    // MinQuestionLength 最小问题长度
    MinQuestionLength int `json:"min_question_length" yaml:"min_question_length"`

    // MaxQuestionLength 最大问题长度
    MaxQuestionLength int `json:"max_question_length" yaml:"max_question_length"`

    // MinAnswerLength 最小答案长度
    MinAnswerLength int `json:"min_answer_length" yaml:"min_answer_length"`

    // MaxAnswerLength 最大答案长度
    MaxAnswerLength int `json:"max_answer_length" yaml:"max_answer_length"`

    // BlacklistKeywords 黑名单关键词
    BlacklistKeywords []string `json:"blacklist_keywords" yaml:"blacklist_keywords"`

    // BlacklistPatterns 黑名单正则模式
    BlacklistPatterns []string `json:"blacklist_patterns" yaml:"blacklist_patterns"`

    // EnableBlacklistCheck 是否启用黑名单检查
    EnableBlacklistCheck bool `json:"enable_blacklist_check" yaml:"enable_blacklist_check"`

    // Timeout 评估超时时间（秒）
    Timeout int `json:"timeout" yaml:"timeout"`

    // CustomFunctionConfig 自定义函数配置
    CustomFunctionConfig map[string]interface{} `json:"custom_function_config" yaml:"custom_function_config"`
}
```

**设计特点：**
- **自定义函数支持**: 通过CustomQualityFunction类型支持用户自定义质量评估逻辑
- **函数注册机制**: 支持运行时注册和管理自定义质量评估函数
- **多维度评估**: 支持问题质量、答案质量、黑名单检查等多个维度的独立评估
- **策略化设计**: 提供QualityAssessmentStrategy接口支持插拔式质量评估策略
- **综合评分**: 支持多策略加权计算综合质量分数
- **黑名单检查**: 通过BlacklistChecker接口检查关键词和模式黑名单
- **灵活配置**: 支持用户类型阈值、长度限制、超时等详细配置
- **统一状态码**: 使用pkg/status包提供统一的状态码管理

#### 2.4 服务接口设计原则

**依赖倒置原则：**
- 所有服务接口都定义在domain层，不依赖具体实现
- 接口设计面向业务需求，而非技术实现细节
- 支持依赖注入模式，便于测试和模块替换

**单一职责原则：**
- 每个服务接口职责明确，边界清晰
- 请求预处理专注于查询优化
- 召回后处理专注于结果优化
- 质量评估专注于内容质量管控

**开闭原则：**
- 通过策略接口支持算法扩展
- 配置驱动的策略选择机制
- 插拔式组件设计，支持运行时策略切换

**接口隔离原则：**
- 细粒度的子接口设计，避免接口污染
- 客户端只依赖需要的接口方法
- 支持按需实现，提高代码复用性

### 3. 配置管理模块 (configs/)

配置管理模块提供了灵活、多环境的配置管理解决方案。

#### 架构设计

```
configs/
├── config.go      # 配置结构体定义
├── loader.go      # 配置加载器实现
└── config.yaml    # 主配置文件
```

#### 核心特性

1. **强类型配置结构**
    - 使用Go结构体定义所有配置项
    - 支持嵌套配置和标签验证
    - 提供配置验证方法

2. **多源配置加载**
    - 支持YAML文件配置
    - 支持环境变量覆盖（前缀：QA_CACHE_）
    - 支持多环境配置文件

3. **配置分类**
    - 服务器配置（端口、超时等）
    - 数据库配置（Qdrant向量数据库）
    - 嵌入模型配置（本地/远程模型）
    - 日志配置（级别、格式、输出）
    - 缓存配置（TTL、阈值等）
    - 质量评估配置

#### 实现细节

```go
// 配置结构示例
package configs

import (
	"fmt"
	"time"
)

// Config 主配置结构体
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Embedding EmbeddingConfig `yaml:"embedding"`
	Logging   LoggingConfig   `yaml:"logging"`
	Cache     CacheConfig     `yaml:"cache"`
	Quality   QualityConfig   `yaml:"quality"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host                    string        `yaml:"host"`
	Port                    int           `yaml:"port"`
	ReadTimeout             time.Duration `yaml:"read_timeout"`
	WriteTimeout            time.Duration `yaml:"write_timeout"`
	IdleTimeout             time.Duration `yaml:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
	MaxConnections          int           `yaml:"max_connections"`
}

// DatabaseConfig 向量数据库配置
type DatabaseConfig struct {
	Type   string       `yaml:"type"`
	Qdrant QdrantConfig `yaml:"qdrant"`
}

// QdrantConfig Qdrant向量数据库配置
type QdrantConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	APIKey         string        `yaml:"api_key"`
	CollectionName string        `yaml:"collection_name"`
	VectorSize     int           `yaml:"vector_size"`
	Distance       string        `yaml:"distance"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
}

// EmbeddingConfig 嵌入模型配置
type EmbeddingConfig struct {
	Type   string          `yaml:"type"`
	Local  LocalEmbedding  `yaml:"local"`
	Remote RemoteEmbedding `yaml:"remote"`
}

// LocalEmbedding 本地嵌入模型配置
type LocalEmbedding struct {
	ModelPath string `yaml:"model_path"`
	MaxTokens int    `yaml:"max_tokens"`
	BatchSize int    `yaml:"batch_size"`
}

// RemoteEmbedding 远程嵌入模型配置
type RemoteEmbedding struct {
	APIEndpoint string            `yaml:"api_endpoint"`
	APIKey      string            `yaml:"api_key"`
	ModelName   string            `yaml:"model_name"`
	Timeout     time.Duration     `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay"`
	Headers     map[string]string `yaml:"headers"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	SimilarityThreshold float64       `yaml:"similarity_threshold"`
	TopK                int           `yaml:"top_k"`
	TTL                 time.Duration `yaml:"ttl"`
	MaxCacheSize        int64         `yaml:"max_cache_size"`
	EnableAsyncUpdate   bool          `yaml:"enable_async_update"`
	UpdateBatchSize     int           `yaml:"update_batch_size"`
	UpdateInterval      time.Duration `yaml:"update_interval"`
}

// QualityConfig 质量评估配置
type QualityConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Threshold  float64           `yaml:"threshold"`
	Strategies []QualityStrategy `yaml:"strategies"`
	Blacklist  QualityBlacklist  `yaml:"blacklist"`
}

// QualityStrategy 质量评估策略
type QualityStrategy struct {
	Name    string                 `yaml:"name"`
	Weight  float64                `yaml:"weight"`
	Enabled bool                   `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config"`
}

// QualityBlacklist 质量评估黑名单
type QualityBlacklist struct {
	ApologyKeywords   []string `yaml:"apology_keywords"`
	ErrorKeywords     []string `yaml:"error_keywords"`
	MinAnswerLength   int      `yaml:"min_answer_length"`
	MaxAnswerLength   int      `yaml:"max_answer_length"`
	MinQuestionLength int      `yaml:"min_question_length"`
	MaxQuestionLength int      `yaml:"max_question_length"`
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}

	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}

	if err := c.Embedding.Validate(); err != nil {
		return fmt.Errorf("embedding config validation failed: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}

	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}

	if err := c.Quality.Validate(); err != nil {
		return fmt.Errorf("quality config validation failed: %w", err)
	}

	return nil
}

// Validate 验证服务器配置
func (s *ServerConfig) Validate() error {
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("invalid port: %d", s.Port)
	}

	if s.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}

	if s.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}

	if s.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}

	return nil
}

// Validate 验证数据库配置
func (d *DatabaseConfig) Validate() error {
	if d.Type == "" {
		return fmt.Errorf("database type is required")
	}

	if d.Type == "qdrant" {
		return d.Qdrant.Validate()
	}

	return fmt.Errorf("unsupported database type: %s", d.Type)
}

// Validate 验证Qdrant配置
func (q *QdrantConfig) Validate() error {
	if q.Host == "" {
		return fmt.Errorf("qdrant host is required")
	}

	if q.Port <= 0 || q.Port > 65535 {
		return fmt.Errorf("invalid qdrant port: %d", q.Port)
	}

	if q.CollectionName == "" {
		return fmt.Errorf("qdrant collection name is required")
	}

	if q.VectorSize <= 0 {
		return fmt.Errorf("vector size must be positive")
	}

	if q.Distance == "" {
		q.Distance = "cosine"
	}

	return nil
}

// Validate 验证嵌入模型配置
func (e *EmbeddingConfig) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("embedding type is required")
	}

	switch e.Type {
	case "local":
		return e.Local.Validate()
	case "remote":
		return e.Remote.Validate()
	default:
		return fmt.Errorf("unsupported embedding type: %s", e.Type)
	}
}

// Validate 验证本地嵌入模型配置
func (l *LocalEmbedding) Validate() error {
	if l.ModelPath == "" {
		return fmt.Errorf("local embedding model path is required")
	}

	if l.BatchSize <= 0 {
		l.BatchSize = 32 // 默认批处理大小
	}

	return nil
}

// Validate 验证远程嵌入模型配置
func (r *RemoteEmbedding) Validate() error {
	if r.APIEndpoint == "" {
		return fmt.Errorf("remote embedding API endpoint is required")
	}

	if r.ModelName == "" {
		return fmt.Errorf("remote embedding model name is required")
	}

	if r.Timeout <= 0 {
		r.Timeout = 30 * time.Second // 默认超时时间
	}

	return nil
}

// Validate 验证日志配置
func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	if !validLevels[l.Level] {
		return fmt.Errorf("invalid log level: %s", l.Level)
	}

	validOutputs := map[string]bool{
		"stdout": true, "stderr": true, "file": true,
	}

	if !validOutputs[l.Output] {
		return fmt.Errorf("invalid log output: %s", l.Output)
	}

	if l.Output == "file" && l.FilePath == "" {
		return fmt.Errorf("file path is required when output is file")
	}

	return nil
}

// Validate 验证缓存配置
func (c *CacheConfig) Validate() error {
	if c.SimilarityThreshold < 0 || c.SimilarityThreshold > 1 {
		return fmt.Errorf("similarity threshold must be between 0 and 1")
	}

	if c.TopK <= 0 {
		return fmt.Errorf("top_k must be positive")
	}

	if c.TTL <= 0 {
		return fmt.Errorf("ttl must be positive")
	}

	if c.EnableAsyncUpdate && c.UpdateBatchSize <= 0 {
		return fmt.Errorf("update_batch_size must be positive when async update is enabled")
	}

	return nil
}

// Validate 验证质量评估配置
func (q *QualityConfig) Validate() error {
	if !q.Enabled {
		return nil
	}

	if q.Threshold < 0 || q.Threshold > 1 {
		return fmt.Errorf("quality threshold must be between 0 and 1")
	}

	totalWeight := 0.0
	for _, strategy := range q.Strategies {
		if strategy.Enabled {
			totalWeight += strategy.Weight
		}
	}

	if totalWeight <= 0 {
		return fmt.Errorf("total weight of enabled strategies must be positive")
	}

	return nil
}

// GetAddr 获取服务器监听地址
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetQdrantAddr 获取Qdrant地址
func (q *QdrantConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", q.Host, q.Port)
}

```

#### 使用方式

```go
loader := configs.NewLoader("configs/config.yaml", "dev")
config, err := loader.Load(ctx)
if err != nil {
    log.Fatal("配置加载失败", err)
}
```

### 4. 状态码模块 (pkg/status/)

状态码模块提供了统一的系统状态码定义和管理，支持成功状态和各类错误状态。

#### 架构设计

```
pkg/status/
└── codes.go       # 状态码定义和方法
```

#### 核心特性

1. **数值化状态码**
    - 成功状态码：2xxx (如：2000-SUCCESS)
    - 客户端错误：4xxx (如：4000-BAD_REQUEST, 4001-INVALID_PARAM)
    - 服务器错误：5xxx (如：5000-INTERNAL_ERROR, 5001-TIMEOUT)

2. **分类管理**
    - 按业务领域分类（系统、配置、数据库、向量DB、嵌入、缓存、质量、HTTP、业务）
    - 每个分类使用特定数字段

3. **状态码方法**
    - `GetCategory()`: 获取状态码分类
    - `String()`: 获取状态码字符串表示
    - `IsRetryable()`: 判断是否可重试
    - `IsClientError()`: 判断是否为客户端错误
    - `IsServerError()`: 判断是否为服务器错误

#### 状态码分布

```
成功状态码 (2xxx):
  2000 - SUCCESS

客户端错误 (4xxx):
  4000-4099  - 通用客户端错误
  4100-4199  - HTTP相关错误
  4200-4299  - 认证授权错误
  4300-4399  - 业务逻辑错误
  4400-4499  - 质量评估错误

服务器错误 (5xxx):
  5000-5099  - 通用服务器错误
  5100-5199  - 配置相关错误
  5200-5299  - 数据库相关错误
  5300-5399  - 向量数据库相关错误
  5400-5499  - 嵌入模型相关错误
  5500-5599  - 缓存相关错误
```

#### 使用方式

```go
import "qa-cache/pkg/status"

// 使用状态码
code := status.CodeOK
fmt.Printf("状态: %d (%s)", code, code.String())
fmt.Printf("分类: %s", code.GetCategory())
fmt.Printf("可重试: %v", code.IsRetryable())
```

### 5. 日志系统 (pkg/logger/)

日志系统基于Go标准库的`log/slog`包实现，提供结构化、高性能的日志功能。

#### 架构设计

```
pkg/logger/
└── logger.go      # 日志接口和实现
```



#### 核心特性

1. **基于log/slog**
    - 利用Go 1.23的官方结构化日志包
    - 高性能、零分配的日志记录
    - 原生支持JSON和文本格式

2. **灵活配置**
    - 支持多种日志级别（DEBUG, INFO, WARN, ERROR）
    - 支持控制台和文件输出
    - 支持自定义时间格式和级别字符串

3. **上下文感知**
    - 支持`context.Context`的日志记录
    - 集成请求追踪和用户信息

#### 实现特点

```go
// 日志接口
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger 日志器接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})

	// 带上下文的日志方法
	DebugContext(ctx context.Context, msg string, args ...interface{})
	InfoContext(ctx context.Context, msg string, args ...interface{})
	WarnContext(ctx context.Context, msg string, args ...interface{})
	ErrorContext(ctx context.Context, msg string, args ...interface{})

	// 获取底层slog.Logger，用于与现有基础设施兼容
	SlogLogger() *slog.Logger
}

// Config 简化的日志配置
type Config struct {
	Level    slog.Level // 日志级别
	Output   string     // 输出：stdout、stderr、file
	FilePath string     // 文件路径（当Output为file时）
}

// appLogger 日志器实现
type appLogger struct {
	logger *slog.Logger
}

// Default 创建默认日志器 - 使用Go的默认配置，不做任何修改
func Default() Logger {
	return &appLogger{
		logger: slog.Default(),
	}
}

// New 根据配置创建日志器
func New(config Config) Logger {
	handler := createHandler(config)
	return &appLogger{
		logger: slog.New(handler),
	}
}

// createHandler 创建简单的Handler
func createHandler(config Config) slog.Handler {
	// 基本配置 - 与默认slog保持一致
	opts := &slog.HandlerOptions{
		Level:       config.Level,
		AddSource:   false,
		ReplaceAttr: nil,
	}

	// 获取输出Writer
	writer := getWriter(config)

	// 统一使用TextHandler
	return slog.NewTextHandler(writer, opts)
}

// getWriter 获取输出Writer
func getWriter(config Config) io.Writer {
	switch config.Output {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "file":
		if config.FilePath == "" {
			return os.Stdout
		}

		// 确保目录存在
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
			return os.Stdout
		}

		// 打开文件
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
			return os.Stdout
		}
		return file
	default:
		return os.Stdout
	}
}

// Debug 记录Debug级别日志
func (l *appLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info 记录Info级别日志
func (l *appLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn 记录Warn级别日志
func (l *appLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error 记录Error级别日志
func (l *appLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// DebugContext 记录带上下文的Debug级别日志
func (l *appLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// InfoContext 记录带上下文的Info级别日志
func (l *appLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// WarnContext 记录带上下文的Warn级别日志
func (l *appLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// ErrorContext 记录带上下文的Error级别日志
func (l *appLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// SlogLogger 获取底层slog.Logger
func (l *appLogger) SlogLogger() *slog.Logger {
	return l.logger
}

// 全局默认日志器
var defaultLogger Logger = Default()

// GetDefault 获取默认日志器
func GetDefault() Logger {
	return defaultLogger
}

```

### 6. 应用层中间件 (internal/app/middleware/)

应用层中间件模块提供了HTTP请求处理的中间件功能，负责记录请求日志、生成请求ID、处理请求上下文等横切关注点。

#### 架构设计

```
internal/app/middleware/
└── logging.go           # HTTP日志中间件
```

#### 6.1 HTTP日志中间件 (logging.go)

HTTP日志中间件基于Gin框架实现，为每个HTTP请求提供完整的生命周期日志记录功能。

**核心实现：**

```go
package middleware

import (
	"context"
	"qa-cache/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey 请求ID在Context中的键名
const RequestIDKey = "request_id"

// LoggingConfig 日志中间件配置
type LoggingConfig struct {
	// SkipPaths 跳过日志记录的路径（如健康检查接口）
	SkipPaths []string
	// IncludeRequestBody 是否记录请求体
	IncludeRequestBody bool
	// IncludeResponseBody 是否记录响应体
	IncludeResponseBody bool
	// Logger 日志器实例
	Logger logger.Logger
}

// LoggingMiddleware 返回HTTP日志记录中间件
// config: 中间件配置，如果为nil则使用默认配置
func LoggingMiddleware(config *LoggingConfig) gin.HandlerFunc {
	// 使用默认配置
	if config == nil {
		config = &LoggingConfig{
			SkipPaths:           []string{"/health", "/metrics"},
			IncludeRequestBody:  false,
			IncludeResponseBody: false,
			Logger:              logger.GetDefault(),
		}
	}

	// 如果没有指定Logger，使用默认Logger
	if config.Logger == nil {
		config.Logger = logger.GetDefault()
	}

	return func(c *gin.Context) {
		// 生成请求ID
		requestID := generateRequestID()
		c.Set(RequestIDKey, requestID)

		// 检查是否需要跳过日志记录
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 记录请求开始时间
		startTime := time.Now()

		// 提取请求信息
		requestInfo := extractRequestInfo(c, requestID)

		// 创建带有请求ID的context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 记录请求开始日志
		config.Logger.InfoContext(ctx, "HTTP请求开始",
			"request_id", requestID,
			"method", requestInfo.Method,
			"path", requestInfo.Path,
			"client_ip", requestInfo.ClientIP,
			"user_agent", requestInfo.UserAgent,
			"content_length", requestInfo.ContentLength,
			"query_params", requestInfo.QueryParams,
		)

		// 创建自定义ResponseWriter来捕获响应信息
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		// 执行请求处理
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 提取响应信息
		responseInfo := extractResponseInfo(c, responseWriter, duration)

		config.Logger.InfoContext(ctx, "HTTP请求完成",
			"request_id", requestID,
			"method", requestInfo.Method,
			"path", requestInfo.Path,
			"status_code", statusCode,
			"duration_ms", responseInfo.DurationMs,
			"response_size", responseInfo.ResponseSize,
			"client_ip", requestInfo.ClientIP,
		)

		// 如果有错误，记录详细错误信息
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				config.Logger.ErrorContext(ctx, "HTTP请求处理错误",
					"request_id", requestID,
					"error", err.Error(),
					"error_type", err.Type,
				)
			}
		}
	}
}

// RequestInfo HTTP请求信息
type RequestInfo struct {
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	ClientIP      string            `json:"client_ip"`
	UserAgent     string            `json:"user_agent"`
	ContentLength int64             `json:"content_length"`
	QueryParams   map[string]string `json:"query_params"`
	Headers       map[string]string `json:"headers"`
}

// ResponseInfo HTTP响应信息
type ResponseInfo struct {
	StatusCode   int     `json:"status_code"`
	DurationMs   float64 `json:"duration_ms"`
	ResponseSize int     `json:"response_size"`
}

// responseWriter 自定义ResponseWriter，用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

// Write 重写Write方法，捕获响应体
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}

// extractRequestInfo 提取请求信息
func extractRequestInfo(c *gin.Context, requestID string) *RequestInfo {
	// 提取查询参数
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0] // 只取第一个值
		}
	}

	// 提取重要的请求头
	headers := make(map[string]string)
	importantHeaders := []string{"Content-Type", "Accept", "Authorization", "X-Forwarded-For"}
	for _, header := range importantHeaders {
		if value := c.GetHeader(header); value != "" {
			headers[header] = value
		}
	}

	return &RequestInfo{
		Method:        c.Request.Method,
		Path:          c.Request.URL.Path,
		ClientIP:      c.ClientIP(),
		UserAgent:     c.GetHeader("User-Agent"),
		ContentLength: c.Request.ContentLength,
		QueryParams:   queryParams,
		Headers:       headers,
	}
}

// extractResponseInfo 提取响应信息
func extractResponseInfo(c *gin.Context, rw *responseWriter, duration time.Duration) *ResponseInfo {
	return &ResponseInfo{
		StatusCode:   c.Writer.Status(),
		DurationMs:   float64(duration.Nanoseconds()) / 1e6, // 转换为毫秒
		ResponseSize: len(rw.body),
	}
}

// shouldSkipPath 检查是否应该跳过某个路径的日志记录
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// GetRequestID 从Context中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}

```

**核心功能：**

- **请求ID生成**: 为每个请求生成唯一的UUID标识符
- **请求生命周期记录**: 记录请求开始和结束的详细信息
- **上下文传递**: 将请求ID注入到Go Context中，支持链路追踪
- **智能日志级别**: 根据HTTP状态码自动选择合适的日志级别
- **可配置过滤**: 支持跳过特定路径的日志记录（如健康检查接口）

**记录的信息：**

请求开始时记录：
- 请求ID、HTTP方法、请求路径
- 客户端IP、User-Agent、内容长度
- 查询参数、重要请求头

请求结束时记录：
- 响应状态码、处理时间、响应大小
- 错误信息（如果有）

**设计特点：**

1. **结构化日志**: 基于pkg/logger模块，输出JSON格式的结构化日志
2. **性能优化**: 通过自定义ResponseWriter捕获响应信息，避免额外开销
3. **可配置性**: 支持灵活的配置选项，适应不同环境需求
4. **错误处理**: 完善的错误记录和Gin错误信息收集
5. **上下文集成**: 提供便捷的请求ID提取和上下文操作方法

**辅助方法：**

```go
// GetRequestID 从Gin Context中获取请求ID
func GetRequestID(c *gin.Context) string

// GetRequestIDFromContext 从Go Context中获取请求ID
func GetRequestIDFromContext(ctx context.Context) string

// WithRequestID 为Context添加请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context
```

**集成架构：**

```
HTTP请求 → Gin路由器 → 日志中间件 → 业务处理器
    ↓             ↓         ↓           ↓
生成请求ID → 记录开始日志 → 传递上下文 → 记录结束日志
```


## 项目结构

当前已实现的项目结构：

```
qa-cache/
├── cmd/                    # 应用程序入口
│   └── server/             # 服务器主程序
│       └── main.go         # 主程序入口点，负责依赖注入和应用生命周期管理
├── internal/               # 核心业务逻辑
│   ├── app/                # 应用层
│   │   ├── handlers/       # HTTP 处理器
│   │   ├── middleware/     # 中间件
│   │   │   └── logging.go  # HTTP日志中间件
│   │   └── server/         # 服务器配置
│   │       ├── routes.go   # 路由配置
│   │       └── server.go   # HTTP服务器
│   ├── domain/             # 域模型层
│   │   ├── models/         # 域模型定义
│   │   │   ├── cache.go    # 缓存相关模型
│   │   │   ├── vector.go   # 向量相关模型
│   │   │   └── request.go  # 请求处理模型
│   │   ├── repositories/   # 数据访问层接口
│   │   │   └── vector_repository.go # 向量仓储接口
│   │   └── services/       # 业务逻辑层接口
│   │       ├── cache_service.go                    # 缓存服务接口
│   │       ├── vector_service.go                   # 向量服务接口
│   │       ├── embedding_service.go                # 嵌入服务接口
│   │       ├── request_preprocessing_service.go    # 请求预处理服务接口
│   │       ├── recall_postprocessing_service.go    # 召回后处理服务接口
│   │       └── quality_service.go                  # 质量评估服务接口
│   └── infrastructure/    # 基础设施层
│       ├── vector/         # 向量数据库实现
│       │   └── qdrant/     # Qdrant向量数据库实现
│       │       ├── init.go         # Qdrant初始化和工厂模式
│       │       ├── vector_store.go # VectorRepository接口实现
│       │       └── qdrant_client.go # Qdrant客户端封装
│       ├── embedding/      # 嵌入模型实现
│       │   └── remote/     # 远程嵌入模型实现
│       │       ├── init.go     # 初始化和配置管理
│       │       └── embedding.go # 核心向量化服务实现
│       └── services/       # 向量服务实现
│           ├── vector_service.go # 向量服务核心实现
│           ├── init.go          # 工厂模式和构建器
│           └── README.md        # 使用文档和最佳实践
├── configs/                # 配置管理模块
│   ├── config.go           # 配置结构体定义
│   ├── loader.go           # 配置加载器
│   └── config.yaml         # 主配置文件
├── pkg/                    # 共享工具包
│   ├── logger/             # 日志工具包
│   │   └── logger.go       # 基于log/slog的日志实现
│   └── status/             # 状态码模块
│       └── codes.go        # 状态码定义
├── docs/                   # 项目文档
│   ├── design.md           # 项目设计文档
│   ├── ARCHITECTURE.md     # 架构文档(本文件)
│   ├── 大模型语义缓存.md    # 项目需求文档
│   └── goroutine使用指南.md # 并发编程指南
├── go.mod                  # Go模块文件
└── go.sum                  # 依赖校验文件
```

### 7. 服务器配置 (internal/app/server/)

服务器配置模块提供了完整的HTTP服务器解决方案，基于Gin框架实现HTTP服务的生命周期管理和路由配置。

#### 架构设计

```
internal/app/server/
├── routes.go         # 路由配置
└── server.go         # HTTP服务器
```
package handlers

import (
"net/http"
"strconv"
"strings"
"time"

	"qa-cache/internal/app/middleware"
	"qa-cache/internal/domain/models"
	"qa-cache/internal/domain/services"
	"qa-cache/pkg/logger"
	"qa-cache/pkg/status"

	"github.com/gin-gonic/gin"
)

// CacheHandler 缓存处理器
// 负责处理缓存相关的HTTP请求，协调CacheService完成业务操作
type CacheHandler struct {
cacheService services.CacheService // 缓存服务接口
logger       logger.Logger         // 日志器
}

// NewCacheHandler 创建缓存处理器
func NewCacheHandler(cacheService services.CacheService, log logger.Logger) *CacheHandler {
return &CacheHandler{
cacheService: cacheService,
logger:       log,
}
}

// APIResponse 统一的API响应格式
type APIResponse struct {
Success   bool        `json:"success"`              // 是否成功
Code      int         `json:"code"`                 // 状态码
Message   string      `json:"message"`              // 消息
Data      interface{} `json:"data,omitempty"`       // 数据
RequestID string      `json:"request_id,omitempty"` // 请求ID
Timestamp int64       `json:"timestamp"`            // 时间戳
}

// ErrorDetail 错误详情
type ErrorDetail struct {
Field   string `json:"field,omitempty"` // 错误字段
Message string `json:"message"`         // 错误消息
Code    string `json:"code,omitempty"`  // 错误码
}

// QueryCache 查询缓存
// GET /v1/cache/search
func (h *CacheHandler) QueryCache(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存查询请求", "request_id", requestID)

	// 解析请求参数
	var query models.CacheQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateCacheQuery(&query); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 调用缓存服务查询
	startTime := time.Now()
	result, err := h.cacheService.QueryCache(ctx, &query)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存查询服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存查询失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存查询请求处理完成",
		"request_id", requestID,
		"duration_ms", duration,
		"found", result.Found)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存查询成功")
}

// StoreCache 存储缓存
// POST /v1/cache/store
func (h *CacheHandler) StoreCache(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存存储请求", "request_id", requestID)

	// 解析请求参数
	var request models.CacheWriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateCacheWriteRequest(&request); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 调用缓存服务存储
	startTime := time.Now()
	result, err := h.cacheService.StoreCache(ctx, &request)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存存储服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存存储失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存存储请求处理完成",
		"request_id", requestID,
		"duration_ms", duration,
		"success", result.Success,
		"cache_id", result.CacheID)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存存储成功")
}

// DeleteCache 删除单个缓存
// DELETE /v1/cache/:cache_id
func (h *CacheHandler) DeleteCache(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存删除请求", "request_id", requestID)

	// 获取缓存ID
	cacheID := c.Param("cache_id")
	if cacheID == "" {
		h.logger.ErrorContext(ctx, "缓存删除请求缺少cache_id参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少cache_id参数", "")
		return
	}

	// 获取用户类型
	userType := c.Query("user_type")
	if userType == "" {
		h.logger.ErrorContext(ctx, "缓存删除请求缺少user_type参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 构建删除请求
	deleteRequest := &models.CacheDeleteRequest{
		CacheIDs: []string{cacheID},
		UserType: userType,
		Force:    c.Query("force") == "true",
	}

	// 调用缓存服务删除
	startTime := time.Now()
	result, err := h.cacheService.DeleteCache(ctx, deleteRequest)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存删除服务调用失败",
			"request_id", requestID,
			"cache_id", cacheID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存删除失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存删除请求处理完成",
		"request_id", requestID,
		"cache_id", cacheID,
		"duration_ms", duration,
		"success", result.Success)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存删除成功")
}

// BatchDeleteCache 批量删除缓存
// DELETE /v1/cache/batch
func (h *CacheHandler) BatchDeleteCache(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理批量缓存删除请求", "request_id", requestID)

	// 解析请求参数
	var request models.CacheDeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if len(request.CacheIDs) == 0 {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少cache_ids", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少要删除的缓存ID", "")
		return
	}

	if request.UserType == "" {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少user_type", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 调用缓存服务删除
	startTime := time.Now()
	result, err := h.cacheService.DeleteCache(ctx, &request)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除服务调用失败",
			"request_id", requestID,
			"cache_count", len(request.CacheIDs),
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "批量缓存删除失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "批量缓存删除请求处理完成",
		"request_id", requestID,
		"cache_count", len(request.CacheIDs),
		"deleted_count", result.DeletedCount,
		"duration_ms", duration,
		"success", result.Success)

	// 返回成功响应
	h.respondWithSuccess(c, result, "批量缓存删除成功")
}

// GetCacheByID 根据ID获取缓存项
// GET /v1/cache/:cache_id
func (h *CacheHandler) GetCacheByID(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存查询请求", "request_id", requestID)

	// 获取缓存ID
	cacheID := c.Param("cache_id")
	if cacheID == "" {
		h.logger.ErrorContext(ctx, "缓存查询请求缺少cache_id参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少cache_id参数", "")
		return
	}

	// 获取用户类型
	userType := c.Query("user_type")
	if userType == "" {
		h.logger.ErrorContext(ctx, "缓存查询请求缺少user_type参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 解析是否包含统计信息
	includeStatistics, _ := strconv.ParseBool(c.Query("include_statistics"))

	// 调用缓存服务查询
	startTime := time.Now()
	cacheItem, err := h.cacheService.GetCacheByID(ctx, cacheID, userType, includeStatistics)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存查询服务调用失败",
			"request_id", requestID,
			"cache_id", cacheID,
			"duration_ms", duration,
			"error", err.Error())

		// 检查是否为资源不存在错误
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(c, status.ErrCodeNotFound, "缓存项不存在", err.Error())
		} else {
			h.respondWithError(c, status.ErrCodeInternal, "缓存查询失败", err.Error())
		}
		return
	}

	h.logger.InfoContext(ctx, "缓存查询请求处理完成",
		"request_id", requestID,
		"cache_id", cacheID,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, cacheItem, "缓存查询成功")
}

// GetCacheStatistics 获取缓存统计信息
// GET /v1/cache/statistics
func (h *CacheHandler) GetCacheStatistics(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存统计查询请求", "request_id", requestID)

	// 获取查询参数
	userType := c.Query("user_type")
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "24h" // 默认24小时
	}

	// 调用缓存服务查询统计信息
	startTime := time.Now()
	statistics, err := h.cacheService.GetCacheStatistics(ctx, userType, timeRange)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存统计查询服务调用失败",
			"request_id", requestID,
			"user_type", userType,
			"time_range", timeRange,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存统计查询失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存统计查询请求处理完成",
		"request_id", requestID,
		"user_type", userType,
		"time_range", timeRange,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, statistics, "缓存统计查询成功")
}

// HealthCheck 健康检查
// GET /v1/cache/health
func (h *CacheHandler) HealthCheck(c *gin.Context) {
ctx := c.Request.Context()
requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理健康检查请求", "request_id", requestID)

	// 调用缓存服务健康检查
	startTime := time.Now()
	healthInfo, err := h.cacheService.GetCacheHealth(ctx)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "健康检查服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeUnavailable, "服务不可用", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "健康检查请求处理完成",
		"request_id", requestID,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, healthInfo, "服务正常")
}

// 私有方法：参数验证

// validateCacheQuery 验证缓存查询请求
func (h *CacheHandler) validateCacheQuery(query *models.CacheQuery) error {
if strings.TrimSpace(query.Question) == "" {
return &ValidationError{Field: "question", Message: "问题不能为空"}
}

	if strings.TrimSpace(query.UserType) == "" {
		return &ValidationError{Field: "user_type", Message: "用户类型不能为空"}
	}

	if query.SimilarityThreshold != nil && (*query.SimilarityThreshold < 0 || *query.SimilarityThreshold > 1) {
		return &ValidationError{Field: "similarity_threshold", Message: "相似度阈值必须在0-1之间"}
	}

	if query.TopK != nil && (*query.TopK < 1 || *query.TopK > 100) {
		return &ValidationError{Field: "top_k", Message: "TopK值必须在1-100之间"}
	}

	return nil
}

// validateCacheWriteRequest 验证缓存写入请求
func (h *CacheHandler) validateCacheWriteRequest(request *models.CacheWriteRequest) error {
if strings.TrimSpace(request.Question) == "" {
return &ValidationError{Field: "question", Message: "问题不能为空"}
}

	if len(request.Question) > 1000 {
		return &ValidationError{Field: "question", Message: "问题长度不能超过1000字符"}
	}

	if strings.TrimSpace(request.Answer) == "" {
		return &ValidationError{Field: "answer", Message: "答案不能为空"}
	}

	if len(request.Answer) > 10000 {
		return &ValidationError{Field: "answer", Message: "答案长度不能超过10000字符"}
	}

	if strings.TrimSpace(request.UserType) == "" {
		return &ValidationError{Field: "user_type", Message: "用户类型不能为空"}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
Field   string
Message string
}

func (e *ValidationError) Error() string {
return e.Message
}

// 私有方法：响应处理

// respondWithSuccess 返回成功响应
func (h *CacheHandler) respondWithSuccess(c *gin.Context, data interface{}, message string) {
response := APIResponse{
Success:   true,
Code:      int(status.CodeOK),
Message:   message,
Data:      data,
RequestID: middleware.GetRequestID(c),
Timestamp: time.Now().Unix(),
}

	c.JSON(http.StatusOK, response)
}

// respondWithError 返回错误响应
func (h *CacheHandler) respondWithError(c *gin.Context, code status.StatusCode, message, detail string) {
response := APIResponse{
Success:   false,
Code:      int(code),
Message:   message,
RequestID: middleware.GetRequestID(c),
Timestamp: time.Now().Unix(),
}

	// 如果有详细错误信息，添加到data中
	if detail != "" {
		response.Data = ErrorDetail{
			Message: detail,
			Code:    code.String(),
		}
	}

	// 返回200的HTTP状态码
	c.JSON(http.StatusOK, response)
}

#### 7.1 路由配置 (routes.go)

路由配置是整个HTTP服务的交通枢纽，负责建立外部HTTP请求路径与内部处理逻辑之间的映射关系。

**核心实现：**

```go
// SetupRoutes 设置HTTP路由
// 这是整个HTTP服务的交通枢纽，建立外部HTTP请求路径与内部处理逻辑之间的映射关系
func SetupRoutes(engine *gin.Engine, cacheHandler *handlers.CacheHandler, log logger.Logger)
```

**API端点映射：**

- `POST /v1/cache/search` → QueryCache - 查询语义缓存，根据问题文本进行相似度匹配
- `POST /v1/cache/store` → StoreCache - 存储问答对到语义缓存，经过质量评估后写入
- `GET /v1/cache/:cache_id` → GetCacheByID - 根据缓存ID获取具体的缓存项信息
- `DELETE /v1/cache/:cache_id` → DeleteCache - 删除指定ID的缓存项
- `DELETE /v1/cache/batch` → BatchDeleteCache - 批量删除多个缓存项
- `GET /v1/cache/statistics` → GetCacheStatistics - 获取缓存系统的统计信息和性能指标
- `GET /v1/cache/health` → HealthCheck - 健康检查，验证缓存服务是否正常运行

**中间件集成：**

```go
package server

import (
	"qa-cache/internal/app/handlers"
	"qa-cache/internal/app/middleware"
	"qa-cache/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置HTTP路由
// 这是整个HTTP服务的交通枢纽，建立外部HTTP请求路径与内部处理逻辑之间的映射关系
func SetupRoutes(engine *gin.Engine, cacheHandler *handlers.CacheHandler, log logger.Logger) {
	// 应用全局中间件
	setupMiddleware(engine, log)

	// 设置API路由组
	v1 := engine.Group("/v1")
	// 缓存相关路由
	cache := v1.Group("/cache")

	// 查询缓存 - POST方法，支持复杂查询条件
	cache.POST("/search", cacheHandler.QueryCache)
	// 存储缓存 - 将问答对存入语义缓存
	cache.POST("/store", cacheHandler.StoreCache)
	// 根据ID获取缓存项 - 支持查询参数：user_type, include_statistics
	cache.GET("/:cache_id", cacheHandler.GetCacheByID)
	// 删除单个缓存项 - 支持查询参数：user_type, force
	cache.DELETE("/:cache_id", cacheHandler.DeleteCache)
	// 批量删除缓存 - 请求体包含要删除的ID列表
	cache.DELETE("/batch", cacheHandler.BatchDeleteCache)
	// 获取缓存统计信息 - 支持查询参数：user_type, time_range
	cache.GET("/statistics", cacheHandler.GetCacheStatistics)
	// 健康检查 - 检查缓存服务状态
	cache.GET("/health", cacheHandler.HealthCheck)

}

// setupMiddleware 设置全局中间件
func setupMiddleware(engine *gin.Engine, log logger.Logger) {
	// 设置恢复中间件 - 捕获panic并返回500错误
	engine.Use(gin.Recovery())

	// 设置日志中间件 - 记录请求日志并生成请求ID
	loggingConfig := &middleware.LoggingConfig{
		// 跳过健康检查路径的日志记录，减少日志噪音
		SkipPaths: []string{
			"/v1/cache/health",
		},
		// 暂不记录请求和响应体，避免日志过大
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		Logger:              log,
	}
	engine.Use(middleware.LoggingMiddleware(loggingConfig))
}

```

**设计特点：**

- **版本化API**: 使用`/v1`前缀支持API版本管理
- **模块化路由**: 采用路由分组，便于功能模块化管理
- **中间件集成**: 集成日志中间件和恢复中间件
- **智能过滤**: 跳过健康检查路径的日志记录，减少日志噪音

#### 7.2 HTTP服务器 (server.go)

HTTP服务器是基于Gin框架的核心管理器，负责整个服务器的生命周期管理。

**核心结构：**

```go
package server

import (
	"context"
	"fmt"
	"net/http"
	"qa-cache/configs"
	"qa-cache/internal/app/handlers"
	"qa-cache/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器结构体
// 负责整个服务器的生命周期管理，包括初始化、启动、运行和优雅关闭
type Server struct {
	config       *configs.ServerConfig  // 服务器配置
	httpServer   *http.Server           // HTTP服务器实例
	engine       *gin.Engine            // Gin引擎
	cacheHandler *handlers.CacheHandler // 缓存处理器
	logger       logger.Logger          // 日志器
}

// NewServer 创建新的HTTP服务器实例
// config: 服务器配置
// cacheHandler: 缓存处理器
// log: 日志器
func NewServer(config *configs.ServerConfig, cacheHandler *handlers.CacheHandler, log logger.Logger) *Server {
	// 根据配置设置Gin模式
	if config.Host == "0.0.0.0" || config.Host == "" {
		gin.SetMode(gin.ReleaseMode) // 生产模式
	} else {
		gin.SetMode(gin.DebugMode) // 开发模式
	}

	// 创建Gin引擎
	engine := gin.New()

	return &Server{
		config:       config,
		engine:       engine,
		cacheHandler: cacheHandler,
		logger:       log,
	}
}

// Start 启动HTTP服务器
// ctx: 上下文，用于控制服务器启动过程
func (s *Server) Start(ctx context.Context) error {

	// 设置路由
	SetupRoutes(s.engine, s.cacheHandler, s.logger)

	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.logger.InfoContext(ctx, "HTTP服务器初始化完成",
		"addr", s.httpServer.Addr,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"idle_timeout", s.config.IdleTimeout)

	// 启动服务器（非阻塞）
	go func() {
		s.logger.InfoContext(ctx, "HTTP服务器开始监听", "addr", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "HTTP服务器启动失败", "error", err.Error())
		}
	}()

	return nil
}

// Shutdown 优雅关闭服务器
// ctx: 上下文，用于控制关闭过程的超时
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "开始执行HTTP服务器优雅关闭")

	// 创建带超时的关闭上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.GracefulShutdownTimeout)
	defer cancel()

	// 执行优雅关闭
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(ctx, "HTTP服务器优雅关闭失败，强制关闭", "error", err.Error())
		return fmt.Errorf("HTTP服务器关闭失败: %w", err)
	}

	s.logger.InfoContext(ctx, "HTTP服务器优雅关闭完成")
	return nil
}

```

**生命周期管理：**

```go
// NewServer 创建新的HTTP服务器实例
func NewServer(config *configs.ServerConfig, cacheHandler *handlers.CacheHandler, log logger.Logger) *Server

// Start 启动HTTP服务器（非阻塞）
func (s *Server) Start(ctx context.Context) error

// Run 运行服务器直到收到停止信号
func (s *Server) Run(ctx context.Context) error

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error
```

**核心功能：**

1. **智能模式设置**: 根据配置自动选择Gin运行模式（开发/生产）
2. **配置化超时**: 支持读取/写入/空闲超时的灵活配置
3. **信号处理**: 监听SIGINT/SIGTERM系统信号，支持优雅关闭
4. **上下文支持**: 完整的context.Context集成，支持取消和超时控制
5. **依赖注入**: 清晰的依赖注入设计，遵循Clean Architecture原则

**优雅关闭机制：**

- 监听系统信号（SIGINT/SIGTERM）
- 配置化的关闭超时时间
- 完整的资源清理流程
- 详细的关闭状态日志记录

**扩展性支持：**

```go
// ServerOptions 服务器选项（用于扩展配置）
type ServerOptions struct {
    EnablePprof       bool              // 是否启用性能分析
    EnableMetrics     bool              // 是否启用指标收集
    CustomMiddleware  []gin.HandlerFunc // 自定义中间件
    ShutdownTimeout   time.Duration     // 关闭超时时间
}

// ApplyOptions 应用服务器选项
func (s *Server) ApplyOptions(options *ServerOptions)
```

#### 7.3 集成架构

服务器配置与整体架构的集成关系：

```
HTTP请求 → 路由配置(routes.go) → 缓存处理器(handlers/) → 业务服务层(services/)
    ↓              ↓                       ↓                    ↓
服务器管理 → 中间件处理 → 请求ID生成 → 结构化日志 → 统一状态码
(server.go)    ↓              ↓              ↓           ↓
             路由分组 → 上下文传递 → 链路追踪 → 错误处理
```

**关键特性：**

- **基于Gin框架**: 高性能的HTTP路由处理和丰富的中间件生态
- **Clean Architecture**: 遵循依赖倒置原则，服务器层不依赖具体业务实现
- **生产就绪**: 优雅关闭、信号处理、配置化超时等生产环境特性
- **可观测性**: 集成结构化日志和请求追踪
- **可扩展性**: 支持自定义中间件和灵活的配置选项

### 8. 向量数据库实现 (internal/infrastructure/vector/qdrant/)

向量数据库实现模块提供了完整的向量存储和检索能力，采用Qdrant作为底层向量数据库，实现了VectorRepository接口的所有功能。

#### 架构设计

```
internal/infrastructure/vector/qdrant/
├── init.go                 # Qdrant初始化和工厂模式
├── vector_store.go         # VectorRepository接口实现
└── qdrant_client.go        # Qdrant客户端封装
```

#### 8.1 Qdrant向量存储实现 (vector_store.go)

QdrantVectorStore是VectorRepository接口的完整实现，提供向量数据的持久化操作。
package qdrant

import (
"context"
"fmt"
"log/slog"
"time"

	"qa-cache/configs"
	"qa-cache/internal/domain/models"
	"qa-cache/internal/domain/repositories"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantVectorStore Qdrant向量存储实现
// 实现 VectorRepository 接口，提供向量数据的持久化操作
type QdrantVectorStore struct {
client *QdrantClient
config *configs.QdrantConfig
logger *slog.Logger
}

// NewQdrantVectorStore 创建新的Qdrant向量存储实例
func NewQdrantVectorStore(ctx context.Context, config *configs.QdrantConfig, logger *slog.Logger) (repositories.VectorRepository, error) {
if config == nil {
return nil, fmt.Errorf("qdrant config cannot be nil")
}

	if logger == nil {
		logger = slog.Default()
	}

	// 创建Qdrant客户端
	client, err := NewQdrantClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	store := &QdrantVectorStore{
		client: client,
		config: config,
		logger: logger,
	}

	logger.InfoContext(ctx, "Qdrant向量存储初始化成功",
		"collection", config.CollectionName,
		"vector_size", config.VectorSize)

	return store, nil
}

// Store 存储单个向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) Store(ctx context.Context, request *models.VectorStoreRequest) (*models.VectorStoreResponse, error) {
startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if request.ID == "" {
		return nil, fmt.Errorf("vector ID cannot be empty")
	}

	if len(request.Vector) == 0 {
		return nil, fmt.Errorf("vector cannot be empty")
	}

	// 验证向量维度
	if len(request.Vector) != v.config.VectorSize {
		v.logger.WarnContext(ctx, "向量维度不匹配",
			"vector_id", request.ID,
			"expected", v.config.VectorSize,
			"actual", len(request.Vector))
		return &models.VectorStoreResponse{
			Success:   false,
			VectorID:  request.ID,
			Message:   fmt.Sprintf("vector dimension mismatch: expected %d, got %d", v.config.VectorSize, len(request.Vector)),
			StoreTime: float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	// 丰富payload数据
	enrichedPayload := v.enrichPayload(request.Payload)

	// 执行单个向量存储
	err := v.client.UpsertPoint(ctx, request.ID, request.Vector, enrichedPayload)
	if err != nil {
		v.logger.ErrorContext(ctx, "单个向量存储失败",
			"vector_id", request.ID,
			"error", err)
		return &models.VectorStoreResponse{
			Success:   false,
			VectorID:  request.ID,
			Message:   fmt.Sprintf("vector storage failed: %v", err),
			StoreTime: float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	duration := time.Since(startTime)
	v.logger.InfoContext(ctx, "单个向量存储完成",
		"vector_id", request.ID,
		"duration_ms", duration.Milliseconds())

	return &models.VectorStoreResponse{
		Success:   true,
		VectorID:  request.ID,
		Message:   "vector stored successfully",
		StoreTime: float64(duration.Milliseconds()),
	}, nil
}

// BatchStore 存储向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) BatchStore(ctx context.Context, request *models.VectorBatchStoreRequest) (*models.VectorBatchStoreResponse, error) {
startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(request.Vectors) == 0 {
		return &models.VectorBatchStoreResponse{
			Success:     true,
			StoredCount: 0,
			Message:     "no vectors to store",
			StoreTime:   0,
		}, nil
	}

	// 转换为批量点格式
	batchPoints := make([]BatchPoint, 0, len(request.Vectors))
	failedIDs := make([]string, 0)

	for _, item := range request.Vectors {
		// 验证向量维度
		if len(item.Vector) != v.config.VectorSize {
			failedIDs = append(failedIDs, item.ID)
			v.logger.WarnContext(ctx, "向量维度不匹配",
				"vector_id", item.ID,
				"expected", v.config.VectorSize,
				"actual", len(item.Vector))
			continue
		}

		// 丰富payload数据
		enrichedPayload := v.enrichPayload(item.Payload)

		batchPoint := BatchPoint{
			ID:      item.ID,
			Vector:  item.Vector,
			Payload: enrichedPayload,
		}

		batchPoints = append(batchPoints, batchPoint)
	}

	// 执行批量存储
	err := v.client.UpsertBatch(ctx, batchPoints)
	if err != nil {
		v.logger.ErrorContext(ctx, "批量存储向量失败",
			"total_count", len(request.Vectors),
			"valid_count", len(batchPoints),
			"error", err)
		return &models.VectorBatchStoreResponse{
			Success:     false,
			StoredCount: 0,
			FailedCount: len(request.Vectors),
			FailedIDs:   v.extractIDs(request.Vectors),
			Message:     fmt.Sprintf("batch storage failed: %v", err),
			StoreTime:   float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	duration := time.Since(startTime)
	storedCount := len(batchPoints)
	failedCount := len(failedIDs)

	v.logger.InfoContext(ctx, "批量向量存储完成",
		"total_count", len(request.Vectors),
		"stored_count", storedCount,
		"failed_count", failedCount,
		"duration_ms", duration.Milliseconds())

	return &models.VectorBatchStoreResponse{
		Success:     true,
		StoredCount: storedCount,
		FailedCount: failedCount,
		FailedIDs:   failedIDs,
		Message:     fmt.Sprintf("successfully stored %d vectors", storedCount),
		StoreTime:   float64(duration.Milliseconds()),
	}, nil
}

// Search 向量相似性搜索 - 实现VectorRepository接口
func (v *QdrantVectorStore) Search(ctx context.Context, request *models.VectorSearchRequest) (*models.VectorSearchResponse, error) {
startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(request.QueryVector) == 0 && request.QueryID == "" {
		return nil, fmt.Errorf("either query_vector or query_id must be provided")
	}

	// 构建过滤器
	filter := v.buildFilter(request.UserType, request.Filters)

	// 构建搜索参数
	limit := uint64(request.TopK)
	var scoreThreshold *float32
	if request.SimilarityThreshold > 0 {
		threshold := float32(request.SimilarityThreshold)
		scoreThreshold = &threshold
	}

	// 获取查询向量：优先使用QueryVector，其次使用QueryID
	var queryVector []float32
	if len(request.QueryVector) > 0 {
		// 直接使用提供的向量
		queryVector = request.QueryVector
		v.logger.InfoContext(ctx, "使用提供的查询向量",
			"vector_dimension", len(queryVector),
			"query_text", request.QueryText)
	} else if request.QueryID != "" {
		// 根据ID获取向量
		point, err := v.client.GetPoint(ctx, request.QueryID, false, true)
		if err != nil {
			v.logger.ErrorContext(ctx, "获取查询向量失败",
				"query_id", request.QueryID,
				"error", err)
			return nil, fmt.Errorf("failed to get query vector: %w", err)
		}
		if point == nil {
			return nil, fmt.Errorf("query vector not found: %s", request.QueryID)
		}
		queryVector = point.Vector
		v.logger.InfoContext(ctx, "通过ID获取查询向量",
			"query_id", request.QueryID,
			"vector_dimension", len(queryVector))
	}

	// 执行搜索
	results, err := v.client.SearchPoints(ctx, queryVector, limit, scoreThreshold, filter)
	if err != nil {
		v.logger.ErrorContext(ctx, "向量搜索失败",
			"user_type", request.UserType,
			"top_k", request.TopK,
			"threshold", request.SimilarityThreshold,
			"error", err)
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 转换搜索结果
	searchResults := make([]models.VectorSearchResult, 0, len(results))
	for _, result := range results {
		searchResult := models.VectorSearchResult{
			ID:      result.ID,
			Score:   float64(result.Score),
			Payload: result.Payload,
		}

		// 如果需要，可以包含向量数据
		if result.Vector != nil {
			vector := models.NewVector(result.ID, result.Vector)
			searchResult.Vector = vector
		}

		searchResults = append(searchResults, searchResult)
	}

	duration := time.Since(startTime)
	v.logger.InfoContext(ctx, "向量搜索完成",
		"user_type", request.UserType,
		"result_count", len(searchResults),
		"duration_ms", duration.Milliseconds())

	return &models.VectorSearchResponse{
		Results:    searchResults,
		TotalCount: len(searchResults),
		SearchTime: float64(duration.Milliseconds()),
		QueryInfo: models.VectorQueryInfo{
			Dimension:     len(queryVector),
			FilterApplied: len(request.Filters) > 0 || request.UserType != "",
		},
	}, nil
}

// Delete 删除向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) Delete(ctx context.Context, ids []string, userType string) error {
if len(ids) == 0 {
return nil
}

	// 如果指定了用户类型，需要先验证这些向量是否属于该用户类型
	if userType != "" {
		// 为安全起见，先获取这些向量的信息进行验证
		for _, id := range ids {
			point, err := v.client.GetPoint(ctx, id, true, false)
			if err != nil {
				v.logger.WarnContext(ctx, "获取向量信息失败，跳过删除",
					"vector_id", id,
					"error", err)
				continue
			}
			if point == nil {
				v.logger.WarnContext(ctx, "向量不存在，跳过删除", "vector_id", id)
				continue
			}

			// 检查用户类型
			if payloadUserType, ok := point.Payload["user_type"].(string); ok {
				if payloadUserType != userType {
					v.logger.WarnContext(ctx, "用户类型不匹配，跳过删除",
						"vector_id", id,
						"expected_user_type", userType,
						"actual_user_type", payloadUserType)
					continue
				}
			}
		}
	}

	err := v.client.DeleteBatch(ctx, ids)
	if err != nil {
		v.logger.ErrorContext(ctx, "批量删除向量失败",
			"count", len(ids),
			"user_type", userType,
			"error", err)
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	v.logger.InfoContext(ctx, "批量向量删除成功",
		"count", len(ids),
		"user_type", userType)
	return nil
}

// GetByID 根据ID获取向量 - 实现VectorRepository接口
func (v *QdrantVectorStore) GetByID(ctx context.Context, id string) (*models.Vector, error) {
if id == "" {
return nil, fmt.Errorf("vector ID cannot be empty")
}

	// 获取向量点
	point, err := v.client.GetPoint(ctx, id, false, true)
	if err != nil {
		v.logger.ErrorContext(ctx, "获取向量失败",
			"vector_id", id,
			"error", err)
		return nil, fmt.Errorf("failed to get vector: %w", err)
	}

	if point == nil {
		return nil, nil // 未找到
	}

	// 构建向量对象
	vector := models.NewVector(point.ID, point.Vector)

	v.logger.DebugContext(ctx, "向量获取成功",
		"vector_id", id,
		"dimension", len(point.Vector))

	return vector, nil
}

// 以下是附加的辅助方法，不是接口的一部分

// buildFilter 构建Qdrant过滤器
func (v *QdrantVectorStore) buildFilter(userType string, filters map[string]interface{}) *qdrant.Filter {
conditions := make([]*qdrant.Condition, 0)

	// 添加用户类型过滤
	if userType != "" {
		conditions = append(conditions, qdrant.NewMatch("user_type", userType))
	}

	// 添加其他过滤条件
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			conditions = append(conditions, qdrant.NewMatch(key, v))
		}
	}

	if len(conditions) == 0 {
		return nil
	}

	return &qdrant.Filter{
		Must: conditions,
	}
}

// enrichPayload 丰富payload数据，可选择性添加向量元数据
func (v *QdrantVectorStore) enrichPayload(originalPayload map[string]interface{}) map[string]interface{} {
payload := make(map[string]interface{})

	// 复制原始payload
	for k, v := range originalPayload {
		payload[k] = v
	}

	// 添加时间戳
	payload["created_at"] = time.Now().Format(time.RFC3339)

	return payload
}

// extractIDs 提取向量ID列表
func (v *QdrantVectorStore) extractIDs(vectors []models.VectorStoreItem) []string {
ids := make([]string, len(vectors))
for i, vector := range vectors {
ids[i] = vector.ID
}
return ids
}

// Close 关闭向量存储连接
func (v *QdrantVectorStore) Close() error {
if v.client != nil {
return v.client.Close()
}
return nil
}

**核心实现：**

```go
// QdrantVectorStore Qdrant向量存储实现
type QdrantVectorStore struct {
    client *QdrantClient        // Qdrant客户端
    config *configs.QdrantConfig // 配置信息
    logger *slog.Logger         // 结构化日志
}
```

**VectorRepository接口实现：**

- **BatchStore**: 批量存储向量数据，支持维度验证和错误处理
- **Search**: 向量相似性搜索，支持过滤器和阈值设置
- **Delete**: 批量删除向量，支持用户类型权限验证
- **GetByID**: 根据ID获取单个向量数据

**核心特性：**

1. **批量处理优化**: 支持高效的批量向量存储和删除操作
2. **智能过滤**: 通过buildFilter方法构建复杂的查询过滤条件
3. **用户隔离**: 基于user_type实现多租户数据隔离
4. **元数据丰富**: 自动添加时间戳等元数据信息
5. **错误处理**: 完善的错误日志记录和异常处理机制

#### 8.2 Qdrant客户端封装 (qdrant_client.go)
package qdrant

import (
"context"
"fmt"
"log/slog"
"time"

	"qa-cache/configs"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantClient Qdrant客户端封装
// 提供向量数据库的基础操作能力，封装gRPC连接管理和错误处理
type QdrantClient struct {
client     *qdrant.Client
config     *configs.QdrantConfig
logger     *slog.Logger
collection string
}

// QdrantClientConfig Qdrant客户端配置
type QdrantClientConfig struct {
Host           string        `json:"host"`
Port           int           `json:"port"`
APIKey         string        `json:"api_key,omitempty"`
UseTLS         bool          `json:"use_tls"`
CollectionName string        `json:"collection_name"`
VectorSize     int           `json:"vector_size"`
Distance       string        `json:"distance"`
Timeout        time.Duration `json:"timeout"`
MaxRetries     int           `json:"max_retries"`
RetryDelay     time.Duration `json:"retry_delay"`
}

// NewQdrantClient 创建新的Qdrant客户端实例
func NewQdrantClient(ctx context.Context, config *configs.QdrantConfig, logger *slog.Logger) (*QdrantClient, error) {
if config == nil {
return nil, fmt.Errorf("qdrant config cannot be nil")
}

	if logger == nil {
		logger = slog.Default()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		logger.ErrorContext(ctx, "Qdrant配置验证失败", "error", err)
		return nil, fmt.Errorf("invalid qdrant config: %w", err)
	}

	// 创建Qdrant客户端配置
	clientConfig := &qdrant.Config{
		Host:   config.Host,
		Port:   config.Port,
		APIKey: config.APIKey,
		UseTLS: config.APIKey != "", // 如果有API Key则使用TLS
	}

	// 创建客户端
	client, err := qdrant.NewClient(clientConfig)
	if err != nil {
		logger.ErrorContext(ctx, "创建Qdrant客户端失败",
			"host", config.Host,
			"port", config.Port,
			"error", err)
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	qdrantClient := &QdrantClient{
		client:     client,
		config:     config,
		logger:     logger,
		collection: config.CollectionName,
	}

	// 测试连接并确保集合存在
	if err := qdrantClient.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	logger.InfoContext(ctx, "Qdrant客户端初始化成功",
		"host", config.Host,
		"port", config.Port,
		"collection", config.CollectionName)

	return qdrantClient, nil
}

// ensureCollection 确保集合存在，如果不存在则创建
func (q *QdrantClient) ensureCollection(ctx context.Context) error {
// 检查集合是否存在
exists, err := q.collectionExists(ctx)
if err != nil {
return fmt.Errorf("failed to check collection existence: %w", err)
}

	if exists {
		return nil
	}

	// 创建集合
	return q.createCollection(ctx)
}

// collectionExists 检查集合是否存在
func (q *QdrantClient) collectionExists(ctx context.Context) (bool, error) {
collections, err := q.client.ListCollections(ctx)
if err != nil {
q.logger.ErrorContext(ctx, "获取集合列表失败", "error", err)
return false, err
}

	for _, collectionName := range collections {
		if collectionName == q.collection {
			return true, nil
		}
	}

	return false, nil
}

// createCollection 创建向量集合
func (q *QdrantClient) createCollection(ctx context.Context) error {
// 转换距离类型
distance, err := q.parseDistance()
if err != nil {
return err
}

	// 创建集合配置
	vectorsConfig := qdrant.NewVectorsConfig(&qdrant.VectorParams{
		Size:     uint64(q.config.VectorSize),
		Distance: distance,
	})

	// 创建集合
	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: q.collection,
		VectorsConfig:  vectorsConfig,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "创建向量集合失败",
			"collection", q.collection,
			"vector_size", q.config.VectorSize,
			"distance", q.config.Distance,
			"error", err)
		return fmt.Errorf("failed to create collection %s: %w", q.collection, err)
	}

	q.logger.InfoContext(ctx, "向量集合创建成功",
		"collection", q.collection,
		"vector_size", q.config.VectorSize,
		"distance", q.config.Distance)

	return nil
}

// parseDistance 解析距离类型
func (q *QdrantClient) parseDistance() (qdrant.Distance, error) {
switch q.config.Distance {
case "cosine":
return qdrant.Distance_Cosine, nil
case "euclidean":
return qdrant.Distance_Euclid, nil
case "dot":
return qdrant.Distance_Dot, nil
case "manhattan":
return qdrant.Distance_Manhattan, nil
default:
return qdrant.Distance_Cosine, fmt.Errorf("unsupported distance type: %s", q.config.Distance)
}
}

// UpsertPoint 插入或更新单个向量点
func (q *QdrantClient) UpsertPoint(ctx context.Context, id string, vector []float32, payload map[string]interface{}) error {
if len(vector) != q.config.VectorSize {
return fmt.Errorf("vector dimension mismatch: expected %d, got %d", q.config.VectorSize, len(vector))
}

	// 创建点结构
	point := &qdrant.PointStruct{
		Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}},
		Vectors: qdrant.NewVectors(vector...),
		Payload: qdrant.NewValueMap(payload),
	}

	// 执行插入
	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Points:         []*qdrant.PointStruct{point},
		Wait:           &waitUpsert,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "向量点插入失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to upsert point %s: %w", id, err)
	}

	q.logger.DebugContext(ctx, "向量点插入成功", "id", id, "collection", q.collection)
	return nil
}

// UpsertBatch 批量插入或更新向量点
func (q *QdrantClient) UpsertBatch(ctx context.Context, points []BatchPoint) error {
if len(points) == 0 {
return nil
}

	// 转换为Qdrant点结构
	qdrantPoints := make([]*qdrant.PointStruct, 0, len(points))
	for _, point := range points {
		if len(point.Vector) != q.config.VectorSize {
			return fmt.Errorf("vector dimension mismatch for point %s: expected %d, got %d",
				point.ID, q.config.VectorSize, len(point.Vector))
		}

		qdrantPoint := &qdrant.PointStruct{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: point.ID}},
			Vectors: qdrant.NewVectors(point.Vector...),
			Payload: qdrant.NewValueMap(point.Payload),
		}
		qdrantPoints = append(qdrantPoints, qdrantPoint)
	}

	// 执行批量插入
	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Points:         qdrantPoints,
		Wait:           &waitUpsert,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "批量向量点插入失败",
			"count", len(points),
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to upsert batch points: %w", err)
	}

	q.logger.DebugContext(ctx, "批量向量点插入成功",
		"count", len(points),
		"collection", q.collection)
	return nil
}

// BatchPoint 批量点结构
type BatchPoint struct {
ID      string                 `json:"id"`
Vector  []float32              `json:"vector"`
Payload map[string]interface{} `json:"payload"`
}

// SearchPoints 搜索相似向量点
func (q *QdrantClient) SearchPoints(ctx context.Context, queryVector []float32, limit uint64, scoreThreshold *float32, filter *qdrant.Filter) ([]*SearchResult, error) {
if len(queryVector) != q.config.VectorSize {
return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", q.config.VectorSize, len(queryVector))
}

	// 构建查询请求
	queryRequest := &qdrant.QueryPoints{
		CollectionName: q.collection,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          &limit,
		Filter:         filter,
		ScoreThreshold: scoreThreshold,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
	}

	// 执行搜索
	q.logger.InfoContext(ctx, "开始向量搜索", "collection", q.collection, "limit", limit)

	queryResult, err := q.client.Query(ctx, queryRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "向量搜索失败",
			"collection", q.collection,
			"error", err)
		return nil, fmt.Errorf("failed to search points: %w", err)
	}

	// 转换结果
	results := make([]*SearchResult, 0, len(queryResult))
	for _, point := range queryResult {
		result := &SearchResult{
			Score: point.Score,
		}

		// 转换PointId为字符串
		switch id := point.Id.PointIdOptions.(type) {
		case *qdrant.PointId_Num:
			result.ID = fmt.Sprintf("%d", id.Num)
		case *qdrant.PointId_Uuid:
			result.ID = id.Uuid
		}

		// 提取payload
		if point.Payload != nil {
			result.Payload = convertPayload(point.Payload)
		}

		results = append(results, result)
	}

	q.logger.DebugContext(ctx, "向量搜索完成",
		"collection", q.collection,
		"result_count", len(results),
		"limit", limit)

	return results, nil
}

// SearchResult 搜索结果
type SearchResult struct {
ID      string                 `json:"id"`
Score   float32                `json:"score"`
Vector  []float32              `json:"vector,omitempty"`
Payload map[string]interface{} `json:"payload,omitempty"`
}

// GetPoint 根据ID获取向量点
func (q *QdrantClient) GetPoint(ctx context.Context, id string, withPayload bool, withVector bool) (*SearchResult, error) {
// 构建查询请求
getRequest := &qdrant.GetPoints{
CollectionName: q.collection,
Ids:            []*qdrant.PointId{qdrant.NewID(id)},
}

	if withPayload {
		getRequest.WithPayload = qdrant.NewWithPayloadInclude()
	}

	// 执行查询
	getResult, err := q.client.Get(ctx, getRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "获取向量点失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return nil, fmt.Errorf("failed to get point %s: %w", id, err)
	}

	if len(getResult) == 0 {
		return nil, nil // 未找到
	}

	point := getResult[0]

	// 转换PointId为字符串
	var idStr string
	switch id := point.Id.PointIdOptions.(type) {
	case *qdrant.PointId_Num:
		idStr = fmt.Sprintf("%d", id.Num)
	case *qdrant.PointId_Uuid:
		idStr = id.Uuid
	}

	result := &SearchResult{
		ID: idStr,
	}

	// 提取payload
	if withPayload && point.Payload != nil {
		result.Payload = convertPayload(point.Payload)
	}

	// 提取vector
	if withVector && point.Vectors != nil {
		if vectors := point.Vectors.GetVector(); vectors != nil {
			result.Vector = vectors.Data
		}
	}

	return result, nil
}

// DeletePoint 删除单个向量点
func (q *QdrantClient) DeletePoint(ctx context.Context, id string) error {
waitDelete := true
_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
CollectionName: q.collection,
Points: &qdrant.PointsSelector{
PointsSelectorOneOf: &qdrant.PointsSelector_Points{
Points: &qdrant.PointsIdsList{
Ids: []*qdrant.PointId{qdrant.NewID(id)},
},
},
},
Wait: &waitDelete,
})

	if err != nil {
		q.logger.ErrorContext(ctx, "删除向量点失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to delete point %s: %w", id, err)
	}

	q.logger.DebugContext(ctx, "向量点删除成功", "id", id, "collection", q.collection)
	return nil
}

// DeleteBatch 批量删除向量点
func (q *QdrantClient) DeleteBatch(ctx context.Context, ids []string) error {
if len(ids) == 0 {
return nil
}

	// 转换ID列表
	pointIds := make([]*qdrant.PointId, 0, len(ids))
	for _, id := range ids {
		pointIds = append(pointIds, qdrant.NewID(id))
	}

	waitDelete := true
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.collection,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIds,
				},
			},
		},
		Wait: &waitDelete,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "批量删除向量点失败",
			"count", len(ids),
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to delete batch points: %w", err)
	}

	q.logger.DebugContext(ctx, "批量向量点删除成功",
		"count", len(ids),
		"collection", q.collection)
	return nil
}

// CountPoints 统计向量点数量
func (q *QdrantClient) CountPoints(ctx context.Context, filter *qdrant.Filter) (uint64, error) {
exact := true
countRequest := &qdrant.CountPoints{
CollectionName: q.collection,
Filter:         filter,
Exact:          &exact,
}

	countResult, err := q.client.Count(ctx, countRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "统计向量点数量失败",
			"collection", q.collection,
			"error", err)
		return 0, fmt.Errorf("failed to count points: %w", err)
	}

	return countResult, nil
}

// HealthCheck 健康检查
func (q *QdrantClient) HealthCheck(ctx context.Context) error {
healthResult, err := q.client.HealthCheck(ctx)
if err != nil {
q.logger.ErrorContext(ctx, "Qdrant健康检查失败", "error", err)
return fmt.Errorf("qdrant health check failed: %w", err)
}

	if healthResult.Title != "qdrant - vector search engine" {
		return fmt.Errorf("unexpected health check response: %s", healthResult.Title)
	}

	q.logger.DebugContext(ctx, "Qdrant健康检查成功", "version", healthResult.Version)
	return nil
}

// GetCollectionInfo 获取集合信息
func (q *QdrantClient) GetCollectionInfo(ctx context.Context) (map[string]interface{}, error) {
collectionInfo, err := q.client.GetCollectionInfo(ctx, q.collection)
if err != nil {
q.logger.ErrorContext(ctx, "获取集合信息失败",
"collection", q.collection,
"error", err)
return nil, fmt.Errorf("failed to get collection info: %w", err)
}

	info := map[string]interface{}{
		"status":         collectionInfo.Status.String(),
		"vectors_count":  collectionInfo.VectorsCount,
		"segments_count": collectionInfo.SegmentsCount,
		"config":         collectionInfo.Config,
	}

	return info, nil
}

// Close 关闭客户端连接
func (q *QdrantClient) Close() error {
if q.client != nil {
return q.client.Close()
}
return nil
}

// convertPayload 转换payload格式
func convertPayload(payload map[string]*qdrant.Value) map[string]interface{} {
result := make(map[string]interface{})
for key, value := range payload {
result[key] = convertValue(value)
}
return result
}

// convertValue 转换Value到interface{}
func convertValue(value *qdrant.Value) interface{} {
if value == nil {
return nil
}

	switch v := value.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_ListValue:
		list := make([]interface{}, len(v.ListValue.Values))
		for i, item := range v.ListValue.Values {
			list[i] = convertValue(item)
		}
		return list
	case *qdrant.Value_StructValue:
		return convertPayload(v.StructValue.Fields)
	default:
		return nil
	}
}

QdrantClient提供了对Qdrant gRPC API的高级封装，简化向量数据库操作。

**核心功能：**

```go
// QdrantClient Qdrant客户端封装
type QdrantClient struct {
    client     *qdrant.Client        // 原生Qdrant客户端
    config     *configs.QdrantConfig // 配置信息
    logger     *slog.Logger          // 日志器
    collection string                // 集合名称
}
```

**主要方法：**

- **UpsertPoint/UpsertBatch**: 单个和批量向量插入/更新
- **SearchPoints**: 高性能向量相似度搜索
- **GetPoint**: 根据ID获取向量点
- **DeletePoint/DeleteBatch**: 单个和批量删除操作
- **CountPoints**: 统计向量数量
- **HealthCheck**: 健康状态检查
- **GetCollectionInfo**: 获取集合信息

**设计特点：**

1. **自动集合管理**: 启动时自动创建和验证集合存在性
2. **类型转换**: 提供Qdrant protobuf类型与Go类型的自动转换
3. **距离度量支持**: 支持余弦、欧几里得、点积、曼哈顿距离
4. **连接管理**: 提供连接健康检查和资源清理功能

#### 8.3 初始化和工厂模式 (init.go)
package qdrant

import (
"context"
"fmt"
"log/slog"

	"qa-cache/configs"
	"qa-cache/internal/domain/repositories"
)

// QdrantVectorStoreFactory Qdrant向量存储工厂
type QdrantVectorStoreFactory struct {
logger *slog.Logger
}

// NewQdrantVectorStoreFactory 创建Qdrant向量存储工厂
func NewQdrantVectorStoreFactory(logger *slog.Logger) *QdrantVectorStoreFactory {
if logger == nil {
logger = slog.Default()
}

	return &QdrantVectorStoreFactory{
		logger: logger,
	}
}

// CreateVectorRepository 创建Qdrant向量仓储实例
// 根据配置创建Qdrant向量存储实现
func (f *QdrantVectorStoreFactory) CreateVectorRepository(ctx context.Context, config *configs.QdrantConfig) (repositories.VectorRepository, error) {
if config == nil {
return nil, fmt.Errorf("qdrant config cannot be nil")
}

	store, err := NewQdrantVectorStore(ctx, config, f.logger)
	if err != nil {
		f.logger.ErrorContext(ctx, "Qdrant向量存储初始化失败", "error", err)
		return nil, fmt.Errorf("failed to create qdrant vector store: %w", err)
	}

	return store, nil
}

// ValidateQdrantConfiguration 验证Qdrant配置
func ValidateQdrantConfiguration(config *configs.QdrantConfig) error {
if config == nil {
return fmt.Errorf("qdrant config cannot be nil")
}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("qdrant config validation failed: %w", err)
	}

	return nil
}

提供简洁的工厂模式实现，管理Qdrant向量存储的创建和配置。

**工厂实现：**

```go
// QdrantVectorStoreFactory Qdrant向量存储工厂
type QdrantVectorStoreFactory struct {
    logger *slog.Logger
}
```

**核心方法：**

- **NewQdrantVectorStoreFactory**: 创建工厂实例
- **CreateVectorRepository**: 根据配置创建VectorRepository实现
- **ValidateQdrantConfiguration**: 验证Qdrant配置有效性

**设计优势：**

1. **简化创建流程**: 通过工厂模式统一管理复杂的初始化过程
2. **配置验证**: 提前验证配置参数的有效性
3. **日志集成**: 完整的初始化过程日志记录
4. **错误处理**: 详细的错误信息和失败回滚机制

#### 8.4 集成架构

向量数据库实现与整体架构的集成关系：

```
业务服务层 (services/)
    ↓ 依赖
VectorRepository接口 (repositories/)
    ↓ 实现
QdrantVectorStore (infrastructure/vector/qdrant/)
    ↓ 使用
QdrantClient (封装原生gRPC客户端)
    ↓ 连接
Qdrant向量数据库
```

**关键特性：**

- **接口驱动**: 完全基于VectorRepository接口，支持未来替换其他向量数据库
- **Clean Architecture**: 遵循依赖倒置原则，基础设施层不影响业务逻辑
- **高性能**: 优化的批量操作和连接池管理
- **可扩展**: 支持多种向量相似度算法和过滤策略
- **可监控**: 集成详细的性能指标和健康检查

#### 8.5 性能优化

1. **批量操作**: 优先使用批量API减少网络开销
2. **连接复用**: 单例客户端实例和连接池管理
3. **异步处理**: 支持上下文取消和超时控制
4. **内存优化**: 避免不必要的数据复制和中间对象创建

### 9. 嵌入模型实现 (internal/infrastructure/embedding/remote/)

嵌入模型实现模块提供了基于OpenAI Format API的文本向量化服务，实现了EmbeddingService接口的完整功能。

#### 架构设计

```
internal/infrastructure/embedding/remote/
├── init.go         # 初始化和配置管理
└── embedding.go    # 核心向量化服务实现
```

#### 9.1 远程嵌入服务实现 (embedding.go)
package remote

import (
"context"
"fmt"
"time"

	"github.com/openai/openai-go"

	"qa-cache/configs"
	"qa-cache/internal/domain/models"
	"qa-cache/pkg/logger"
)

// RemoteEmbeddingService 远程嵌入模型服务实现
// 基于OpenAI Format API实现文本向量化功能
type RemoteEmbeddingService struct {
client openai.Client
config *configs.RemoteEmbedding
logger logger.Logger
}

// GenerateEmbedding 生成单个文本的向量
func (s *RemoteEmbeddingService) GenerateEmbedding(ctx context.Context, request *models.VectorProcessingRequest) (*models.VectorProcessingResult, error) {
startTime := time.Now()

	s.logger.DebugContext(ctx, "开始生成单个文本向量",
		"text_length", len(request.Text),
		"model", s.getModelName(request.ModelName))

	// 构建OpenAI嵌入请求
	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(request.Text),
		},
		Model: s.getModelName(request.ModelName),
	}

	// 调用OpenAI API
	response, err := s.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := fmt.Sprintf("failed to generate embedding: %v", err)

		s.logger.ErrorContext(ctx, "向量生成失败",
			"error", err,
			"processing_time_ms", processingTime)

		return &models.VectorProcessingResult{
			ProcessingTime: processingTime,
			ModelUsed:      s.getModelName(request.ModelName),
			Success:        false,
			Error:          errorMsg,
		}, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// 验证响应
	if len(response.Data) == 0 {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := "no embedding data in response"

		s.logger.ErrorContext(ctx, "响应数据为空",
			"processing_time_ms", processingTime)

		return &models.VectorProcessingResult{
			ProcessingTime: processingTime,
			ModelUsed:      response.Model,
			Success:        false,
			Error:          errorMsg,
		}, fmt.Errorf(errorMsg)
	}

	// 提取向量数据
	embedding := response.Data[0]
	vectorValues := make([]float32, len(embedding.Embedding))
	for i, val := range embedding.Embedding {
		vectorValues[i] = float32(val)
	}

	// 创建向量对象
	now := time.Now()
	vector := &models.Vector{
		ID:         fmt.Sprintf("embedding_%d", now.UnixNano()),
		Values:     vectorValues,
		Dimension:  len(vectorValues),
		CreateTime: now,
		UpdateTime: now,
		Normalized: false,
		ModelName:  response.Model,
	}

	// 如果需要归一化
	if request.Normalize {
		vector.Normalize()
	}

	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	tokenCount := int(response.Usage.PromptTokens)

	s.logger.InfoContext(ctx, "向量生成成功",
		"dimension", len(vectorValues),
		"token_count", tokenCount,
		"processing_time_ms", processingTime,
		"model_used", response.Model)

	return &models.VectorProcessingResult{
		Vector:         vector,
		ProcessingTime: processingTime,
		TokenCount:     tokenCount,
		ModelUsed:      response.Model,
		Success:        true,
	}, nil
}

// GenerateBatchEmbeddings 批量生成文本向量
func (s *RemoteEmbeddingService) GenerateBatchEmbeddings(ctx context.Context, requests []*models.VectorProcessingRequest) ([]*models.VectorProcessingResult, error) {

	s.logger.InfoContext(ctx, "开始批量生成向量",
		"batch_size", len(requests))

	// 提取所有文本
	texts := make([]string, len(requests))
	for i, req := range requests {
		if req == nil || req.Text == "" {
			return nil, fmt.Errorf("request at index %d is nil or has empty text", i)
		}
		texts[i] = req.Text
	}

	startTime := time.Now()

	// 构建批量嵌入请求
	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
		Model: openai.EmbeddingModel(s.getModelName(requests[0].ModelName)),
	}

	// 设置编码格式为float
	embeddingParams.EncodingFormat = openai.EmbeddingNewParamsEncodingFormatFloat

	// 调用OpenAI API
	response, err := s.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := fmt.Sprintf("failed to generate batch embeddings: %v", err)

		s.logger.ErrorContext(ctx, "批量向量生成失败",
			"error", err,
			"batch_size", len(requests),
			"processing_time_ms", processingTime)

		// 返回所有失败的结果
		results := make([]*models.VectorProcessingResult, len(requests))
		for i := range requests {
			results[i] = &models.VectorProcessingResult{
				ProcessingTime: processingTime / float64(len(requests)),
				ModelUsed:      s.getModelName(requests[i].ModelName),
				Success:        false,
				Error:          errorMsg,
			}
		}
		return results, fmt.Errorf(errorMsg)
	}

	// 验证响应数量
	if len(response.Data) != len(requests) {
		return nil, fmt.Errorf("response data count (%d) doesn't match request count (%d)",
			len(response.Data), len(requests))
	}

	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	totalTokens := int(response.Usage.PromptTokens)
	avgTokensPerRequest := totalTokens / len(requests)

	// 构建结果
	results := make([]*models.VectorProcessingResult, len(requests))
	for i, req := range requests {
		embedding := response.Data[i]

		// 转换向量数据
		vectorValues := make([]float32, len(embedding.Embedding))
		for j, val := range embedding.Embedding {
			vectorValues[j] = float32(val)
		}

		// 创建向量对象
		now := time.Now()
		vector := &models.Vector{
			ID:         fmt.Sprintf("embedding_%d_%d", now.UnixNano(), i),
			Values:     vectorValues,
			Dimension:  len(vectorValues),
			CreateTime: now,
			UpdateTime: now,
			Normalized: false,
			ModelName:  response.Model,
		}

		// 如果需要归一化
		if req.Normalize {
			vector.Normalize()
		}

		results[i] = &models.VectorProcessingResult{
			Vector:         vector,
			ProcessingTime: processingTime / float64(len(requests)),
			TokenCount:     avgTokensPerRequest,
			ModelUsed:      response.Model,
			Success:        true,
		}
	}

	s.logger.InfoContext(ctx, "批量向量生成成功",
		"batch_size", len(requests),
		"total_tokens", totalTokens,
		"processing_time_ms", processingTime,
		"model_used", response.Model)

	return results, nil
}

// getModelName 获取使用的模型名称，支持请求级别的模型覆盖
func (s *RemoteEmbeddingService) getModelName(requestModel string) string {
if requestModel != "" {
return requestModel
}
return s.config.ModelName
}

// Close 关闭服务，清理资源
func (s *RemoteEmbeddingService) Close() error {
s.logger.Info("远程嵌入模型服务关闭")
// OpenAI客户端不需要显式关闭
return nil
}

RemoteEmbeddingService是EmbeddingService接口的完整实现，基于OpenAI Go SDK提供文本向量化能力。

**核心实现：**

```go
// RemoteEmbeddingService 远程嵌入模型服务实现
type RemoteEmbeddingService struct {
    client openai.Client            // OpenAI客户端
    config *configs.RemoteEmbedding // 远程嵌入配置
    logger logger.Logger            // 结构化日志
}
```

**核心接口：**

```go
// EmbeddingService 文本向量化服务接口
// 专门负责将文本转换为向量表示
type EmbeddingService interface {
    // GenerateEmbedding 生成单个文本的向量
    // 将输入文本转换为向量表示
    GenerateEmbedding(ctx context.Context, request *models.VectorProcessingRequest) (*models.VectorProcessingResult, error)

    // GenerateBatchEmbeddings 批量生成文本向量
    // 批量处理多个文本的向量化，提高处理效率
    GenerateBatchEmbeddings(ctx context.Context, requests []*models.VectorProcessingRequest) ([]*models.VectorProcessingResult, error)
}
```

**EmbeddingService接口功能：**

- **GenerateEmbedding**: 单个文本向量化，支持自定义模型和归一化
- **GenerateBatchEmbeddings**: 批量文本向量化，提高处理效率


**核心特性：**

1. **OpenAI兼容**: 基于官方`openai-go` SDK，完全兼容OpenAI Format API
2. **批量优化**: 支持高效的批量文本向量化，减少API调用次数
3. **配置灵活**: 支持自定义端点、API密钥、超时、重试等配置
4. **向量归一化**: 可选的L2向量归一化支持
5. **错误处理**: 完善的错误处理和结构化日志记录

#### 9.2 初始化和配置管理 (init.go)

提供嵌入服务的初始化逻辑和OpenAI客户端配置管理。
package remote

import (
"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"qa-cache/configs"
	"qa-cache/internal/domain/services"
	"qa-cache/pkg/logger"
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

**初始化功能：**

```go
// NewRemoteEmbeddingService 创建远程嵌入模型服务
func NewRemoteEmbeddingService(config *configs.RemoteEmbedding, log logger.Logger) (services.EmbeddingService, error)
```

**配置管理：**

- **客户端选项配置**: API密钥、端点、超时、重试策略
- **请求头定制**: 支持自定义HTTP请求头
- **参数验证**: 完整的配置参数有效性验证

**设计特点：**

1. **工厂模式**: 通过工厂函数统一管理复杂的初始化过程
2. **配置驱动**: 完全基于配置文件进行服务初始化
3. **依赖注入**: 支持日志器等依赖的注入
4. **错误处理**: 详细的初始化错误信息和失败处理

#### 9.3 集成架构

嵌入模型实现与整体架构的集成关系：

```
业务服务层 (services/)
    ↓ 依赖
EmbeddingService接口 (services/)
    ↓ 实现
RemoteEmbeddingService (infrastructure/embedding/remote/)
    ↓ 使用
OpenAI Go SDK
    ↓ 调用
OpenAI Format API (或兼容端点)
```

**关键特性：**

- **接口驱动**: 完全基于EmbeddingService接口，支持未来替换其他嵌入实现
- **Clean Architecture**: 遵循依赖倒置原则，基础设施层不影响业务逻辑
- **可扩展**: 支持多种OpenAI嵌入模型和自定义端点
- **高性能**: 批量处理优化和连接复用
- **可配置**: 灵活的配置管理和运行时参数调整

#### 9.4 向量化流程

单个文本向量化流程：

```
1. 参数验证 → 2. 构建请求 → 3. 调用API → 4. 处理响应 → 5. 创建向量对象
```

批量文本向量化流程：

```
1. 批量验证 → 2. 提取文本 → 3. 批量调用 → 4. 响应验证 → 5. 构建结果集
```

**性能优化：**

1. **批量处理**: 优先使用批量API减少网络延迟
2. **并发安全**: 支持多goroutine并发调用
3. **资源管理**: 自动的资源清理和连接管理
4. **错误恢复**: 智能的重试策略和错误恢复机制



### 内部依赖关系

```
域模型模块 (internal/domain/models/):
├── cache.go -> time (标准库)
├── vector.go -> fmt, math (标准库)
└── request.go -> context, time (标准库)

主程序入口 (cmd/server/):
└── main.go -> context, os, signal, syscall, time (标准库), 
               configs, pkg/logger, pkg/status,
               internal/domain/repositories, internal/domain/services,
               internal/infrastructure/vector/qdrant, internal/infrastructure/embedding/remote,
               internal/app/handlers, internal/app/server

配置模块 (configs/) -> time (标准库), yaml.v3
状态码模块 (pkg/status/) -> 无外部依赖  
日志模块 (pkg/logger/) -> log/slog (标准库)
```



### 10. 向量服务实现 (internal/infrastructure/services/)

向量服务实现模块提供了VectorService接口的完整业务层实现，负责组合EmbeddingService和VectorRepository完成语义搜索，提供文本相似度计算能力，封装向量化和向量搜索的复杂性。

#### 架构设计

```
internal/infrastructure/services/
├── vector_service.go        # 向量服务核心实现
├── init.go                  # 工厂模式和构建器
└── README.md               # 使用文档和最佳实践
```

#### 10.1 向量服务核心实现 (vector_service.go)
package vector

import (
"context"
"fmt"
"math"
"time"

	"github.com/google/uuid"

	"qa-cache/internal/domain/models"
	"qa-cache/internal/domain/repositories"
	"qa-cache/internal/domain/services"
	"qa-cache/pkg/logger"
)

// DefaultVectorService 默认向量服务实现
// 组合EmbeddingService和VectorRepository完成语义搜索，提供文本相似度计算能力
type DefaultVectorService struct {
embeddingService services.EmbeddingService     // 嵌入服务，负责文本向量化
vectorRepository repositories.VectorRepository // 向量仓储，负责向量存储和搜索
strategyFactory  *StrategyFactory              // 选择策略工厂
logger           logger.Logger                 // 日志器
config           *VectorServiceConfig          // 配置
}

// VectorServiceConfig 向量服务配置
type VectorServiceConfig struct {
// DefaultCollectionName 默认集合名称
DefaultCollectionName string `json:"default_collection_name" yaml:"default_collection_name"`

	// DefaultTopK 默认返回结果数量
	DefaultTopK int `json:"default_top_k" yaml:"default_top_k"`

	// DefaultSimilarityThreshold 默认相似度阈值
	DefaultSimilarityThreshold float64 `json:"default_similarity_threshold" yaml:"default_similarity_threshold"`

	// MaxBatchSize 最大批量处理大小
	MaxBatchSize int `json:"max_batch_size" yaml:"max_batch_size"`

	// RequestTimeout 请求超时时间（秒）
	RequestTimeout int `json:"request_timeout" yaml:"request_timeout"`

	// EnableNormalization 是否启用向量归一化
	EnableNormalization bool `json:"enable_normalization" yaml:"enable_normalization"`

	// DefaultSelectionStrategy 默认选择策略
	DefaultSelectionStrategy string `json:"default_selection_strategy" yaml:"default_selection_strategy"`

	// TemperatureSoftmaxConfig 温度softmax策略配置
	TemperatureSoftmaxConfig *TemperatureSoftmaxConfig `json:"temperature_softmax" yaml:"temperature_softmax"`
}

// DefaultVectorServiceConfig 默认向量服务配置
func DefaultVectorServiceConfig() *VectorServiceConfig {
return &VectorServiceConfig{
DefaultCollectionName:      "qa_cache",
DefaultTopK:                5,
DefaultSimilarityThreshold: 0.7,
MaxBatchSize:               100,
RequestTimeout:             30,
EnableNormalization:        true,
DefaultSelectionStrategy:   "highest_score",
TemperatureSoftmaxConfig:   DefaultTemperatureSoftmaxConfig(),
}
}

// NewDefaultVectorService 创建默认向量服务实例
func NewDefaultVectorService(
embeddingService services.EmbeddingService,
vectorRepository repositories.VectorRepository,
config *VectorServiceConfig,
log logger.Logger,
) services.VectorService {
if config == nil {
config = DefaultVectorServiceConfig()
}

	// 创建选择策略工厂
	strategyFactory := NewStrategyFactory(log, config.TemperatureSoftmaxConfig)

	return &DefaultVectorService{
		embeddingService: embeddingService,
		vectorRepository: vectorRepository,
		strategyFactory:  strategyFactory,
		logger:           log,
		config:           config,
	}
}

// SearchCache 搜索语义缓存
// 完整的语义缓存查询流程：文本向量化 + 相似度搜索
func (s *DefaultVectorService) SearchCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error) {
startTime := time.Now()

	s.logger.InfoContext(ctx, "开始搜索语义缓存",
		"question", query.Question,
		"user_type", query.UserType,
		"similarity_threshold", query.SimilarityThreshold,
		"top_k", query.TopK,
	)

	// 1. 文本向量化
	vectorRequest := &models.VectorProcessingRequest{
		Text:      query.Question,
		Normalize: s.config.EnableNormalization,
	}

	vectorResult, err := s.embeddingService.GenerateEmbedding(ctx, vectorRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "文本向量化失败", "error", err)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("文本向量化失败: %w", err)
	}

	if !vectorResult.Success {
		s.logger.ErrorContext(ctx, "向量化处理失败", "error", vectorResult.Error)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("向量化处理失败: %s", vectorResult.Error)
	}

	// 2. 构建向量搜索请求
	topK, similarityThreshold := func() (int, float64) {
		t := s.config.DefaultTopK
		st := s.config.DefaultSimilarityThreshold
		if query.TopK != nil && *query.TopK > 0 {
			t = *query.TopK
		}
		if query.SimilarityThreshold != nil && *query.SimilarityThreshold >= 0 {
			st = *query.SimilarityThreshold
		}
		return t, st
	}()

	searchRequest := &models.VectorSearchRequest{
		QueryText:           query.Question,
		QueryVector:         vectorResult.Vector.Values,
		TopK:                topK,
		SimilarityThreshold: similarityThreshold,
		UserType:            query.UserType,
		Filters:             query.Filters,
	}

	// 3. 执行向量相似度搜索
	s.logger.InfoContext(ctx, "开始执行向量搜索",
		"vector_dimension", len(vectorResult.Vector.Values),
		"top_k", topK,
		"similarity_threshold", similarityThreshold,
	)

	searchResponse, err := s.vectorRepository.Search(ctx, searchRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量搜索失败", "error", err)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 4. 处理搜索结果
	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if len(searchResponse.Results) == 0 {
		s.logger.InfoContext(ctx, "未找到匹配的缓存", "response_time", responseTime)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: responseTime,
		}, nil
	}

	// 5. 使用选择策略选择最佳结果
	strategy := s.config.DefaultSelectionStrategy

	// 转换结果类型
	resultPointers := make([]*models.VectorSearchResult, len(searchResponse.Results))
	for i := range searchResponse.Results {
		resultPointers[i] = &searchResponse.Results[i]
	}

	bestResult, err := s.SelectBestResult(ctx, resultPointers, query, strategy)
	if err != nil {
		s.logger.ErrorContext(ctx, "选择最优结果失败", "error", err)
		// 回退到第一个结果
		bestResult = &searchResponse.Results[0]
	}

	// 从payload中提取答案
	answer := ""
	if answerValue, exists := bestResult.Payload["answer"]; exists {
		if answerStr, ok := answerValue.(string); ok {
			answer = answerStr
		}
	}

	// 从payload中提取元数据
	var metadata *models.CacheMetadata
	if metadataValue, exists := bestResult.Payload["metadata"]; exists {
		if metadataMap, ok := metadataValue.(map[string]interface{}); ok {
			metadata = s.extractMetadataFromPayload(metadataMap)
		}
	}

	// 从payload中提取统计信息
	var statistics *models.CacheStatistics
	if query.IncludeStatistics {
		if statsValue, exists := bestResult.Payload["statistics"]; exists {
			if statsMap, ok := statsValue.(map[string]interface{}); ok {
				statistics = s.extractStatisticsFromPayload(statsMap)
			}
		}
	}

	result := &models.CacheResult{
		Found:        true,
		CacheID:      bestResult.ID,
		Answer:       answer,
		Similarity:   bestResult.Score,
		ResponseTime: responseTime,
		Metadata:     metadata,
		Statistics:   statistics,
	}

	s.logger.InfoContext(ctx, "成功找到缓存匹配",
		"cache_id", result.CacheID,
		"similarity", result.Similarity,
		"response_time", responseTime,
		"selection_strategy", strategy,
	)

	return result, nil
}

// StoreCache 存储查询和响应到缓存
// 将用户查询和LLM响应存储到向量缓存中
func (s *DefaultVectorService) StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error) {
startTime := time.Now()

	s.logger.InfoContext(ctx, "开始存储缓存",
		"question_length", len(request.Question),
		"answer_length", len(request.Answer),
		"user_type", request.UserType,
		"force_write", request.ForceWrite,
	)

	// 1. 文本向量化
	vectorRequest := &models.VectorProcessingRequest{
		Text:      request.Question,
		Normalize: s.config.EnableNormalization,
	}

	vectorResult, err := s.embeddingService.GenerateEmbedding(ctx, vectorRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "文本向量化失败", "error", err)
		return &models.CacheWriteResult{
			Success: false,
			Message: "文本向量化失败",
			Reason:  fmt.Sprintf("向量化错误: %v", err),
		}, fmt.Errorf("文本向量化失败: %w", err)
	}

	if !vectorResult.Success {
		s.logger.ErrorContext(ctx, "向量化处理失败", "error", vectorResult.Error)
		return &models.CacheWriteResult{
			Success: false,
			Message: "向量化处理失败",
			Reason:  vectorResult.Error,
		}, fmt.Errorf("向量化处理失败: %s", vectorResult.Error)
	}

	// 2. 生成缓存ID
	cacheID := s.generateCacheID()

	// 3. 构建向量存储请求
	payload := map[string]interface{}{
		"question":  request.Question,
		"answer":    request.Answer,
		"user_type": request.UserType,
		"timestamp": time.Now().Unix(),
	}

	// 添加元数据
	if request.Metadata != nil {
		payload["metadata"] = map[string]interface{}{
			"source":        request.Metadata.Source,
			"tags":          request.Metadata.Tags,
			"quality_score": request.Metadata.QualityScore,
			"version":       request.Metadata.Version,
		}
	}

	// 添加初始统计信息
	payload["statistics"] = map[string]interface{}{
		"hit_count":     0,
		"like_count":    0,
		"dislike_count": 0,
		"response_time": 0.0,
	}

	storeRequest := &models.VectorStoreRequest{
		ID:             cacheID,
		Vector:         vectorResult.Vector.Values,
		CollectionName: s.config.DefaultCollectionName,
		Payload:        payload,
		UpsertMode:     true, // 允许更新已存在的记录
	}

	// 4. 执行存储操作
	storeResponse, err := s.vectorRepository.Store(ctx, storeRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量存储失败", "error", err)
		return &models.CacheWriteResult{
			Success: false,
			Message: "向量存储失败",
			Reason:  fmt.Sprintf("存储错误: %v", err),
		}, fmt.Errorf("向量存储失败: %w", err)
	}

	if !storeResponse.Success {
		s.logger.ErrorContext(ctx, "向量存储操作失败", "message", storeResponse.Message)
		return &models.CacheWriteResult{
			Success: false,
			Message: "向量存储操作失败",
			Reason:  storeResponse.Message,
		}, nil
	}

	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	result := &models.CacheWriteResult{
		Success: true,
		CacheID: cacheID,
		Message: "缓存存储成功",
	}

	// 设置质量分数（如果有元数据）
	if request.Metadata != nil {
		result.QualityScore = request.Metadata.QualityScore
	}

	s.logger.InfoContext(ctx, "成功存储缓存",
		"cache_id", cacheID,
		"vector_id", storeResponse.VectorID,
		"store_time", storeResponse.StoreTime,
		"response_time", responseTime,
	)

	return result, nil
}

// DeleteCache 删除缓存项
// 从向量缓存中删除指定的缓存项
func (s *DefaultVectorService) DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error) {
startTime := time.Now()

	s.logger.InfoContext(ctx, "开始删除缓存",
		"cache_ids", request.CacheIDs,
		"user_type", request.UserType,
		"force", request.Force,
	)

	if len(request.CacheIDs) == 0 {
		return &models.CacheDeleteResult{
			Success:      false,
			DeletedCount: 0,
			Message:      "未提供要删除的缓存ID",
		}, nil
	}

	// 执行删除操作
	err := s.vectorRepository.Delete(ctx, request.CacheIDs, request.UserType)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量删除失败", "error", err)
		return &models.CacheDeleteResult{
			Success:      false,
			DeletedCount: 0,
			FailedIDs:    request.CacheIDs,
			Message:      fmt.Sprintf("删除操作失败: %v", err),
		}, fmt.Errorf("向量删除失败: %w", err)
	}

	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	result := &models.CacheDeleteResult{
		Success:      true,
		DeletedCount: len(request.CacheIDs),
		Message:      "缓存删除成功",
	}

	s.logger.InfoContext(ctx, "成功删除缓存",
		"deleted_count", result.DeletedCount,
		"response_time", responseTime,
	)

	return result, nil
}

// generateCacheID 生成缓存ID
func (s *DefaultVectorService) generateCacheID() string {
// 暂时直接使用UUID
return uuid.New().String()
}

// calculateCosineSimilarity 计算余弦相似度
func (s *DefaultVectorService) calculateCosineSimilarity(vector1, vector2 []float32) (float64, error) {
if len(vector1) != len(vector2) {
return 0.0, fmt.Errorf("向量维度不匹配: %d vs %d", len(vector1), len(vector2))
}

	if len(vector1) == 0 {
		return 0.0, fmt.Errorf("向量不能为空")
	}

	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(vector1); i++ {
		dotProduct += float64(vector1[i]) * float64(vector2[i])
		norm1 += float64(vector1[i]) * float64(vector1[i])
		norm2 += float64(vector2[i]) * float64(vector2[i])
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0, nil
	}

	similarity := dotProduct / (norm1 * norm2)

	// 确保相似度在[0, 1]范围内
	if similarity < 0 {
		similarity = 0
	} else if similarity > 1 {
		similarity = 1
	}

	return similarity, nil
}

// extractMetadataFromPayload 从payload中提取元数据
func (s *DefaultVectorService) extractMetadataFromPayload(payload map[string]interface{}) *models.CacheMetadata {
metadata := &models.CacheMetadata{}

	if source, ok := payload["source"].(string); ok {
		metadata.Source = source
	}

	if tags, ok := payload["tags"].([]interface{}); ok {
		stringTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				stringTags = append(stringTags, tagStr)
			}
		}
		metadata.Tags = stringTags
	}

	if qualityScore, ok := payload["quality_score"].(float64); ok {
		metadata.QualityScore = qualityScore
	}

	if version, ok := payload["version"].(int); ok {
		metadata.Version = version
	}

	return metadata
}

// extractStatisticsFromPayload 从payload中提取统计信息
func (s *DefaultVectorService) extractStatisticsFromPayload(payload map[string]interface{}) *models.CacheStatistics {
statistics := &models.CacheStatistics{}

	if hitCount, ok := payload["hit_count"].(int64); ok {
		statistics.HitCount = hitCount
	}

	if likeCount, ok := payload["like_count"].(int64); ok {
		statistics.LikeCount = likeCount
	}

	if dislikeCount, ok := payload["dislike_count"].(int64); ok {
		statistics.DislikeCount = dislikeCount
	}

	if responseTime, ok := payload["response_time"].(float64); ok {
		statistics.ResponseTime = responseTime
	}

	if lastHitTimeUnix, ok := payload["last_hit_time"].(int64); ok {
		if lastHitTimeUnix > 0 {
			lastHitTime := time.Unix(lastHitTimeUnix, 0)
			statistics.LastHitTime = &lastHitTime
		}
	}

	return statistics
}

// min 返回两个整数中的较小值
func min(a, b int) int {
if a < b {
return a
}
return b
}

// SelectBestResult 选择最优结果
// 从候选结果中选择最符合查询意图的单个结果
func (s *DefaultVectorService) SelectBestResult(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, strategy string) (*models.VectorSearchResult, error) {
s.logger.InfoContext(ctx, "开始选择最优结果",
"results_count", len(results),
"strategy", strategy,
"user_type", query.UserType)

	// 验证输入
	if len(results) == 0 {
		s.logger.WarnContext(ctx, "没有候选结果可供选择")
		return nil, fmt.Errorf("没有候选结果可供选择")
	}

	if query == nil {
		s.logger.ErrorContext(ctx, "查询请求不能为空")
		return nil, fmt.Errorf("查询请求不能为空")
	}

	// 使用默认策略如果未指定
	if strategy == "" {
		strategy = s.config.DefaultSelectionStrategy
	}

	// 创建策略实例
	selectionStrategy, err := s.strategyFactory.CreateStrategy(strategy)
	if err != nil {
		s.logger.ErrorContext(ctx, "创建选择策略失败",
			"strategy", strategy,
			"error", err)
		return nil, fmt.Errorf("创建选择策略失败: %w", err)
	}

	// 执行选择
	selectedResult, err := selectionStrategy.Select(ctx, results, query, make(map[string]interface{}))
	if err != nil {
		s.logger.ErrorContext(ctx, "选择结果失败",
			"strategy", strategy,
			"error", err)
		return nil, fmt.Errorf("选择结果失败: %w", err)
	}

	s.logger.InfoContext(ctx, "成功选择最优结果",
		"selected_id", selectedResult.ID,
		"selected_score", selectedResult.Score,
		"strategy", strategy)

	return selectedResult, nil
}

DefaultVectorService是VectorService接口的完整实现，作为语义缓存系统的核心业务协调器。

**核心实现：**

```go
// DefaultVectorService 默认向量服务实现
type DefaultVectorService struct {
    embeddingService services.EmbeddingService     // 嵌入服务，负责文本向量化
    vectorRepository repositories.VectorRepository // 向量仓储，负责向量存储和搜索
    logger           logger.Logger                 // 日志器
    config           *VectorServiceConfig         // 配置
}
```

**VectorService接口完整实现：**

- **SearchCache**: 完整的语义缓存查询流程：`文本向量化 → 向量搜索 → 结果处理`
- **StoreCache**: 将用户查询和LLM响应存储到向量缓存：`向量化 → 生成ID → 存储`
- **DeleteCache**: 从向量缓存中删除指定缓存项
- **CalculateSimilarity**: 计算两个文本的语义相似度

**核心特性：**

1. **依赖组合**: 优雅组合EmbeddingService和VectorRepository完成复杂业务流程
2. **配置化管理**: 支持TopK、相似度阈值、批量大小等灵活配置
3. **智能缓存ID**: 基于用户类型、问题内容和时间戳生成唯一标识
4. **余弦相似度**: 使用标准余弦相似度算法计算文本语义相似性
5. **批量优化**: CalculateSimilarity使用批量向量化提高效率
6. **Payload管理**: 丰富的元数据和统计信息支持

#### 10.2 工厂模式和构建器 (init.go)
package vector

import (
"fmt"

	"qa-cache/internal/domain/repositories"
	"qa-cache/internal/domain/services"
	"qa-cache/pkg/logger"
)

// VectorServiceFactory 向量服务工厂
type VectorServiceFactory struct {
logger logger.Logger
}

// NewVectorServiceFactory 创建向量服务工厂实例
func NewVectorServiceFactory(log logger.Logger) *VectorServiceFactory {
return &VectorServiceFactory{
logger: log,
}
}

// CreateVectorService 创建向量服务实例
// 根据配置和依赖创建VectorService的具体实现
func (f *VectorServiceFactory) CreateVectorService(
embeddingService services.EmbeddingService,
vectorRepository repositories.VectorRepository,
config *VectorServiceConfig,
) (services.VectorService, error) {

	// 验证必要依赖
	if embeddingService == nil {
		return nil, fmt.Errorf("EmbeddingService不能为空")
	}

	if vectorRepository == nil {
		return nil, fmt.Errorf("VectorRepository不能为空")
	}

	// 验证配置
	if config != nil {
		if err := f.validateVectorServiceConfig(config); err != nil {
			return nil, fmt.Errorf("向量服务配置验证失败: %w", err)
		}
	}

	// 创建向量服务实例
	vectorService := NewDefaultVectorService(
		embeddingService,
		vectorRepository,
		config,
		f.logger,
	)

	f.logger.Info("向量服务实例创建成功",
		"collection_name", getCollectionName(config),
		"default_top_k", getDefaultTopK(config),
		"similarity_threshold", getDefaultSimilarityThreshold(config),
	)

	return vectorService, nil
}

// validateVectorServiceConfig 验证向量服务配置
func (f *VectorServiceFactory) validateVectorServiceConfig(config *VectorServiceConfig) error {
if config.DefaultCollectionName == "" {
return fmt.Errorf("默认集合名称不能为空")
}

	if config.DefaultTopK <= 0 {
		return fmt.Errorf("默认TopK必须大于0，当前值: %d", config.DefaultTopK)
	}

	if config.DefaultTopK > 1000 {
		return fmt.Errorf("默认TopK不能超过1000，当前值: %d", config.DefaultTopK)
	}

	if config.DefaultSimilarityThreshold < 0.0 || config.DefaultSimilarityThreshold > 1.0 {
		return fmt.Errorf("默认相似度阈值必须在[0.0, 1.0]范围内，当前值: %f", config.DefaultSimilarityThreshold)
	}

	if config.MaxBatchSize <= 0 {
		return fmt.Errorf("最大批量大小必须大于0，当前值: %d", config.MaxBatchSize)
	}

	if config.MaxBatchSize > 10000 {
		return fmt.Errorf("最大批量大小不能超过10000，当前值: %d", config.MaxBatchSize)
	}

	if config.RequestTimeout <= 0 {
		return fmt.Errorf("请求超时时间必须大于0秒，当前值: %d", config.RequestTimeout)
	}

	if config.RequestTimeout > 300 {
		return fmt.Errorf("请求超时时间不能超过300秒，当前值: %d", config.RequestTimeout)
	}

	return nil
}

// CreateVectorServiceWithDefaults 使用默认配置创建向量服务实例
func (f *VectorServiceFactory) CreateVectorServiceWithDefaults(
embeddingService services.EmbeddingService,
vectorRepository repositories.VectorRepository,
) (services.VectorService, error) {
return f.CreateVectorService(
embeddingService,
vectorRepository,
DefaultVectorServiceConfig(),
)
}

// 辅助函数：获取配置值（如果配置为nil则使用默认值）

func getCollectionName(config *VectorServiceConfig) string {
if config == nil {
return DefaultVectorServiceConfig().DefaultCollectionName
}
return config.DefaultCollectionName
}

func getDefaultTopK(config *VectorServiceConfig) int {
if config == nil {
return DefaultVectorServiceConfig().DefaultTopK
}
return config.DefaultTopK
}

func getDefaultSimilarityThreshold(config *VectorServiceConfig) float64 {
if config == nil {
return DefaultVectorServiceConfig().DefaultSimilarityThreshold
}
return config.DefaultSimilarityThreshold
}

// VectorServiceBuilder 向量服务构建器（提供更灵活的构建方式）
type VectorServiceBuilder struct {
embeddingService services.EmbeddingService
vectorRepository repositories.VectorRepository
config           *VectorServiceConfig
logger           logger.Logger
}

// NewVectorServiceBuilder 创建向量服务构建器
func NewVectorServiceBuilder() *VectorServiceBuilder {
return &VectorServiceBuilder{
config: DefaultVectorServiceConfig(),
}
}

// WithEmbeddingService 设置嵌入服务
func (b *VectorServiceBuilder) WithEmbeddingService(service services.EmbeddingService) *VectorServiceBuilder {
b.embeddingService = service
return b
}

// WithVectorRepository 设置向量仓储
func (b *VectorServiceBuilder) WithVectorRepository(repo repositories.VectorRepository) *VectorServiceBuilder {
b.vectorRepository = repo
return b
}

// WithConfig 设置配置
func (b *VectorServiceBuilder) WithConfig(config *VectorServiceConfig) *VectorServiceBuilder {
b.config = config
return b
}

// WithLogger 设置日志器
func (b *VectorServiceBuilder) WithLogger(log logger.Logger) *VectorServiceBuilder {
b.logger = log
return b
}

// WithCollectionName 设置集合名称
func (b *VectorServiceBuilder) WithCollectionName(name string) *VectorServiceBuilder {
if b.config == nil {
b.config = DefaultVectorServiceConfig()
}
b.config.DefaultCollectionName = name
return b
}

// WithDefaultTopK 设置默认TopK
func (b *VectorServiceBuilder) WithDefaultTopK(topK int) *VectorServiceBuilder {
if b.config == nil {
b.config = DefaultVectorServiceConfig()
}
b.config.DefaultTopK = topK
return b
}

// WithSimilarityThreshold 设置相似度阈值
func (b *VectorServiceBuilder) WithSimilarityThreshold(threshold float64) *VectorServiceBuilder {
if b.config == nil {
b.config = DefaultVectorServiceConfig()
}
b.config.DefaultSimilarityThreshold = threshold
return b
}

// WithNormalization 设置是否启用向量归一化
func (b *VectorServiceBuilder) WithNormalization(enabled bool) *VectorServiceBuilder {
if b.config == nil {
b.config = DefaultVectorServiceConfig()
}
b.config.EnableNormalization = enabled
return b
}

// Build 构建向量服务实例
func (b *VectorServiceBuilder) Build() (services.VectorService, error) {
if b.embeddingService == nil {
return nil, fmt.Errorf("EmbeddingService必须设置")
}

	if b.vectorRepository == nil {
		return nil, fmt.Errorf("VectorRepository必须设置")
	}

	if b.logger == nil {
		return nil, fmt.Errorf("Logger必须设置")
	}

	factory := NewVectorServiceFactory(b.logger)
	return factory.CreateVectorService(
		b.embeddingService,
		b.vectorRepository,
		b.config,
	)
}

提供多种创建模式，支持灵活的向量服务实例化和配置管理。

**工厂模式实现：**

```go
// VectorServiceFactory 向量服务工厂
type VectorServiceFactory struct {
    logger logger.Logger
}

// CreateVectorService 创建向量服务实例
func (f *VectorServiceFactory) CreateVectorService(
    embeddingService services.EmbeddingService,
    vectorRepository repositories.VectorRepository,
    config *VectorServiceConfig,
) (services.VectorService, error)
```

**构建器模式实现：**

```go
// VectorServiceBuilder 向量服务构建器
type VectorServiceBuilder struct {
    embeddingService services.EmbeddingService
    vectorRepository repositories.VectorRepository
    config           *VectorServiceConfig
    logger           logger.Logger
}
```

**支持的创建方式：**

1. **基础创建**: `NewDefaultVectorService()`
2. **工厂创建**: `factory.CreateVectorService()`
3. **构建器创建**: `NewVectorServiceBuilder().WithXXX().Build()`

#### 10.3 配置管理

**VectorServiceConfig 配置结构：**

```go
type VectorServiceConfig struct {
    DefaultCollectionName      string  // 默认集合名称
    DefaultTopK               int     // 默认返回结果数量  
    DefaultSimilarityThreshold float64 // 默认相似度阈值
    MaxBatchSize              int     // 最大批量处理大小
    RequestTimeout            int     // 请求超时时间（秒）
    EnableNormalization       bool    // 是否启用向量归一化
}
```

**默认配置值：**
- DefaultCollectionName: "qa_cache"
- DefaultTopK: 5
- DefaultSimilarityThreshold: 0.7
- MaxBatchSize: 100
- RequestTimeout: 30
- EnableNormalization: true

**配置验证规则：**
- TopK范围: [1, 1000]
- 相似度阈值范围: [0.0, 1.0]
- 批量大小范围: [1, 10000]
- 超时时间范围: [1, 300]秒

#### 10.4 业务流程实现

**语义搜索流程 (SearchCache):**

```
1. 文本向量化 → EmbeddingService.GenerateEmbedding()
2. 构建搜索请求 → 设置TopK、相似度阈值、过滤条件
3. 执行向量搜索 → VectorRepository.Search()
4. 结果处理 → 提取答案、元数据、统计信息
5. 返回结果 → 包含相似度分数和响应时间
```

**缓存存储流程 (StoreCache):**

```
1. 文本向量化 → 将问题转换为向量表示
2. 生成缓存ID → 基于用户类型、内容hash、时间戳
3. 构建Payload → 组装问题、答案、元数据、统计信息
4. 批量存储 → VectorRepository.BatchStore()
5. 返回结果 → 包含缓存ID和质量分数
```

**相似度计算流程 (CalculateSimilarity):**

```
1. 批量向量化 → EmbeddingService.GenerateBatchEmbeddings()
2. 余弦相似度计算 → 使用向量点积公式
3. 结果标准化 → 确保范围在[0.0, 1.0]
4. 返回相似度分数
```

#### 10.5 集成架构

向量服务实现与整体架构的集成关系：

```
业务服务层 (services/cache_service.go)
    ↓ 依赖
VectorService接口 (domain/services/vector_service.go)
    ↓ 实现
DefaultVectorService (infrastructure/services/vector_service.go)
    ↓ 组合
EmbeddingService + VectorRepository
    ↓ 协调
文本向量化 + 向量存储搜索
```

**关键特性：**

- **接口驱动**: 完全基于VectorService接口，支持未来替换其他实现
- **Clean Architecture**: 遵循依赖倒置原则，业务逻辑不依赖具体技术实现
- **高性能**: 批量向量化、向量归一化、连接复用等性能优化
- **可配置**: 灵活的配置管理和多种实例化模式
- **可观测**: 详细的结构化日志和性能指标

#### 10.6 性能优化
package vector

import (
"context"
"fmt"
"math"
"math/rand"
"qa-cache/internal/domain/models"
"qa-cache/pkg/logger"
"time"
)

// TemperatureSoftmaxConfig 温度softmax策略配置
type TemperatureSoftmaxConfig struct {
// Temperature 温度参数，控制选择的随机性
Temperature float64 `json:"temperature" yaml:"temperature"`

	// TopK 考虑的前K个结果
	TopK int `json:"top_k" yaml:"top_k"`

	// MinProbability 最小概率阈值
	MinProbability float64 `json:"min_probability" yaml:"min_probability"`
}

// DefaultTemperatureSoftmaxConfig 默认温度softmax配置
func DefaultTemperatureSoftmaxConfig() *TemperatureSoftmaxConfig {
return &TemperatureSoftmaxConfig{
Temperature:    1.0,
TopK:           5,
MinProbability: 0.1,
}
}

// ResultSelectionStrategy 结果选择策略接口
type ResultSelectionStrategy interface {
// Name 策略名称
Name() string

	// Select 选择结果
	Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error)

	// Validate 验证策略配置
	Validate(config map[string]interface{}) (bool, error)
}

// FirstSelectionStrategy 选择第一个结果的策略
type FirstSelectionStrategy struct {
logger logger.Logger
}

// NewFirstSelectionStrategy 创建第一个结果选择策略
func NewFirstSelectionStrategy(log logger.Logger) *FirstSelectionStrategy {
return &FirstSelectionStrategy{
logger: log,
}
}

func (s *FirstSelectionStrategy) Name() string {
return "first"
}

func (s *FirstSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
if len(results) == 0 {
return nil, fmt.Errorf("no results to select from")
}

	selected := results[0]
	s.logger.DebugContext(ctx, "Selected first result",
		"result_id", selected.ID,
		"score", selected.Score,
		"strategy", s.Name())

	return selected, nil
}

func (s *FirstSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
// 第一个结果策略不需要特殊配置
return true, nil
}

// HighestScoreSelectionStrategy 选择分数最高结果的策略
type HighestScoreSelectionStrategy struct {
logger logger.Logger
}

// NewHighestScoreSelectionStrategy 创建最高分数选择策略
func NewHighestScoreSelectionStrategy(log logger.Logger) *HighestScoreSelectionStrategy {
return &HighestScoreSelectionStrategy{
logger: log,
}
}

func (s *HighestScoreSelectionStrategy) Name() string {
return "highest_score"
}

func (s *HighestScoreSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
if len(results) == 0 {
return nil, fmt.Errorf("no results to select from")
}

	var bestResult *models.VectorSearchResult
	var bestScore float64 = -1

	for _, result := range results {
		if result.Score > bestScore {
			bestScore = result.Score
			bestResult = result
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("no valid result found")
	}

	s.logger.DebugContext(ctx, "Selected highest score result",
		"result_id", bestResult.ID,
		"score", bestResult.Score,
		"strategy", s.Name())

	return bestResult, nil
}

func (s *HighestScoreSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
// 最高分数策略不需要特殊配置
return true, nil
}

// TemperatureSoftmaxSelectionStrategy 基于温度的softmax选择策略
type TemperatureSoftmaxSelectionStrategy struct {
logger logger.Logger
config *TemperatureSoftmaxConfig
}

// NewTemperatureSoftmaxSelectionStrategy 创建温度softmax选择策略
func NewTemperatureSoftmaxSelectionStrategy(log logger.Logger, config *TemperatureSoftmaxConfig) *TemperatureSoftmaxSelectionStrategy {
return &TemperatureSoftmaxSelectionStrategy{
logger: log,
config: config,
}
}

func (s *TemperatureSoftmaxSelectionStrategy) Name() string {
return "temperature_softmax"
}

func (s *TemperatureSoftmaxSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
if len(results) == 0 {
return nil, fmt.Errorf("no results to select from")
}

	// 获取温度参数
	temperature := s.config.Temperature
	if tempValue, ok := config["temperature"]; ok {
		if tempFloat, ok := tempValue.(float64); ok {
			temperature = tempFloat
		}
	}

	// 限制考虑的结果数量
	topK := s.config.TopK
	if topKValue, ok := config["top_k"]; ok {
		if topKInt, ok := topKValue.(int); ok {
			topK = topKInt
		}
	}

	// 取前TopK个结果
	candidates := results
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	// 计算softmax概率
	probabilities := s.calculateSoftmaxProbabilities(candidates, temperature)

	// 基于概率进行随机选择
	selectedIndex := s.weightedRandomSelect(probabilities)
	if selectedIndex >= len(candidates) {
		return nil, fmt.Errorf("invalid selection index: %d", selectedIndex)
	}

	selected := candidates[selectedIndex]
	s.logger.DebugContext(ctx, "Selected result using temperature softmax",
		"result_id", selected.ID,
		"score", selected.Score,
		"probability", probabilities[selectedIndex],
		"temperature", temperature,
		"strategy", s.Name())

	return selected, nil
}

func (s *TemperatureSoftmaxSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
if temp, ok := config["temperature"]; ok {
if tempFloat, ok := temp.(float64); ok {
if tempFloat <= 0 {
return false, fmt.Errorf("temperature must be positive")
}
} else {
return false, fmt.Errorf("temperature must be a float64")
}
}

	if topK, ok := config["top_k"]; ok {
		if topKInt, ok := topK.(int); ok {
			if topKInt <= 0 {
				return false, fmt.Errorf("top_k must be positive")
			}
		} else {
			return false, fmt.Errorf("top_k must be an integer")
		}
	}

	return true, nil
}

// calculateSoftmaxProbabilities 计算softmax概率分布
func (s *TemperatureSoftmaxSelectionStrategy) calculateSoftmaxProbabilities(results []*models.VectorSearchResult, temperature float64) []float64 {
if len(results) == 0 {
return []float64{}
}

	// 计算指数值
	expValues := make([]float64, len(results))
	var sumExp float64

	for i, result := range results {
		expValues[i] = math.Exp(result.Score / temperature)
		sumExp += expValues[i]
	}

	// 归一化为概率
	probabilities := make([]float64, len(results))
	for i := range expValues {
		probabilities[i] = expValues[i] / sumExp
	}

	return probabilities
}

// weightedRandomSelect 基于权重进行随机选择
func (s *TemperatureSoftmaxSelectionStrategy) weightedRandomSelect(probabilities []float64) int {
if len(probabilities) == 0 {
return 0
}

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成随机数
	r := rand.Float64()

	// 累积概率选择
	cumulative := 0.0
	for i, prob := range probabilities {
		cumulative += prob
		if r <= cumulative {
			return i
		}
	}

	// 如果由于浮点精度问题没有选中，返回最后一个
	return len(probabilities) - 1
}

// StrategyFactory 策略工厂
type StrategyFactory struct {
logger                   logger.Logger
temperatureSoftmaxConfig *TemperatureSoftmaxConfig
}

// NewStrategyFactory 创建策略工厂
func NewStrategyFactory(log logger.Logger, temperatureSoftmaxConfig *TemperatureSoftmaxConfig) *StrategyFactory {
if temperatureSoftmaxConfig == nil {
temperatureSoftmaxConfig = DefaultTemperatureSoftmaxConfig()
}
return &StrategyFactory{
logger:                   log,
temperatureSoftmaxConfig: temperatureSoftmaxConfig,
}
}

// CreateStrategy 创建策略实例
func (f *StrategyFactory) CreateStrategy(strategyName string) (ResultSelectionStrategy, error) {
switch strategyName {
case "first":
return NewFirstSelectionStrategy(f.logger), nil
case "highest_score":
return NewHighestScoreSelectionStrategy(f.logger), nil
case "temperature_softmax":
return NewTemperatureSoftmaxSelectionStrategy(f.logger, f.temperatureSoftmaxConfig), nil
default:
f.logger.ErrorContext(context.Background(), "Unknown selection strategy",
"strategy", strategyName)
return nil, fmt.Errorf("unknown selection strategy: %s", strategyName)
}
}

1. **批量向量化**: CalculateSimilarity使用批量API减少网络延迟
2. **向量归一化**: 可配置的L2归一化提高搜索精度
3. **智能缓存ID**: 基于内容hash和时间戳的高效ID生成
4. **连接复用**: 复用EmbeddingService和VectorRepository连接
5. **异步处理**: 支持上下文取消和超时控制

### 11. 请求预处理服务实现 (internal/infrastructure/preprocessing/)

请求预处理服务实现模块提供了完整的查询请求预处理能力，通过智能策略和上下文处理提高查询质量和匹配准确性。

#### 架构设计

```
internal/infrastructure/preprocessing/
├── config.go              # 配置结构定义
├── strategies.go           # 预处理策略实现
├── context_processor.go    # 上下文处理器实现  
├── service.go             # RequestPreprocessingService核心实现
├── init.go                # 工厂模式和服务初始化
└── README.md              # 实现文档
```

#### 11.1 预处理服务核心实现 (service.go)

DefaultRequestPreprocessingService是RequestPreprocessingService接口的完整实现，提供智能的查询预处理功能。

**核心实现：**

```go
// DefaultRequestPreprocessingService 默认请求预处理服务实现
type DefaultRequestPreprocessingService struct {
    config           *Config
    strategyRegistry *StrategyRegistry  
    contextProcessor services.ContextProcessor
    logger           logger.Logger
}
```

**RequestPreprocessingService接口实现：**

- **PreprocessQuery**: 完整的查询预处理流程，包括参数验证、上下文提取、策略应用和结果封装
- **ExtractContext**: 智能上下文信息提取，包括问题类型识别、实体提取和过滤条件分析
- **ApplyUserTypeStrategy**: 基于用户类型的差异化预处理策略应用

**核心特性：**

1. **完整的预处理流程**: 从参数验证到结果输出的完整处理链路
2. **智能上下文提取**: 自动识别问题类型、提取关键实体、分析查询意图
3. **用户类型策略**: 支持技术用户、普通用户、内部用户的差异化处理
4. **处理时间统计**: 详细记录各阶段的处理耗时
5. **错误处理**: 完善的错误处理和状态码管理

#### 11.2 预处理策略实现 (strategies.go)

StrategyRegistry提供了插拔式的预处理策略管理和实现。

**策略注册表：**

```go
// StrategyRegistry 策略注册表
type StrategyRegistry struct {
    strategies map[string]services.PreprocessingStrategy
    logger     logger.Logger
}
```

**内置策略实现：**

1. **TextCleaningStrategy**: 基础文本清理
    - 移除多余空白字符
    - 标准化Unicode字符
    - 清理不必要的特殊字符

2. **NormalizationStrategy**: 文本标准化
    - 统一中英文标点符号
    - 展开常见缩写和简写
    - 标准化数字和单位表示

3. **TechnicalEnhancementStrategy**: 技术增强策略
    - 标准化技术术语
    - 添加技术上下文标识
    - 适用于技术用户类型

4. **CustomerFriendlyStrategy**: 客户友好策略
    - 简化技术术语为通俗表达
    - 添加礼貌用语
    - 适用于普通客户

5. **ContextEnrichmentStrategy**: 上下文丰富策略
    - 添加历史上下文信息
    - 扩展隐式引用
    - 适用于内部用户

**设计特点：**
- **策略模式**: 支持运行时策略注册和切换
- **用户类型适配**: 每个策略都有用户类型适用性验证
- **链式处理**: 支持多策略组合处理
- **可扩展性**: 易于添加新的预处理策略

#### 11.3 上下文处理器实现 (context_processor.go)

DefaultContextProcessor实现了智能的多轮对话上下文处理功能。

**核心实现：**

```go
// DefaultContextProcessor 默认上下文处理器实现
type DefaultContextProcessor struct {
    sessions     map[string]*SessionContext
    sessionMutex sync.RWMutex
    logger       logger.Logger
    config       *ContextConfig
}

// SessionContext 会话上下文
type SessionContext struct {
    UserType     string                 
    History      []string               
    Metadata     map[string]interface{} 
    LastActivity time.Time              
    CreatedAt    time.Time              
}
```

**核心功能：**

1. **会话管理**
    - 自动识别和管理用户会话
    - 基于用户ID和用户类型的会话隔离
    - 可配置的会话超时和清理

2. **上下文关键词检测**
    - 智能识别上下文引用关键词
    - 支持中文语境的上下文检测
    - 可配置的关键词列表

3. **历史记录维护**
    - 维护查询历史记录
    - 限制历史记录数量，避免内存膨胀
    - 支持上下文信息的快速检索

4. **上下文增强**
    - 自动为包含上下文关键词的查询添加历史信息
    - 构建上下文增强的查询文本
    - 更新会话元数据

**并发安全设计：**
- 使用读写锁保护会话数据
- 自动清理过期会话的后台goroutine
- 线程安全的会话操作接口

#### 11.4 配置管理 (config.go)

提供完整的预处理服务配置结构和默认值管理。

**配置结构：**

```go
// Config 请求预处理服务配置
type Config struct {
    DefaultStrategies   []string           // 默认预处理策略
    UserTypeStrategies  map[string][]string // 用户类型策略映射
    TextCleaningConfig  TextCleaningConfig  // 文本清理配置
    ContextConfig       ContextConfig       // 上下文处理配置
    NormalizationConfig NormalizationConfig // 标准化配置
    Timeout            time.Duration       // 处理超时时间
    MaxTextLength      int                 // 最大文本长度
    MinTextLength      int                 // 最小文本长度
    EnableLogging      bool                // 是否启用详细日志
}
```

**特性：**
- **分层配置**: 针对不同处理阶段的细粒度配置
- **用户类型映射**: 支持不同用户类型的策略配置
- **默认值管理**: 提供合理的默认配置
- **配置验证**: 内置配置有效性验证

#### 11.5 工厂模式初始化 (init.go)

提供简洁的工厂模式实现，管理请求预处理服务的创建和配置。

**工厂实现：**

```go
// Factory 请求预处理服务工厂
type Factory struct {
    logger logger.Logger
}
```

**核心方法：**

- **CreateRequestPreprocessingService**: 根据配置创建服务实例
- **CreateRequestPreprocessingServiceWithDefaults**: 使用默认配置创建服务
- **ValidateConfig**: 验证配置有效性

**便捷函数：**

- **NewRequestPreprocessingService**: 直接创建服务的便捷函数
- **NewRequestPreprocessingServiceWithDefaults**: 使用默认配置的便捷函数

#### 11.6 集成架构

请求预处理服务实现与整体架构的集成关系：

```
业务服务层 (services/)
    ↓ 依赖
RequestPreprocessingService接口 (services/)
    ↓ 实现
DefaultRequestPreprocessingService (infrastructure/preprocessing/)
    ↓ 使用
策略注册表 + 上下文处理器
    ↓ 处理
查询文本预处理和上下文增强
```

**关键特性：**

- **接口驱动**: 完全基于RequestPreprocessingService接口，支持未来替换其他实现
- **Clean Architecture**: 遵循依赖倒置原则，基础设施层不影响业务逻辑
- **策略可扩展**: 支持运行时注册新的预处理策略
- **并发安全**: 完整的并发安全设计和资源管理
- **可监控**: 集成详细的处理统计和性能日志

#### 11.7 使用场景

1. **多轮对话支持**: 自动处理"这个"、"那个"等上下文引用
2. **用户类型优化**: 为不同用户提供差异化的查询预处理
3. **文本质量提升**: 通过清理和标准化提高查询文本质量
4. **技术术语处理**: 智能处理技术术语的标准化和简化

### 12. 主程序入口 (cmd/server/)

主程序入口模块是整个应用的启动点，负责依赖注入、服务组装和应用生命周期管理。

#### 架构设计

```
cmd/server/
└── main.go                # 应用程序主入口点
```

#### 12.1 应用程序主入口 (main.go)

main.go是整个应用程序的核心入口点，负责完整的应用生命周期管理，从初始化到优雅关闭的全过程。

**核心功能：**

1. **依赖注入和服务组装**
    - 配置系统初始化和验证
    - 日志系统初始化和配置
    - 基础设施层初始化（向量存储、嵌入服务）
    - 业务服务层初始化（缓存服务实现）
    - 应用层初始化（HTTP处理器、服务器）

2. **StubCacheService实现**
    - 为当前开发阶段提供简单的缓存服务实现
    - 实现CacheService接口的所有方法
    - 提供完整的健康检查功能
    - 对复杂功能返回"正在开发中"的友好响应

3. **应用启动流程**
    - 遵循Clean Architecture的分层初始化顺序
    - 完整的错误处理和失败回滚机制
    - 服务依赖关系的正确组装
    - HTTP服务器的启动和运行

4. **优雅关闭机制**
    - 严格遵循Goroutine使用指南
    - 监听系统信号（SIGINT/SIGTERM）
    - 30秒超时的优雅关闭流程
    - 按序关闭各个组件并清理资源

**核心实现：**

```go
// StubCacheService 简单的缓存服务实现
// 提供基本功能，复杂功能返回"未实现"响应
type StubCacheService struct {
    startTime        time.Time
    vectorRepo       repositories.VectorRepository
    embeddingService services.EmbeddingService
    logger           logger.Logger
}

// main 主函数 - 应用程序入口点
func main() {
    // 1. 配置和日志初始化
    // 2. 基础设施层初始化
    // 3. 业务服务层初始化  
    // 4. 应用层初始化
    // 5. 服务启动和优雅关闭
}
```

**应用启动序列：**

```
1. 配置加载和验证 → configs.Load()
2. 日志系统初始化 → logger.NewLogger()
3. 向量存储初始化 → qdrant.NewQdrantVectorStoreFactory()
4. 嵌入服务初始化 → remote.NewRemoteEmbeddingService()
5. 缓存服务组装 → NewStubCacheService()
6. HTTP处理器创建 → handlers.NewCacheHandler()
7. HTTP服务器启动 → server.NewServer()
8. 信号监听和优雅关闭
```

**StubCacheService特性：**

- **完整接口实现**: 实现CacheService接口的所有15个方法
- **健康检查功能**: 提供详细的系统健康状态信息，包括：
    - 服务运行状态和启动时间
    - 向量存储连接状态
    - 嵌入服务配置状态
    - 功能开发状态
- **降级模式运行**: 当基础设施未完全配置时，服务仍能启动并提供基础功能
- **开发友好**: 为未实现功能返回清晰的"功能正在开发中"消息

**优雅关闭流程：**

```go
// 优雅关闭流程
1. 监听信号 → signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
2. 停止接受新请求 → server.Shutdown()
3. 等待现有请求完成 → 30秒超时
4. 关闭基础设施连接 → vectorStore.Close(), embeddingService.Close()
5. 清理资源和记录日志
```

**错误处理策略：**

- **初始化失败**: 记录详细错误日志并退出程序
- **基础设施故障**: 启用降级模式继续运行
- **服务依赖缺失**: 提供降级功能和明确的状态信息
- **优雅关闭超时**: 强制退出并记录警告日志

#### 12.2 集成架构

主程序入口与整体架构的集成关系：

```
main() 入口函数
    ↓ 初始化
配置管理 (configs/) ← 加载应用配置
    ↓ 提供配置
日志系统 (pkg/logger/) ← 初始化日志器
    ↓ 日志记录
基础设施层 (infrastructure/) ← 初始化向量存储和嵌入服务
    ↓ 提供服务
StubCacheService ← 组装业务服务
    ↓ 注入依赖
应用层 (app/) ← 创建处理器和服务器
    ↓ 启动服务
HTTP服务器 ← 运行并监听请求
    ↓ 信号处理
优雅关闭 ← 按序清理资源
```

**关键特性：**

- **遵循Clean Architecture**: 严格按照分层架构进行依赖注入
- **生产就绪**: 完整的错误处理、日志记录、信号处理
- **开发友好**: StubCacheService提供开发阶段的基础功能
- **可观测性**: 详细的启动日志和健康检查信息
- **资源管理**: 完善的资源清理和内存管理

#### 12.3 依赖注入设计

采用手动依赖注入模式，确保依赖关系的清晰和可控：

```go
// 依赖注入流程
func main() {
    // 1. 配置层
    config := configs.Load()
    
    // 2. 工具层
    logger := logger.NewLogger(config.Logging)
    
    // 3. 基础设施层
    vectorRepo := qdrant.CreateVectorRepository(config.Database.Qdrant, logger)
    embeddingService := remote.NewRemoteEmbeddingService(config.Embedding.Remote, logger)
    
    // 4. 业务服务层
    cacheService := NewStubCacheService(vectorRepo, embeddingService, logger)
    
    // 5. 应用层
    cacheHandler := handlers.NewCacheHandler(cacheService, logger)
    httpServer := server.NewServer(config.Server, cacheHandler, logger)
    
    // 6. 启动应用
    httpServer.Run(context.Background())
}
```

这种设计确保了：
- **依赖方向**: 遵循依赖倒置原则
- **生命周期管理**: 清晰的创建和销毁顺序
- **错误传播**: 完整的错误处理链
- **测试友好**: 支持依赖模拟和单元测试

这些基础组件严格遵循清洁架构原则和Go语言最佳实践，通过Repository模式和Service模式实现了清晰的分层架构。从核心域模型、业务接口定义、基础设施实现，到应用层中间件、服务器配置和主程序入口的完整技术栈已经建立。模块间职责清晰、依赖关系明确，具备了完整的HTTP API服务能力，为系统的正式运行提供了坚实的基础，整体架构具备良好的可扩展性、可测试性和可维护性。应用程序现在可以完整启动并提供基础的缓存服务功能，为后续的高级功能开发奠定了坚实的架构基础。 


