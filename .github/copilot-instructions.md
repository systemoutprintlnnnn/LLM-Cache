# LLM-Cache Copilot Instructions

## Architecture Overview

LLM-Cache is an **LLM semantic caching middleware** built on the **Eino framework** (ByteDance's LLM orchestration library). The architecture uses Eino's graph-based composition for processing pipelines.

```
cmd/server/main.go           → Dependency wiring: Eino components → Graphs → Handlers → Server
internal/eino/
  components/                → Eino component factories (Embedder, Retriever, Indexer)
  flows/                     → Graph definitions: CacheQueryGraph, CacheStoreGraph
  nodes/                     → Lambda nodes: preprocessing, quality_check, result_select
  config/                    → Eino-specific config (EinoConfig struct)
internal/app/                → HTTP layer (Gin handlers, routes)
configs/                     → YAML loading; main Config embeds EinoConfig
pkg/                         → Shared utilities (logger, status codes)
```

## Core Data Flow (Eino Graphs)

**Query Flow** (`flows/cache_query.go`):
```
START → preprocess → retrieve → select → postprocess → END
```
- `preprocess`: Normalize whitespace, remove control chars (`nodes/preprocessing.go`)
- `retrieve`: Qdrant vector search via Eino Retriever component
- `select`: Result ranking strategy (first/highest_score/temperature_softmax)
- `postprocess`: Extract question/answer from Document metadata

**Store Flow** (`flows/cache_store.go`):
```
START → quality_check → embedding → indexing → END
```
- Quality check can reject low-quality Q&A pairs before storage

## Eino Component Pattern

Components are created via factory functions in `internal/eino/components/`:

```go
// Create components with config
embedder, _ := components.NewEmbedder(ctx, &einoCfg.Embedder)  // supports openai, ark, ollama
retriever, _ := components.NewRetriever(ctx, &einoCfg.Retriever, embedder)  // qdrant, milvus, redis
indexer, _ := components.NewIndexer(ctx, &einoCfg.Indexer, embedder)

// Build and compile Graph → Runnable
queryGraph := flows.NewCacheQueryGraph(embedder, retriever, &einoCfg.Query)
queryRunner, _ := queryGraph.Compile(ctx)  // compose.Runnable[*Input, *Output]

// Invoke in handler
result, _ := queryRunner.Invoke(ctx, input)
```

## Essential Commands

```bash
# Development
go run cmd/server/main.go           # Requires configs/config.yaml
go mod tidy                         # Sync deps (uses Eino, Eino-ext)

# Testing
go test ./...                       # Unit tests
go test -cover ./...                # With coverage

# Infrastructure
docker run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant:latest
```

## Config Structure (`configs/config.yaml`)

The `eino` section drives all AI components:

```yaml
eino:
  embedder:
    provider: "openai"              # openai | ark | ollama | dashscope
    api_key: "sk-..."
    model: "text-embedding-3-small"
  retriever:
    provider: "qdrant"              # qdrant | milvus | redis | es8
    collection: "llm_cache"
    top_k: 5
    score_threshold: 0.75
    qdrant:
      host: "localhost"
      port: 6334                    # gRPC port, not HTTP 6333
  query:
    selection_strategy: "highest_score"  # first | highest_score | temperature_softmax
    preprocess_enabled: true
```

## Adding New Functionality

### New Embedding Provider
1. Add case in `components/embedder.go` switch
2. Implement `newXxxEmbedder()` using Eino-ext component
3. Add config fields in `eino/config/config.go` under `EmbedderConfig`

### New Graph Node
1. Create Lambda function in `nodes/` (see `preprocessing.go` pattern)
2. Add to Graph in `flows/` using `compose.InvokableLambda()`
3. Wire with `graph.AddLambdaNode()` and `graph.AddEdge()`

### New Result Selection Strategy
Add case in `nodes/result_select.go` (`ResultSelector.Select` method)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/cache/search` | Semantic cache query |
| POST | `/v1/cache/store` | Store Q&A pair (with quality check) |
| GET | `/v1/cache/:cache_id` | Get by ID |
| DELETE | `/v1/cache/:cache_id` | Delete single |
| DELETE | `/v1/cache/batch` | Batch delete |
| GET | `/v1/cache/health` | Health check |

## Testing Patterns

Tests use table-driven style (see `nodes/preprocessing_test.go`):
```go
tests := []struct {
    name     string
    input    string
    expected string
}{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {...})
}
```

## Common Pitfalls

- **Qdrant port**: Use gRPC 6334, not HTTP 6333
- **Vector dimension**: Must match embedding model (1536 for OpenAI text-embedding-3-small)
- **Graph compilation**: Call `Compile(ctx)` once, reuse returned `Runnable`
- **Eino-ext imports**: Some providers require uncommenting imports and adding deps
