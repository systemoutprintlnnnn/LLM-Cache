# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/` — application entrypoint wiring configs, services, and HTTP server.
- `configs/` — YAML config schema (`config.go`) and loader stubs; keep runtime overrides in `configs/config.yaml`.
- `internal/app/` — Gin routes, middleware, and HTTP handlers.
- `internal/domain/` — domain models plus service and repository interfaces.
- `internal/infrastructure/` — concrete implementations: cache pipeline, preprocessing/postprocessing, quality checks, vector store (Qdrant), and embedding clients.
- `pkg/` — shared utilities (logger, status codes); `test/` holds example/integration scaffolding. Architectural docs live in `docs/project/`.

## Build, Test, and Development Commands
- `go mod tidy` — sync dependencies when modules change.
- `go fmt ./...` and `go vet ./...` — format and static checks before pushing.
- `go build ./cmd/server` — compile the API server.
- `go run cmd/server/main.go` — start locally; ensure `configs/config.yaml` points to a running Qdrant and embedding provider.
- `go test ./...` or `go test -cover ./...` — run unit tests; add `-short` for fast suites if added.
- Docker helpers from README: `docker-compose up -d` starts Qdrant + server; `docker-compose down` stops.

## Coding Style & Naming Conventions
- Follow idiomatic Go: gofmt formatting, lower_snake_case YAML fields, UpperCamelCase exported types, short receiver names.
- Keep HTTP layers thin; put business rules in `internal/domain` services and infra details under `internal/infrastructure`.
- Prefer context-aware functions (`ctx` first) and structured logging via `pkg/logger`.
- Align new configs with `configs.Config` validation; default safe values instead of panics.

## Testing Guidelines
- Place tests beside code as `*_test.go`; table-driven tests where possible.
- Stub external calls (Qdrant, embeddings) when unit-testing; use a local Qdrant instance for integration suites.
- Validate config and error paths: unauthorized access, timeouts, and empty payloads.
- Capture coverage for new packages; keep tests deterministic and independent of remote APIs.

## Commit & Pull Request Guidelines
- Use concise, imperative commits; Conventional-style prefixes (`feat:`, `fix:`, `docs:`) are welcome for clarity.
- Include what changed and why in the body when behavior or configs shift.
- PRs should state scope, testing performed, config/env impacts (e.g., new YAML keys), and any deployment considerations; attach logs or screenshots for observable changes.
- Link to related issues or docs in `docs/project/` when relevant to ease review.
