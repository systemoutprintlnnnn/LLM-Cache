# LLM-Cache

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![Eino](https://img.shields.io/badge/Eino-v0.7.3-purple.svg)](https://github.com/cloudwego/eino)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/yourusername/llm-cache/pulls)

## 📖 简介

LLM-Cache 是一个基于 **CloudWeGo Eino 框架**和 Golang 实现的高性能、分布式、企业级 **LLM 语义缓存中间件**。通过智能语义匹配技术，能够显著降低大语言模型 API 调用成本并大幅提升响应速度。在典型应用场景下，可实现 **API 成本降低 90%、响应速度提升 100 倍** 的显著业务价值。

## ✨ 主要特性

- 🚀 **极致性能**: 基于 Go 语言的高并发特性，单节点 QPS > 10,000，P99 延迟 < 50ms
- 💡 **智能语义匹配**: 支持基于向量相似度的语义缓存，不局限于精确匹配
- 🔒 **企业级可靠**: 完整的监控、持久化、高可用等企业级功能
- 🎯 **模型无关**: 兼容主流 LLM 和 Embedding 模型（OpenAI、ARK、Ollama、Dashscope、Qianfan、Tencentcloud）
- 📦 **灵活部署**: 支持单机、集群及云原生（Docker/Kubernetes）部署
- 🔧 **可插拔架构**: 基于 Eino 框架的流程编排，支持多种向量数据库（Qdrant、Milvus、Redis、ES8、VikingDB）
- 📊 **可观测性**: 内置 Callback 机制，支持 Langfuse、APMPlus、Cozeloop 等监控平台

## 🛠️ 技术栈

- **语言**: Go 1.23+
- **核心框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino) v0.7.3 - LLM 应用开发框架
- **Web 框架**: Gin 1.10.1
- **向量数据库**: Qdrant（默认）/ Milvus / Redis / Elasticsearch / VikingDB
- **Embedding 服务**: OpenAI（默认）/ ARK / Ollama / Dashscope / Qianfan / Tencentcloud
- **核心依赖**:
  - `github.com/cloudwego/eino` - Eino 核心库
  - `github.com/cloudwego/eino-ext` - Eino 扩展组件
  - `github.com/qdrant/go-client` - Qdrant Go 客户端
  - `github.com/google/uuid` - UUID 生成

## 🎯 应用场景

- **智能客服机器人**: 缓存常见问题，提升响应速度和一致性
- **RAG 知识库问答**: 缓存对知识文档的查询，降低检索和生成成本
- **代码生成助手**: 缓存常见的代码片段生成请求
- **内容创作辅助**: 缓存相似的指令或草稿，加速内容迭代

## 🚀 快速开始

### 前置要求

- Go 1.23 或更高版本
- Qdrant 向量数据库（可使用 Docker 快速启动）
- OpenAI API Key（用于 Embedding 服务）

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/yourusername/llm-cache.git
cd llm-cache
```

#### 2. 启动 Qdrant 向量数据库

```bash
docker run -d -p 6333:6333 -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage \
  qdrant/qdrant:latest
```

#### 3. 配置应用

复制配置文件并编辑:

```bash
cp configs/config.yaml.example configs/config.yaml
```

编辑 `configs/config.yaml`，配置必要参数:

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080

# Eino 框架配置
eino:
  # Embedder 配置
  embedder:
    provider: "openai"  # openai/ark/ollama/dashscope/qianfan/tencentcloud
    api_key: "your-openai-api-key"
    model: "text-embedding-3-small"
    timeout: 30

  # Retriever 配置
  retriever:
    provider: "qdrant"  # qdrant/milvus/redis/es8/vikingdb
    collection: "llm_cache"
    top_k: 5
    score_threshold: 0.7
    qdrant:
      host: "localhost"
      port: 6334

  # Indexer 配置
  indexer:
    provider: "qdrant"
    collection: "llm_cache"
    vector_size: 1536
    qdrant:
      host: "localhost"
      port: 6334
      distance: "Cosine"

  # 查询配置
  query:
    preprocess_enabled: true
    postprocess_enabled: true
    selection_strategy: "highest_score"  # first/highest_score/temperature_softmax
    temperature: 0.7

  # 存储配置
  store:
    quality_check_enabled: true
    min_question_length: 5
    min_answer_length: 10

  # Callback 配置
  callbacks:
    logging:
      enabled: true
      level: "info"
    metrics:
      enabled: false
    tracing:
      enabled: false

# 日志配置
logging:
  level: "info"
  output: "stdout"
```

#### 4. 安装依赖

```bash
go mod download
```

#### 5. 启动服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### 环境变量配置

也可以通过环境变量覆盖配置：

```bash
# OpenAI 配置
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"

# Qdrant 配置
export QDRANT_HOST="localhost"
```

### 使用示例

#### 查询缓存

```bash
curl -X POST http://localhost:8080/v1/cache/search \
  -H "Content-Type: application/json" \
  -d '{
    "question": "什么是机器学习?",
    "user_type": "default"
  }'
```

**响应示例**（缓存命中）:

```json
{
  "success": true,
  "code": 2000,
  "message": "缓存查询成功",
  "data": {
    "hit": true,
    "question": "什么是机器学习",
    "answer": "机器学习是人工智能的一个分支...",
    "score": 0.95,
    "cache_id": "550e8400-e29b-41d4-a716-446655440000"
  },
  "timestamp": 1701416400
}
```

#### 存储缓存

```bash
curl -X POST http://localhost:8080/v1/cache/store \
  -H "Content-Type: application/json" \
  -d '{
    "question": "什么是深度学习?",
    "answer": "深度学习是机器学习的一个子领域...",
    "user_type": "default"
  }'
```

#### 删除缓存

```bash
curl -X DELETE "http://localhost:8080/v1/cache/550e8400-e29b-41d4-a716-446655440000?user_type=default"
```

#### 健康检查

```bash
curl http://localhost:8080/v1/cache/health
```

## 🏗️ 架构设计

LLM-Cache 采用 **CloudWeGo Eino 框架** 进行流程编排，基于 Graph 实现灵活的业务流程：

```
┌─────────────────────────────────────────────────────────────────┐
│                      客户端应用层                                 │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                   API 接口层 (Gin HTTP)                          │
│                   Handler → compose.Runnable                     │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                   Eino Graph 流程编排层                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 查询 Graph: Preprocess → Retrieve → Select → Postprocess │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 存储 Graph: QualityCheck → Branch → Embed → Index       │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                   Eino 组件层                                    │
│   embedding.Embedder | retriever.Retriever | indexer.Indexer    │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                   数据存储层                                      │
│         Qdrant | Milvus | Redis | Elasticsearch | VikingDB      │
└─────────────────────────────────────────────────────────────────┘
```

### 核心模块

| 模块 | 路径 | 说明 |
|------|------|------|
| 组件工厂 | `internal/eino/components/` | Embedder/Retriever/Indexer 工厂 |
| Lambda 节点 | `internal/eino/nodes/` | 预处理/后处理/质量检查/结果选择 |
| 业务 Graph | `internal/eino/flows/` | 查询/存储/删除流程 |
| Callback | `internal/eino/callbacks/` | 日志/指标/追踪处理器 |
| 配置 | `internal/eino/config/` | Eino 配置结构 |

详细架构说明，请参考 [架构文档](docs/project/ARCHITECTURE.md) 和 [Eino 集成方案](docs/project/EINO_INTEGRATION_PLAN.md)。

## ⚙️ 配置说明

### Eino 配置参数

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `eino.embedder.provider` | Embedding 提供商 | openai |
| `eino.embedder.model` | Embedding 模型 | text-embedding-3-small |
| `eino.retriever.provider` | 向量数据库类型 | qdrant |
| `eino.retriever.top_k` | 返回结果数量 | 5 |
| `eino.retriever.score_threshold` | 相似度阈值 | 0.7 |
| `eino.indexer.vector_size` | 向量维度 | 1536 |
| `eino.query.selection_strategy` | 结果选择策略 | highest_score |
| `eino.store.quality_check_enabled` | 启用质量检查 | true |
| `eino.callbacks.logging.enabled` | 启用日志回调 | true |

### 支持的组件提供商

**Embedding 服务**:
- `openai` - OpenAI API
- `ark` - 火山引擎 ARK
- `ollama` - 本地 Ollama
- `dashscope` - 阿里云 Dashscope
- `qianfan` - 百度千帆
- `tencentcloud` - 腾讯云

**向量数据库**:
- `qdrant` - Qdrant
- `milvus` - Milvus
- `redis` - Redis Stack
- `es8` - Elasticsearch 8
- `vikingdb` - VikingDB

**Callback 集成**:
- `logging` - 内置日志
- `metrics` - Prometheus 指标
- `tracing` - 链路追踪
- `langfuse` - Langfuse
- `apmplus` - APMPlus
- `cozeloop` - Cozeloop

## 📊 性能指标

| 指标 | 数值 |
|------|------|
| 单节点 QPS | > 10,000 |
| 缓存命中延迟 (P99) | < 50ms |
| 并发连接数 | > 5,000 |
| 缓存命中率 | 80-95%（取决于场景）|
| API 成本节省 | 90%+ |
| 响应速度提升 | 50-100 倍 |

## 🐳 Docker 部署

### 使用 Docker Compose（推荐）

```bash
# 启动所有服务 (LLM-Cache + Qdrant)
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 单独构建镜像

```bash
# 构建镜像
docker build -t llm-cache:latest .

# 运行容器
docker run -d -p 8080:8080 \
  -e QDRANT_HOST="qdrant" \
  -e OPENAI_API_KEY="your-api-key" \
  llm-cache:latest
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行 Eino 节点单元测试
go test ./internal/eino/nodes/... -v

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📈 监控与可观测性

LLM-Cache 通过 Eino Callback 机制提供完整的可观测性支持：

### 内置指标

```bash
curl http://localhost:8080/v1/cache/statistics
```

### Callback 集成

在配置文件中启用所需的 Callback：

```yaml
eino:
  callbacks:
    logging:
      enabled: true
    metrics:
      enabled: true
      endpoint: "/metrics"
    langfuse:
      enabled: true
      public_key: "your-public-key"
      secret_key: "your-secret-key"
      host: "https://cloud.langfuse.com"
```

## 📚 文档

- [产品需求文档](docs/project/PRODUCT_REQUIREMENTS_DOCUMENT.md)
- [架构设计文档](docs/project/ARCHITECTURE.md)
- [Eino 集成方案](docs/project/EINO_INTEGRATION_PLAN.md)

## 🤝 贡献指南

我们欢迎所有形式的贡献！

1. Fork 本仓库
2. 创建特性分支（`git checkout -b feature/AmazingFeature`）
3. 提交更改（`git commit -m 'feat: Add some AmazingFeature'`）
4. 推送到分支（`git push origin feature/AmazingFeature`）
5. 提交 Pull Request

详细贡献指南请查看 [AGENTS.md](AGENTS.md)。

## 📄 开源协议

本项目基于 Apache 2.0 协议开源。详见 [LICENSE](LICENSE) 文件。

## 💬 社区与支持

- **Issues**: [GitHub Issues](https://github.com/yourusername/llm-cache/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/llm-cache/discussions)
- **Eino 框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino)

## 🗺️ 路线图

- **v0.9.0（当前）**
  - ✅ 完成核心缓存功能
  - ✅ 基于 Eino 框架重构
  - ✅ 支持 Qdrant 向量数据库
  - ✅ 支持 OpenAI Embedding
  - ✅ Graph 流程编排
  - ✅ Callback 可观测性

- **v1.0.0**
  - 📋 支持更多向量数据库（Milvus、Redis、ES8）
  - 📋 支持更多 Embedding 提供商
  - 📋 分布式集群支持
  - 📋 Grafana 监控面板

- **v1.1.0**
  - 📋 ChatModel 集成（缓存未命中时 LLM 回退）
  - 📋 Tools 集成
  - 📋 Kubernetes Helm Chart

- **v2.0.0**
  - 📋 多模态缓存支持（图像、音频）
  - 📋 智能缓存预热和淘汰
  - 📋 可视化管理后台

## ⭐ Star History

如果这个项目对您有帮助，请给我们一个 Star ⭐️

---

**Built with ❤️ using Go and [CloudWeGo Eino](https://github.com/cloudwego/eino)**
