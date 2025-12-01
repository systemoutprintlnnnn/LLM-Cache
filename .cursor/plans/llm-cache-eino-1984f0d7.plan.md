<!-- 1984f0d7-23ce-401b-ba32-a7c8ab1f37b3 7863e38b-d50d-4794-82e7-4b318b60f295 -->
# LLM-Cache Eino 框架完全重构计划

## 重构原则

1. **删除所有自定义接口** - `internal/domain/services/` 和 `internal/domain/repositories/` 下的所有接口定义将被删除
2. **直接使用 Eino 类型** - 业务代码直接依赖 `embedding.Embedder`、`retriever.Retriever`、`compose.Runnable` 等 Eino 原生类型
3. **不创建适配器** - 不存在"包装 Eino 实现自定义接口"的情况
4. **Handler 层重构** - HTTP Handler 直接依赖 `compose.Runnable` 而非自定义 Service 接口

## 重构范围

### 保留层（调整依赖类型）

- `cmd/server/main.go` - 重写依赖注入，使用 Eino 组件
- `internal/app/handlers/` - 调整依赖类型为 Eino 原生类型
- `internal/app/server/` - 路由保留
- `internal/app/middleware/` - 保留
- `internal/domain/models/` - 保留领域模型（CacheItem 等）
- `pkg/logger/` 和 `pkg/status/` - 保留
- `configs/` - 扩展配置结构，支持 Eino 配置

### 删除层（完全移除）

- `internal/domain/services/` - 6 个文件（embedding_service.go, vector_service.go, cache_service.go, quality_service.go, request_preprocessing_service.go, recall_postprocessing_service.go）
- `internal/domain/repositories/` - vector_repository.go
- `internal/infrastructure/embedding/` - 远程 Embedding 实现
- `internal/infrastructure/stores/qdrant/` - Qdrant 存储实现
- `internal/infrastructure/vector/` - 向量服务实现
- `internal/infrastructure/cache/` - 缓存服务实现
- `internal/infrastructure/quality/` - 质量评估实现
- `internal/infrastructure/preprocessing/` - 预处理实现
- `internal/infrastructure/postprocessing/` - 后处理实现

## 阶段一：基础组件替换（2-3周）

### 1.1 添加 Eino 依赖

- 更新 `go.mod` 添加 `github.com/cloudwego/eino` 和 `github.com/cloudwego/eino-ext`
- 添加必要的向量数据库客户端依赖（qdrant、milvus 等）

### 1.2 创建 Eino 配置结构

- 新建 `internal/eino/config/config.go`，定义 Embedder/Retriever/Indexer 配置结构
- 扩展 `configs/config.go` 添加 `EinoConfig` 字段

### 1.3 创建组件工厂

- `internal/eino/components/embedder.go` - 支持 OpenAI/ARK/Ollama/Dashscope/Qianfan/Tencentcloud 等
- `internal/eino/components/retriever.go` - 支持 Qdrant/Milvus/Redis/ES8/VikingDB 等
- `internal/eino/components/indexer.go` - 对应 Indexer 工厂

### 1.4 删除旧实现

- 删除 `internal/domain/services/embedding_service.go`
- 删除 `internal/domain/repositories/vector_repository.go`
- 删除 `internal/infrastructure/embedding/` 目录
- 删除 `internal/infrastructure/stores/` 目录

## 阶段二：流程编排重构（3-4周）

### 2.1 创建 Lambda 节点

- `internal/eino/nodes/preprocessing.go` - 查询预处理节点
- `internal/eino/nodes/postprocessing.go` - 结果后处理节点
- `internal/eino/nodes/quality_check.go` - 质量检查节点
- `internal/eino/nodes/result_select.go` - 结果选择节点（支持 first/highest_score/temperature_softmax）

### 2.2 创建业务 Graph

- `internal/eino/flows/cache_query.go` - 缓存查询 Graph
- 流程：START -> Preprocess -> Embedding -> Retrieve -> Select -> Postprocess -> END
- `internal/eino/flows/cache_store.go` - 缓存存储 Graph（带条件分支）
- 流程：START -> QualityCheck -> Branch(通过/拒绝) -> Embedding -> Index -> END
- `internal/eino/flows/cache_delete.go` - 删除流程

### 2.3 重构 Handler 层

- 修改 `internal/app/handlers/cache_handler.go`：
- 改为依赖 `compose.Runnable[*QueryInput, *QueryOutput]`
- 改为依赖 `compose.Runnable[*StoreInput, *StoreOutput]`
- 改为依赖 `compose.Runnable[*DeleteInput, *DeleteOutput]`

### 2.4 重写 main.go

- 使用 Eino 组件初始化（Embedder/Retriever/Indexer）
- 编译 Graph 为 Runnable 传入 Handler
- 删除所有旧的 Service 初始化代码

### 2.5 删除所有旧 Service 接口和实现

- 删除 `internal/domain/services/*.go`（6个文件）
- 删除 `internal/infrastructure/vector/` 目录
- 删除 `internal/infrastructure/cache/` 目录
- 删除 `internal/infrastructure/quality/` 目录
- 删除 `internal/infrastructure/preprocessing/` 目录
- 删除 `internal/infrastructure/postprocessing/` 目录

## 阶段三：可观测性增强（1-2周）

### 3.1 创建 Callback 处理器

- `internal/eino/callbacks/logging.go` - 组件执行日志
- `internal/eino/callbacks/metrics.go` - Prometheus 指标
- `internal/eino/callbacks/tracing.go` - 链路追踪
- `internal/eino/callbacks/factory.go` - Callback 工厂（支持 Langfuse/APMPlus/Cozeloop）

### 3.2 集成 Callback 到 Graph

- 在 Graph 编译时注入 Callback 处理器
- 配置化 Callback 启用/禁用

## 阶段四：功能扩展（持续）

### 4.1 ChatModel 集成（可选）

- 缓存未命中时 LLM 回退
- 支持 OpenAI/ARK/Ollama/Qwen 等

### 4.2 Tools 集成（可选）

- 缓存内容摘要工具
- 相似度分析工具

## 关键技术决策

| 决策项 | 选择 | 理由 |
|-------|------|------|
| 向量数据库 | Qdrant (默认) | 现有基础，Eino-ext 支持良好 |
| Embedding | OpenAI (默认) | 现有配置，支持多提供商切换 |
| 流程编排 | Eino Graph | 支持条件分支和并行执行 |
| 结果选择 | Lambda 节点 | 保留现有 3 种策略 |

## 验收标准

- [ ] 所有 API 端点功能正常
- [ ] 单元测试覆盖率 >= 80%
- [ ] P99 延迟：查询 < 100ms，存储 < 200ms
- [ ] 项目中无任何自定义 Embedding/Vector/Cache 接口定义
- [ ] Handler 层仅依赖 Eino 原生类型

## 文件变更清单

### 新建文件

- `internal/eino/config/config.go`
- `internal/eino/components/embedder.go`
- `internal/eino/components/retriever.go`
- `internal/eino/components/indexer.go`
- `internal/eino/nodes/preprocessing.go`
- `internal/eino/nodes/postprocessing.go`
- `internal/eino/nodes/quality_check.go`
- `internal/eino/nodes/result_select.go`
- `internal/eino/flows/cache_query.go`
- `internal/eino/flows/cache_store.go`
- `internal/eino/flows/cache_delete.go`
- `internal/eino/callbacks/logging.go`
- `internal/eino/callbacks/metrics.go`
- `internal/eino/callbacks/tracing.go`
- `internal/eino/callbacks/factory.go`

### 修改文件

- `go.mod` - 添加 Eino 依赖
- `configs/config.go` - 扩展 Eino 配置
- `cmd/server/main.go` - 重写依赖注入
- `internal/app/handlers/cache_handler.go` - 改为依赖 compose.Runnable

### 删除文件/目录

- `internal/domain/services/*.go` (6个文件)
- `internal/domain/repositories/vector_repository.go`
- `internal/infrastructure/embedding/`
- `internal/infrastructure/stores/`
- `internal/infrastructure/vector/`
- `internal/infrastructure/cache/`
- `internal/infrastructure/quality/`
- `internal/infrastructure/preprocessing/`
- `internal/infrastructure/postprocessing/`

### To-dos

- [ ] 添加 Eino 核心依赖 (eino + eino-ext) 到 go.mod
- [ ] 创建 internal/eino/config/ 目录和 Eino 配置结构
- [ ] 扩展 configs/config.go 添加 EinoConfig 配置段
- [ ] 创建 internal/eino/components/embedder.go 组件工厂（支持 OpenAI/ARK/Ollama/Dashscope/Qianfan/Tencentcloud）
- [ ] 创建 internal/eino/components/retriever.go 组件工厂（支持 Qdrant/Milvus/Redis/ES8/VikingDB）
- [ ] 创建 internal/eino/components/indexer.go 组件工厂
- [ ] 删除 domain/services/embedding_service.go 和 infrastructure/embedding/
- [ ] 删除 domain/repositories/ 和 infrastructure/stores/
- [ ] 创建 internal/eino/nodes/ 目录及 Lambda 节点实现（preprocessing/postprocessing/quality_check/result_select）
- [ ] 创建 internal/eino/flows/cache_query.go 查询 Graph
- [ ] 创建 internal/eino/flows/cache_store.go 存储 Graph（带条件分支）
- [ ] 创建 internal/eino/flows/cache_delete.go 删除流程
- [ ] 重构 cache_handler.go 依赖 compose.Runnable
- [ ] 重写 cmd/server/main.go 使用 Eino 组件初始化
- [ ] 删除剩余的 domain/services/ 和 infrastructure/ 目录（vector/cache/quality/preprocessing/postprocessing）
- [ ] 创建 internal/eino/callbacks/ 目录及 Callback 实现（logging/metrics/tracing/factory）
- [ ] 在 Graph 编译时集成 Callback 处理器
- [ ] 为新的 Eino 组件和 Graph 添加单元测试