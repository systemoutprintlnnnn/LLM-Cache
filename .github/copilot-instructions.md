# LLM-Cache Copilot Instructions

## Architecture Overview

LLM-Cache is a **semantic caching middleware** for LLM APIs using **DDD + Clean Architecture**:

```
cmd/server/main.go     → Dependency injection & wiring (start here for service initialization flow)
internal/app/          → HTTP layer (Gin handlers, routes, middleware)
internal/domain/       → Interfaces & models ONLY (services/, repositories/, models/)
internal/infrastructure/  → Concrete implementations (cache, vector, embedding, quality, stores)
pkg/                   → Shared utilities (logger, status codes)
configs/               → YAML config schema and loading
```

**Key Design Principle**: Domain layer defines interfaces; infrastructure implements them. Never import infrastructure from domain.

## Core Data Flow

Query path: `Handler → CacheService → VectorService → EmbeddingService + VectorRepository(Qdrant)`

1. **Preprocessing** (`preprocessing/`) - Normalize query text
2. **Embedding** (`embedding/remote/`) - Convert text → vector via OpenAI API
3. **Vector Search** (`stores/qdrant/`) - Find similar cached Q&A pairs
4. **Postprocessing** (`postprocessing/`) - Filter/rank results
5. **Quality Check** (`quality/`) - Validate before storage

## Essential Commands

```powershell
# Development
go run cmd/server/main.go           # Start server (requires configs/config.yaml)
go mod tidy                         # Sync dependencies

# Testing & Quality
go test ./...                       # Run all tests
go test -cover ./...                # With coverage
go fmt ./... ; go vet ./...         # Format + lint before commits

# Infrastructure
docker run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant:latest  # Local Qdrant
```

## Config Structure (`configs/config.yaml`)

```yaml
server:
  port: 8080
database:
  type: "qdrant"           # Only qdrant supported currently
  qdrant:
    host: "localhost"
    port: 6334             # gRPC port, not HTTP
    collection_name: "llm_cache"
    vector_size: 1536      # Must match embedding model dimension
embedding:
  type: "remote"           # Only remote/OpenAI supported
  remote:
    api_key: "sk-..."
    model_name: "text-embedding-3-small"
```

## Code Patterns

### Factory Pattern for Services
All infrastructure services use factory pattern (see `cmd/server/main.go`):
```go
factory := qdrant.NewQdrantVectorStoreFactory(log.SlogLogger())
vectorRepo, err := factory.CreateVectorRepository(ctx, &config.Database.Qdrant)
```

### Interface-First Design
Add new capabilities by:
1. Define interface in `internal/domain/services/` or `repositories/`
2. Implement in `internal/infrastructure/`
3. Wire in `cmd/server/main.go` via factory

### Context & Logging
Always pass `ctx` as first parameter; use structured logging:
```go
s.logger.InfoContext(ctx, "operation started", "key", value)
```

### Status Codes (`pkg/status/codes.go`)
Use `status.CodeOK`, `status.ErrCodeInvalidParam`, `status.ErrCodeInternal`, etc. for consistent API responses.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/cache/search` | Query semantic cache |
| POST | `/v1/cache/store` | Store Q&A pair |
| GET | `/v1/cache/:cache_id` | Get by ID |
| DELETE | `/v1/cache/:cache_id` | Delete single |
| DELETE | `/v1/cache/batch` | Batch delete |
| GET | `/v1/cache/health` | Health check |

## Testing Approach

- Place tests as `*_test.go` beside implementation
- Stub external calls (Qdrant, OpenAI) in unit tests
- Key test scenarios: similarity thresholds, quality filtering, empty/invalid payloads
- Use local Qdrant container for integration tests

## Common Pitfalls

- **Vector dimension mismatch**: `vector_size` in config must match embedding model output (1536 for OpenAI small)
- **gRPC vs HTTP port**: Qdrant client uses gRPC port 6334, not HTTP 6333
- **User type isolation**: `user_type` field creates namespace isolation in cache queries
