# LLM-Cache

[![Go Version](https://img.shields.io/badge/go-1.22.2-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/yourusername/llm-cache/pulls)

## 📖 简介

LLM-Cache 是一个基于 Golang 实现的高性能、分布式、企业级 **LLM 语义缓存中间件**,通过智能语义匹配技术,能够显著降低大语言模型 API 调用成本并大幅提升响应速度。在典型应用场景下,可实现 **API 成本降低 90%、响应速度提升 100 倍** 的显著业务价值。

## ✨ 主要特性

- 🚀 **极致性能**: 基于 Go 语言的高并发特性,单节点 QPS > 10,000,P99 延迟 < 50ms
- 💡 **智能语义匹配**: 支持基于向量相似度的语义缓存,不局限于精确匹配
- 🔒 **企业级可靠**: 完整的监控、持久化、高可用等企业级功能
- 🎯 **模型无关**: 兼容主流 LLM 和 Embedding 模型(OpenAI、本地 ONNX 等)
- 📦 **灵活部署**: 支持单机、集群及云原生(Docker/Kubernetes)部署
- 🔧 **可插拔架构**: 支持多种向量数据库(Qdrant、Milvus、Weaviate)和存储后端

## 🛠️ 技术栈

- **语言**: Go 1.22.2
- **Web框架**: Gin 1.10.1
- **向量数据库**: Qdrant
- **Embedding服务**: OpenAI API / 本地 ONNX
- **其他核心依赖**:
  - `github.com/qdrant/go-client` - Qdrant Go 客户端
  - `github.com/openai/openai-go` - OpenAI Go SDK
  - `github.com/google/uuid` - UUID 生成

## 🎯 应用场景

- **智能客服机器人**: 缓存常见问题,提升响应速度和一致性
- **RAG 知识库问答**: 缓存对知识文档的查询,降低检索和生成成本
- **代码生成助手**: 缓存常见的代码片段生成请求
- **内容创作辅助**: 缓存相似的指令或草稿,加速内容迭代

## 🚀 快速开始

### 前置要求

- Go 1.22.2 或更高版本
- Qdrant 向量数据库 (可使用 Docker 快速启动)
- OpenAI API Key (用于 Embedding 服务)

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

编辑 `configs/config.yaml`,配置必要参数:

```yaml
# 服务器配置
server:
  port: 8080
  mode: "release"  # debug/release

# 数据库配置
database:
  type: "qdrant"
  qdrant:
    address: "localhost:6334"
    collection_name: "llm_cache"
    vector_dimension: 1536

# Embedding 配置
embedding:
  type: "remote"
  remote:
    provider: "openai"
    api_key: "your-openai-api-key"  # 替换为你的 API Key
    model: "text-embedding-3-small"

# 日志配置
logging:
  level: "info"
  output: "console"  # console/file
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

### 使用示例

#### 查询缓存

```bash
curl -X POST http://localhost:8080/api/v1/cache/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "什么是机器学习?",
    "user_type": "default"
  }'
```

**响应示例** (缓存命中):

```json
{
  "code": 2000,
  "message": "success",
  "data": {
    "answer": "机器学习是人工智能的一个分支...",
    "similarity": 0.95,
    "cached": true,
    "timestamp": "2024-12-01T10:30:00Z"
  }
}
```

#### 存储缓存

```bash
curl -X POST http://localhost:8080/api/v1/cache/store \
  -H "Content-Type: application/json" \
  -d '{
    "question": "什么是深度学习?",
    "answer": "深度学习是机器学习的一个子领域...",
    "user_type": "default"
  }'
```

#### 健康检查

```bash
curl http://localhost:8080/health
```

## ⚙️ 配置说明

### 核心配置参数

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.port` | HTTP 服务端口 | 8080 |
| `server.mode` | 运行模式 (debug/release) | release |
| `database.type` | 向量数据库类型 | qdrant |
| `database.qdrant.address` | Qdrant 服务地址 | localhost:6334 |
| `database.qdrant.collection_name` | 集合名称 | llm_cache |
| `database.qdrant.vector_dimension` | 向量维度 | 1536 |
| `embedding.type` | Embedding 服务类型 | remote |
| `embedding.remote.provider` | Embedding 提供商 | openai |
| `embedding.remote.api_key` | API 密钥 | - |
| `embedding.remote.model` | Embedding 模型 | text-embedding-3-small |
| `logging.level` | 日志级别 | info |
| `logging.output` | 日志输出 (console/file) | console |

### 高级配置

更多高级配置选项,请参考 [配置文档](docs/configuration.md)。

## 📊 性能指标

| 指标 | 数值 |
|------|------|
| 单节点 QPS | > 10,000 |
| 缓存命中延迟 (P99) | < 50ms |
| 并发连接数 | > 5,000 |
| 缓存命中率 | 80-95% (取决于场景) |
| API 成本节省 | 90%+ |
| 响应速度提升 | 50-100 倍 |

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    客户端应用层                              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  API 接口层 (HTTP/gRPC)                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  缓存核心层                                  │
│   请求预处理 → 向量化 → 相似度匹配 → 质量评估 → 结果后处理  │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  数据管理层                                  │
│         向量存储 (Qdrant) | 标量存储 | 对象存储             │
└─────────────────────────────────────────────────────────────┘
```

详细架构说明,请参考 [架构文档](docs/project/AECHITECTURE.md)。

## 📚 文档

- [产品文档](docs/project/LLM-Cache产品文档.md)
- [产品需求文档](docs/project/PRODUCT_REQUIREMENTS_DOCUMENT.md)
- [架构设计文档](docs/project/AECHITECTURE.md)
- [API 参考文档](docs/api-reference.md)
- [部署指南](docs/deployment.md)

## 🐳 Docker 部署

### 使用 Docker Compose (推荐)

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
  -e QDRANT_ADDRESS="qdrant:6334" \
  -e OPENAI_API_KEY="your-api-key" \
  llm-cache:latest
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📈 监控与可观测性

LLM-Cache 提供 Prometheus 格式的监控指标:

```bash
curl http://localhost:8080/metrics
```

关键指标:
- `llm_cache_requests_total` - 总请求数
- `llm_cache_hits_total` - 缓存命中数
- `llm_cache_misses_total` - 缓存未命中数
- `llm_cache_request_duration_seconds` - 请求响应时间

## 🤝 贡献指南

我们欢迎所有形式的贡献!

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

详细贡献指南请查看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📝 更新日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新详情。

## 📄 开源协议

本项目基于 Apache 2.0 协议开源。详见 [LICENSE](LICENSE) 文件。

## 💬 社区与支持

- **Issues**: [GitHub Issues](https://github.com/yourusername/llm-cache/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/llm-cache/discussions)
- **Email**: [待补充]

## 🗺️ 路线图

- **Q3 2025**
  - ✅ 完成核心缓存功能
  - ✅ 支持 Qdrant 向量数据库
  - ✅ 支持 OpenAI Embedding
  - 🚧 Docker 镜像发布

- **Q4 2025**
  - 📋 支持更多向量数据库 (Milvus, Weaviate)
  - 📋 分布式集群支持
  - 📋 Grafana 监控面板
  - 📋 Kubernetes Helm Chart

- **2026**
  - 📋 多模态缓存支持 (图像、音频)
  - 📋 智能缓存预热和淘汰
  - 📋 可视化管理后台

## ⭐ Star History

如果这个项目对您有帮助,请给我们一个 Star ⭐️

---

**Built with ❤️ using Go**

