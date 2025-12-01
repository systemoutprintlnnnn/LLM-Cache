# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/` – API entrypoint; wires configs, services, and HTTP server startup.
- `configs/` – `config.go` schema plus loader stubs; runtime values live in `configs/config.yaml`.
- `internal/app/` – Gin routes, middleware, and HTTP handlers.
- `internal/domain/` – core models with service and repository interfaces.
- `internal/infrastructure/` – concrete adapters: cache pipeline, preprocessing/postprocessing, quality checks, vector store, embedding clients.
- `pkg/` – shared utilities (logger, status codes); integration scaffolding in `test/`; architectural notes in `docs/project/`.

## Build, Test, and Development Commands
- `go mod tidy` – sync dependencies after module changes.
- `go fmt ./...` and `go vet ./...` – format and static analysis before committing.
- `go build ./cmd/server` – compile the API server.
- `go run cmd/server/main.go` – run locally; ensure `configs/config.yaml` points to reachable Qdrant and embedding provider.
- `go test ./...` or `go test -cover ./...` – run unit tests; add `-short` for fast suites if introduced.
- Docker helpers: `docker-compose up -d` to start Qdrant + server; `docker-compose down` to stop.

## Coding Style & Naming Conventions
- Go idioms with `gofmt`; keep functions context-aware (`ctx` first) and prefer structured logging via `pkg/logger`.
- YAML keys use `lower_snake_case`; exported Go types use UpperCamelCase.
- Keep HTTP handlers thin; business logic in `internal/domain`, infra details in `internal/infrastructure`.

## Testing Guidelines
- Co-locate tests as `*_test.go`; favor table-driven cases and deterministic behavior.
- Stub external calls (Qdrant, embeddings) for unit tests; use local Qdrant for integration.
- Cover config validation, error paths (timeouts, empty payloads, unauthorized).
- Aim for meaningful coverage; keep suites independent of remote APIs.

## Commit & Pull Request Guidelines
- Use concise, imperative commits; Conventional-style prefixes (`feat:`, `fix:`, `docs:`) are welcome.
- PRs should describe scope, testing performed, config/env impacts (e.g., new YAML keys), and deployment considerations; attach relevant logs or screenshots.
- Link related issues or docs in `docs/project/` when applicable.

## Security & Configuration Tips
- Align new settings with `configs.Config` validation; provide safe defaults rather than panics.
- Keep secrets out of `configs/config.yaml`; use environment overrides when applicable.
